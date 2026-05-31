package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Valkey   ValkeyConfig
	MotherDuck MotherDuckConfig
}

type ServerConfig struct {
	Port string
}

type ValkeyConfig struct {
	// URL is a full Redis/Valkey URI, e.g.:
	//   redis://user:pass@host:port/db
	//   rediss://user:pass@host:port        (TLS — Aiven, Upstash, etc.)
	URL string

	// Stream settings
	StreamName    string
	MaxStreamLen  int64
	ConsumerGroup string
	ConsumerName  string
}

type MotherDuckConfig struct {
	Token string
	DB    string
}

func Load() (*Config, error) {
	valkeyURL := getEnv("VALKEY_URL", "redis://localhost:6379")
	if valkeyURL == "" {
		return nil, fmt.Errorf("VALKEY_URL must not be empty")
	}

	maxLen, err := strconv.ParseInt(getEnv("VALKEY_STREAM_MAX_LEN", "100000"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid VALKEY_STREAM_MAX_LEN: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Valkey: ValkeyConfig{
			URL:           valkeyURL,
			StreamName:    getEnv("VALKEY_STREAM_NAME", "analytics:events"),
			MaxStreamLen:  maxLen,
			ConsumerGroup: getEnv("VALKEY_CONSUMER_GROUP", "analytics-processors"),
			ConsumerName:  getEnv("VALKEY_CONSUMER_NAME", "processor-1"),
		},
		MotherDuck: MotherDuckConfig{
			Token: getEnv("MOTHERDUCK_TOKEN", ""),
			DB:    getEnv("MOTHERDUCK_DB", "analytics"),
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
