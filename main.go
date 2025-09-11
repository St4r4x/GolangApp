package main

import (
	"log"
	"net/http"
	"os"
)

var version string = "0.0.0-local"

func main() {
    Logger.Info("Starting the server")

    app := newApp()

    // Get port from environment variable, default to 8080
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    server := http.Server{
        Addr:    ":" + port,
        Handler: app,
    }

    log.Printf("HTTP server listening on %v", server.Addr)
    server.ListenAndServe()
}