package config

import (
	"errors"
	"github.com/spf13/viper"
	"log"
	"time"
)

type Config struct {
	App             AppConfig
	Ws              WsConfig
	NotificationBus NotificationBusConfig `mapstructure:"notification-bus"`
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
	UserTopic string `mapstructure:"user-topic"`
	Single    *RedisInstanceConfig
	Cluster   *RedisClusterConfig
}

type RedisInstanceConfig struct {
	Id   string
	Host string
	Port int
}

type RedisClusterConfig struct {
	Consul ConsulConfig
}

func (rc *RedisConfig) validate() error {
	if rc.Single == nil && rc.Cluster == nil {
		return errors.New("No defined config for redis bus, should be either single or cluster")
	}

	return nil
}

type ConsulConfig struct {
	Host             string
	Port             int
	RedisServiceName string `mapstructure:"redis-service-name"`
}

func ReadConfig() Config {
	var config Config
	readFromConfigFile()
	viper.AutomaticEnv()

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("unable to read configuration: ", err)
	}

	if err = validateConfig(config); err != nil {
		log.Fatal("unable to read configuration: ", err)
	}

	return config
}

func readFromConfigFile() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

func validateConfig(config Config) error {
	if err := config.NotificationBus.Redis.validate(); err != nil {
		return err
	}
	return nil
}
