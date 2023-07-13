package main

import (
	"context"
	"fmt"
	websocket2 "github.com/gorilla/websocket"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/config"
	"online-chat-go/discovery"
	"online-chat-go/notifications"
	notificationsconfig "online-chat-go/notifications/config"
	"online-chat-go/websocket"
	"strings"
)

func main() {
	cfg := config.ReadConfig()
	fmt.Println(cfg)
	wss := websocket.NewWSServer()
	authorizer := &auth.DummyAuthorizer{}
	notificationBus := notificationsconfig.NewNotificationBus(&cfg.NotificationBus)
	_ = discovery.NewConsul(cfg.App.Port, &cfg.Discovery.Consul)

	SetUpNotificationHandlers(wss, notificationBus)
	SetUpWsMessageHandlers(wss, notificationBus)

	http.HandleFunc("/", websocket.NewWsHandler(wss, authorizer, &cfg.Ws))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.App.Port), nil); err != nil {
		log.Fatal("Unable to bind server: ", err)
	}
}

func SetUpNotificationHandlers(wss *websocket.WSServer, bus notifications.NotificationBus) {
	bus.SetMessageHandler(func(topic string, msg []byte) {
		id, _ := strings.CutPrefix(topic, bus.UserTopic())
		wss.SendMessage(id, msg, websocket2.TextMessage)
	})
}

func SetUpWsMessageHandlers(wss *websocket.WSServer, bus notifications.NotificationBus) {
	wss.SetOnUserConnected(func(id string) {
		bus.Subscribe(context.Background(), fmt.Sprintf("%s%s", bus.UserTopic(), id))
	})
	wss.SetOnUserDisconnected(func(id string) {
		bus.Unsubscribe(context.Background(), fmt.Sprintf("%s%s", bus.UserTopic(), id))
	})
}
