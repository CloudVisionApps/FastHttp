package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	// Resolve the request path to filesystem path
	urlPath := req.URL.Path
	if urlPath == "" {
		urlPath = "/"
	}
	fullPath := filepath.Join(virtualHost.DocumentRoot, filepath.Clean(urlPath))
	
	// Check for location-based configuration
	// For Directory blocks, we need to check if the filesystem path is within the Directory path
	var location *config.Location
	var matchRule *config.MatchRule
	hasLocation := false
	
	// Check each location to see if the request path falls within it
	for i := range virtualHost.Locations {
		loc := &virtualHost.Locations[i]
		// Check if this location's path (filesystem path from Directory) contains the request path
		// For Directory blocks, the path is a filesystem path
		// Normalize paths for comparison
		locPath := filepath.Clean(loc.Path)
		reqPath := filepath.Clean(fullPath)
		
		// Check if request path is within the Directory path
		if strings.HasPrefix(reqPath, locPath) || reqPath == locPath {
			location = loc
			hasLocation = true
			
			// Check match rules against the URL path or resolved index file
			if len(loc.MatchRules) > 0 {
				// First check against the URL path filename
				urlFilename := filepath.Base(urlPath)
				if urlFilename == "/" || urlFilename == "" {
					urlFilename = ""
				}
				
				// If it's a directory, check against index file
				if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
					effectiveDirectoryIndex := r.config.GetDirectoryIndex(virtualHost)
					if loc.DirectoryIndex != "" {
						effectiveDirectoryIndex = loc.DirectoryIndex
					}
					indexFile := utils.FindIndexFile(fullPath, effectiveDirectoryIndex)
					if indexFile != "" {
						urlFilename = indexFile
					}
				}
				
				// Check match rules
				if urlFilename != "" {
					for j := range loc.MatchRules {
						rule := &loc.MatchRules[j]
						if rule.Matches(urlFilename) {
							matchRule = rule
							break
						}
					}
				}
			}
			break
		}
	}
	
	var effectiveDirectoryIndex string
	var handler string
	var proxyUnixSocket string
	var proxyType string
	
	if hasLocation {
		// Use location's directoryIndex if set, otherwise virtual host, then global
		if location.DirectoryIndex != "" {
			effectiveDirectoryIndex = location.DirectoryIndex
		} else {
			effectiveDirectoryIndex = r.config.GetDirectoryIndex(virtualHost)
		}
		
		// If a match rule was found, use its handler/proxy config
		if matchRule != nil {
			handler = matchRule.Handler
			proxyUnixSocket = matchRule.ProxyUnixSocket
			proxyType = matchRule.ProxyType
			utils.WebServerLog("Using location: %s, match rule: %s (handler: %s)", location.Path, matchRule.Path, handler)
		} else if len(location.MatchRules) > 0 {
			// Location has match rules but none matched the URL path
			// Check if this is a directory request and try to resolve index file
			urlPath := req.URL.Path
			if urlPath == "" || urlPath == "/" {
				urlPath = "/"
			}
			fullPath := filepath.Join(virtualHost.DocumentRoot, filepath.Clean(urlPath))
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				// It's a directory, check for index file
				indexFile := utils.FindIndexFile(fullPath, effectiveDirectoryIndex)
				if indexFile != "" {
					// Check match rules against the resolved index file name
					for j := range location.MatchRules {
						rule := &location.MatchRules[j]
						// Match rules always match against filename
						if rule.Matches(indexFile) {
							matchRule = rule
							handler = rule.Handler
							proxyUnixSocket = rule.ProxyUnixSocket
							proxyType = rule.ProxyType
							utils.WebServerLog("Using location: %s, match rule: %s (handler: %s) for index file: %s", location.Path, rule.Path, handler, indexFile)
							break
						}
					}
				}
			}
			
			// If still no match rule, use location's own handler
			if matchRule == nil {
				handler = location.Handler
				proxyUnixSocket = location.ProxyUnixSocket
				proxyType = location.ProxyType
				utils.WebServerLog("Using location: %s (handler: %s)", location.Path, handler)
			}
		} else {
			// Use location's own handler/proxy config
			handler = location.Handler
			proxyUnixSocket = location.ProxyUnixSocket
			proxyType = location.ProxyType
			utils.WebServerLog("Using location: %s (handler: %s)", location.Path, handler)
		}
	} else {
		effectiveDirectoryIndex = r.config.GetDirectoryIndex(virtualHost)
	}
	
	// If location/match rule specifies a handler, route to that handler type
	if hasLocation && handler != "" {
		r.handleLocationRequest(w, req, virtualHost, location, matchRule, handler, proxyUnixSocket, proxyType, effectiveDirectoryIndex)
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
func (r *Router) handleLocationRequest(w http.ResponseWriter, req *http.Request, virtualHost *config.VirtualHost, location *config.Location, matchRule *config.MatchRule, handler string, proxyUnixSocket string, proxyType string, effectiveDirectoryIndex string) {
	switch handler {
	case "proxy":
		proxyHandler := NewProxyHandler()
		// Create a temporary virtual host with proxy config (from match rule or location)
		tempVHost := *virtualHost
		tempVHost.ProxyUnixSocket = proxyUnixSocket
		tempVHost.ProxyType = proxyType
		tempVHost.ProxyPath = location.Path
		if err := proxyHandler.Handle(w, req, &tempVHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location proxy handler error: %v", err)
		}
	case "cgi":
		cgiHandler := NewCGIHandler()
		// Create a temporary virtual host with location's CGI config
		tempVHost := *virtualHost
		if matchRule != nil && matchRule.CGIPath != "" {
			tempVHost.CGIPath = matchRule.CGIPath
		} else if location.CGIPath != "" {
			tempVHost.CGIPath = location.CGIPath
		} else {
			tempVHost.CGIPath = location.Path
		}
		if err := cgiHandler.Handle(w, req, &tempVHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location CGI handler error: %v", err)
		}
	case "php":
		phpHandler := NewPHPHandler()
		// Create a temporary virtual host with PHP config (from match rule or location)
		tempVHost := *virtualHost
		if matchRule != nil && matchRule.PHPProxyFCGI != "" {
			tempVHost.PHPProxyFCGI = matchRule.PHPProxyFCGI
		} else {
			tempVHost.PHPProxyFCGI = location.PHPProxyFCGI
		}
		if err := phpHandler.Handle(w, req, &tempVHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location PHP handler error: %v", err)
		}
	case "static":
		staticHandler := NewStaticFileHandler()
		if err := staticHandler.Handle(w, req, virtualHost, effectiveDirectoryIndex); err != nil {
			utils.ErrorLog("Location static handler error: %v", err)
		}
	default:
		utils.ErrorLog("Unknown location handler type: %s", handler)
		http.Error(w, "Configuration error", http.StatusInternalServerError)
	}
}
