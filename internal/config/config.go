package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

type ClusterConfig struct {
	Kubeconfig string
	Context    string
}

type Config struct {
	Clusters    []ClusterConfig
	OrgID       string
	SnykToken   string
	Rate        float64
	Burst       int
	Concurrency int
	MaxRetries  int
}

type multiClusterFlag []ClusterConfig

func (m *multiClusterFlag) String() string { return "" }
func (m *multiClusterFlag) Set(v string) error {
	parts := strings.Split(v, ",")
	c := ClusterConfig{Kubeconfig: parts[0]}
	if len(parts) > 1 {
		c.Context = parts[1]
	}
	*m = append(*m, c)
	return nil
}

func Load() Config {
	var clusters multiClusterFlag
	var orgID string
	var rate float64
	var burst, concurrency, maxRetries int
	if home := os.Getenv("HOME"); home != "" {
		flag.StringVar(&clusters[0].Kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "")
	}
	flag.Var(&clusters, "cluster", "")
	flag.StringVar(&orgID, "org", "", "")
	flag.Float64Var(&rate, "rate", 2, "")
	flag.IntVar(&burst, "burst", 2, "")
	flag.IntVar(&concurrency, "concurrency", 5, "")
	flag.IntVar(&maxRetries, "retries", 3, "")
	flag.Parse()

	var finalClusters []ClusterConfig
	for _, c := range clusters {
		if c.Kubeconfig != "" {
			finalClusters = append(finalClusters, c)
		}
	}
	if len(finalClusters) == 0 {
		finalClusters = []ClusterConfig{{Kubeconfig: filepath.Join(os.Getenv("HOME"), ".kube", "config")}}
	}

	return Config{
		Clusters:    finalClusters,
		OrgID:       orgID,
		SnykToken:   os.Getenv("SNYK_API_TOKEN"),
		Rate:        rate,
		Burst:       burst,
		Concurrency: concurrency,
		MaxRetries:  maxRetries,
	}
}
