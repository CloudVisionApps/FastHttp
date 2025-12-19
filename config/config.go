package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

// LogEntry represents a log file entry with optional format
type LogEntry struct {
	Path   string `json:"path"`   // Log file path
	Format string `json:"format"` // Log format name (e.g., "combined", "common") - optional
}

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
	ErrorLog        []LogEntry `json:"errorLog"`       // Array of error log entries (can have multiple)
	CustomLog       []LogEntry `json:"customLog"`      // Array of custom log entries with formats (can have multiple)
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
	Include               interface{}   `json:"include"`               // Single file path (string) or array of file paths ([]string)
	Includes              interface{}   `json:"includes"`              // Alternative field name for multiple includes (array of file paths)
	User                  string        `json:"user"`
	Group                 string        `json:"group"`
	ServerAdmin           string        `json:"serverAdmin"`
	Listen                []string      `json:"listen"`                // Global ports to listen on (applies to all virtual hosts)
	VirtualHosts          []VirtualHost `json:"virtualHosts"`
	MimeTypes             []MimeType    `json:"mimeTypes"`
	DirectoryIndex        string        `json:"directoryIndex"`        // Global default directory index
	RateLimitRequests     int           `json:"rateLimitRequests"`
	RateLimitWindowSeconds int          `json:"rateLimitWindowSeconds"`
	AdminPort             string        `json:"adminPort"`             // Port for admin API (default: "8080")
	AdminEnabled           bool         `json:"adminEnabled"`          // Enable admin API (default: false)
	AdminAuthEnabled       bool         `json:"adminAuthEnabled"`        // Enable admin authentication (default: true)
	AdminUsername          string        `json:"adminUsername"`          // Admin username (default: "admin")
	AdminPassword          string        `json:"adminPassword"`          // Admin password (MUST be changed!)
	AdminIPWhitelist       []string      `json:"adminIPWhitelist"`       // IP whitelist for admin access (empty = allow all)
	LogFile                string        `json:"logFile"`                // Log file path for web server (empty = stdout)
	AdminLogFile           string        `json:"adminLogFile"`           // Log file path for admin API (empty = stdout)
	ErrorLogFile           string        `json:"errorLogFile"`           // Error log file path (empty = stderr)
}

// Load loads configuration from a file, handling includes recursively
func Load(configFilePath string) (*Config, error) {
	return loadWithDepth(configFilePath, 0, make(map[string]bool))
}

// loadWithDepth loads config with include depth tracking to prevent circular includes
func loadWithDepth(configFilePath string, depth int, loaded map[string]bool) (*Config, error) {
	// Prevent infinite recursion (max depth: 10)
	if depth > 10 {
		return nil, fmt.Errorf("maximum include depth exceeded (circular include?)")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving config path: %w", err)
	}

	// Check for circular includes
	if loaded[absPath] {
		return nil, fmt.Errorf("circular include detected: %s", absPath)
	}
	loaded[absPath] = true

	// Open and parse config file
	configFile, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("error opening FastHTTP JSON file: %w", err)
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error parsing FastHTTP JSON configuration: %w", err)
	}

	// Get base directory for resolving relative include paths
	baseDir := filepath.Dir(absPath)

	// Process includes - support both "include" and "includes" fields
	var includeFiles []string

	// Handle "include" field (single or multiple)
	if config.Include != nil {
		files := parseIncludeField(config.Include)
		includeFiles = append(includeFiles, files...)
	}

	// Handle "includes" field (always multiple)
	if config.Includes != nil {
		files := parseIncludeField(config.Includes)
		includeFiles = append(includeFiles, files...)
	}

	// Load and merge each included file
	for _, includeFile := range includeFiles {
		// Resolve relative paths relative to the current config file
		var includePath string
		if filepath.IsAbs(includeFile) {
			includePath = includeFile
		} else {
			includePath = filepath.Join(baseDir, includeFile)
		}

		// Load included config
		includedConfig, err := loadWithDepth(includePath, depth+1, loaded)
		if err != nil {
			return nil, fmt.Errorf("error loading included config %s: %w", includeFile, err)
		}

		// Merge included config into current config
		mergeConfig(&config, includedConfig)
	}

	// Compile regex patterns for all locations
	for i := range config.VirtualHosts {
		if err := config.VirtualHosts[i].CompileLocationRegexes(); err != nil {
			return nil, fmt.Errorf("error compiling location regexes: %w", err)
		}
	}

	return &config, nil
}

