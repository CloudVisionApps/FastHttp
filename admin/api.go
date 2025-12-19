package admin

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"

	"fasthttp/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// AdminAPI handles the admin panel REST API
type AdminAPI struct {
	config      *config.Config
	configPath  string
	configMutex sync.RWMutex
	app         *fiber.App
}

// NewAdminAPI creates a new AdminAPI instance
func NewAdminAPI(cfg *config.Config, configPath string) *AdminAPI {
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,HEAD",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length,Content-Type",
	}))

	// Authentication middleware (if enabled)
	// Only use values from config file, no environment variables
	authConfig := NewAuthConfig(
		cfg.AdminAuthEnabled,
		cfg.AdminUsername,
		cfg.AdminPassword,
		"", // Token not supported via config for now
	)

	// Apply IP whitelist if configured
	if len(cfg.AdminIPWhitelist) > 0 {
		app.Use(IPWhitelistMiddleware(cfg.AdminIPWhitelist))
	}

	// Apply authentication to all routes except health check
	app.Use(func(c *fiber.Ctx) error {
		// Skip auth for health check
		if c.Path() == "/api/v1/health" {
			return c.Next()
		}
		return BasicAuthMiddleware(authConfig)(c)
	})

	api := &AdminAPI{
		config:     cfg,
		configPath: configPath,
		app:        app,
	}

	// Setup routes
	api.setupRoutes()

	return api
}

// setupRoutes configures all API routes
func (a *AdminAPI) setupRoutes() {
	// API routes - define before static files to ensure they work
	api := a.app.Group("/api/v1")

	// Health check
	api.Get("/health", a.healthCheck)

	// Config endpoints
	api.Get("/config", a.getConfig)
	api.Put("/config", a.updateConfig)
	api.Post("/config/reload", a.reloadConfig)

	// Virtual hosts endpoints
	api.Get("/virtualhosts", a.getVirtualHosts)
	api.Get("/virtualhosts/:serverName", a.getVirtualHost)
	api.Post("/virtualhosts", a.createVirtualHost)
	api.Put("/virtualhosts/:serverName", a.updateVirtualHost)
	api.Delete("/virtualhosts/:serverName", a.deleteVirtualHost)

	// Location endpoints
	api.Get("/virtualhosts/:serverName/locations", a.getLocations)
	api.Post("/virtualhosts/:serverName/locations", a.createLocation)
	api.Put("/virtualhosts/:serverName/locations/:index", a.updateLocation)
	api.Delete("/virtualhosts/:serverName/locations/:index", a.deleteLocation)

	// Server control endpoints
	api.Get("/server/status", a.getServerStatus)
	api.Post("/server/reload", a.reloadServer)
	api.Post("/server/restart", a.restartServer)

	// Stats endpoints
	api.Get("/stats", a.getStats)

	// Serve static files from admin-ui/dist (if it exists)
	// IMPORTANT: Static files must be registered AFTER API routes
	// Fiber matches routes in order, so API routes defined above will be checked first
	adminUIDir := "admin-ui/dist"
	if info, err := os.Stat(adminUIDir); err == nil && info.IsDir() {
		// Serve static files with explicit path exclusions for API routes
		a.app.Static("/", adminUIDir, fiber.Static{
			Index:         "index.html",
			Browse:        false,
			Download:      false,
			MaxAge:        3600,
			CacheDuration: 0,
		})
		log.Printf("Serving admin UI from: %s", adminUIDir)
		
		// Catch-all for React Router (SPA routing)
		// This must be last to allow API routes to work
		a.app.Get("*", func(c *fiber.Ctx) error {
			path := c.Path()
			// If it's an API route, return 404 (shouldn't happen as API routes are defined first)
			if len(path) >= 4 && path[:4] == "/api" {
				return c.Status(404).JSON(fiber.Map{
					"error": "Not found",
				})
			}
			// Otherwise serve index.html for client-side routing
			return c.SendFile(filepath.Join(adminUIDir, "index.html"))
		})
	} else {
		// If admin-ui/dist doesn't exist, serve a simple info page
		a.app.Get("/", a.serveAdminInfo)
		a.app.Get("*", func(c *fiber.Ctx) error {
			path := c.Path()
			// Don't interfere with API routes
			if len(path) >= 4 && path[:4] == "/api" {
				return c.Next()
			}
			return a.serveAdminInfo(c)
		})
		log.Printf("Admin UI not found at %s, serving info page", adminUIDir)
	}
}

// Start starts the admin API server
func (a *AdminAPI) Start(port string) error {
	log.Printf("Starting admin API on port %s", port)
	return a.app.Listen(":" + port)
}

// Shutdown gracefully shuts down the admin API
func (a *AdminAPI) Shutdown() error {
	return a.app.Shutdown()
}

// GetApp returns the Fiber app instance (for testing or advanced usage)
func (a *AdminAPI) GetApp() *fiber.App {
	return a.app
}
