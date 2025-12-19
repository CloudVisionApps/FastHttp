package admin

import (
	"encoding/json"
	"os"
	"strconv"

	"fasthttp/config"

	"github.com/gofiber/fiber/v2"
)

// serveAdminInfo serves a simple info page when admin UI is not built
func (a *AdminAPI) serveAdminInfo(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>FastHTTP Admin API</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
		h1 { color: #667eea; }
		.api-endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
		code { background: #e0e0e0; padding: 2px 6px; border-radius: 3px; }
	</style>
</head>
<body>
	<h1>FastHTTP Admin API</h1>
	<p>Admin API is running. To use the web interface, build the React admin panel:</p>
	<pre><code>cd admin-ui
npm install
npm run build</code></pre>
	<p>API endpoints are available at:</p>
	<div class="api-endpoint">
		<strong>GET</strong> <code>/api/v1/health</code> - Health check
	</div>
	<div class="api-endpoint">
		<strong>GET</strong> <code>/api/v1/config</code> - Get configuration
	</div>
	<div class="api-endpoint">
		<strong>GET</strong> <code>/api/v1/virtualhosts</code> - List virtual hosts
	</div>
	<div class="api-endpoint">
		<strong>GET</strong> <code>/api/v1/stats</code> - Get server statistics
	</div>
	<p>See <code>admin/README.md</code> for full API documentation.</p>
</body>
</html>`
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// Health check endpoint
func (a *AdminAPI) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Admin API is running",
	})
}

// Config endpoints
func (a *AdminAPI) getConfig(c *fiber.Ctx) error {
	a.configMutex.RLock()
	defer a.configMutex.RUnlock()

	return c.JSON(a.config)
}

func (a *AdminAPI) updateConfig(c *fiber.Ctx) error {
	var newConfig config.Config
	if err := c.BodyParser(&newConfig); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate and compile location regexes
	for i := range newConfig.VirtualHosts {
		if err := newConfig.VirtualHosts[i].CompileLocationRegexes(); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid location regex",
				"details": err.Error(),
			})
		}
	}

	// Save to file
	configJSON, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize config",
		})
	}

	if err := os.WriteFile(a.configPath, configJSON, 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save config",
			"details": err.Error(),
		})
	}

	// Update in-memory config
	a.configMutex.Lock()
	a.config = &newConfig
	a.configMutex.Unlock()

	return c.JSON(fiber.Map{
		"message": "Configuration updated successfully",
		"config":  newConfig,
	})
}

func (a *AdminAPI) reloadConfig(c *fiber.Ctx) error {
	// Reload config from file
	newConfig, err := config.Load(a.configPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reload config",
			"details": err.Error(),
		})
	}

	a.configMutex.Lock()
	a.config = newConfig
	a.configMutex.Unlock()

	return c.JSON(fiber.Map{
		"message": "Configuration reloaded successfully",
	})
}

// Virtual host endpoints
func (a *AdminAPI) getVirtualHosts(c *fiber.Ctx) error {
	a.configMutex.RLock()
	defer a.configMutex.RUnlock()

	return c.JSON(a.config.VirtualHosts)
}

func (a *AdminAPI) getVirtualHost(c *fiber.Ctx) error {
	serverName := c.Params("serverName")
	
	a.configMutex.RLock()
	defer a.configMutex.RUnlock()

	vhost := a.config.GetVirtualHostByServerName(serverName)
	if vhost == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Virtual host not found",
		})
	}

	return c.JSON(vhost)
}

func (a *AdminAPI) createVirtualHost(c *fiber.Ctx) error {
	var vhost config.VirtualHost
	if err := c.BodyParser(&vhost); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate
	if vhost.ServerName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "serverName is required",
		})
	}

	// Compile location regexes
	if err := vhost.CompileLocationRegexes(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid location regex",
			"details": err.Error(),
		})
	}

	a.configMutex.Lock()
	// Check if already exists
	if a.config.GetVirtualHostByServerName(vhost.ServerName) != nil {
		a.configMutex.Unlock()
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Virtual host already exists",
		})
	}

	a.config.VirtualHosts = append(a.config.VirtualHosts, vhost)
	a.configMutex.Unlock()

	// Save to file
	if err := a.saveConfig(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save config",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(vhost)
}

func (a *AdminAPI) updateVirtualHost(c *fiber.Ctx) error {
	serverName := c.Params("serverName")
	
	var vhost config.VirtualHost
	if err := c.BodyParser(&vhost); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Compile location regexes
	if err := vhost.CompileLocationRegexes(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid location regex",
			"details": err.Error(),
		})
	}

	a.configMutex.Lock()
	defer a.configMutex.Unlock()

	// Find and update
	found := false
	for i := range a.config.VirtualHosts {
		if a.config.VirtualHosts[i].ServerName == serverName {
			a.config.VirtualHosts[i] = vhost
			found = true
			break
		}
	}

	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Virtual host not found",
		})
	}

	// Save to file
	if err := a.saveConfig(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save config",
			"details": err.Error(),
		})
	}

	return c.JSON(vhost)
}

func (a *AdminAPI) deleteVirtualHost(c *fiber.Ctx) error {
	serverName := c.Params("serverName")

	a.configMutex.Lock()
	defer a.configMutex.Unlock()

	// Find and remove
	found := false
	for i, vhost := range a.config.VirtualHosts {
		if vhost.ServerName == serverName {
			a.config.VirtualHosts = append(a.config.VirtualHosts[:i], a.config.VirtualHosts[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Virtual host not found",
		})
	}

	// Save to file
	if err := a.saveConfig(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save config",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Virtual host deleted successfully",
	})
}

// Location endpoints
func (a *AdminAPI) getLocations(c *fiber.Ctx) error {
	serverName := c.Params("serverName")

	a.configMutex.RLock()
	defer a.configMutex.RUnlock()

	vhost := a.config.GetVirtualHostByServerName(serverName)
	if vhost == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Virtual host not found",
		})
	}

	return c.JSON(vhost.Locations)
}

func (a *AdminAPI) createLocation(c *fiber.Ctx) error {
	serverName := c.Params("serverName")
	
	var location config.Location
	if err := c.BodyParser(&location); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	a.configMutex.Lock()
	defer a.configMutex.Unlock()

	vhost := a.config.GetVirtualHostByServerName(serverName)
	if vhost == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Virtual host not found",
		})
	}

	vhost.Locations = append(vhost.Locations, location)
	
	// Recompile all location regexes for this virtual host
	if err := vhost.CompileLocationRegexes(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid location regex",
			"details": err.Error(),
		})
	}

	// Save to file
	if err := a.saveConfig(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save config",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(location)
}

func (a *AdminAPI) updateLocation(c *fiber.Ctx) error {
	serverName := c.Params("serverName")
	indexStr := c.Params("index")
	
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid location index",
		})
	}

	var location config.Location
	if err := c.BodyParser(&location); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	a.configMutex.Lock()
	defer a.configMutex.Unlock()

	vhost := a.config.GetVirtualHostByServerName(serverName)
	if vhost == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Virtual host not found",
		})
	}

	if index >= len(vhost.Locations) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Location not found",
		})
	}

	vhost.Locations[index] = location
	
	// Recompile all location regexes for this virtual host
	if err := vhost.CompileLocationRegexes(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid location regex",
			"details": err.Error(),
		})
	}

	// Save to file
	if err := a.saveConfig(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save config",
			"details": err.Error(),
		})
	}

	return c.JSON(location)
}

func (a *AdminAPI) deleteLocation(c *fiber.Ctx) error {
	serverName := c.Params("serverName")
	indexStr := c.Params("index")
	
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid location index",
		})
	}

	a.configMutex.Lock()
	defer a.configMutex.Unlock()

	vhost := a.config.GetVirtualHostByServerName(serverName)
	if vhost == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Virtual host not found",
		})
	}

	if index >= len(vhost.Locations) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Location not found",
		})
	}

	vhost.Locations = append(vhost.Locations[:index], vhost.Locations[index+1:]...)

	// Save to file
	if err := a.saveConfig(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save config",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Location deleted successfully",
	})
}

// Server control endpoints
func (a *AdminAPI) getServerStatus(c *fiber.Ctx) error {
	// TODO: Get actual server status
	return c.JSON(fiber.Map{
		"status": "running",
		"ports":  a.config.GetAllListenPorts(),
	})
}

func (a *AdminAPI) reloadServer(c *fiber.Ctx) error {
	// Reload config
	if err := a.reloadConfig(c); err != nil {
		return err
	}

	// TODO: Signal server to reload configuration
	return c.JSON(fiber.Map{
		"message": "Server reload initiated",
	})
}

func (a *AdminAPI) restartServer(c *fiber.Ctx) error {
	// TODO: Implement server restart
	return c.JSON(fiber.Map{
		"message": "Server restart initiated",
	})
}

// Stats endpoint
func (a *AdminAPI) getStats(c *fiber.Ctx) error {
	a.configMutex.RLock()
	defer a.configMutex.RUnlock()

	return c.JSON(fiber.Map{
		"virtualHosts": len(a.config.VirtualHosts),
		"ports":        a.config.GetAllListenPorts(),
		"mimeTypes":    len(a.config.MimeTypes),
	})
}

// Helper function to save config to file
func (a *AdminAPI) saveConfig() error {
	configJSON, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.configPath, configJSON, 0644)
}
