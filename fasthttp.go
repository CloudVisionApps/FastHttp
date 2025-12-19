package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"fasthttp/config"
	"fasthttp/handlers"
	"fasthttp/process"
	"fasthttp/ratelimit"
	"fasthttp/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fasthttp <command>")
		os.Exit(1)
	}

	command := os.Args[1]
	configFilePath := "fasthttp.json"

	cfg, err := config.Load(configFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Printf("Configuration loaded successfully from %s", configFilePath)

	switch command {
	case "start":
		startServer(cfg)
	case "stop":
		if err := process.Stop(); err != nil {
			log.Fatal(err)
		}
	case "status":
		if err := process.Status(cfg.HttpPort); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}
}

func startServer(cfg *config.Config) {
	// Initialize rate limiter
	maxRequests, windowSeconds := cfg.GetRateLimitConfig()
	rateLimiter := ratelimit.New(maxRequests, windowSeconds)

	// Create request handler with rate limiting middleware
	handler := handlers.New(cfg)
	rateLimitHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check rate limit
		clientIP := utils.GetClientIP(r)
		if !rateLimiter.Allow(clientIP) {
// 			log.Printf("Rate limit exceeded for IP: %s", clientIP)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		handler.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    ":" + cfg.HttpPort,
		Handler: rateLimitHandler,
	}

	// Write PID file
	if err := process.WritePID(); err != nil {
		log.Fatal(err)
	}

	log.Println("Starting FastHTTP server on port: " + cfg.HttpPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
