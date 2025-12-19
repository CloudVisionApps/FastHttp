package handlers

import (
	"net/http"

	"fasthttp/config"
)

// RequestHandler is the interface that all request handlers must implement
type RequestHandler interface {
	// CanHandle returns true if this handler can handle the given request
	CanHandle(r *http.Request, virtualHost *config.VirtualHost) bool

	// Handle processes the HTTP request
	Handle(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost) error
}
