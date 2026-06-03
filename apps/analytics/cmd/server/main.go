package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"analytics/internal/api"
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
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	valkeyClient, err := queue.NewClient(cfg.Valkey, logger)
	if err != nil {
		return fmt.Errorf("connect to valkey: %w", err)
	}
	// This will now execute safely AFTER the Fiber server completely finishes draining requests
	defer func() {
		logger.Info("disconnecting from valkey backend pool")
		valkeyClient.Close()
	}()

	app := api.NewHandler(valkeyClient, logger).App()

	// 1. Start the Fiber engine on a background thread so it doesn't block signal interception
	serverErrors := make(chan error, 1)
	addr := ":" + cfg.Server.Port

	go func() {
		logger.Info("analytics server starting", "addr", addr)
		if err := app.Listen(addr); err != nil {
			serverErrors <- err
		}
	}()

	// 2. Set up our main signal trap for orchestration
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// 3. Block execution until the server fails naturally OR we capture a termination event
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server unexpected runtime failure: %w", err)

	case sig := <-shutdownSignal:
		logger.Warn("lifecycle interruption event captured", "signal", sig.String())
		logger.Info("initiating graceful shutdown; draining active request traffic")

		// Enforce a hard timeout limit for the graceful shutdown phase
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Pass the context down to Fiber to safely wind down active listeners
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			return fmt.Errorf("forced server shutdown failure: %w", err)
		}
	}

	logger.Info("server HTTP listeners shut down cleanly")
	return nil
}