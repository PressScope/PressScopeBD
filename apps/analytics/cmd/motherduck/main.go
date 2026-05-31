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

	_ "github.com/duckdb/duckdb-go/v2"

	"analytics/internal/config"
	"analytics/internal/queue"
)

func main() {
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

	// Connect to DuckDB (using local file for demonstration)
	// In production, this would connect to MotherDuck
	dbPath := "analytics.duckdb"
	if cfg.MotherDuck.Token != "" {
		// Try to connect to MotherDuck if token is provided
		db, err := sql.Open("duckdb", fmt.Sprintf("md:%s?motherduck_token=%s", cfg.MotherDuck.DB, cfg.MotherDuck.Token))
		if err == nil {
			// Test the connection
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err == nil {
				logger.Info("connected to motherduck", "database", cfg.MotherDuck.DB)
				defer db.Close()
				return runWithDB(context.Background(), db, valkeyClient, logger, cfg)
			}
			db.Close()
		}
		logger.Warn("failed to connect to motherduck, falling back to local duckdb", "err", err)
	}

	// Fallback to local DuckDB file
	logger.Info("using local duckdb file", "path", dbPath)
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("open duckdb connection: %w", err)
	}
	defer db.Close()

	return runWithDB(context.Background(), db, valkeyClient, logger, cfg)
}

func runWithDB(ctx context.Context, db *sql.DB, valkeyClient *queue.Client, logger *slog.Logger, cfg *config.Config) error {
	// Ensure the events table exists
	if err := ensureEventsTable(ctx, db); err != nil {
		return fmt.Errorf("ensure events table: %w", err)
	}

	logger.Info("motherduck worker starting",
		"stream", cfg.Valkey.StreamName,
		"group", cfg.Valkey.ConsumerGroup,
		"consumer", cfg.Valkey.ConsumerName)

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Main processing loop
	go func() {
		ticker := time.NewTicker(5 * time.Second) // For timeout if we don't reach 1000 events
		defer ticker.Stop()

		var batch []queue.EventMessage
		var streamIDs []string

		for {
			select {
			case <-quit:
				logger.Info("shutdown signal received")
				// Process any remaining events in the batch
				if len(batch) > 0 {
					if err := processBatch(context.Background(), db, batch, streamIDs, valkeyClient, logger); err != nil {
						logger.Error("failed to process final batch", "err", err)
					}
				}
				return
			case <-ticker.C:
				// Timeout: process batch if we have any events
				if len(batch) > 0 {
					if err := processBatch(context.Background(), db, batch, streamIDs, valkeyClient, logger); err != nil {
						logger.Error("failed to process batch on timeout", "err", err)
					}
					batch = nil
					streamIDs = nil
				}
			default:
				// Read from Valkey stream
				entries, err := valkeyClient.Read(context.Background(), 1000, 500*time.Millisecond)
				if err != nil {
					logger.Error("failed to read from stream", "err", err)
					time.Sleep(1 * time.Second) // Back off on error
					continue
				}

				if len(entries) == 0 {
					// No new messages, continue
					continue
				}

				for _, entry := range entries {
					batch = append(batch, entry.Message)
					streamIDs = append(streamIDs, entry.StreamID)
				}

				// If we've reached 1000 events, process the batch
				if len(batch) >= 1000 {
					if err := processBatch(context.Background(), db, batch, streamIDs, valkeyClient, logger); err != nil {
						logger.Error("failed to process batch", "err", err)
					}
					batch = nil
					streamIDs = nil
				}
			}
		}
	}()

	// Wait for shutdown signal
	<-quit
	logger.Info("motherduck worker shut down cleanly")
	return nil
}

func processBatch(ctx context.Context, db *sql.DB, events []queue.EventMessage, streamIDs []string, valkeyClient *queue.Client, logger *slog.Logger) error {
	if len(events) == 0 {
		return nil
	}

	logger.Info("processing batch", "count", len(events))

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Prepare the insert statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO events (
			event_id, event_type, source, occurred_at, received_at, session_id, user_id, properties, meta
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert each event
	for _, evt := range events {
		propertiesJSON, err := json.Marshal(evt.Properties)
		if err != nil {
			logger.Warn("failed to marshal properties", "event_id", evt.ID, "err", err)
			propertiesJSON = []byte(`{}`)
		}
		metaJSON, err := json.Marshal(evt.Meta)
		if err != nil {
			logger.Warn("failed to marshal meta", "event_id", evt.ID, "err", err)
			metaJSON = []byte(`{}`)
		}

		_, err = stmt.ExecContext(ctx,
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
			logger.Error("failed to insert event", "event_id", evt.ID, "err", err)
			tx.Rollback()
			return fmt.Errorf("insert event: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Acknowledge the processed events
	if err := valkeyClient.Ack(ctx, streamIDs...); err != nil {
		logger.Error("failed to acknowledge events", "err", err)
		// Note: We don't return the error here because the events are already inserted into DuckDB.
		// The unacknowledged messages will be redelivered, but we have already processed them.
		// We could choose to not acknowledge and rely on deduplication, but for simplicity,
		// we'll log the error and continue.
	}

	logger.Info("batch processed and acknowledged", "count", len(events))
	return nil
}

func ensureEventsTable(ctx context.Context, db *sql.DB) error {
	// Check if the table exists
	var exists int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'main' 
		AND table_name = 'events'
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check table existence: %w", err)
	}

	if exists > 0 {
		return nil
	}

	// Create the table
	query := `
		CREATE TABLE events (
			event_id VARCHAR,
			event_type VARCHAR,
			source VARCHAR,
			occurred_at TIMESTAMP,
			received_at TIMESTAMP,
			session_id VARCHAR,
			user_id VARCHAR,
			properties JSON,
			meta JSON
		)
	`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	return nil
}