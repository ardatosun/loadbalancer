package main

import (
	"loadbalancer/backends"
	"loadbalancer/health"
	"loadbalancer/serverpool"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var pool serverpool.ServerPool

func main() {
	hardcodedBackends := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	// Parse backend URLs and add them to the server pool
	for _, backend := range hardcodedBackends {
		url, err := url.Parse(backend)
		if err != nil {
			log.Fatal(err)
		}
		pool.AddBackend(&backends.Backend{URL: url, Alive: true})
	}

	// Start health checking in a separate goroutine
	go health.HealthCheck(&pool)

	// Handle incoming requests
	http.HandleFunc("/", handleRequest)
	log.Println("Load Balancer started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Get the next backend server using round-robin
	backend := pool.GetNextPeer()
	if backend != nil {
		// Create a reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(backend.URL)
		// Serve the request using the reverse proxy
		proxy.ServeHTTP(w, r)
	} else {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
	}
}
