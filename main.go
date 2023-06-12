package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/db"
	"online-chat-go/db/repository"
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
	authorizer := &auth.DummyAuthorizer{}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/ws", websocket.NewWsHandler(wss, authorizer))

	// TODO: remove hardcoded variables and move to config files
	const dbUrl = "postgres://online-chat:jumanji@localhost:5432/online-chat?sslmode=disable"
	err := db.RunMigrations(os.DirFS("."), "db/migrations", dbUrl)
	if err != nil {
		log.Fatal("Unable to run migrations: ", err)
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}

	userRepo := repository.NewPgUserRepository(pool)
	fmt.Println(userRepo.GetByUserName(context.Background(), "user"))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Unable to bind server: ", err)
	}
}
