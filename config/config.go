package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Location represents a location/directory block within a virtual host
type Location struct {
	Path           string `json:"path"`           // Path prefix to match (e.g., "/api", "/cgi-bin")
	Handler        string `json:"handler"`        // Handler type: "proxy", "cgi", "php", "static"
	ProxyUnixSocket string `json:"proxyUnixSocket"` // Unix socket for proxy handler
	ProxyType      string `json:"proxyType"`      // Proxy type: "http" or "fcgi"
	CGIPath        string `json:"cgiPath"`        // CGI path (usually same as path)
	PHPProxyFCGI   string `json:"phpProxyFCGI"`   // PHP FastCGI address (TCP)
	DirectoryIndex string `json:"directoryIndex"`  // Directory index for this location
}

type VirtualHost struct {
	PortType        string     `json:"portType"`
	Listen          []string   `json:"listen"`
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
	Listen                []string      `json:"listen"`
	VirtualHosts          []VirtualHost `json:"virtualHosts"`
	HttpPort              string        `json:"httpPort"`
	HttpsPort             string        `json:"httpsPort"`
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

	return &config, nil
}

func (c *Config) GetVirtualHostByServerName(serverName string) *VirtualHost {
	for i, v := range c.VirtualHosts {
		if v.ServerName == serverName {
			return &c.VirtualHosts[i]
		}
	}
	return nil
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

// GetLocationForPath finds the matching location block for a given path
// Returns the location and true if found, nil and false otherwise
// Locations are matched by longest path prefix
func (v *VirtualHost) GetLocationForPath(path string) (*Location, bool) {
	var bestMatch *Location
	longestMatch := 0

	for i := range v.Locations {
		loc := &v.Locations[i]
		if strings.HasPrefix(path, loc.Path) {
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
