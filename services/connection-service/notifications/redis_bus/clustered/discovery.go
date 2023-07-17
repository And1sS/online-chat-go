package clustered

import (
	"fmt"
	set "github.com/deckarep/golang-set/v2"
	capi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"log"
	"online-chat-go/config"
	"sync"
)

type RedisNode struct {
	NodeId   string
	NodeHost string
	NodePort int
}

type RedisClusterEvent struct {
	Added   set.Set[RedisNode]
	Removed set.Set[RedisNode]
}

type RedisWatcher interface {
	Start()
	ClusterWatcher() <-chan RedisClusterEvent
	Close()
}

type ConsulAgent struct {
	address     string
	serviceName string
	plan        *watch.Plan
	nodes       set.Set[RedisNode]
	events      chan RedisClusterEvent
	mut         *sync.Mutex
}

func (c *ConsulAgent) Start() {
	query := map[string]any{
		"type":        "service",
		"service":     c.serviceName,
		"passingonly": true,
	}

	plan, err := watch.Parse(query)
	if err != nil {
		log.Fatal("Could not start watching for redis notification bus: ", err)
	}

	plan.HybridHandler = func(index watch.BlockingParamVal, result any) {
		switch msg := result.(type) {
		case []*capi.ServiceEntry:
			c.events <- c.updateNodes(msg)
		}
	}

	c.mut.Lock()
	c.plan = plan
	c.mut.Unlock()

	go func() {
		err = plan.RunWithConfig(c.address, &capi.Config{})
		if err != nil {
			c.Close()
			log.Println("Could not start watching for redis notification bus: ", err)
		}
	}()
}

func (c *ConsulAgent) ClusterWatcher() <-chan RedisClusterEvent {
	return c.events
}

func (c *ConsulAgent) Close() {
	c.mut.Lock()
	defer c.mut.Unlock()

	select {
	case <-c.events:
		return
	default:
		close(c.events)
	}

	if c.plan != nil && !c.plan.IsStopped() {
		c.plan.Stop()
	}
}

func (c *ConsulAgent) updateNodes(msg []*capi.ServiceEntry) RedisClusterEvent {
	c.mut.Lock()
	defer c.mut.Unlock()

	newNodes := set.NewThreadUnsafeSet[RedisNode]()
	for _, entry := range msg {
		newNodes.Add(RedisNode{
			NodeId:   entry.Node.Node,
			NodeHost: entry.Node.Address,
			NodePort: entry.Service.Port,
		})
	}

	added := newNodes.Difference(c.nodes)
	removed := c.nodes.Difference(newNodes)
	c.nodes = newNodes

	return RedisClusterEvent{Added: added, Removed: removed}
}

func NewConsul(config *config.ConsulConfig) *ConsulAgent {
	return &ConsulAgent{
		address:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		serviceName: config.RedisServiceName,
		nodes:       set.NewThreadUnsafeSet[RedisNode](),
		events:      make(chan RedisClusterEvent),
		mut:         &sync.Mutex{},
	}
}
