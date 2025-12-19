package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Location represents a location/directory block within a virtual host
type Location struct {
	Path            string `json:"path"`            // Path prefix to match (e.g., "/api", "/cgi-bin") OR regex pattern if matchType is "regex"
	MatchType       string `json:"matchType"`       // Match type: "prefix" (default), "regex", "regexCaseInsensitive"
	Handler         string `json:"handler"`         // Handler type: "proxy", "cgi", "php", "static"
	ProxyUnixSocket string `json:"proxyUnixSocket"` // Unix socket for proxy handler
	ProxyType       string `json:"proxyType"`       // Proxy type: "http" or "fcgi"
	CGIPath         string `json:"cgiPath"`         // CGI path (usually same as path)
	PHPProxyFCGI    string `json:"phpProxyFCGI"`   // PHP FastCGI address (TCP)
	DirectoryIndex  string `json:"directoryIndex"`  // Directory index for this location
	
	// Internal: compiled regex (not in JSON)
	regex *regexp.Regexp
}

type VirtualHost struct {
	Listen          []string   `json:"listen"`          // Ports this virtual host listens on (empty = all ports)
	ServerName      string     `json:"serverName"`
	ServerAlias     []string   `json:"serverAlias"`
	DocumentRoot    string     `json:"documentRoot"`
	User            string     `json:"user"`
	Group           string     `json:"group"`
	ServerAdmin     string     `json:"serverAdmin"`
	ErrorLog        string     `json:"errorLog"`
	CustomLog       string     `json:"customLog"`
	DirectoryIndex  string     `json:"directoryIndex"`
	PHPProxyFCGI    string     `json:"phpProxyFCGI"`
	CGIPath         string     `json:"cgiPath"`         // Path prefix for CGI scripts (e.g., "/cgi-bin")
	ProxyUnixSocket string     `json:"proxyUnixSocket"` // Unix socket path for proxy (e.g., "/var/run/app.sock")
	ProxyPath       string     `json:"proxyPath"`       // URL path prefix to proxy (e.g., "/api")
	ProxyType       string     `json:"proxyType"`       // Proxy type: "http" or "fcgi" (default: "http")
	Locations       []Location `json:"locations"`       // Location/directory blocks (like nginx/httpd)
}

type MimeType struct {
	Ext  string `json:"ext"`
	Type string `json:"type"`
}

type Config struct {
	User                  string        `json:"user"`
	Group                 string        `json:"group"`
	ServerAdmin           string        `json:"serverAdmin"`
	Listen                []string      `json:"listen"`                // Global ports to listen on (applies to all virtual hosts)
	VirtualHosts          []VirtualHost `json:"virtualHosts"`
	MimeTypes             []MimeType    `json:"mimeTypes"`
	DirectoryIndex        string        `json:"directoryIndex"`        // Global default directory index
	RateLimitRequests     int           `json:"rateLimitRequests"`
	RateLimitWindowSeconds int          `json:"rateLimitWindowSeconds"`
}

func Load(configFilePath string) (*Config, error) {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening FastHTTP JSON file: %w", err)
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error parsing FastHTTP JSON configuration: %w", err)
	}

	// Compile regex patterns for all locations
	for i := range config.VirtualHosts {
		if err := config.VirtualHosts[i].CompileLocationRegexes(); err != nil {
			return nil, fmt.Errorf("error compiling location regexes: %w", err)
		}
	}

	return &config, nil
}

// GetVirtualHostByServerName finds a virtual host by server name
// If port is provided, also matches against the virtual host's Listen ports
func (c *Config) GetVirtualHostByServerName(serverName string) *VirtualHost {
	return c.GetVirtualHostByServerNameAndPort(serverName, "")
}

