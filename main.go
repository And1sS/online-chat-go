package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	websocket2 "github.com/gorilla/websocket"
	capi "github.com/hashicorp/consul/api"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/notifications"
	"online-chat-go/websocket"
	"strings"
	"time"
)

func main() {
	// TODO: remove hardcoded variables and move to config files
	wss := websocket.NewWSServer()
	authorizer := &auth.DummyAuthorizer{}
	notificationBus := StartNotificationBus()

	StartConsul()
	SetUpNotificationHandlers(wss, notificationBus)
	SetUpWsMessageHandlers(wss, notificationBus)

	http.HandleFunc("/", websocket.NewWsHandler(wss, authorizer))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Unable to bind server: ", err)
	}
}

func StartNotificationBus() notifications.NotificationBus {
	rc := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       -1,
	})
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

func StartConsul() *capi.Client {
	consul, err := capi.NewClient(capi.DefaultConfig())
	if err != nil {
		log.Fatal("Unable to create consul client: ", err)
	}

	ttl := time.Second * 30
	checkId := "alive-check"
	register := &capi.AgentServiceRegistration{
		ID:   fmt.Sprintf("connection-service-%s", uuid.New().String()),
		Name: "connection-service",
		Tags: []string{"connection"},
		Port: 8080,
		Check: &capi.AgentServiceCheck{
			CheckID:       checkId,
			TLSSkipVerify: true,
			TTL:           ttl.String(),
		},
	}

	err = consul.Agent().ServiceRegister(register)
	if err != nil {
		log.Fatal("Unable to register service in consul: ", err)
	}

	go func() {
		ticker := time.NewTicker(ttl / 2)
		for {
			<-ticker.C
			err := consul.Agent().UpdateTTL(checkId, "online", capi.HealthPassing)
			if err != nil {
				panic(err)
			}
		}
	}()

	return consul
}
