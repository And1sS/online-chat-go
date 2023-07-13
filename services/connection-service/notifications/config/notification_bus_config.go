package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"online-chat-go/config"
	"online-chat-go/notifications"
	redisnotifications "online-chat-go/notifications/redis_bus"
)

func NewNotificationBus(config *config.NotificationBusConfig) notifications.NotificationBus {
	rc := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
			Password: "",
			DB:       -1,
		},
	)
	pubsub := rc.Subscribe(context.Background())
	return redisnotifications.NewRedisNotificationBus(rc, pubsub)
}
