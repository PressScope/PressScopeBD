package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"analytics/internal/api"
	"analytics/internal/config"
	"analytics/internal/queue"
)

func main() {
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
	defer valkeyClient.Close()

	app := api.NewHandler(valkeyClient, logger).App()

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutdown signal received")
		if err := app.Shutdown(); err != nil {
			logger.Error("error during shutdown", "err", err)
		}
	}()

	addr := ":" + cfg.Server.Port
	logger.Info("analytics server starting", "addr", addr)
	if err := app.Listen(addr); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	logger.Info("server shut down cleanly")
	return nil
}
