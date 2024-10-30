// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// StaticResponse holds the configuration for static responses
type StaticResponse struct {
	URL          string
	StatusCode   int
	ResponseBody string
}

// Global map for storing static responses
var staticResponses = map[string]StaticResponse{
	"/static1": {URL: "/static1", StatusCode: http.StatusOK, ResponseBody: "This is static response 1."},
	"/static2": {URL: "/static2", StatusCode: http.StatusNotFound, ResponseBody: "Static resource not found."},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	http.HandleFunc("/", requestHandler)

	log.Printf("HTTP Server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// requestHandler handles incoming HTTP requests
func requestHandler(w http.ResponseWriter, r *http.Request) {
	// Check for static responses first
	if staticResponse, exists := staticResponses[r.URL.Path]; exists {
		w.WriteHeader(staticResponse.StatusCode)
		fmt.Fprint(w, staticResponse.ResponseBody)
		return
	}

	// Forward to the backend or return a default response
	fmt.Fprintf(w, "Hello from HTTP Server on port %s\n", os.Getenv("PORT"))
}
