package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"fasthttp/admin"
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
		ports := cfg.GetAllListenPorts()
		portStr := "80"
		if len(ports) > 0 {
			portStr = ports[0]
		}
		if err := process.Status(portStr); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}
}

func startServer(cfg *config.Config) {
	// Start admin API if enabled
	if cfg.AdminEnabled {
		adminPort := cfg.AdminPort
		if adminPort == "" {
			adminPort = "8080"
		}
		go func() {
			admin.StartAdminPanel(cfg, "fasthttp.json", adminPort)
		}()
		log.Printf("[Web Server] Admin API enabled on port: %s", adminPort)
	}

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

	// Get all ports to listen on
	listenPorts := cfg.GetAllListenPorts()
	
	// If no ports configured, use default port 80
	if len(listenPorts) == 0 {
		listenPorts = []string{"80"}
	}

	// Write PID file
	if err := process.WritePID(); err != nil {
		log.Fatal(err)
	}

	// Start listening on all ports
	if len(listenPorts) == 1 {
		// Single port - simple case
		server := &http.Server{
			Addr:    ":" + listenPorts[0],
			Handler: rateLimitHandler,
		}
		log.Printf("[Web Server] Starting FastHTTP server on port: %s", listenPorts[0])
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		// Multiple ports - start goroutines for each
		for _, port := range listenPorts {
			go func(p string) {
				server := &http.Server{
					Addr:    ":" + p,
					Handler: rateLimitHandler,
				}
				log.Printf("[Web Server] Starting FastHTTP server on port: %s", p)
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("Server failed on port %s: %v", p, err)
				}
			}(port)
		}
		// Keep main goroutine alive
		select {}
	}
}
