package main

import (
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/websocket"
)

func handleIndex(response http.ResponseWriter, request *http.Request) {
	_, err := response.Write([]byte("test"))
	if err != nil {
		log.Fatal("error sending response: ", err)
	}
}

func main() {
	wss := websocket.NewWSConnections()
	authorizer := &auth.DummyAuthenticator{}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/ws", websocket.NewWsHandler(wss, authorizer))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