// parseIncludeField parses the include field which can be a string or array
func parseIncludeField(include interface{}) []string {
	var includeFiles []string

	// Handle both string and []string formats
	switch v := include.(type) {
	case string:
		includeFiles = []string{v}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				includeFiles = append(includeFiles, str)
			}
		}
	case []string:
		includeFiles = v
	}

	return includeFiles
}

// mergeConfig merges an included config into the base config
// Arrays are appended, other fields override if set in included config
func mergeConfig(base *Config, included *Config) {
	baseValue := reflect.ValueOf(base).Elem()
	includedValue := reflect.ValueOf(included).Elem()
	baseType := baseValue.Type()

	for i := 0; i < baseType.NumField(); i++ {
		field := baseType.Field(i)
		fieldName := field.Name

		// Skip include fields to prevent recursive includes
		if fieldName == "Include" || fieldName == "Includes" {
			continue
		}

		baseField := baseValue.Field(i)
		includedField := includedValue.Field(i)

		// Skip if included field is zero/empty
		if !includedField.IsValid() || includedField.IsZero() {
			continue
		}

		// Handle arrays (append, avoid duplicates)
		if baseField.Kind() == reflect.Slice {
			mergeSlice(baseField, includedField, fieldName)
			continue
		}

		// Handle scalar fields (override if set)
		if baseField.CanSet() {
			// For strings: override if not empty
			if baseField.Kind() == reflect.String {
				if includedField.String() != "" {
					baseField.SetString(includedField.String())
				}
			} else if baseField.Kind() == reflect.Int || baseField.Kind() == reflect.Int64 {
				// For integers: override if greater than 0 (or any non-zero for bool-like)
				if includedField.Int() > 0 {
					baseField.SetInt(includedField.Int())
				}
			} else if baseField.Kind() == reflect.Bool {
				// For booleans: override if set to true (to handle false defaults)
				// This means if included has true, use it; if false, keep base value
				if includedField.Bool() {
					baseField.SetBool(true)
				}
			}
		}
	}
}

// mergeSlice merges slice fields, appending and avoiding duplicates
func mergeSlice(baseField, includedField reflect.Value, fieldName string) {
	if !baseField.IsValid() || !includedField.IsValid() {
		return
	}

	baseLen := baseField.Len()
	includedLen := includedField.Len()

	if includedLen == 0 {
		return
	}

	// Special handling for MimeTypes (deduplicate by Ext)
	if fieldName == "MimeTypes" {
		mimeMap := make(map[string]string)
		for i := 0; i < baseLen; i++ {
			mt := baseField.Index(i)
			ext := mt.FieldByName("Ext").String()
			typ := mt.FieldByName("Type").String()
			mimeMap[ext] = typ
		}

		for i := 0; i < includedLen; i++ {
			mt := includedField.Index(i)
			ext := mt.FieldByName("Ext").String()
			if _, exists := mimeMap[ext]; !exists {
				baseField.Set(reflect.Append(baseField, mt))
			}
		}
		return
	}

	// For other slices (VirtualHosts, Listen, AdminIPWhitelist), use map for deduplication
	if baseField.Type().Elem().Kind() == reflect.String {
		// String slices: use map to track duplicates
		valueMap := make(map[string]bool)
		for i := 0; i < baseLen; i++ {
			valueMap[baseField.Index(i).String()] = true
		}

		for i := 0; i < includedLen; i++ {
			val := includedField.Index(i).String()
			if !valueMap[val] {
				baseField.Set(reflect.Append(baseField, includedField.Index(i)))
				valueMap[val] = true
			}
		}
	} else {
		// For struct slices (VirtualHosts), just append (no deduplication by default)
		for i := 0; i < includedLen; i++ {
			baseField.Set(reflect.Append(baseField, includedField.Index(i)))
		}
	}
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
