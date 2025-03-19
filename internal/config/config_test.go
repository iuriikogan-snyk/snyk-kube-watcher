package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("SNYK_API_TOKEN", "test-token")
	os.Args = []string{"cmd", "-org", "test-org", "-cluster", "test-kubeconfig,test-context", "-rate", "10", "-burst", "10", "-concurrency", "10", "-retries", "10"}
	defer os.Clearenv()
	c := Load()
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

func TestLoadConfigDefaultCluster(t *testing.T) {
	os.Setenv("SNYK_API_TOKEN", "test-token")
	os.Args = []string{"cmd", "-org", "test-org"}
	defer os.Clearenv()
	c := Load()
	if len(c.Clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(c.Clusters))
	}
	expectedKubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if c.Clusters[0].Kubeconfig != expectedKubeconfig {
		t.Errorf("Expected kubeconfig to be %s, got %s", expectedKubeconfig, c.Clusters[0].Kubeconfig)
	}
}

func TestLoadConfigNoCluster(t *testing.T) {
	os.Args = []string{"cmd", "-org", "test-org", "-cluster", ""}
	defer os.Clearenv()
}
