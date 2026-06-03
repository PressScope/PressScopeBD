package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/joho/godotenv"

	"analytics/internal/config"
	"analytics/internal/queue"
)

// batchJob bundles the pulled data so a background worker can ingest it.
type batchJob struct {
	events    []queue.EventMessage
	streamIDs []string
	createdAt time.Time
}

// Global synchronization pools to eliminate heap allocations under extreme load
var (
	bufferPool = sync.Pool{
		New: func() any { return &bytes.Buffer{} },
	}
	// Reuses the event message slices to prevent constant GC sweeping
	eventSlicePool = sync.Pool{
		New: func() any {
			b := make([]queue.EventMessage, 0, 1000)
			return &b
		},
	}
	// Reuses the stream ID string slices
	streamIDSlicePool = sync.Pool{
		New: func() any {
			s := make([]string, 0, 1000)
			return &s
		},
	}
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	logger.Info("starting application lifecycle initialization")

	if err := run(logger); err != nil {
		logger.Error("motherduck worker exited with critical error", "err", err)
		os.Exit(1)
	}
	logger.Info("application lifecycle terminated successfully")
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger.Debug("configuration values successfully loaded into memory")

	logger.Info("attempting connection to valkey stream backend", "url", cfg.Valkey.StreamName)
	valkeyClient, err := queue.NewClient(cfg.Valkey, logger)
	if err != nil {
		return fmt.Errorf("connect to valkey: %w", err)
	}
	defer valkeyClient.Close()
	logger.Info("valkey client connection verified and active")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPath := "analytics.duckdb"
	if cfg.MotherDuck.Token != "" && cfg.MotherDuck.DB != "" {
		logger.Info("motherduck credentials detected; attempting cloud handshake", "database", cfg.MotherDuck.DB)
		db, err := sql.Open("duckdb", fmt.Sprintf("md:%s?motherduck_token=%s", cfg.MotherDuck.DB, cfg.MotherDuck.Token))
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := db.PingContext(pingCtx)
			cancel()
			if err == nil {
				logger.Info("connected to cloud motherduck server instance successfully", "database", cfg.MotherDuck.DB)
				defer db.Close()
				return runWithDB(ctx, db, valkeyClient, logger, cfg)
			}
		}
		logger.Warn("failed connection interface to motherduck cloud; falling back to persistent local storage", "err", err)
	}

	logger.Info("initializing local duckdb embedded engine storage", "path", dbPath)
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("open duckdb connection: %w", err)
	}
	defer db.Close()

	return runWithDB(ctx, db, valkeyClient, logger, cfg)
}

