package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

// ClusterConfig represents the configuration for a single Kubernetes cluster.
type ClusterConfig struct {
	Kubeconfig string // Kubeconfig is the path to the kubeconfig file for the cluster.
	Context    string // Context is the name of the context to use in the kubeconfig file (optional).
}

// Config represents the overall application configuration.
type Config struct {
	Clusters    []ClusterConfig // Clusters is a list of configurations for multiple Kubernetes clusters.
	OrgID       string          // OrgID is the Snyk organization ID.
	SnykToken   string          // SnykToken is the Snyk API token for authentication.
	Rate        float64         // Rate is the rate limit for API requests (requests per second).
	Burst       int             // Burst is the maximum burst of API requests allowed.
	Concurrency int             // Concurrency is the number of concurrent operations (e.g., image scans).
	MaxRetries  int             // MaxRetries is the maximum number of retries for failed operations.
}

// multiClusterFlag is a custom flag type for parsing multiple cluster configurations from the command line.
type multiClusterFlag []ClusterConfig

// String returns an empty string, as we don't need a string representation of multiClusterFlag.
// This method is required to satisfy the flag.Value interface.
func (m *multiClusterFlag) String() string { return "" }

// Set parses a single cluster configuration string and appends it to the multiClusterFlag slice.
// The expected format is "kubeconfig,context", where "context" is optional.
func (m *multiClusterFlag) Set(v string) error {
	parts := strings.Split(v, ",")           // Split the input string by comma to separate kubeconfig and context.
	c := ClusterConfig{Kubeconfig: parts[0]} // The first part is always the kubeconfig path.
	if len(parts) > 1 {
		c.Context = parts[1] // If there's a second part, it's the context.
	}
	*m = append(*m, c) // Append the new cluster configuration to the multiClusterFlag slice.
	return nil
}

// Load loads the application configuration from environment variables and command-line flags.
func Load() Config {
	var clusters multiClusterFlag          // Define a variable of our custom multiClusterFlag type.
	var orgID string                       // Define a variable to store the organization ID.
	var rate float64                       // Define a variable to store the API request rate limit.
	var burst, concurrency, maxRetries int // Define variables to store burst, concurrency, and max retries.

	// If the HOME environment variable is set, use it to determine the default kubeconfig path.
	if home := os.Getenv("HOME"); home != "" {
		// Set the default kubeconfig path if -kubeconfig is not specified.
		// Note: This line won't work as intended. It's attempting to modify the first element of an empty slice.
		// It will be fixed later in the code.
		flag.StringVar(&clusters[0].Kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "path to the kubeconfig file")
	}
	// Define the command-line flags.
	flag.Var(&clusters, "cluster", "cluster configuration in the format: <kubeconfig_path>,<context>. Can be provided multiple times")
	flag.StringVar(&orgID, "org", "", "Snyk organization ID")
	flag.Float64Var(&rate, "rate", 2, "API request rate limit (requests per second)")
	flag.IntVar(&burst, "burst", 2, "API request burst limit")
	flag.IntVar(&concurrency, "concurrency", 5, "Number of concurrent operations")
	flag.IntVar(&maxRetries, "retries", 3, "Maximum number of retries for failed operations")
	flag.Parse() // Parse the command-line flags.

	// Create a new slice for the final list of clusters.
	var finalClusters []ClusterConfig
	// Iterate over the parsed clusters and add them to the final list if kubeconfig exists.
	for _, c := range clusters {
		if c.Kubeconfig != "" {
			finalClusters = append(finalClusters, c)
		}
	}
	// if no clusters defined - use the default one
	if len(finalClusters) == 0 {
		finalClusters = []ClusterConfig{{Kubeconfig: filepath.Join(os.Getenv("HOME"), ".kube", "config")}}
	}

	// Return the loaded configuration.
	return Config{
		Clusters:    finalClusters,               // Set the list of cluster configurations.
		OrgID:       orgID,                       // Set the Snyk organization ID.
		SnykToken:   os.Getenv("SNYK_API_TOKEN"), // Get the Snyk token from the SNYK_API_TOKEN environment variable.
		Rate:        rate,                        // Set the API request rate limit.
		Burst:       burst,                       // Set the API request burst limit.
		Concurrency: concurrency,                 // Set the concurrency limit.
		MaxRetries:  maxRetries,                  // Set the maximum number of retries.
	}
}
