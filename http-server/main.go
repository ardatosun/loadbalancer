package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type StaticResponse struct {
	URL          string `yaml:"url"`
	StatusCode   int    `yaml:"status_code"`
	ResponseBody string `yaml:"response_body"`
}

// Backend holds configuration for each backend server
type Backend struct {
	URL                      string `yaml:"url"`
	BackendRequestsPerSecond int    `yaml:"backend_requests_per_second"`
}

// Config holds all configurations for the server
type Config struct {
	StaticResponses         []StaticResponse `yaml:"static_responses"`
	GlobalRequestsPerSecond int              `yaml:"global_requests_per_second"`
	Backends                []Backend        `yaml:"backends"`
	HealthCheckInterval     string           `yaml:"health_check_interval"`
	MaxLatency              string           `yaml:"max_latency"`
	HealthCheckTimeout      string           `yaml:"health_check_timeout"`
	MaxRetries              int              `yaml:"max_retries"`
	Port                    int              `yaml:"port"`
}

// Global map for storing static responses, with mutex for concurrent access
var (
	staticResponses = make(map[string]StaticResponse)
	mu              sync.RWMutex // Mutex to protect staticResponses map
)

// LoadConfig reads the configuration file and unmarshals it into the Config struct
func LoadConfig(filepath string) (*Config, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// updateStaticResponses safely updates the staticResponses map with new values
func updateStaticResponses(newResponses []StaticResponse) {
	mu.Lock()
	defer mu.Unlock()
	staticResponses = make(map[string]StaticResponse) // Clear existing map
	for _, resp := range newResponses {
		staticResponses[resp.URL] = resp
	}
	log.Println("Static responses updated.")
}

// reloadConfig loads and applies the configuration file to update server behavior
func reloadConfig(filepath string) error {
	config, err := LoadConfig(filepath)
	if err != nil {
		return err
	}
	updateStaticResponses(config.StaticResponses)
	return nil
}

// WatchConfig watches the configuration file for changes and reloads configurations when changes are detected
func WatchConfig(filepath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(filepath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Config file changed, reloading...")
				if err := reloadConfig(filepath); err != nil {
					log.Printf("Error reloading config: %v\n", err)
				}
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v\n", err)
		}
	}
}

// requestHandler handles incoming HTTP requests
func requestHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	if staticResponse, exists := staticResponses[r.URL.Path]; exists {
		w.WriteHeader(staticResponse.StatusCode)
		fmt.Fprint(w, staticResponse.ResponseBody)
		return
	}

	fmt.Fprintf(w, "Hello from HTTP Server on port %s\n", os.Getenv("PORT"))
}

func main() {

	config, err := LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	updateStaticResponses(config.StaticResponses)

	go WatchConfig("config.yml")

	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", config.Port)
	}

	// setup new http mux for incoming requests
	httpEngine := http.NewServeMux()
	httpEngine.HandleFunc("/", requestHandler)

	httpSrv := &http.Server{
		Addr:    port,
		Handler: httpEngine,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server experienced a shutdown")
		}
	}()

	log.Printf("HTTP Server started on port %s\n", port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	shutdownSignal := <-signalChan
	log.Printf("Received signal: %s. Shutting down", shutdownSignal)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("Failed to shut down server gracefully")
	} else {
		log.Printf("Server shut down gracefully")
	}
}
