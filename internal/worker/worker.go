package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/config"
	snyk "github.com/iuriikogan-snyk/snyk-kube-watcher/internal/snyk"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/tasks"

	"golang.org/x/time/rate"
)

// StartPool initializes a pool of workers that will process image scan tasks concurrently.
// It takes a context, configuration, and a channel of tasks as input.
// It returns a WaitGroup that can be used to wait for all workers to finish.
// The number of workers is determined by the Concurrency field in the configuration.
// Each worker runs in its own goroutine and processes tasks from the taskCh channel.

func StartPool(ctx context.Context, c config.Config, taskCh <-chan tasks.ImageTask) *sync.WaitGroup {
	var wg sync.WaitGroup
	for i := 0; i < c.Concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runWorker(ctx, id, c, taskCh)
		}(i)
	}
	return &wg
}

// runWorker is a worker function that processes image scan tasks.
// It takes a context, worker ID, configuration, and a channel of tasks as input.
// It logs the start and stop of the worker.
// It uses a rate limiter to control the rate of task processing.
// It maintains a map of scanned images to avoid processing the same image multiple times.
// It processes tasks from the taskCh channel until it is closed.
// For each task, it calls the process function to perform the actual scanning.
// If the task fails, it logs the error and retries the task up to MaxRetries times.
// The retry logic uses exponential backoff, starting with a 1-second delay and doubling the delay for each subsequent retry.
// If the task is successful, it adds the image to the scanned map to avoid reprocessing.
// The function returns when the context is done or the taskCh channel is closed.

func runWorker(ctx context.Context, id int, c config.Config, taskCh <-chan tasks.ImageTask) {
	slog.Info("Worker started", "id", id)
	defer slog.Info("Worker stopped", "id", id)
	limiter := rate.NewLimiter(rate.Limit(c.Rate), c.Burst)
	scanned := make(map[string]struct{})
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-taskCh:
			if !ok {
				return
			}
			if _, found := scanned[t.Image]; found {
				continue
			}
			err := process(ctx, t, c, limiter)
			if err == nil {
				scanned[t.Image] = struct{}{}
			} else {
				slog.Error("Scan failed", "image", t.Image, "err", err)
			}
		}
	}
}

func process(ctx context.Context, t tasks.ImageTask, c config.Config, limiter *rate.Limiter) error {
	var err error
	backoff := time.Second
	for i := 1; i <= c.MaxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if e := limiter.Wait(ctx); e != nil {
			return e
		}
		err = snyk.MonitorImage(ctx, t.Image, t.OrgID, c.SnykToken)
		if err == nil {
			return nil
		}
		if i == c.MaxRetries {
			return fmt.Errorf("failed after retries: %w", err)
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	return errors.New("unreachable")
}
