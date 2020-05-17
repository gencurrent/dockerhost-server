package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	uuid "github.com/satori/go.uuid"

	Application "./application"
	FeedbackHandlers "./feedbackhandlers"
	Handlers "./handlers"
	Types "./types"
)

var address = flag.String("addr", ":8000", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { log.Print("New connection detected"); return true },
}

// PopRequest: Pop a RequestStructure from the client's request queue
func PopRequest(client *Types.Client) (string, error) {

	log.Printf("The request queue = %v", client.RequestQueue)
	RequestDefault := Handlers.Status

	request := client.PopRequestStructure()
	if request == nil {
		return RequestDefault()
	}
	var result string
	var err error

	if request.Name == "Image.Run" {
		result, err = Handlers.RequestToRunImage(string(request.Arguments["Image.Name"].(string)))
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else if request.Name == "Image.Remove" {
		// TODO: this is just WRONG // Need the mutex usage
		result, err = Handlers.RequestToRemoveImage(request.Arguments["Image.ID"].(string))
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else if request.Name == "Image.Pull" {
		// TODO: this is just WRONG // Need the mutex usage
		result, err = Handlers.RequestToPullImage(request.Arguments["Image.Tag"].(string))
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else if request.Name == "Container.Start" {
		result, err = Handlers.RequestToStartContainer(request.Arguments["Container.ID"].(string))
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else if request.Name == "Container.Pause" {
		result, err = Handlers.RequestToPauseContainer(request.Arguments["Container.ID"].(string))
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else if request.Name == "Container.Stop" {
		result, err = Handlers.RequestToStopContainer(request.Arguments["Container.ID"].(string))
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else if request.Name == "Container.Remove" {
		result, err = Handlers.RequestToRemoveContainer(request.Arguments["Container.ID"].(string))
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else if request.Name == "Status" {
		// Request a status
		result, err = Handlers.Status()
		if err != nil {
			log.Printf("Could not process a request %s", request.Name)
			panic(err)
		}
	} else {
		log.Printf("Could not process a request %s", request.Name)
		return "", nil
	}

	return result, nil
}

func rpc(w http.ResponseWriter, r *http.Request) {
	// Step 1: Create a connection
	log.Printf("RPC -> Connection : %s ", r.URL)
	clientAddress := r.RemoteAddr
	splitted := strings.Split(clientAddress, ":")
	clientIP, clientPort := splitted[0], splitted[1]

	newClient := Types.NewClient(clientIP, clientPort)
	client, clientAddError := Application.ClientMap.AddClient(newClient)
	if clientAddError != nil {
		log.Fatalf("The client error: %s", clientAddError)
	}
	log.Printf("The new client is : %s, %s:%d", client.UUID, client.IP, client.Port)
	for _, cmItem := range Application.ClientMap.ClientList {
		log.Printf("The client list item #%s = %s:%d", cmItem.UUID, cmItem.IP, cmItem.Port)

	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	i := 0
	for {
		// Step 2: Make up a request to client
		// TODO: Wrap all the requests and responses to Request type and Response Type
		functionResult, err := PopRequest(client)
		if err != nil {
			log.Printf("The wrong function result: %s", err)
			panic(err)
		}

		// Step 3: send encoded request to the connected client
		err = c.WriteMessage(websocket.TextMessage, []byte(functionResult))
		if err != nil {
			log.Printf("Write message to the client error: %s", err)
			break
		}

		// Step 4: Read the feedback on the request from the client
		_, clientResponseBody, err := c.ReadMessage()
		if err != nil {
			for id, cmClient := range Application.ClientMap.ClientList {
				if client.Port == cmClient.Port && client.IP == cmClient.IP {
					// delete(Application.ClientMap.ClientList[id])
					log.Printf("Id", id)
				}
			}
			log.Print("Reading message error:", err)

			break
		}

		// Step 5: Handle the feedback from the client
		FeedbackHandlers.HandleClientFeedback(client, clientResponseBody)
		i++

	}

}

var ListOfImages = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Printf("ListOfImages -> Called")
	Handlers.UpdateLocalImageList()
	imageList := Handlers.HostImageList
	encoded, err := json.Marshal(imageList)
	if err != nil {
		errorResponse := Types.ApiResponseError{
			Error: `Could not encode the list of images`,
		}
		encodedError, err := json.Marshal(errorResponse)
		if err != nil {
			log.Fatalf("Error encoding error")
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(encodedError))
		return
	}
	w.Write(encoded)
	return
}

var ListOfClients = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	js, err := json.Marshal(Application.ClientMap.ClientList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Internal Server error"))
		return
	}
	w.Write(js)
}

var GetClient = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	vars := mux.Vars(r)
	uuidParsed, err := uuid.FromString(vars["uuid"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong UUID supplied"))
		return
	}
	client, err := Application.ClientMap.ByUUID(uuidParsed)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("404 - The client is not found by UUID: %s", uuidParsed)))
		return
	}
	js, err := json.Marshal(client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Internal Server error"))
		return
	}
	w.Write(js)
}

// ClientAddRequest adds a Request Structure for a client's queue
var ClientAddRequest = func(w http.ResponseWriter, r *http.Request) {
	log.Printf(`CliendAddAddress called`)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	vars := mux.Vars(r)
	uuidParsed, err := uuid.FromString(vars["uuid"])
	if err != nil {
		log.Infof("Could not parse a UUID `%v`", vars["uuid"])
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong UUID supplied"))
		return
	}
	client, err := Application.ClientMap.ByUUID(uuidParsed)
	if err != nil {
		log.Infof("Client with UUID `%s` is not found", vars["uuid"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("404 - The client is not found by UUID: %s", uuidParsed)))
		return
	}

	// Request Structure
	rStructure := Types.RequestStructure{}
	// Body decoding
	var metaRS interface{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&metaRS)
	if err != nil {
		panic(err)
	}
	// Convert to an interface
	metaRSMap, ok := metaRS.(map[string]interface{})
	if !ok {
		log.Printf("The error converting RequestStructure: %s", err)
	}
	reqName, ok := metaRSMap["name"].(string)
	if !ok {
		log.Printf(`Could not parse the reqName, ok := strstr["name"].(string)`)
	}
	rStructure.Name = reqName
	rStructure.Arguments = make(map[string]interface{})
	log.Printf("The name = %s", reqName)
	argumentList, ok := metaRSMap["arguments"].([]interface{})
	if !ok {
		errorResponse := Types.ApiResponseError{
			Error: `"arguments" field must be an array: "arguments": [{"key": "value", ...}]`,
		}
		encodedError, err := json.Marshal(errorResponse)
		if err != nil {
			log.Fatalf("Error encoding error")
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(encodedError))
	}
	for _, v := range argumentList {
		arg := v.(map[string]interface{})
		for argeKey, argValue := range arg {
			rStructure.Arguments[argeKey] = argValue
		}
	}

	client.PushRequestStructure(&rStructure)
	w.Write([]byte("Ok"))
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	log.Println("New connection detected")
	homeTemplate.Execute(w, "ws://"+r.Host+"/rpc")
}

func main() {
	flag.Parse()
	router := mux.NewRouter()

	router.HandleFunc("/", home).Methods("GET")
	router.HandleFunc("/rpc", rpc)

	// Images
	router.HandleFunc("/api/image", ListOfImages).Methods("GET")

	// Clients
	router.HandleFunc("/api/client/{uuid}/request-queue", ClientAddRequest).Methods("POST")
	router.HandleFunc("/api/client", ListOfClients).Methods("GET")
	router.HandleFunc("/api/client/{uuid}", GetClient).Methods("GET")

	// Requests
	// router.HandleFunc("/api/client/{uuid}/request-queue/{requestUUID}", ClientHandleSingleRequest)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})
	// c.allowedOriginsAll = true
	handler := c.Handler(router)
	// http.Handle("/", router)

	log.Printf("DockerHost Server -> started at address %v", *address)
	log.Fatal(http.ListenAndServe(*address, handler))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
