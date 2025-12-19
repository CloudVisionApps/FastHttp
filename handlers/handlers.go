package handlers

import (
	"html"
	"log"
	"net/http"

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
		router: NewRouter(),
	}
}

// ServeHTTP is the main entry point for HTTP requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	virtualHost := h.config.GetVirtualHostByServerName(r.Host)
	if virtualHost != nil {
		log.Printf("Request from %s", r.RemoteAddr)
		log.Printf("Host: %s", html.EscapeString(r.Host))
		log.Printf("Method: %s", html.EscapeString(r.Method))
		log.Printf("URI: %s", r.RequestURI)

		h.router.HandleRequest(w, r, virtualHost)
	} else {
		// Default fallback for unknown virtual hosts
		http.FileServer(http.Dir("/var/www/html")).ServeHTTP(w, r)
	}
}