// GetVirtualHostByServerNameAndPort finds a virtual host by server name and port
func (c *Config) GetVirtualHostByServerNameAndPort(serverName, port string) *VirtualHost {
	for i, v := range c.VirtualHosts {
		// Check server name
		if v.ServerName == serverName {
			// If port is specified, check if virtual host listens on that port
			if port != "" && len(v.Listen) > 0 {
				for _, listenPort := range v.Listen {
					if listenPort == port {
						return &c.VirtualHosts[i]
					}
				}
				// Virtual host has Listen ports but this port doesn't match, skip
				continue
			}
			// No port specified or virtual host has no Listen restriction
			return &c.VirtualHosts[i]
		}
		// Check server aliases
		for _, alias := range v.ServerAlias {
			if alias == serverName {
				if port != "" && len(v.Listen) > 0 {
					for _, listenPort := range v.Listen {
						if listenPort == port {
							return &c.VirtualHosts[i]
						}
					}
					continue
				}
				return &c.VirtualHosts[i]
			}
		}
	}
	return nil
}

// GetAllListenPorts returns all unique ports that should be listened on
// Combines global Listen and virtual host Listen ports
func (c *Config) GetAllListenPorts() []string {
	portMap := make(map[string]bool)
	var ports []string

	// Add global Listen ports
	for _, port := range c.Listen {
		if port != "" && !portMap[port] {
			portMap[port] = true
			ports = append(ports, port)
		}
	}

	// Add virtual host Listen ports
	for _, vhost := range c.VirtualHosts {
		for _, port := range vhost.Listen {
			if port != "" && !portMap[port] {
				portMap[port] = true
				ports = append(ports, port)
			}
		}
	}

	return ports
}

func (c *Config) GetRateLimitConfig() (maxRequests, windowSeconds int) {
	maxRequests = c.RateLimitRequests
	if maxRequests <= 0 {
		maxRequests = 100 // Default: 100 requests per window
	}
	windowSeconds = c.RateLimitWindowSeconds
	if windowSeconds <= 0 {
		windowSeconds = 60 // Default: 60 seconds window
	}
	return maxRequests, windowSeconds
}

// GetDirectoryIndex returns the directory index for a virtual host
// Uses virtual host setting if set, otherwise falls back to global default
func (c *Config) GetDirectoryIndex(virtualHost *VirtualHost) string {
	if virtualHost != nil && virtualHost.DirectoryIndex != "" {
		return virtualHost.DirectoryIndex
	}
	return c.DirectoryIndex
}

// CompileLocationRegexes compiles regex patterns for all locations
// Should be called after loading configuration
func (v *VirtualHost) CompileLocationRegexes() error {
	for i := range v.Locations {
		loc := &v.Locations[i]
		matchType := strings.ToLower(loc.MatchType)
		if matchType == "" {
			matchType = "prefix" // Default to prefix matching
		}

		if matchType == "regex" || matchType == "regexcaseinsensitive" {
			var err error
			if matchType == "regexcaseinsensitive" {
				loc.regex, err = regexp.Compile("(?i)" + loc.Path)
			} else {
				loc.regex, err = regexp.Compile(loc.Path)
			}
			if err != nil {
				return fmt.Errorf("invalid regex pattern in location %s: %w", loc.Path, err)
			}
		}
	}
	return nil
}

// GetLocationForPath finds the matching location block for a given path
// Returns the location and true if found, nil and false otherwise
// Priority: regex matches first (in order), then longest path prefix
func (v *VirtualHost) GetLocationForPath(path string) (*Location, bool) {
	// First, check regex matches (they have higher priority)
	for i := range v.Locations {
		loc := &v.Locations[i]
		matchType := strings.ToLower(loc.MatchType)
		if matchType == "" {
			matchType = "prefix"
		}

		if (matchType == "regex" || matchType == "regexcaseinsensitive") && loc.regex != nil {
			if loc.regex.MatchString(path) {
				return loc, true
			}
		}
	}

	// Then, check prefix matches (longest prefix wins)
	var bestMatch *Location
	longestMatch := 0

	for i := range v.Locations {
		loc := &v.Locations[i]
		matchType := strings.ToLower(loc.MatchType)
		if matchType == "" {
			matchType = "prefix"
		}

		// Only check prefix matches (skip regex locations)
		if matchType == "prefix" && strings.HasPrefix(path, loc.Path) {
			if len(loc.Path) > longestMatch {
				longestMatch = len(loc.Path)
				bestMatch = loc
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, true
	}
	return nil, false
}
