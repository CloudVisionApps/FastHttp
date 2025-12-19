package admin

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fasthttp/config"
)

// StartAdminPanel starts the admin panel API server
func StartAdminPanel(cfg *config.Config, configPath, adminPort string) {
	api := NewAdminAPI(cfg, configPath)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("[Admin API] Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := api.Shutdown(); err != nil {
			log.Printf("[Admin API] Error shutting down: %v", err)
		}
		<-ctx.Done()
		os.Exit(0)
	}()

	// Start server
	if err := api.Start(adminPort); err != nil {
		log.Fatalf("[Admin API] Failed to start: %v", err)
	}
}
