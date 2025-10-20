package main

import (
	"fmt"
	"loadbalancer/backends"
	"loadbalancer/serverpool"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config structure to hold the parsed YAML values
type Config struct {
	Backends            []BackendConfig `yaml:"backends"`
	HealthCheckInterval string          `yaml:"health_check_interval"`
	MaxLatency          string          `yaml:"max_latency"`
	HealthCheckTimeout  string          `yaml:"health_check_timeout"`
	MaxRetries          int             `yaml:"max_retries"`
	Port                int             `yaml:"port"`
	GlobalRateLimit     int             `yaml:"global_requests_per_second"` // Global rate limit
}

// BackendConfig represents a backend server configuration
type BackendConfig struct {
	URL             string `yaml:"url"`
	RateLimitPerSec int    `yaml:"backend_requests_per_second,omitempty"` // Backend-specific rate limit
}

var pool serverpool.ServerPool

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// Load config from YAML file
	config := loadConfig()

	// Override with environment variables if available
	overrideWithEnv(&config)

	// Set the global rate limit for all backends
	pool.GlobalRateLimitPerSec = config.GlobalRateLimit

	// Parse backend URLs and add them to the server pool with rate limits
	for _, backend := range config.Backends {
		url, err := url.Parse(backend.URL)
		if err != nil {
			log.Fatal(err)
		}

		// If the backend doesn't have a specific rate limit, use the global one
		rateLimit := backend.RateLimitPerSec
		if rateLimit == 0 {
			rateLimit = config.GlobalRateLimit
		}

		// Add backend with specific or global rate limit
		pool.AddBackend(&backends.Backend{URL: url, Alive: true}, rateLimit)
	}

	// Start health checks with configured intervals
	healthCheckInterval, _ := time.ParseDuration(config.HealthCheckInterval)
	go func() {
		for {
			pool.HealthCheck()
			time.Sleep(healthCheckInterval)
		}
	}()

	// Handle incoming requests
	http.HandleFunc("/", handleRequest)

	port := fmt.Sprintf(":%d", config.Port)
	log.Printf("Load Balancer started on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// loadConfig parses the YAML configuration file
func loadConfig() Config {
	var config Config
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	return config
}

// overrideWithEnv overrides config values with environment variables if present
func overrideWithEnv(config *Config) {
	if envURLs := os.Getenv("BACKEND_URLS"); envURLs != "" {
		config.Backends = parseBackendURLs(envURLs)
	}

	if interval := os.Getenv("HEALTH_CHECK_INTERVAL"); interval != "" {
		config.HealthCheckInterval = interval
	}

	if port := os.Getenv("PORT"); port != "" {
		config.Port, _ = strconv.Atoi(port)
	}
}

// parseBackendURLs parses backend URLs from a comma-separated string
func parseBackendURLs(backendURLs string) []BackendConfig {
	var backends []BackendConfig
	for _, url := range strings.Split(backendURLs, ",") {
		backends = append(backends, BackendConfig{URL: strings.TrimSpace(url)})
	}
	return backends
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Process request
	loadBalance(w, r)

	log.Printf("Request processed in %v", time.Since(startTime))
}

func loadBalance(w http.ResponseWriter, r *http.Request) {
	backend := pool.GetNextPeer()
	if backend == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Check the rate limit for the backend
	if !pool.CheckRateLimit(backend, w) {
		return
	}

	// Create a reverse proxy to forward the request
	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ServeHTTP(w, r)

}
