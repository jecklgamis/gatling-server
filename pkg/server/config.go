package server

import (
	"fmt"
	"log"
	"time"
)
import "github.com/spf13/viper"

type HttpServerConfig struct {
	Port int
}

type HttpsServerConfig struct {
	Port     int
	KeyFile  string
	CertFile string
}

type ListenerConfig struct {
	Http  *HttpServerConfig
	Https *HttpsServerConfig
}

type EventNotifierConfig struct {
	Type      string
	Enabled   bool
	ConfigMap map[string]string
}

type UploaderConfig struct {
	Type      string
	Enabled   bool
	ConfigMap map[string]string
}

type DownloaderConfig struct {
	Type      string
	Enabled   bool
	ConfigMap map[string]string
}

type HeartbeatConfig struct {
	Enabled   bool
	Frequency time.Duration
}

type Config struct {
	Server         ListenerConfig
	GatlingDir     string
	WorkspaceDir   string
	UploadDir      string
	EventNotifiers []EventNotifierConfig
	Uploaders      []UploaderConfig
	Heartbeat      HeartbeatConfig
	Downloaders    map[string]DownloaderConfig
}

func ReadConfig(env string) *Config {
	configFile := fmt.Sprintf("config-%s.yaml", env)
	log.Printf("Loading %s\n", configFile)
	viper.SetConfigName(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic("Unable to read config file", err)
	}
	config := &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		log.Fatalf("Unable to umarshall config file: %s\n", err)
	}
	return config
}
