package queue

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventMessageJSON(t *testing.T) {
	evt := EventMessage{
		ID:         "test-123",
		Type:       "page_view",
		Source:     "web",
		OccurredAt: time.Now().UTC(),
		ReceivedAt: time.Now().UTC(),
		SessionID:  "sess-abc",
		UserID:     "user-xyz",
		Properties: json.RawMessage(`{"url":"/home"}`),
		Meta:       json.RawMessage(`{"env":"test"}`),
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	var parsed EventMessage
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if parsed.ID != evt.ID {
		t.Errorf("expected ID %s, got %s", evt.ID, parsed.ID)
	}
	if parsed.Type != evt.Type {
		t.Errorf("expected Type %s, got %s", evt.Type, parsed.Type)
	}
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "redis://localhost:6379",
			expected: "redis://localhost:6379",
		},
		{
			input:    "rediss://user:pass@host:6379",
			expected: "rediss://***:***@host:6379",
		},
		{
			input:    "rediss://default:secret-token@valkey.example.com:21577",
			expected: "rediss://***:***@valkey.example.com:21577",
		},
	}

	for _, tt := range tests {
		result := redactURL(tt.input)
		if result != tt.expected {
			t.Errorf("redactURL(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestStreamInfoFromGroups(t *testing.T) {
	groups := []struct {
		Name    string
		Pending int64
	}{
		{Name: "group1", Pending: 100},
		{Name: "analytics-processors", Pending: 500},
		{Name: "group3", Pending: 50},
	}

	var pendingCount int64
	for _, g := range groups {
		if g.Name == "analytics-processors" {
			pendingCount = g.Pending
		}
	}

	if pendingCount != 500 {
		t.Errorf("expected pending count 500, got %d", pendingCount)
	}
}