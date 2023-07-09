package main

import (
	"context"
	"fmt"
	websocket2 "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/discovery"
	"online-chat-go/notifications"
	"online-chat-go/websocket"
	"strings"
)

// TODO: remove hardcoded variables and move to config files
const (
	applicationPort = 8080
	redisPort       = 6379
)

func main() {
	wss := websocket.NewWSServer()
	authorizer := &auth.DummyAuthorizer{}
	notificationBus := StartNotificationBus()
	_ = discovery.NewConsul(applicationPort)

	SetUpNotificationHandlers(wss, notificationBus)
	SetUpWsMessageHandlers(wss, notificationBus)

	http.HandleFunc("/", websocket.NewWsHandler(wss, authorizer))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", applicationPort), nil); err != nil {
		log.Fatal("Unable to bind server: ", err)
	}
}

func StartNotificationBus() notifications.NotificationBus {
	rc := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("localhost:%s", redisPort),
			Password: "",
			DB:       -1,
		},
	)
	pubsub := rc.Subscribe(context.Background())
	return notifications.NewRedisNotificationBus(rc, pubsub)
}

func SetUpNotificationHandlers(wss *websocket.WSServer, bus notifications.NotificationBus) {
	bus.SetMessageHandler(func(topic string, msg []byte) {
		id, _ := strings.CutPrefix(topic, "/to/user/")
		wss.SendMessage(id, msg, websocket2.TextMessage)
	})
}

func SetUpWsMessageHandlers(wss *websocket.WSServer, bus notifications.NotificationBus) {
	wss.SetOnUserConnected(func(id string) {
		bus.Subscribe(context.Background(), fmt.Sprintf("/to/user/%s", id))
	})
	wss.SetOnUserDisconnected(func(id string) {
		bus.Unsubscribe(context.Background(), fmt.Sprintf("/to/user/%s", id))
	})
}
