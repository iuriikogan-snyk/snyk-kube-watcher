package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/config"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/snyk"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/tasks"

	"golang.org/x/time/rate"
)

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
