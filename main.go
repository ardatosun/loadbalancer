package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"loadbalancer/backends"
	"loadbalancer/health"
	"loadbalancer/serverpool"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config structure to hold the parsed YAML values
type Config struct {
	Backends            []BackendConfig `yaml:"backends"`
	HealthCheckInterval string          `yaml:"health_check_interval"`
	MaxLatency          string          `yaml:"max_latency"`
	HealthCheckTimeout  string          `yaml:"health_check_timeout"`
	MaxRetries          int             `yaml:"max_retries"`
	Port                int             `yaml:"port"`
}

// BackendConfig represents a backend server configuration
type BackendConfig struct {
	URL string `yaml:"url"`
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

	// Parse backend URLs and add them to the server pool
	for _, backend := range config.Backends {
		url, err := url.Parse(backend.URL)
		if err != nil {
			log.Fatal(err)
		}
		pool.AddBackend(&backends.Backend{URL: url, Alive: true})
	}

	// Start health checks with configured intervals
	healthCheckInterval, _ := time.ParseDuration(config.HealthCheckInterval)
	go health.CheckHealth(&pool, healthCheckInterval)

	// Handle incoming requests
	http.HandleFunc("/", handleRequest)

	port := fmt.Sprintf(":%d", config.Port)
	log.Printf("Load Balancer started on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// loadConfig parses the YAML configuration file
func loadConfig() Config {
	var config Config
	file, err := ioutil.ReadFile("config.yaml")
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

	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ServeHTTP(w, r)
}
