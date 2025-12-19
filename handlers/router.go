package handlers

import (
	"net/http"

	"fasthttp/config"
	"fasthttp/utils"
)

// Router selects and executes the appropriate handler for a request
type Router struct {
	handlers []RequestHandler
	config   *config.Config
}

// NewRouter creates a new Router with all available handlers
func NewRouter(cfg *config.Config) *Router {
	return &Router{
		handlers: []RequestHandler{
			NewPHPHandler(),
			NewCGIHandler(),
			NewProxyHandler(),
			NewStaticFileHandler(), // Always last as fallback
		},
		config: cfg,
	}
}

// HandleRequest routes the request to the appropriate handler
func (r *Router) HandleRequest(w http.ResponseWriter, req *http.Request, virtualHost *config.VirtualHost) {
	// Check for location-based configuration first
	location, hasLocation := virtualHost.GetLocationForPath(req.URL.Path)
	
	var effectiveDirectoryIndex string
	
	if hasLocation {
		// Use location's directoryIndex if set, otherwise virtual host, then global
		if location.DirectoryIndex != "" {
			effectiveDirectoryIndex = location.DirectoryIndex
		} else {
			effectiveDirectoryIndex = r.config.GetDirectoryIndex(virtualHost)
		}
		utils.WebServerLog("Using location: %s (handler: %s)", location.Path, location.Handler)
	} else {
		effectiveDirectoryIndex = r.config.GetDirectoryIndex(virtualHost)
	}
	
	// If location specifies a handler, route to that handler type
	if hasLocation && location.Handler != "" {
		r.handleLocationRequest(w, req, virtualHost, location, effectiveDirectoryIndex)
		return
	}
	
	// Otherwise, use default handler selection logic
	// Try each handler in order until one can handle the request
	for _, handler := range r.handlers {
		if handler.CanHandle(req, virtualHost) {
			utils.WebServerLog("Using handler: %T", handler)
			// Pass both virtualHost and effective directoryIndex to handler
			if err := handler.Handle(w, req, virtualHost, effectiveDirectoryIndex); err != nil {
				utils.ErrorLog("Handler error: %v", err)
				// Don't write error response here - handler should have already handled it
				// Only write if response hasn't been written yet
			}
			return
		}
	}

	// Fallback: should never reach here as StaticFileHandler always returns true
	utils.WebServerLog("No handler found, using default file server")
	http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, req)
}

// handleLocationRequest handles requests for location blocks
func (r *Router) handleLocationRequest(w http.ResponseWriter, req *http.Request, virtualHost *config.VirtualHost, location *config.Location, effectiveDirectoryIndex string) {
	switch location.Handler {
	case "proxy":
		handler := NewProxyHandler()
		// Create a temporary virtual host with location's proxy config
		tempVHost := *virtualHost
		tempVHost.ProxyUnixSocket = location.ProxyUnixSocket
		tempVHost.ProxyType = location.ProxyType
		tempVHost.ProxyPath = location.Path
		if err := handler.Handle(w, req, &tempVHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location proxy handler error: %v", err)
		}
	case "cgi":
		handler := NewCGIHandler()
		// Create a temporary virtual host with location's CGI config
		tempVHost := *virtualHost
		tempVHost.CGIPath = location.CGIPath
		if location.CGIPath == "" {
			tempVHost.CGIPath = location.Path
		}
		if err := handler.Handle(w, req, &tempVHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location CGI handler error: %v", err)
		}
	case "php":
		handler := NewPHPHandler()
		// Create a temporary virtual host with location's PHP config
		tempVHost := *virtualHost
		tempVHost.PHPProxyFCGI = location.PHPProxyFCGI
		if err := handler.Handle(w, req, &tempVHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location PHP handler error: %v", err)
		}
	case "static":
		handler := NewStaticFileHandler()
		if err := handler.Handle(w, req, virtualHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location static handler error: %v", err)
		}
	default:
		utils.ErrorLog("Unknown location handler type: %s", location.Handler)
		http.Error(w, "Configuration error", http.StatusInternalServerError)
	}
}
