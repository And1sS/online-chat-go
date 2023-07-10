package notifications

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"online-chat-go/config"
)

type NotificationBus interface {
	Publish(ctx context.Context, topic string, msg []byte) error
	Subscribe(ctx context.Context, topic string)
	Unsubscribe(ctx context.Context, topic string)
	SetMessageHandler(func(topic string, msg []byte))
	UserTopic() string
}

func NewNotificationBus(config *config.NotificationBusConfig) NotificationBus {
	rc := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
			Password: "",
			DB:       -1,
		},
	)
	pubsub := rc.Subscribe(context.Background())
	return NewRedisNotificationBus(rc, pubsub)
}
