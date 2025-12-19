package handlers

import (
	"html"
	"log"
	"net/http"
	"strings"

	"fasthttp/config"
)

// Handler is the main HTTP handler that routes requests to appropriate handlers
type Handler struct {
	config *config.Config
	router *Router
}

// New creates a new Handler with modular architecture
func New(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
		router: NewRouter(cfg),
	}
}

// ServeHTTP is the main entry point for HTTP requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract port from Host header or RemoteAddr
	port := ""
	host := r.Host
	
	// Check if port is in Host header (e.g., "example.com:8080")
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		port = host[idx+1:]
		host = host[:idx]
	} else {
		// Extract port from RemoteAddr or use defaults
		if r.TLS != nil {
			port = "443"
		} else {
			port = "80"
		}
	}

	// Try to find virtual host by server name and port
	virtualHost := h.config.GetVirtualHostByServerNameAndPort(host, port)
	if virtualHost == nil {
		// Try with full Host header (in case it includes port)
		virtualHost = h.config.GetVirtualHostByServerNameAndPort(r.Host, port)
	}
	
	if virtualHost == nil {
		// Try without port (for backward compatibility)
		virtualHost = h.config.GetVirtualHostByServerName(host)
		if virtualHost == nil {
			virtualHost = h.config.GetVirtualHostByServerName(r.Host)
		}
	}

	if virtualHost != nil {
	    log.Printf("Request from %s", r.RemoteAddr)
        log.Printf("Host: %s", html.EscapeString(r.Host))
		log.Printf("Port: %s", port)
        log.Printf("Method: %s", html.EscapeString(r.Method))
		log.Printf("URI: %s", r.RequestURI)

		h.router.HandleRequest(w, r, virtualHost)
	} else {
		// Default fallback for unknown virtual hosts
		http.FileServer(http.Dir("/var/www/html")).ServeHTTP(w, r)
	}
}
