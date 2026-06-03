package api

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"

	"analytics/internal/queue"
)

// IngestRequest accepts raw JSON bytes for unstructured fields.
type IngestRequest struct {
	Type       string          `json:"type"`
	Source     string          `json:"source"`
	OccurredAt *time.Time      `json:"occurred_at,omitempty"`
	SessionID  string          `json:"session_id,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	Properties json.RawMessage `json:"properties,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
}

// IngestResponse is returned on a successful POST /events.
type IngestResponse struct {
	EventID  string `json:"event_id"`
	StreamID string `json:"stream_id"`
	Message  string `json:"message"`
}

// ErrorResponse wraps API errors.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// StreamStatsResponse is returned by GET /health/stream.
type StreamStatsResponse struct {
	StreamName   string `json:"stream_name"`
	Length       int64  `json:"length"`
	PendingCount int64  `json:"pending_count"`
	Groups       int64  `json:"groups"`
}

type Handler struct {
	queue  *queue.Client
	logger *slog.Logger
}

func NewHandler(q *queue.Client, logger *slog.Logger) *Handler {
	return &Handler{queue: q, logger: logger}
}

func (h *Handler) App() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "analytics",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: h.errorHandler,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: `{"time":"${time}","request_id":"${locals:requestid}","method":"${method}","path":"${path}","status":${status},"latency":"${latency}","bytes_out":${bytesSent}}` + "\n",
	}))

	app.Post("/events", h.IngestEvent)
	app.Post("/events/batch", h.IngestBatch)
	app.Get("/health", h.Health)
	app.Get("/health/stream", h.StreamStats)

	return app
}

// IngestEvent handles POST /events.
func (h *Handler) IngestEvent(c *fiber.Ctx) error {
	var req IngestRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body: "+err.Error())
	}

	if req.Type == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "missing required field: type")
	}
	if req.Source == "" {
		req.Source = "unknown"
	}

	occurredAt := time.Now().UTC()
	if req.OccurredAt != nil {
		occurredAt = req.OccurredAt.UTC()
	}

	evt := queue.EventMessage{
		ID:         uuid.NewString(),
		Type:       req.Type,
		Source:     req.Source,
		OccurredAt: occurredAt,
		ReceivedAt: time.Now().UTC(),
		SessionID:  req.SessionID,
		UserID:     req.UserID,
		Properties: req.Properties,
		Meta:       req.Meta,
	}

	streamID, err := h.queue.Publish(c.Context(), evt)
	if err != nil {
		h.logger.Error("failed to publish event", "err", err, "type", req.Type)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to publish event")
	}

	return c.Status(fiber.StatusAccepted).JSON(IngestResponse{
		EventID:  evt.ID,
		StreamID: streamID,
		Message:  "event accepted",
	})
}

// IngestBatch handles POST /events/batch with native Redis/Valkey pipelining.
func (h *Handler) IngestBatch(c *fiber.Ctx) error {
	var reqs []IngestRequest
	if err := c.BodyParser(&reqs); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body — expected array: "+err.Error())
	}

	if len(reqs) == 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "batch must contain at least one event")
	}
	if len(reqs) > 500 {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "batch limit is 500 events")
	}

	// Pre-allocate slice capacity to ensure memory packing efficiency
	eventsToPublish := make([]queue.EventMessage, len(reqs))
	now := time.Now().UTC()

	for i, req := range reqs {
		if req.Type == "" {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "all events require a type field")
		}
		if req.Source == "" {
			req.Source = "unknown"
		}

		occurredAt := now
		if req.OccurredAt != nil {
			occurredAt = req.OccurredAt.UTC()
		}

		eventsToPublish[i] = queue.EventMessage{
			ID:         uuid.NewString(),
			Type:       req.Type,
			Source:     req.Source,
			OccurredAt: occurredAt,
			ReceivedAt: now,
			SessionID:  req.SessionID,
			UserID:     req.UserID,
			Properties: req.Properties,
			Meta:       req.Meta,
		}
	}

	// Dynamic Change: Publish entire slice in a single round-trip over the internet
	if err := h.queue.PublishBatch(c.Context(), eventsToPublish); err != nil {
		h.logger.Error("pipelined batch publish failed", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to publish event batch")
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message":  "batch accepted and pipelined",
		"accepted": len(eventsToPublish),
	})
}

// Health handles GET /health.
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// StreamStats handles GET /health/stream.
func (h *Handler) StreamStats(c *fiber.Ctx) error {
	info, err := h.queue.Info(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch stream info")
	}

	return c.JSON(StreamStatsResponse{
		StreamName:   info.StreamName,
		Length:       info.Length,
		PendingCount: info.PendingCount,
		Groups:       info.Groups,
	})
}

// errorHandler formats all fiber errors as JSON.
func (h *Handler) errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "internal server error"

	if fe, ok := err.(*fiber.Error); ok {
		code = fe.Code
		msg = fe.Message
	}

	return c.Status(code).JSON(ErrorResponse{Error: msg})
}