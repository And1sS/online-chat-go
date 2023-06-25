package notifications

import "context"

type NotificationBus interface {
	Publish(ctx context.Context, topic string, msg []byte) error
	Subscribe(ctx context.Context, topic string)
	Unsubscribe(ctx context.Context, topic string)
	SetMessageHandler(func(topic string, msg []byte))
}
