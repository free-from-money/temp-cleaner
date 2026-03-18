package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/free-from-money/temp-cleaner/internal/checker"
	"github.com/free-from-money/temp-cleaner/internal/deleter"
)

func main() {
	var targetDir string
	var intervalStr string
	var timeoutStr string

	flag.StringVar(&targetDir, "target", "", "Target directory to remove")
	flag.StringVar(&intervalStr, "interval", "5s", "Polling interval (e.g., 1s, 5s)")
	flag.StringVar(&timeoutStr, "timeout", "0", "Total timeout (e.g., 1h, 30m). 0 means no timeout")
	flag.Parse()

	if targetDir == "" {
		slog.Error("target flag is required")
		os.Exit(1)
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		slog.Error("invalid interval", slog.Any("error", err))
		os.Exit(1)
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		slog.Error("invalid timeout", slog.Any("error", err))
		os.Exit(1)
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("received termination signal")
		cancel()
	}()

	slog.Info("starting temp-cleaner", slog.String("target", targetDir), slog.Duration("interval", interval), slog.Duration("timeout", timeout))

	// Run the cleanup process
	if err := runCleanup(ctx, targetDir, interval); err != nil {
		slog.Error("cleanup failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("cleanup completed successfully", slog.String("target", targetDir))
}

func runCleanup(ctx context.Context, targetDir string, interval time.Duration) error {
	chk := checker.New()
	del := deleter.New()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// Check if directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			slog.Info("target directory does not exist, nothing to do", slog.String("target", targetDir))
			return nil
		}

		inUse, err := chk.IsInUse(ctx, targetDir)
		if err != nil {
			slog.Warn("failed to check if directory is in use", slog.String("target", targetDir), slog.Any("error", err))
			// Continue to retry, maybe it's a transient error
		} else if !inUse {
			slog.Info("directory is not in use, attempting to delete", slog.String("target", targetDir))
			if err := del.RemoveAll(ctx, targetDir); err != nil {
				return fmt.Errorf("failed to remove directory: %w", err)
			}
			return nil
		} else {
			slog.Debug("directory is currently in use, waiting...", slog.String("target", targetDir))
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("cleanup timed out or cancelled: %w", ctx.Err())
		case <-ticker.C:
			// Next check
		}
	}
}
