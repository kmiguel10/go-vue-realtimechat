package main

//A simple HTTP server that handles requests

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//Global variables

//a map where the key is a pointer to a Websocket... the value is boolean
// we are using a map instead of an array because it is easier to append and delete
var clients = make(map[*websocket.Conn]bool) //connected clients

//A channel that will act as a queue for messages sent by clients
var broadcast = make(chan Message) //broadcast channel

//configure the upgrader which is an object with methods for taking a
//normal http connection and upgrading it to a websocket
var upgrader = websocket.Upgrader{}

//define message object
//the test surrounded by `` are metagata which helps go serialize and
//undserialize the Message object to and from JSON
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	//create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	//configure websocket route
	//this will handle any request for initiating a websocket
	http.HandleFunc("/ws", handleConnections)

	//Start listening for incoming chaat messages
	go handleMessages()

	//start server on port localhost:8000 and log any errors
	log.Println("http server started on :8000")

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	//upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	//Make sure we close the connection when the function returns
	defer ws.Close()

	//register our new client
	clients[ws] = true

	//an infinilite loop that constantly waits for a new message to be written
	//to the Websocket
	for {
		var msg Message
		//Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		//send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

//A loop that constantly reads from the "broadcast" channel and then relays
// the message to all of the clients over their respective WebSocket connection
func handleMessages() {
	for {
		//grab the message from the broadcast channel
		msg := <-broadcast
		//Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
