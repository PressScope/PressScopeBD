package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env        string // e.g., "production", "staging", "development"
	Server     ServerConfig
	Valkey     ValkeyConfig
	MotherDuck MotherDuckConfig
}

type ServerConfig struct {
	Port string
}

type ValkeyConfig struct {
	URL           string
	StreamName    string
	MaxStreamLen  int64
	ConsumerGroup string
	ConsumerName  string
}

type MotherDuckConfig struct {
	Token string
	DB    string
}

// Load reads, parses, validates, and initializes the configuration architecture.
func Load() (*Config, error) {
	appEnv := getEnv("APP_ENV", "development")

	valkeyURL := getEnv("VALKEY_URL", "redis://localhost:6379")
	if valkeyURL == "" {
		return nil, errors.New("environment configuration invalid: VALKEY_URL is required")
	}

	maxLen, err := strconv.ParseInt(getEnv("VALKEY_STREAM_MAX_LEN", "100000"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("environment configuration invalid: VALKEY_STREAM_MAX_LEN parsing failed: %w", err)
	}

	// Enterprise Fix: Dynamic Unique Consumer Naming
	// Checks for cloud metadata envs (K8s pod name, ECS task ID, or hostname fallback)
	consumerName := os.Getenv("VALKEY_CONSUMER_NAME")
	if consumerName == "" {
		consumerName = os.Getenv("HOSTNAME") // Auto-assigned to pod/container ID natively
		if consumerName == "" {
			consumerName = "processor-local-fallback"
		}
	}

	cfg := &Config{
		Env: appEnv,
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Valkey: ValkeyConfig{
			URL:           valkeyURL,
			StreamName:    getEnv("VALKEY_STREAM_NAME", "analytics:events"),
			MaxStreamLen:  maxLen,
			ConsumerGroup: getEnv("VALKEY_CONSUMER_GROUP", "analytics-processors"),
			ConsumerName:  consumerName,
		},
		MotherDuck: MotherDuckConfig{
			Token: getEnv("MOTHERDUCK_TOKEN", ""),
			DB:    getEnv("MOTHERDUCK_DB", "analytics"),
		},
	}

	// Run structural validation rules immediately before giving back control
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation guard failed: %w", err)
	}

	return cfg, nil
}

// Validate handles semantic security checks for critical properties depending on the runtime stage.
func (c *Config) Validate() error {
	// If running locally, we can let things slide. If we're in production, enforce strict structural rules.
	if c.Env == "production" {
		if c.MotherDuck.Token == "" {
			return errors.New("security violation: MOTHERDUCK_TOKEN must be specified in production environments")
		}
		if c.Valkey.ConsumerGroup == "analytics-processors" && os.Getenv("VALKEY_CONSUMER_GROUP") == "" {
			return errors.New("operational warning: default consumer group name detected in production payload")
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}