package main

import (
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var gameStarted = false
var startTime time.Time

type Message struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			break
		}

		switch msg.Type {
		case "startGame":
			if !gameStarted {
				startTime = time.Now()
				gameStarted = true
				go countdown()
			}
		case "clicked":
			if gameStarted {
				duration := time.Since(startTime).Seconds()
				broadcast <- Message{Type: "result", Message: fmt.Sprintf("%.2f seconds", duration)}
				gameStarted = false
			}
		}
	}
}

func countdown() {
	time.Sleep(5 * time.Second) // Adjust the countdown duration as needed
	broadcast <- Message{Type: "timeout", Message: "Time's up!"}
	gameStarted = false
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/ws")
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Click Game</title>
    <script>
        var socket = new WebSocket("{{.}}");

        socket.onmessage = function (event) {
            var msg = JSON.parse(event.data);
            switch (msg.type) {
                case "result":
                    document.getElementById("result").innerHTML = "Your time: " + msg.message;
                    break;
                case "timeout":
                    document.getElementById("result").innerHTML = "Game over - " + msg.message;
                    break;
            }
        };

        function startGame() {
            var message = { type: "startGame" };
            socket.send(JSON.stringify(message));
        }

        function clicked() {
            var message = { type: "clicked" };
            socket.send(JSON.stringify(message));
        }
    </script>
</head>
<body>
    <h1>Click Game</h1>
    <button onclick="startGame()">Start Game</button>
    <button onclick="clicked()">Click Me!</button>
    <div id="result"></div>
</body>
</html>
`))

