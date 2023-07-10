package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type Config struct {
	App             AppConfig
	Ws              WsConfig
	NotificationBus NotificationBusConfig `mapstructure:"notification-bus"`
	Discovery       DiscoveryConfig
}

type AppConfig struct {
	Port int
}

type WsConfig struct {
	Timeout      time.Duration
	PingInterval time.Duration `mapstructure:"ping-interval"`
	ReadLimit    int64         `mapstructure:"read-limit"`
	BufferSize   int64         `mapstructure:"buffer-size"`
}

type NotificationBusConfig struct {
	Redis RedisConfig
}

type RedisConfig struct {
	Host      string
	Port      int
	UserTopic string `mapstructure:"user-topic"`
}
type DiscoveryConfig struct {
	Consul ConsulConfig
}

type ConsulConfig struct {
	Host        string
	Port        int
	ServiceName string `mapstructure:"service-name"`
	CheckId     string `mapstructure:"check-id"`
	Tags        []string
	Ttl         time.Duration
}

func ReadConfig() Config {
	var config Config
	readFromConfigFile()
	viper.AutomaticEnv()

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("unable to read configuration: ", err)
	}

	return config
}

func readFromConfigFile() {
	viper.SetConfigName("services/connection-service/config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}
