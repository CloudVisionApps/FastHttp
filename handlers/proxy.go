package handlers

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"fasthttp/config"
	"fasthttp/utils"

	"github.com/yookoala/gofast"
)

// ProxyHandler handles proxying requests to Unix socket backends
type ProxyHandler struct{}

// NewProxyHandler creates a new ProxyHandler
func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{}
}

// CanHandle returns true if this is a proxy request
func (h *ProxyHandler) CanHandle(r *http.Request, virtualHost *config.VirtualHost) bool {
	// If ProxyPath is set, only handle requests matching that path
	if virtualHost.ProxyPath != "" {
		return strings.HasPrefix(r.URL.Path, virtualHost.ProxyPath)
	}

	// If ProxyUnixSocket is set without ProxyPath, only proxy PHP/dynamic requests
	// Static files should be handled by StaticFileHandler
	if virtualHost.ProxyUnixSocket != "" {
		urlPath := r.URL.Path
		
		// Don't handle root or directory paths - let StaticFileHandler show directory listing
		if urlPath == "/" || strings.HasSuffix(urlPath, "/") {
			return false
		}
		
		// Check if it's a static file (has extension that's not PHP)
		isFile, _ := utils.IsFileRequest(urlPath)
		if isFile {
			// If it's a file with extension, check if it's PHP
			if strings.HasSuffix(urlPath, ".php") {
				return true
			}
			// Static files (.txt, .css, .js, .jpg, etc.) should be served as static
			return false
		}
		
		// No extension - could be PHP or dynamic, proxy it
		// But only if proxyType is fcgi (for PHP)
		if virtualHost.ProxyType == "fcgi" {
			return true
		}
		
		// For HTTP proxy, only proxy if it looks like an API/dynamic request
		// Otherwise let StaticFileHandler try to serve it
		return false
	}

	return false
}

// Handle proxies requests to Unix socket backend
func (h *ProxyHandler) Handle(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, effectiveDirectoryIndex string) error {
	unixSocket := virtualHost.ProxyUnixSocket
	if unixSocket == "" {
		http.Error(w, "Proxy not configured", http.StatusBadGateway)
		return nil
	}

	// Determine proxy type (default to http, but check for fcgi)
	proxyType := strings.ToLower(virtualHost.ProxyType)
	if proxyType == "" {
		proxyType = "http"
	}

	// Handle FastCGI proxy
	if proxyType == "fcgi" {
		return h.handleFCGIProxy(w, r, virtualHost, unixSocket, effectiveDirectoryIndex)
	}

	// Handle HTTP proxy (default)
	return h.handleHTTPProxy(w, r, virtualHost, unixSocket)
}

// handleFCGIProxy handles FastCGI proxying over Unix socket
func (h *ProxyHandler) handleFCGIProxy(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, unixSocket string, effectiveDirectoryIndex string) error {
	log.Printf("Proxying FCGI request to Unix socket: %s", unixSocket)

	// Determine the script path
	scriptPath := r.URL.Path
	if virtualHost.ProxyPath != "" && strings.HasPrefix(scriptPath, virtualHost.ProxyPath) {
		// Strip the proxy path prefix
		scriptPath = strings.TrimPrefix(scriptPath, virtualHost.ProxyPath)
		if !strings.HasPrefix(scriptPath, "/") {
			scriptPath = "/" + scriptPath
		}
	}

	// If script path is empty or root, try to find a default file
	if scriptPath == "/" || scriptPath == "" {
		// Use Apache-style index file lookup
		indexFile := utils.FindIndexFile(virtualHost.DocumentRoot, effectiveDirectoryIndex)
		if indexFile != "" {
			scriptPath = "/" + indexFile
		}
	}

	// Get the file name from the path
	fileName := utils.GetFileName(scriptPath)
	if fileName == "/" || fileName == "" {
		fileName = "index.php"
	}

	log.Printf("FCGI proxy script: %s", fileName)

	// Create Unix socket connection factory for FastCGI
	connFactory := gofast.SimpleConnFactory("unix", unixSocket)

	// Create FastCGI handler
	gofastHandler := gofast.NewHandler(
		gofast.NewFileEndpoint(virtualHost.DocumentRoot+"/"+fileName)(gofast.BasicSession),
		gofast.SimpleClientFactory(connFactory),
	)

	// Serve the request
	http.HandlerFunc(gofastHandler.ServeHTTP).ServeHTTP(w, r)
	return nil
}

// handleHTTPProxy handles HTTP proxying over Unix socket
func (h *ProxyHandler) handleHTTPProxy(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, unixSocket string) error {
	log.Printf("Proxying HTTP request to Unix socket: %s", unixSocket)

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
