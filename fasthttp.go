package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			fmt.Fprintf(w, "Hello, %q! Host: %s", html.EscapeString(r.URL.Path))
			log.Printf("Request from %s", r.RemoteAddr)

            log.Printf("Host: %s", html.EscapeString(r.Host))

		}),
	}

	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		log.Println("Starting server on :80")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Block until a signal is received
	<-stop
	log.Println("Shutting down server...")

	// Create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
