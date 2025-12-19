package admin

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fasthttp/config"
	"fasthttp/utils"
)

// StartAdminPanel starts the admin panel API server
func StartAdminPanel(cfg *config.Config, configPath, adminPort string) {
	api := NewAdminAPI(cfg, configPath)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		utils.AdminLog("[Admin API] Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := api.Shutdown(); err != nil {
			utils.ErrorLog("[Admin API] Error shutting down: %v", err)
		}
		<-ctx.Done()
		os.Exit(0)
	}()

	// Start server
	if err := api.Start(adminPort); err != nil {
		utils.ErrorLog("[Admin API] Failed to start: %v", err)
		os.Exit(1)
	}
}
