package clustered

import (
	"context"
	set "github.com/deckarep/golang-set/v2"
	"log"
	"online-chat-go/config"
	"online-chat-go/notifications"
	"online-chat-go/notifications/redis_bus/single"
	"online-chat-go/util"
)

type ClusteredRedisNotificationBus struct {
	cluster      *util.SafeMap[string, *single.RedisNotificationBus]
	topics       set.Set[string]
	patterns     set.Set[string]
	nodesWatcher RedisWatcher
	msgHandler   func(topic string, msg []byte)
	done         chan bool
}

func (cnb *ClusteredRedisNotificationBus) Start() {
	cnb.nodesWatcher.Start()
	go cnb.watcher()
}

func (cnb *ClusteredRedisNotificationBus) Publish(ctx context.Context, topic string, msg []byte) error {
	cnb.cluster.ForRandomEntry(func(nodeId string, bus *single.RedisNotificationBus) {
		log.Println("published to node: ", nodeId)
		go bus.Publish(ctx, topic, msg)
	})
	return nil
}

func (cnb *ClusteredRedisNotificationBus) Subscribe(ctx context.Context, topic string) {
	cnb.topics.Add(topic)
	cnb.cluster.ForEach(func(_ string, bus *single.RedisNotificationBus) { go bus.Subscribe(ctx, topic) })
}

func (cnb *ClusteredRedisNotificationBus) Unsubscribe(ctx context.Context, topic string) {
	cnb.topics.Remove(topic)
	cnb.cluster.ForEach(func(nodeId string, bus *single.RedisNotificationBus) { go bus.Unsubscribe(ctx, topic) })
}

func (cnb *ClusteredRedisNotificationBus) PatternSubscribe(ctx context.Context, pattern string) {
	cnb.patterns.Add(pattern)
	cnb.cluster.ForEach(func(_ string, bus *single.RedisNotificationBus) { go bus.PatternSubscribe(ctx, pattern) })
}

func (cnb *ClusteredRedisNotificationBus) PatternUnsubscribe(ctx context.Context, pattern string) {
	cnb.patterns.Remove(pattern)
	cnb.cluster.ForEach(func(_ string, bus *single.RedisNotificationBus) { go bus.PatternUnsubscribe(ctx, pattern) })
}

func (cnb *ClusteredRedisNotificationBus) SetMessageHandler(handler func(topic string, msg []byte)) {
	cnb.msgHandler = handler
}

func (cnb *ClusteredRedisNotificationBus) Close() {
	cnb.cluster.ForEach(func(_ string, bus *single.RedisNotificationBus) { bus.Close() })
	cnb.nodesWatcher.Close()
}

func (cnb *ClusteredRedisNotificationBus) add(id string, host string, port int) {
	bus := single.NewRedisNotificationBus(host, port)
	bus.SetMessageHandler(cnb.msgHandler)
	bus.SetDisconnectHandler(func() { cnb.cluster.Delete(id) })
	bus.Start()

	for _, topic := range cnb.topics.ToSlice() {
		go bus.Subscribe(context.Background(), topic)
	}
	for _, pattern := range cnb.patterns.ToSlice() {
		go bus.PatternSubscribe(context.Background(), pattern)
	}

	cnb.cluster.Set(id, bus)
	log.Println("connected to redis node: ", id)
}

func (cnb *ClusteredRedisNotificationBus) remove(id string) {
	if bus, ok := cnb.cluster.Get(id); ok {
		cnb.cluster.Delete(id)
		bus.Close()
		log.Println("disconnected from redis node: ", id)
	}
}

func (cnb *ClusteredRedisNotificationBus) watcher() {
	for {
		select {
		case event, ok := <-cnb.nodesWatcher.ClusterWatcher():
			if !ok {
				cnb.Close()
				return
			}

			log.Printf("New update from consul: \nadded nodes %+v\nremoved nodes:%+v\n", event.Added.ToSlice(), event.Removed.ToSlice())
			for _, node := range event.Added.ToSlice() {
				cnb.add(node.NodeId, node.NodeHost, node.NodePort)
			}
			for _, node := range event.Removed.ToSlice() {
				cnb.remove(node.NodeId)
			}
		}
	}
}

func NewClusteredRedisNotificationBus(config *config.RedisClusterConfig) notifications.NotificationBus {
	cnb := &ClusteredRedisNotificationBus{
		cluster:      util.NewSafeMap[string, *single.RedisNotificationBus](),
		topics:       set.NewSet[string](),
		patterns:     set.NewSet[string](),
		nodesWatcher: NewConsul(&config.Consul),
		msgHandler: func(topic string, msg []byte) {
			log.Printf("message arrived on topic: %s, msg: %s", topic, string(msg))
		},
	}

	return cnb
}
