package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var (
	backends = []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}
	current uint32
)

func main() {
	http.HandleFunc("/", handleRequest)
	log.Println("Load Balancer started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic("Server Failure")
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Get the next backend server using round-robin
	backendURL := getNextBackend()

	// Parse the backend URL
	url, err := url.Parse(backendURL)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func getNextBackend() string {
	// Round-robin load balancing
	next := atomic.AddUint32(&current, 1)
	return backends[next%uint32(len(backends))]
}
