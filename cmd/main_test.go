package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/config"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestMainMissingConfig(t *testing.T) {
	os.Setenv("SNYK_API_TOKEN", "")
	os.Args = []string{"cmd", "-org", ""}
	defer os.Clearenv()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	main()
}

func TestMainHappyPath(t *testing.T) {
	os.Setenv("SNYK_API_TOKEN", "test-token")
	os.Args = []string{"cmd", "-org", "test-org", "-cluster", "test-kubeconfig"}
	defer os.Clearenv()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	go func() {
		main()
	}()
	<-ctx.Done()
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("SNYK_API_TOKEN", "test-token")
	os.Args = []string{"cmd", "-org", "test-org", "-cluster", "test-kubeconfig,test-context", "-rate", "10", "-burst", "10", "-concurrency", "10", "-retries", "10"}
	defer os.Clearenv()
	c := config.Load()
	if c.OrgID != "test-org" {
		t.Errorf("Expected orgID to be test-org, got %s", c.OrgID)
	}
	if c.SnykToken != "test-token" {
		t.Errorf("Expected SnykToken to be test-token, got %s", c.SnykToken)
	}
	if len(c.Clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(c.Clusters))
	}
	if c.Clusters[0].Kubeconfig != "test-kubeconfig" {
		t.Errorf("Expected kubeconfig to be test-kubeconfig, got %s", c.Clusters[0].Kubeconfig)
	}
	if c.Clusters[0].Context != "test-context" {
		t.Errorf("Expected context to be test-context, got %s", c.Clusters[0].Context)
	}
	if c.Rate != 10 {
		t.Errorf("Expected rate to be 10, got %f", c.Rate)
	}
	if c.Burst != 10 || c.Concurrency != 10 || c.MaxRetries != 10 {
		t.Errorf("Expected burst, concurrency and retries to be 10, got %d, %d, %d", c.Burst, c.Concurrency, c.MaxRetries)
	}
}
