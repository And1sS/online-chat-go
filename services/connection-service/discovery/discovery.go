package discovery

import (
	"github.com/google/uuid"
	capi "github.com/hashicorp/consul/api"
	"log"
	"time"
)

const (
	ttl         = 30 * time.Second
	checkId     = "alive-check"
	serviceName = "connection-service"
	online      = "online"
	connection  = "connection"
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

func NewConsul(port int) *ConsulAgent {
	consul, err := capi.NewClient(capi.DefaultConfig())
	if err != nil {
		log.Fatal("Unable to create consul client: ", err)
	}

	agent := &ConsulAgent{
		api:         consul,
		Id:          uuid.New(),
		ServiceName: serviceName,
		Tags:        []string{connection},
		Port:        port,
		Ttl:         ttl,
		CheckId:     checkId,
	}

	if err = agent.registerService(); err != nil {
		log.Fatal("Unable to register service in consul: ", err)
	}

	go agent.healthChecker(func(err error) { panic(err) })
	return agent
}
