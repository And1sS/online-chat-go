package main

import (
	"context"
	"encoding/json"
	"fmt"
	websocket2 "github.com/gorilla/websocket"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/config"
	"online-chat-go/notifications"
	redis "online-chat-go/notifications/redis_bus/clustered"
	"online-chat-go/websocket"
	"strings"
)

func main() {
	cfg := config.ReadConfig()
	cfgJson, _ := json.MarshalIndent(cfg, "", "    ")
	fmt.Println("started app with config:\n", string(cfgJson))

	wss := websocket.NewWSServer()
	authorizer := &auth.DummyAuthorizer{}
	notificationBus := redis.NewClusteredRedisNotificationBus(cfg.NotificationBus.Redis.Cluster)
	notificationBus.Start()

	SetUpNotificationHandlers(wss, notificationBus)

	http.HandleFunc("/", websocket.NewWsHandler(wss, authorizer, &cfg.Ws, MakeWsConnectionHandler(notificationBus)))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.App.Port), nil); err != nil {
		log.Fatal("Unable to bind server: ", err)
	}
}

func SetUpNotificationHandlers(wss *websocket.WSServer, bus notifications.NotificationBus) {
	bus.PatternSubscribe(context.Background(), "/to/user/*")
	bus.SetMessageHandler(func(topic string, msg []byte) {
		id, _ := strings.CutPrefix(topic, "/to/user/")
		go wss.SendMessage(id, msg, websocket2.TextMessage)
	})
}

func MakeWsConnectionHandler(notificationBus notifications.NotificationBus) func(string, websocket.WSConnection) {
	return func(userId string, wsconn websocket.WSConnection) {
		for {
			select {
			case <-wsconn.Done():
				return

			case msg := <-wsconn.ReadPump():
				if err := notificationBus.Publish(context.Background(), "/to/user/1", msg.Data); err != nil {
					log.Println("Error publishing message: ", err)
				}
				log.Println(fmt.Sprintf("new message from user: %s, connection id: %s, msg: %s", userId, wsconn.Id(), string(msg.Data)))
			}
		}
	}
}
