package single

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
)

type RedisNotificationBus struct {
	redis             *redis.Client
	pubsub            *redis.PubSub
	msgHandler        func(topic string, msg []byte)
	disconnectHandler func()
}

func NewRedisNotificationBus(host string, port int) *RedisNotificationBus {
	rc := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: "",
			DB:       -1,
		},
	)

	reb := &RedisNotificationBus{
		redis:  rc,
		pubsub: rc.Subscribe(context.Background()),
	}
	return reb
}

func (reb *RedisNotificationBus) Start() {
	go reb.runReader()
}

func (reb *RedisNotificationBus) Publish(ctx context.Context, topic string, msg []byte) error {
	var err error
	defer func() {
		if p := recover(); p != nil {
			switch perr := p.(type) {
			case error:
				err = perr
			default:
				err = errors.New("Could not publish message due to unknown error")
			}
		}
	}()

	reb.redis.Publish(ctx, topic, msg)
	return err
}

func (reb *RedisNotificationBus) Subscribe(ctx context.Context, topic string) {
	_ = reb.pubsub.Subscribe(ctx, topic)
}

func (reb *RedisNotificationBus) Unsubscribe(ctx context.Context, topic string) {
	_ = reb.pubsub.Unsubscribe(ctx, topic)
}

func (reb *RedisNotificationBus) PatternSubscribe(ctx context.Context, pattern string) {
	_ = reb.pubsub.PSubscribe(ctx, pattern)
}

func (reb *RedisNotificationBus) PatternUnsubscribe(ctx context.Context, pattern string) {
	_ = reb.pubsub.PUnsubscribe(ctx, pattern)
}

func (reb *RedisNotificationBus) SetMessageHandler(handler func(topic string, msg []byte)) {
	reb.msgHandler = handler
}

func (reb *RedisNotificationBus) SetDisconnectHandler(handler func()) {
	reb.disconnectHandler = handler
}

func (reb *RedisNotificationBus) Close() {
	_ = reb.redis.Close()
}

func (reb *RedisNotificationBus) runReader() {
	for {
		select {
		case msg, ok := <-reb.pubsub.Channel():
			if !ok {
				log.Println("Connection with redis closed")
				if reb.disconnectHandler != nil {
					reb.disconnectHandler()
				}
				return
			}
			if reb.msgHandler != nil {
				go reb.msgHandler(msg.Channel, []byte(msg.Payload))
			}
		}
	}
}
