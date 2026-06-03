package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/joho/godotenv"

	"analytics/internal/config"
	"analytics/internal/queue"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("motherduck worker exited with error", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Connect to Valkey
	valkeyClient, err := queue.NewClient(cfg.Valkey, logger)
	if err != nil {
		return fmt.Errorf("connect to valkey: %w", err)
	}
	defer valkeyClient.Close()

	// Handle graceful shutdown via context cancellation
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to DuckDB (using local file for demonstration)
	dbPath := "analytics.duckdb"
	if cfg.MotherDuck.Token != "" && cfg.MotherDuck.DB != "" {
		db, err := sql.Open("duckdb", fmt.Sprintf("md:%s?motherduck_token=%s", cfg.MotherDuck.DB, cfg.MotherDuck.Token))
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := db.PingContext(pingCtx)
			cancel()
			if err == nil {
				logger.Info("connected to motherduck", "database", cfg.MotherDuck.DB)
				defer db.Close()
				return runWithDB(ctx, db, valkeyClient, logger, cfg)
			}
		}
		logger.Warn("failed to connect to motherduck, falling back to local duckdb", "err", err)
	}

	logger.Info("using local duckdb file", "path", dbPath)
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("open duckdb connection: %w", err)
	}
	defer db.Close()

	return runWithDB(ctx, db, valkeyClient, logger, cfg)
}

func runWithDB(ctx context.Context, db *sql.DB, valkeyClient *queue.Client, logger *slog.Logger, cfg *config.Config) error {
	if err := ensureEventsTable(ctx, db); err != nil {
		return fmt.Errorf("ensure events table: %w", err)
	}

	logger.Info("motherduck worker starting",
		"stream", cfg.Valkey.StreamName,
		"group", cfg.Valkey.ConsumerGroup,
		"consumer", cfg.Valkey.ConsumerName)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var batch []queue.EventMessage
	var streamIDs []string

	flush := func() {
		if len(batch) > 0 {
			if err := processBatch(ctx, db, batch, streamIDs, valkeyClient, logger); err != nil {
				logger.Error("failed to process batch", "err", err)
			}
			batch = nil
			streamIDs = nil
		}
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received, flushing remaining records")
			flush()
			logger.Info("motherduck worker shut down cleanly")
			return nil

		case <-ticker.C:
			flush()

		default:
			entries, err := valkeyClient.Read(ctx, 1000, 500*time.Millisecond)
			if err != nil {
				if ctx.Err() != nil {
					continue
				}
				logger.Error("failed to read from stream", "err", err)
				time.Sleep(1 * time.Second)
				continue
			}

			for _, entry := range entries {
				batch = append(batch, entry.Message)
				streamIDs = append(streamIDs, entry.StreamID)
			}

			if len(batch) >= 1000 {
				flush()
			}
		}
	}
}

func processBatch(ctx context.Context, db *sql.DB, events []queue.EventMessage, streamIDs []string, valkeyClient *queue.Client, logger *slog.Logger) error {
	if len(events) == 0 {
		return nil
	}

	logger.Info("processing batch via native appender", "count", len(events))

	// 1. Grab a dedicated connection from the pool
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get connection from pool: %w", err)
	}
	defer conn.Close()

	// 2. Escape driver to leverage native DuckDB driver functionality
	err = conn.Raw(func(driverConn any) error {
		nativeConn, ok := driverConn.(*duckdb.Conn)
		if !ok {
			return fmt.Errorf("unexpected driver connection type")
		}

		// 3. Initialize the appender targeting the main schema and 'events' table
		appender, err := duckdb.NewAppenderFromConn(nativeConn, "", "events")
		if err != nil {
			return fmt.Errorf("initialize appender: %w", err)
		}
		defer appender.Close()

		// 4. Stream rows sequentially into memory
		for _, evt := range events {
			propertiesJSON, err := json.Marshal(evt.Properties)
			if err != nil {
				propertiesJSON = []byte(`{}`)
			}
			metaJSON, err := json.Marshal(evt.Meta)
			if err != nil {
				metaJSON = []byte(`{}`)
			}

			// Column ordering matches your table layout explicitly
			err = appender.AppendRow(
				evt.ID,
				evt.Type,
				evt.Source,
				evt.OccurredAt,
				evt.ReceivedAt,
				evt.SessionID,
				evt.UserID,
				string(propertiesJSON),
				string(metaJSON),
			)
			if err != nil {
				return fmt.Errorf("append row (id: %s): %w", evt.ID, err)
			}
		}

		// 5. Flush all accumulated rows out to the database at once
		if err := appender.Flush(); err != nil {
			return fmt.Errorf("flush appender rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("appender execution failed: %w", err)
	}

	logger.Info("batch appended, now acknowledging", "count", len(events))

	if err := valkeyClient.Ack(ctx, streamIDs...); err != nil {
		logger.Error("failed to acknowledge events", "err", err)
	}

	logger.Info("batch processed and acknowledged", "count", len(events))
	return nil
}

func ensureEventsTable(ctx context.Context, db *sql.DB) error {
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