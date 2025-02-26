package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/config"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/tasks"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/watcher"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/worker"
)

func main() {
	// Configure the default logger to output text-based logs to standard error (stderr).
	// It's set to log at the Info level.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Load the application's configuration from environment variables and command-line flags.
	c := config.Load()

	// Validate the essential configuration parameters: OrgID, SnykToken, and at least one cluster.
	// If any of these are missing, log an error and exit with a non-zero status code (1).
	if c.OrgID == "" || c.SnykToken == "" || len(c.Clusters) == 0 {
		slog.Error("Missing configuration")
		os.Exit(1)
	}

	// Create a new context that will be canceled when either an interrupt signal (Ctrl+C) or
	// a SIGTERM signal is received. This allows for graceful shutdown of the application.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel() // Ensure the context is canceled when main exits.

	// Create a buffered channel to hold image scan tasks.
	// The buffer size of 1000 allows for queuing up a reasonable number of tasks.
	taskCh := make(chan tasks.ImageTask, 1000)

	// Start the worker pool, which will process image scan tasks.
	// This function returns a WaitGroup that we'll use to wait for all workers to finish.
	var wgWorkers = worker.StartPool(ctx, c, taskCh)

	// Start the watchers, which monitor Kubernetes clusters for new or updated pods.
	// Each watcher will enqueue image scan tasks onto the taskCh.
	// This also returns a WaitGroup for waiting on all watchers.
	var wgWatchers = watcher.StartWatchers(ctx, c, taskCh)

	// Wait for all the watchers to finish their work. This typically means that either the
	// context has been canceled (due to a signal) or an unrecoverable error has occurred.
	wgWatchers.Wait()

	// Once all watchers have exited, close the task channel.
	// This signals to the workers that no more tasks will be enqueued.
	close(taskCh)

	// Wait for all the workers to finish processing any remaining tasks.
	wgWorkers.Wait()

	// Log a message indicating that all operations are complete.
	slog.Info("Done.")

	// Introduce a small delay to ensure that all logs are flushed before the application exits.
	time.Sleep(100 * time.Millisecond)
}
