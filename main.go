package main

import (
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/db"
	"online-chat-go/websocket"
	"os"
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

	const dbUrl = "postgres://online-chat:jumanji@localhost:5432/online-chat?sslmode=disable"
	err := db.RunMigrations(os.DirFS("."), "db/migrations", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
