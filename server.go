package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"strings"

	// "reflect"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	Application "./application"
	FeedbackHandlers "./feedbackhandlers"
	Handlers "./handlers"
	Types "./types"
)

var address = flag.String("addr", "0.0.0.0:8000", "http service address")

var upgrader = websocket.Upgrader{}

// MakeRequest: Make up a Request to the connected client
func MakeRequest(i int, client *Types.Client) (string, error) {
	var result string
	var err error
	if (i+1)%10 == 0 {
		// result, err = Handlers.RequestToRunImage(client.ImageList[1])
		result, err = Handlers.RequestToStopContainer(client.ContainerList[0].ID)
		if err != nil {
			log.Printf("Could not get the IDLE ;)")
			panic(err)
		}
	} else if (i+1)%5 == 0 {
		// result, err = Handlers.RequestToRunImage(client.ImageList[1])
		result, err = Handlers.RequestToRunImage(client.ImageList[0].RepoTags[0])
		if err != nil {
			log.Printf("Could not get the IDLE ;)")
			panic(err)
		}
	} else if (i+1)%2 == 0 {
		// TODO: this is just WRONG
		var clientImageTags = []string{}
		for _, clientImage := range client.ImageList {
			clientImageTags = append(clientImageTags, clientImage.RepoTags[0])
		}
		result, err = Handlers.RequestToPullImage(clientImageTags)
		if err != nil {
			log.Printf("Could not get the IDLE ;)")
			panic(err)
		}
	} else {
		// Request a status
		result, err = Handlers.Status()
		if err != nil {
			log.Printf("Could not get the IDLE ;)")
			panic(err)
		}
		// ClientData[images]
	}

	return result, nil
}

func rpc(w http.ResponseWriter, r *http.Request) {
	// Step 1: Create a connection
	log.Printf("A new client connection established: %s ", r.URL)
	clientAddress := r.RemoteAddr
	splitted := strings.Split(clientAddress, ":")
	clientIP, clientPort := splitted[0], splitted[1]

	newClient := Types.NewClient(clientIP, clientPort)
	client, clientAddError := Application.ClientMap.AddClient(*newClient)
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
		functionResult, err := MakeRequest(i, client)
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
			log.Print("Reading message error:", err)
			break
		}
		// log.Printf("Read a message from the client: %s", string(clientResponseBody))

		FeedbackHandlers.HandleClientFeedback(client, clientResponseBody)
		i++

	}

}

var ListOfClients = func(w http.ResponseWriter, r *http.Request) {

	js, err := json.Marshal(Application.ClientMap.ClientList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Internal Server error"))
		return
	}
	w.Write(js)
}

var ClientAddRequest = func(w http.ResponseWriter, r *http.Request) {

}

func home(w http.ResponseWriter, r *http.Request) {
	log.Println("Got a connection")
	homeTemplate.Execute(w, "ws://"+r.Host+"/rpc")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	r := mux.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/rpc", rpc)
	r.HandleFunc("/api/client", ListOfClients)
	http.Handle("/", r)

	log.Printf("DockerHost Server -> started at address %v", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
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
