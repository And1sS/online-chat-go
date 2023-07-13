package discovery

import (
	"fmt"
	"github.com/google/uuid"
	capi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"log"
	"online-chat-go/config"
	"time"
)

// TODO: change hardcoded struct with interface to abstract away consul
// TODO: add ability to retrieve redis nodes addresses from consul and subscribe for updates
const (
	online = "online"
)

type ConsulAgent struct {
	address     string
	api         *capi.Client
	Id          uuid.UUID
	ServiceName string
	Tags        []string
	Port        int
	Ttl         time.Duration
	CheckId     string
}

func (c *ConsulAgent) serviceRegistration() *capi.AgentServiceRegistration {
	return &capi.AgentServiceRegistration{
		ID:   c.Id.String(),
		Name: c.ServiceName,
		Tags: c.Tags,
		Port: c.Port,
		Check: &capi.AgentServiceCheck{
			CheckID:       c.CheckId,
			TLSSkipVerify: true,
			TTL:           c.Ttl.String(),
		},
	}
}

func (c *ConsulAgent) registerService() error {
	return c.api.Agent().ServiceRegister(c.serviceRegistration())
}

func (c *ConsulAgent) healthChecker(onError func(error)) {
	ticker := time.NewTicker(c.Ttl / 2)
	for {
		<-ticker.C
		err := c.api.Agent().UpdateTTL(c.CheckId, online, capi.HealthPassing)
		if err != nil {
			onError(err)
		}
	}
}

func (c *ConsulAgent) RedisWatcher() {
	query := map[string]any{
		"type":        "service",
		"service":     "redis-notification-bus",
		"passingonly": true,
	}

	plan, err := watch.Parse(query)
	if err != nil {
		log.Fatal("Could not start watching for redis notification bus: ", err)
	}

	nodes := make(map[string]struct{})

	plan.HybridHandler = func(index watch.BlockingParamVal, result any) {
		switch msg := result.(type) {
		case []*capi.ServiceEntry:
			newNodes := make(map[string]struct{})

			for _, entry := range msg {
				nodeId := entry.Node.Node
				newNodes[nodeId] = struct{}{}

				if _, ok := nodes[nodeId]; !ok {
					log.Printf("new member joined: %s, address: %s, port: %d", nodeId, entry.Node.Address, entry.Service.Port)
				}
				nodes[nodeId] = struct{}{}
			}

			for key, _ := range nodes {
				if _, ok := newNodes[key]; !ok {
					log.Printf("member removed: %s", key)
				}
			}
			nodes = newNodes
		}
	}

	err = plan.RunWithConfig(c.address, &capi.Config{})
	if err != nil {
		log.Fatal("Could not start watching for redis notification bus: ", err)
	}
}

func NewConsul(applicationPort int, config *config.ConsulConfig) *ConsulAgent {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	consulConfig := capi.DefaultConfig()
	consulConfig.Address = address

	consul, err := capi.NewClient(consulConfig)
	if err != nil {
		log.Fatal("Unable to create consul client: ", err)
	}

	agent := &ConsulAgent{
		address:     address,
		api:         consul,
		Id:          uuid.New(),
		ServiceName: config.ServiceName,
		Tags:        config.Tags,
		Port:        applicationPort,
		Ttl:         config.Ttl,
		CheckId:     config.CheckId,
	}

	if err = agent.registerService(); err != nil {
		log.Fatal("Unable to register service in consul: ", err)
	}

	go agent.healthChecker(func(err error) { panic(err) })
	return agent
}
