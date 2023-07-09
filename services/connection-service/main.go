package main

import (
	"context"
	"fmt"
	websocket2 "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/config"
	"online-chat-go/discovery"
	"online-chat-go/notifications"
	"online-chat-go/websocket"
	"strings"
)

func main() {
	cfg := config.ReadConfig()
	fmt.Println(cfg)
	wss := websocket.NewWSServer()
	authorizer := &auth.DummyAuthorizer{}
	notificationBus := StartNotificationBus(&cfg.NotificationBus)
	_ = discovery.NewConsul(cfg.App.Port, &cfg.Discovery.Consul)

	SetUpNotificationHandlers(wss, notificationBus, &cfg.NotificationBus)
	SetUpWsMessageHandlers(wss, notificationBus, &cfg.NotificationBus)

	http.HandleFunc("/", websocket.NewWsHandler(wss, authorizer, &cfg.Ws))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.App.Port), nil); err != nil {
		log.Fatal("Unable to bind server: ", err)
	}
}

func StartNotificationBus(config *config.NotificationBusConfig) notifications.NotificationBus {
	rc := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
			Password: "",
			DB:       -1,
		},
	)
	pubsub := rc.Subscribe(context.Background())
	return notifications.NewRedisNotificationBus(rc, pubsub)
}

func SetUpNotificationHandlers(wss *websocket.WSServer, bus notifications.NotificationBus, config *config.NotificationBusConfig) {
	bus.SetMessageHandler(func(topic string, msg []byte) {
		id, _ := strings.CutPrefix(topic, config.Redis.UserTopic)
		wss.SendMessage(id, msg, websocket2.TextMessage)
	})
}

func SetUpWsMessageHandlers(wss *websocket.WSServer, bus notifications.NotificationBus, config *config.NotificationBusConfig) {
	wss.SetOnUserConnected(func(id string) {
		bus.Subscribe(context.Background(), fmt.Sprintf("%s%s", config.Redis.UserTopic, id))
	})
	wss.SetOnUserDisconnected(func(id string) {
		bus.Unsubscribe(context.Background(), fmt.Sprintf("%s%s", config.Redis.UserTopic, id))
	})
}
