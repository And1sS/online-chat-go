package notifications

import (
	"context"
)

type NotificationBus interface {
	Start()
	Publish(ctx context.Context, topic string, msg []byte) error
	Subscribe(ctx context.Context, topic string)
	Unsubscribe(ctx context.Context, topic string)
	PatternSubscribe(ctx context.Context, pattern string)
	PatternUnsubscribe(ctx context.Context, pattern string)
	SetMessageHandler(func(topic string, msg []byte))
	Close()
}
