package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("VALKEY_URL", "redis://localhost:6379")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Server.Port)
	}
	if cfg.Valkey.StreamName != "analytics:events" {
		t.Errorf("expected default stream name, got %s", cfg.Valkey.StreamName)
	}
	if cfg.Valkey.MaxStreamLen != 100000 {
		t.Errorf("expected default max stream len 100000, got %d", cfg.Valkey.MaxStreamLen)
	}
	if cfg.Valkey.ConsumerGroup != "analytics-processors" {
		t.Errorf("expected default consumer group, got %s", cfg.Valkey.ConsumerGroup)
	}
}

func TestLoadProductionValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")
	os.Setenv("VALKEY_URL", "redis://localhost:6379")
	os.Setenv("MOTHERDUCK_TOKEN", "test-token")
	os.Setenv("VALKEY_CONSUMER_GROUP", "custom-group")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Env != "production" {
		t.Errorf("expected env production, got %s", cfg.Env)
	}
	if cfg.MotherDuck.Token == "" {
		t.Error("expected MotherDuck token to be set in production")
	}
}

func TestLoadProductionMissingToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")
	os.Setenv("VALKEY_URL", "redis://localhost:6379")

	_, err := Load()
	if err == nil {
		t.Error("expected error when MOTHERDUCK_TOKEN is missing in production")
	}
}

func TestLoadCustomValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("VALKEY_URL", "rediss://user:pass@remote:6379")
	os.Setenv("VALKEY_STREAM_NAME", "custom:stream")
	os.Setenv("VALKEY_STREAM_MAX_LEN", "50000")
	os.Setenv("VALKEY_CONSUMER_GROUP", "custom-group")
	os.Setenv("VALKEY_CONSUMER_NAME", "custom-consumer")
	os.Setenv("MOTHERDUCK_DB", "CustomDB")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
}

	if cfg.Valkey.URL != "rediss://user:pass@remote:6379" {
		t.Errorf("expected custom URL, got %s", cfg.Valkey.URL)
	}
	if cfg.Valkey.StreamName != "custom:stream" {
		t.Errorf("expected custom stream name, got %s", cfg.Valkey.StreamName)
	}
	if cfg.Valkey.MaxStreamLen != 50000 {
		t.Errorf("expected custom max stream len 50000, got %d", cfg.Valkey.MaxStreamLen)
	}
	if cfg.Valkey.ConsumerName != "custom-consumer" {
		t.Errorf("expected custom consumer name, got %s", cfg.Valkey.ConsumerName)
	}
	if cfg.MotherDuck.DB != "CustomDB" {
		t.Errorf("expected custom MotherDuck DB, got %s", cfg.MotherDuck.DB)
	}
}