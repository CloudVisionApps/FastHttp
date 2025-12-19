package handlers

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"fasthttp/config"
)

// ProxyHandler handles proxying requests to Unix socket backends
type ProxyHandler struct{}

// NewProxyHandler creates a new ProxyHandler
func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{}
}

// CanHandle returns true if this is a proxy request
func (h *ProxyHandler) CanHandle(r *http.Request, virtualHost *config.VirtualHost) bool {
	// Check if proxy configuration exists
	if virtualHost.ProxyUnixSocket != "" {
		return true
	}

	// Check if URL path matches proxy path prefix
	if virtualHost.ProxyPath != "" && strings.HasPrefix(r.URL.Path, virtualHost.ProxyPath) {
		return true
	}

	return false
}

// Handle proxies requests to Unix socket backend
func (h *ProxyHandler) Handle(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost) error {
	unixSocket := virtualHost.ProxyUnixSocket
	if unixSocket == "" {
		http.Error(w, "Proxy not configured", http.StatusBadGateway)
		return nil
	}

	log.Printf("Proxying request to Unix socket: %s", unixSocket)

	// Create a custom transport that uses Unix socket
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", unixSocket)
		},
	}

	// Create a URL for the proxy (scheme and host don't matter for Unix socket)
	targetURL, err := url.Parse("http://localhost")
	if err != nil {
		http.Error(w, "Proxy configuration error", http.StatusInternalServerError)
		return err
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = transport

	// Modify the request path if ProxyPath is configured
	if virtualHost.ProxyPath != "" && strings.HasPrefix(r.URL.Path, virtualHost.ProxyPath) {
		// Strip the proxy path prefix
		r.URL.Path = strings.TrimPrefix(r.URL.Path, virtualHost.ProxyPath)
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
	}

	// Set X-Forwarded-* headers
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Forwarded-Proto", "http")
	if r.TLS != nil {
		r.Header.Set("X-Forwarded-Proto", "https")
	}

	// Serve the proxied request
	proxy.ServeHTTP(w, r)
	return nil
}
