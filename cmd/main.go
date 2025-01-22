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
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	c := config.Load()
	if c.OrgID == "" || c.SnykToken == "" || len(c.Clusters) == 0 {
		slog.Error("Missing configuration")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	taskCh := make(chan tasks.ImageTask, 1000)

	var wgWorkers = worker.StartPool(ctx, c, taskCh)
	var wgWatchers = watcher.StartWatchers(ctx, c, taskCh)

	wgWatchers.Wait()
	close(taskCh)
	wgWorkers.Wait()
	slog.Info("Done.")
	time.Sleep(100 * time.Millisecond) // small delay for final logs
}
