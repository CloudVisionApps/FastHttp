package handlers

import (
	"log"
	"net/http"

	"fasthttp/config"
)

// Router selects and executes the appropriate handler for a request
type Router struct {
	handlers []RequestHandler
}

// NewRouter creates a new Router with all available handlers
func NewRouter() *Router {
	return &Router{
		handlers: []RequestHandler{
			NewPHPHandler(),
			NewCGIHandler(),
			NewProxyHandler(),
			NewStaticFileHandler(), // Always last as fallback
		},
	}
}

// HandleRequest routes the request to the appropriate handler
func (r *Router) HandleRequest(w http.ResponseWriter, req *http.Request, virtualHost *config.VirtualHost) {
	// Try each handler in order until one can handle the request
	for _, handler := range r.handlers {
		if handler.CanHandle(req, virtualHost) {
			log.Printf("Using handler: %T", handler)
			if err := handler.Handle(w, req, virtualHost); err != nil {
				log.Printf("Handler error: %v", err)
				// Don't write error response here - handler should have already handled it
				// Only write if response hasn't been written yet
			}
			return
		}
	}

	// Fallback: should never reach here as StaticFileHandler always returns true
	log.Printf("No handler found, using default file server")
	http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, req)
}
