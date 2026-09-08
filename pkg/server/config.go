package server

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
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

// TaskSubmitConfig scopes what /task/submit is allowed to fetch, so a valid
// API token can't be used to pivot the server into reaching arbitrary
// internal/cloud-metadata hosts. AllowedHttpHosts defaults to
// handler.DefaultAllowedHttpHosts when left empty.
type TaskSubmitConfig struct {
	AllowedHttpHosts []string
}

type Config struct {
	Server         ListenerConfig
	ScriptsDir     string
	WorkspaceDir   string
	UploadDir      string
	EventNotifiers []EventNotifierConfig
	Uploaders      []UploaderConfig
	Heartbeat      HeartbeatConfig
	Downloaders    map[string]DownloaderConfig
	TaskTimeout    time.Duration
	LogLevel       string
	TaskSubmit     TaskSubmitConfig
}

// ParseLogLevel maps a config log level string (debug/info/warn/error, case
// insensitive) to a slog.Level, defaulting to Info for empty or unrecognized
// values.
func ParseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "", "info":
		return slog.LevelInfo
	default:
		slog.Warn("Unrecognized log level, defaulting to info", "level", level)
		return slog.LevelInfo
	}
}

func ReadConfig(env string) *Config {
	configFile := fmt.Sprintf("config-%s.yaml", env)
	slog.Info("Loading config", "file", configFile)
	viper.SetConfigName(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	err := viper.ReadInConfig()
	if err != nil {
		slog.Error("Unable to read config file", "error", err)
		panic(err)
	}
	config := &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		slog.Error("Unable to unmarshal config file", "error", err)
		os.Exit(1)
	}
	return config
}
