package main

import (
	"context"
	"fmt"
	websocket2 "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/events"
	"online-chat-go/websocket"
	"strings"
)

func handleIndex(response http.ResponseWriter, request *http.Request) {
	_, err := response.Write([]byte("test"))
	if err != nil {
		log.Fatal("error sending response: ", err)
	}
}

func main() {
	wss := websocket.NewWSServer()
	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       -1,
	})
	pubsub := redis.Subscribe(context.Background())

	eventBus := events.NewRedisEventBus(redis, pubsub)
	eventBus.SetMessageHandler(func(topic string, msg []byte) {
		id, _ := strings.CutPrefix(topic, "/to/user/")
		wss.SendMessage(id, msg, websocket2.TextMessage)
	})

	wss.SetOnUserConnected(func(id string) {
		eventBus.Subscribe(context.Background(), fmt.Sprintf("/to/user/%s", id))
	})
	wss.SetOnUserDisconnected(func(id string) {
		eventBus.Unsubscribe(context.Background(), fmt.Sprintf("/to/user/%s", id))
	})

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
