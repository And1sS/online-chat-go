package notifications

import (
	"context"
	"github.com/redis/go-redis/v9"
	"online-chat-go/config"
)

type RedisEventBus struct {
	redis      *redis.Client
	pubsub     *redis.PubSub
	config     *config.RedisConfig
	msgHandler func(topic string, msg []byte)
}

func NewRedisNotificationBus(redis *redis.Client, pubsub *redis.PubSub) NotificationBus {
	reb := &RedisEventBus{
		redis:  redis,
		pubsub: pubsub,
	}
	go reb.runReader()
	return reb
}

func (reb *RedisEventBus) Publish(ctx context.Context, topic string, msg []byte) error {
	reb.redis.Publish(ctx, topic, msg)
	return nil
}

func (reb *RedisEventBus) Subscribe(ctx context.Context, topic string) {
	_ = reb.pubsub.Subscribe(ctx, topic)
}

func (reb *RedisEventBus) Unsubscribe(ctx context.Context, topic string) {
	_ = reb.pubsub.Unsubscribe(ctx, topic)
}

func (reb *RedisEventBus) SetMessageHandler(handler func(topic string, msg []byte)) {
	reb.msgHandler = handler
}

func (reb *RedisEventBus) UserTopic() string {
	return reb.config.UserTopic
}

func (reb *RedisEventBus) runReader() {
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
