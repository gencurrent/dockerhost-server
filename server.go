package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	Handlers "./handlers"
)


var ClientData map[string]interface{}

var address = flag.String("addr", "0.0.0.0:8000", "http service address")

var upgrader = websocket.Upgrader{}

// MakeRequest: Создать запрос клиенту
func MakeRequest(i int) (string, error) {
	var result string
	var err error
	if (i+1)%5 != 0 {
		// Request a status
		result, err = Handlers.Status()
		if err != nil {
			log.Printf("Could not get the IDLE ;)")
			panic(err)
        }
        // ClientData[images]
	} else {
		result, err = Handlers.RequestToPullImage([]string{"postgres"})
		if err != nil {
			log.Printf("Could not get the IDLE ;)")
			panic(err)
		}
	}

	return result, nil
}

func rpc(w http.ResponseWriter, r *http.Request) {
	// Step 1: Create a connection
	log.Printf("Got the connection :%s ", r.URL)
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
		functionResult, err := MakeRequest(i)
		if err != nil {
			log.Printf("The wrong function result: %s", err)
			panic(err)
		}

		// Step 3: send message to the connected client
		err = c.WriteMessage(websocket.TextMessage, []byte(functionResult))
		if err != nil {
			log.Print("Write message to the client error: %s", err)
			break
		}

		// Step 4: Read the feedback from the client
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Print("Reading message error:", err)
			break
		}
		log.Printf("Read a message from the client: %s", string(message))

		i++

	}

}

func home(w http.ResponseWriter, r *http.Request) {
	log.Println("Got a connection")
	homeTemplate.Execute(w, "ws://"+r.Host+"/rpc")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", home)
	http.HandleFunc("/rpc", rpc)

	log.Print("Started the go server")
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
