package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"

	"analytics/internal/config"
)

// EventMessage is the envelope written to the Valkey stream.
type EventMessage struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Source     string            `json:"source"`
	OccurredAt time.Time         `json:"occurred_at"`
	ReceivedAt time.Time         `json:"received_at"`
	SessionID  string            `json:"session_id,omitempty"`
	UserID     string            `json:"user_id,omitempty"`
	Properties map[string]any    `json:"properties,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
}

// StreamEntry is returned when reading from the stream.
type StreamEntry struct {
	StreamID string
	Message  EventMessage
}

// StreamInfo holds diagnostic data about the stream.
type StreamInfo struct {
	StreamName   string
	Length       int64
	PendingCount int64
	Groups       int64
}

// Client wraps a valkey-go client and exposes stream operations.
type Client struct {
	vk     valkey.Client
	rdb    valkeycompat.Cmdable
	cfg    config.ValkeyConfig
	logger *slog.Logger
}

// NewClient parses the URI, connects, pings, and ensures the consumer group exists.
func NewClient(cfg config.ValkeyConfig, logger *slog.Logger) (*Client, error) {
	opts, err := valkey.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse valkey url: %w", err)
	}

	vk, err := valkey.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("create valkey client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := vk.Do(ctx, vk.B().Ping().Build()).Error(); err != nil {
		vk.Close()
		return nil, fmt.Errorf("valkey ping failed: %w", err)
	}

	rdb := valkeycompat.NewAdapter(vk)
	c := &Client{vk: vk, rdb: rdb, cfg: cfg, logger: logger}

	if err := c.ensureConsumerGroup(ctx); err != nil {
		vk.Close()
		return nil, err
	}

	logger.Info("valkey connected", "url", redactURL(cfg.URL), "stream", cfg.StreamName)
	return c, nil
}

// ensureConsumerGroup issues XGROUP CREATE MKSTREAM, ignoring "already exists".
func (c *Client) ensureConsumerGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("xgroup create: %w", err)
	}
	return nil
}

// Publish writes an EventMessage to the Valkey stream via XADD MAXLEN ~.
func (c *Client) Publish(ctx context.Context, evt EventMessage) (string, error) {
	payload, err := json.Marshal(evt)
	if err != nil {
		return "", fmt.Errorf("marshal event: %w", err)
	}

	streamID, err := c.rdb.XAdd(ctx, valkeycompat.XAddArgs{
		Stream: c.cfg.StreamName,
		MaxLen: c.cfg.MaxStreamLen,
		Approx: true,
		ID:     "*",
		Values: []any{
			"event_id", evt.ID,
			"event_type", evt.Type,
			"source", evt.Source,
			"occurred_at", evt.OccurredAt.UTC().Format(time.RFC3339Nano),
			"payload", string(payload),
		},
	}).Result()

	if err != nil {
		return "", fmt.Errorf("xadd: %w", err)
	}

	c.logger.Debug("event published", "stream_id", streamID, "event_id", evt.ID, "type", evt.Type)
	return streamID, nil
}

// Read fetches up to count new messages from the stream for this consumer.
func (c *Client) Read(ctx context.Context, count int64, block time.Duration) ([]StreamEntry, error) {
	results, err := c.rdb.XReadGroup(ctx, valkeycompat.XReadGroupArgs{
		Group:    c.cfg.ConsumerGroup,
		Consumer: c.cfg.ConsumerName,
		Streams:  []string{c.cfg.StreamName, ">"},
		Count:    count,
		Block:    block,
		NoAck:    false,
	}).Result()

	if err != nil && strings.Contains(err.Error(), "redis: nil") {
		return nil, nil // no new messages
	}
	if err != nil {
		return nil, fmt.Errorf("xreadgroup: %w", err)
	}

	var entries []StreamEntry
	for _, stream := range results {
		for _, msg := range stream.Messages {
			payload, ok := msg.Values["payload"].(string)
			if !ok {
				c.logger.Warn("stream message missing payload", "stream_id", msg.ID)
				continue
			}

			var evt EventMessage
			if err := json.Unmarshal([]byte(payload), &evt); err != nil {
				c.logger.Warn("unmarshal event failed", "stream_id", msg.ID, "err", err)
				continue
			}

			entries = append(entries, StreamEntry{StreamID: msg.ID, Message: evt})
		}
	}

	return entries, nil
}

// Ack removes messages from the PEL after successful processing.
func (c *Client) Ack(ctx context.Context, streamIDs ...string) error {
	if len(streamIDs) == 0 {
		return nil
	}
	return c.rdb.XAck(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, streamIDs...).Err()
}

// Info returns diagnostic metadata about the stream.
func (c *Client) Info(ctx context.Context) (*StreamInfo, error) {
	length, err := c.rdb.XLen(ctx, c.cfg.StreamName).Result()
	if err != nil {
		return nil, fmt.Errorf("xlen: %w", err)
	}

	var pendingCount, groupCount int64
	groups, err := c.rdb.XInfoGroups(ctx, c.cfg.StreamName).Result()
	if err == nil {
		groupCount = int64(len(groups))
		for _, g := range groups {
			if g.Name == c.cfg.ConsumerGroup {
				pendingCount = g.Pending
			}
		}
	}

	return &StreamInfo{
		StreamName:   c.cfg.StreamName,
		Length:       length,
		PendingCount: pendingCount,
		Groups:       groupCount,
	}, nil
}

// Close shuts down the underlying valkey-go client.
func (c *Client) Close() {
	c.vk.Close()
}

// redactURL hides the password in a URI for safe logging.
func redactURL(raw string) string {
	for i := 0; i < len(raw); i++ {
		if raw[i] == '@' {
			for j := 0; j < i; j++ {
				if raw[j] == ':' && j+2 < len(raw) && raw[j+1] == '/' && raw[j+2] == '/' {
					return raw[:j+3] + "***:***@" + raw[i+1:]
				}
			}
		}
	}
	return raw
}
