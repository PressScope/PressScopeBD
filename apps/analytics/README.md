# Analytics Ingestion Service

A lean event ingestion API backed by **Valkey Streams**. Events are written to a durable stream and can be consumed by any downstream processor (MotherDuck worker coming next).

## Architecture

```
Client → POST /events → HTTP Handler → Valkey Stream (XADD)
                                              ↓
                                   Consumer Group (XREADGROUP)
                                              ↓
                                   [Future: MotherDuck worker]
```

## Quick Start

```bash
# 1. Start Valkey
docker compose up valkey -d

# 2. Run the server
cp .env.example .env
go run ./cmd/server

# 3. Send an event
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "type": "page_view",
    "source": "web",
    "session_id": "sess_abc123",
    "user_id": "usr_xyz",
    "properties": {
      "url": "/pricing",
      "referrer": "google.com"
    }
  }'
```

## API Reference

### `POST /events`

Ingest a single event.

**Body**

| Field        | Type              | Required | Description                                |
|--------------|-------------------|----------|--------------------------------------------|
| `type`       | string            | ✅       | Event name e.g. `page_view`, `purchase`    |
| `source`     | string            | ✅       | Origin system e.g. `web`, `ios`, `backend` |
| `occurred_at`| ISO-8601 string   |          | Client-side timestamp (defaults to now)    |
| `session_id` | string            |          | Session grouping key                       |
| `user_id`    | string            |          | Authenticated user identifier              |
| `properties` | object            |          | Arbitrary event payload                    |
| `meta`       | object (str→str)  |          | Routing/filtering key-value pairs          |

**Response `202 Accepted`**
```json
{
  "event_id": "018f4c3a-...",
  "stream_id": "1717161234567-0",
  "message": "event accepted"
}
```

---

### `POST /events/batch`

Ingest up to **500 events** in one request. Body is a JSON array of the same shape as above.

**Response `202 Accepted`**
```json
{
  "message": "batch accepted",
  "accepted": 3,
  "events": [
    { "event_id": "...", "stream_id": "..." },
    ...
  ]
}
```

---

### `GET /health`

Basic liveness probe.

### `GET /health/stream`

Returns Valkey stream diagnostics — useful for monitoring backpressure.

```json
{
  "stream_name": "analytics:events",
  "length": 4210,
  "pending_count": 0,
  "groups": 1
}
```

---

## Configuration

All config is via environment variables. See `.env.example`.

| Variable                  | Default                     | Description                       |
|---------------------------|-----------------------------|-----------------------------------|
| `PORT`                    | `8080`                      | HTTP listen port                  |
| `VALKEY_HOST`             | `localhost`                 | Valkey host                       |
| `VALKEY_PORT`             | `6379`                      | Valkey port                       |
| `VALKEY_PASSWORD`         | _(empty)_                   | Auth password                     |
| `VALKEY_DB`               | `0`                         | DB index                          |
| `VALKEY_STREAM_NAME`      | `analytics:events`          | Stream key                        |
| `VALKEY_STREAM_MAX_LEN`   | `100000`                    | Approximate stream cap (MAXLEN ~) |
| `VALKEY_CONSUMER_GROUP`   | `analytics-processors`      | Consumer group name               |
| `VALKEY_CONSUMER_NAME`    | `processor-1`               | This instance's consumer name     |

---

## Consuming Events

To read from the stream downstream (e.g. your future MotherDuck worker):

```go
entries, err := valkeyClient.Read(ctx, 100, 5*time.Second)
for _, e := range entries {
    // process e.Message ...
    valkeyClient.Ack(ctx, e.StreamID)
}
```

The `Read` method uses `XREADGROUP` with consumer-group semantics so:
- Multiple consumers can process in parallel
- Unacknowledged messages stay in the PEL and can be reclaimed
- `Ack` removes messages from the PEL after successful processing

---

## Docker

```bash
docker compose up --build
```

---

## Roadmap

- [ ] MotherDuck worker — consumes the stream and bulk-inserts into DuckDB
- [ ] Dead-letter stream for poison messages
- [ ] Event schema validation (JSON Schema)
- [ ] Prometheus `/metrics` endpoint
