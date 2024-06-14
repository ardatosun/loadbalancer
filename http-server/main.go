package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from HTTP Server on port %s\n", port)
	})

	log.Printf("HTTP Server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