func runWithDB(ctx context.Context, db *sql.DB, valkeyClient *queue.Client, logger *slog.Logger, cfg *config.Config) error {
	if err := ensureEventsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("ensure events table: %w", err)
	}

	logger.Info("motherduck worker starting concurrent pipeline infrastructure",
		"stream", cfg.Valkey.StreamName,
		"group", cfg.Valkey.ConsumerGroup,
		"consumer", cfg.Valkey.ConsumerName)

	// Concurrency tuning for high-throughput cloud environments
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	jobQueue := make(chan batchJob, 10)
	var wg sync.WaitGroup

	numWorkers := 4 // Scaled out slightly to handle cloud latency variations
	logger.Info("spawning pipeline worker pool", "worker_count", numWorkers, "channel_buffer_capacity", cap(jobQueue))
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerLoop(ctx, db, valkeyClient, jobQueue, logger, workerID)
		}(i)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Fetch pre-allocated slices from our pools right from the start
	batchPtr := eventSlicePool.Get().(*[]queue.EventMessage)
	streamIDsPtr := streamIDSlicePool.Get().(*[]string)

	batch := *batchPtr
	streamIDs := *streamIDsPtr

	flush := func() {
		if len(batch) > 0 {
			count := len(batch)
			logger.Debug("attempting to dispatch accumulated batch to worker channel", "batch_size", count)

			job := batchJob{events: batch, streamIDs: streamIDs, createdAt: time.Now()}

			select {
			case jobQueue <- job:
				logger.Info("successfully enqueued transaction batch into pipeline",
					"batch_size", count,
					"current_queue_depth", len(jobQueue))
			case <-time.After(2 * time.Second):
				logger.Error("pipeline saturation alert: job queue channel blocked for 2s; reverting to synchronous flush", "batch_size", count)
				start := time.Now()
				if err := processBatch(ctx, db, batch, streamIDs, valkeyClient, logger); err != nil {
					logger.Error("synchronous pipeline safety flush failed", "err", err)
				} else {
					logger.Info("synchronous backup safety flush processed successfully", "duration_ms", time.Since(start).Milliseconds())
				}
				// Recycle manually since it didn't pass to a worker
				eventSlicePool.Put(&batch)
				streamIDSlicePool.Put(&streamIDs)
			}

			// Rent completely fresh slices from the pool for the next collection sequence
			// This completely bypasses the runtime slice append reallocations.
			batchPtr = eventSlicePool.Get().(*[]queue.EventMessage)
			streamIDsPtr = streamIDSlicePool.Get().(*[]string)
			batch = (*batchPtr)[:0]
			streamIDs = (*streamIDsPtr)[:0]
		}
	}

	for {
		select {
		case <-ctx.Done():
			logger.Warn("termination runtime event captured; initiating clean application pipeline wind-down")
			flush()
			close(jobQueue)
			wg.Wait()
			logger.Info("all background data pipeline appenders closed cleanly; shutdown sequence complete")
			return nil

		case <-ticker.C:
			if len(batch) > 0 {
				logger.Debug("flush window timer interval reached", "elapsed", "5s", "pending_records", len(batch))
				flush()
			}

		default:
			startRead := time.Now()
			entries, err := valkeyClient.Read(ctx, 1000, 500*time.Millisecond)
			if err != nil {
				if ctx.Err() != nil {
					continue
				}
				logger.Error("failed execution loop when reading from valkey stream backend", "err", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(entries) > 0 {
				logger.Debug("stream records harvested from valkey",
					"count", len(entries),
					"fetch_duration_ms", time.Since(startRead).Milliseconds())
			}

			for _, entry := range entries {
				batch = append(batch, entry.Message)
				streamIDs = append(streamIDs, entry.StreamID)
			}

			if len(batch) >= 1000 {
				logger.Info("batch density limits exceeded; forcing storage append", "max_limit", 1000, "current_count", len(batch))
				flush()
			}
		}
	}
}

func workerLoop(ctx context.Context, db *sql.DB, valkeyClient *queue.Client, jobQueue <-chan batchJob, logger *slog.Logger, id int) {
	logger.Info("background channel processor thread initialized", "worker_id", id)
	for job := range jobQueue {
		queueWaitTime := time.Since(job.createdAt)
		logger.Debug("worker selected pending task from queue channel",
			"worker_id", id,
			"queue_dwell_time_ms", queueWaitTime.Milliseconds())

		start := time.Now()
		if err := processBatch(ctx, db, job.events, job.streamIDs, valkeyClient, logger); err != nil {
			logger.Error("worker failed processing batch completely", "worker_id", id, "err", err)
		} else {
			logger.Info("worker completed batch ingestion lifecycle",
				"worker_id", id,
				"batch_size", len(job.events),
				"total_ingest_duration_ms", time.Since(start).Milliseconds())
		}

		// Enterprise Recycling: Return slices to pools once processing and ACKs are complete
		eventSlicePool.Put(&job.events)
		streamIDSlicePool.Put(&job.streamIDs)
	}
	logger.Info("background channel processor thread stopped cleanly", "worker_id", id)
}

func processBatch(ctx context.Context, db *sql.DB, events []queue.EventMessage, streamIDs []string, valkeyClient *queue.Client, logger *slog.Logger) error {
	if len(events) == 0 {
		return nil
	}

	// Borrow buffer from global pool to minimize garbage collections (GC)
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	enc := json.NewEncoder(buf)

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get connection from pool: %w", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn any) error {
		nativeConn, ok := driverConn.(*duckdb.Conn)
		if !ok {
			return fmt.Errorf("unexpected driver connection type returned from application pool")
		}

		appender, err := duckdb.NewAppenderFromConn(nativeConn, "", "events")
		if err != nil {
			return fmt.Errorf("initialize appender: %w", err)
		}
		defer appender.Close()

		appendStart := time.Now()
		var serializationDuration time.Duration

		for _, evt := range events {
			var propsStr, metaStr string

			// Serialize properties using pooled buffer
			serializationStart := time.Now()
			buf.Reset()
			if err := enc.Encode(evt.Properties); err == nil {
				propsStr = buf.String()
			} else {
				propsStr = "{}\n"
			}

			// Serialize meta using pooled buffer
			buf.Reset()
			if err := enc.Encode(evt.Meta); err == nil {
				metaStr = buf.String()
			} else {
				metaStr = "{}\n"
			}
			serializationDuration += time.Since(serializationStart)

			err = appender.AppendRow(
				evt.ID,
				evt.Type,
				evt.Source,
				evt.OccurredAt,
				evt.ReceivedAt,
				evt.SessionID,
				evt.UserID,
				propsStr,
				metaStr,
			)
			if err != nil {
				return fmt.Errorf("append row error (id: %s): %w", evt.ID, err)
			}
		}

		logger.Debug("batch successfully appended to intermediate memory buffer",
			"row_count", len(events),
			"memory_append_duration_ms", time.Since(appendStart).Milliseconds(),
			"aggregated_json_serialization_duration_ms", serializationDuration.Milliseconds())

		flushStart := time.Now()
		if err := appender.Flush(); err != nil {
			return fmt.Errorf("flush appender rows: %w", err)
		}
		logger.Debug("columnar engine flush complete", "storage_write_duration_ms", time.Since(flushStart).Milliseconds())

		return nil
	})
	if err != nil {
		return fmt.Errorf("appender execution failed: %w", err)
	}

	ackStart := time.Now()
	if err := valkeyClient.Ack(ctx, streamIDs...); err != nil {
		logger.Error("failed to pass bulk confirmation acknowledgements to valkey", "err", err)
	} else {
		logger.Debug("bulk streaming confirmations verified across valkey infrastructure", "ack_duration_ms", time.Since(ackStart).Milliseconds())
	}

	return nil
}

func ensureEventsTable(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	var exists int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_name = 'events'
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check table existence: %w", err)
	}

	if exists > 0 {
		return nil
	}

	logger.Warn("target table structure 'events' missing from schema; running ddl initialization script")
	query := `
		CREATE TABLE IF NOT EXISTS events (
			event_id VARCHAR,
			event_type VARCHAR,
			source VARCHAR,
			occurred_at TIMESTAMP,
			received_at TIMESTAMP,
			session_id VARCHAR,
			user_id VARCHAR,
			properties VARCHAR,
			meta VARCHAR
		)
	`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	return nil
}