package discovery

import (
	"fmt"
	"github.com/google/uuid"
	capi "github.com/hashicorp/consul/api"
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

func NewConsul(applicationPort int, config *config.ConsulConfig) *ConsulAgent {
	consulConfig := capi.DefaultConfig()
	consulConfig.Address = fmt.Sprintf("%s:%d", config.Host, config.Port)

	consul, err := capi.NewClient(consulConfig)
	if err != nil {
		log.Fatal("Unable to create consul client: ", err)
	}

	agent := &ConsulAgent{
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
