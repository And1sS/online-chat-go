package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"online-chat-go/config"
	"online-chat-go/notifications"
)

type RedisNotificationBus struct {
	redis      *redis.Client
	pubsub     *redis.PubSub
	config     *config.RedisConfig
	msgHandler func(topic string, msg []byte)
}

func NewRedisNotificationBus(redis *redis.Client, pubsub *redis.PubSub) notifications.NotificationBus {
	reb := &RedisNotificationBus{
		redis:  redis,
		pubsub: pubsub,
	}
	go reb.runReader()
	return reb
}

func (reb *RedisNotificationBus) Publish(ctx context.Context, topic string, msg []byte) error {
	reb.redis.Publish(ctx, topic, msg)
	return nil
}

func (reb *RedisNotificationBus) Subscribe(ctx context.Context, topic string) {
	_ = reb.pubsub.Subscribe(ctx, topic)
}

func (reb *RedisNotificationBus) Unsubscribe(ctx context.Context, topic string) {
	_ = reb.pubsub.Unsubscribe(ctx, topic)
}

func (reb *RedisNotificationBus) SetMessageHandler(handler func(topic string, msg []byte)) {
	reb.msgHandler = handler
}

func (reb *RedisNotificationBus) UserTopic() string {
	return reb.config.UserTopic
}

func (reb *RedisNotificationBus) runReader() {
	for {
		select {
		case msg, ok := <-reb.pubsub.Channel():
			if !ok {
				return
			}
			if reb.msgHandler != nil {
				go reb.msgHandler(msg.Channel, []byte(msg.Payload))
			}
		}
	}
}
