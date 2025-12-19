package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fasthttp/config"
)

// ApacheHttpdParser parses Apache httpd.conf configuration files
type ApacheHttpdParser struct {
	baseDir string
}

// NewApacheHttpdParser creates a new Apache httpd.conf parser
func NewApacheHttpdParser() *ApacheHttpdParser {
	return &ApacheHttpdParser{}
}

// CanParse checks if this parser can handle the given file
func (p *ApacheHttpdParser) CanParse(filePath string) bool {
	ext := filepath.Ext(filePath)
	name := filepath.Base(filePath)
	return ext == ".conf" || strings.Contains(name, "httpd") || strings.Contains(name, "apache")
}

// Parse reads and parses an Apache httpd.conf file
func (p *ApacheHttpdParser) Parse(filePath string) (*ParsedConfig, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}
	p.baseDir = filepath.Dir(absPath)

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	parsed := &ParsedConfig{
		VirtualHosts: []config.VirtualHost{},
		GlobalConfig: make(map[string]interface{}),
		MimeTypes:    []config.MimeType{},
		Includes:     []string{},
	}

	scanner := bufio.NewScanner(file)
	var currentVHost *config.VirtualHost
	var currentLocation *config.Location
	var inVHost bool
	var inLocation bool
	var inDirectory bool
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Parse directives
		directive, args := p.parseDirective(line)
		if directive == "" {
			continue
		}

		// Handle includes
		if directive == "Include" || directive == "IncludeOptional" {
			for _, arg := range args {
				includePath := p.resolvePath(arg)
				parsed.Includes = append(parsed.Includes, includePath)
			}
			continue
		}

		// Handle global directives
		if !inVHost {
			p.parseGlobalDirective(parsed, directive, args)
		}

		// Handle VirtualHost blocks
		if directive == "<VirtualHost" {
			if len(args) > 0 {
				// Extract port from VirtualHost directive (e.g., "*:80" or "192.168.1.1:443")
				port := p.extractPort(args[0])
				currentVHost = &config.VirtualHost{
					Listen: []string{},
					Locations: []config.Location{},
				}
				if port != "" {
					currentVHost.Listen = []string{port}
				}
				inVHost = true
			}
		} else if directive == "</VirtualHost>" {
			if currentVHost != nil && currentVHost.ServerName != "" {
				parsed.VirtualHosts = append(parsed.VirtualHosts, *currentVHost)
			}
			currentVHost = nil
			inVHost = false
			inLocation = false
			inDirectory = false
			currentLocation = nil
		} else if directive == "<Directory" || directive == "<DirectoryMatch" {
			if currentVHost != nil && len(args) > 0 {
				currentLocation = &config.Location{
					Path:      args[0],
					MatchType: "prefix",
					Handler:   "static",
				}
				if directive == "<DirectoryMatch" {
					currentLocation.MatchType = "regex"
				}
				inDirectory = true
			}
		} else if directive == "</Directory>" || directive == "</DirectoryMatch>" {
			if currentLocation != nil && currentVHost != nil {
				currentVHost.Locations = append(currentVHost.Locations, *currentLocation)
			}
			currentLocation = nil
			inDirectory = false
		} else if directive == "<Location" || directive == "<LocationMatch" {
			if currentVHost != nil && len(args) > 0 {
				currentLocation = &config.Location{
					Path:      args[0],
					MatchType: "prefix",
					Handler:   "static",
				}
				if directive == "<LocationMatch" {
					currentLocation.MatchType = "regex"
				}
				inLocation = true
			}
		} else if directive == "</Location>" || directive == "</LocationMatch>" {
			if currentLocation != nil && currentVHost != nil {
				currentVHost.Locations = append(currentVHost.Locations, *currentLocation)
			}
			currentLocation = nil
			inLocation = false
		} else if inVHost {
			// Parse VirtualHost directives
			p.parseVirtualHostDirective(currentVHost, directive, args)
			
			// Parse Location/Directory directives
			if inLocation || inDirectory {
				if currentLocation != nil {
					p.parseLocationDirective(currentLocation, directive, args)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Handle last virtual host if file ends without closing tag
	if currentVHost != nil && currentVHost.ServerName != "" {
		parsed.VirtualHosts = append(parsed.VirtualHosts, *currentVHost)
	}

	return parsed, nil
}

// parseDirective extracts directive name and arguments from a line
func (p *ApacheHttpdParser) parseDirective(line string) (string, []string) {
	// Handle block directives like <VirtualHost *:80>
	if strings.HasPrefix(line, "<") {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			directive := strings.Trim(parts[0], "<>")
			args := []string{}
			if len(parts) > 1 {
				// Join remaining parts and handle quoted strings
				remaining := strings.Join(parts[1:], " ")
				args = p.parseArguments(remaining)
			}
			return directive, args
		}
		return "", nil
	}

	// Regular directive
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", nil
	}

	directive := parts[0]
	args := []string{}
	if len(parts) > 1 {
		// Join remaining parts and handle quoted strings
		remaining := strings.Join(parts[1:], " ")
		args = p.parseArguments(remaining)
	}

	return directive, args
}

// parseArguments parses arguments, handling quoted strings
func (p *ApacheHttpdParser) parseArguments(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(line); i++ {
		char := line[i]

		if (char == '"' || char == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = char
			continue
		}

		if char == quoteChar && inQuotes {
			inQuotes = false
			quoteChar = 0
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		if !inQuotes && (char == ' ' || char == '\t') {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(char)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// parseGlobalDirective parses global Apache directives
func (p *ApacheHttpdParser) parseGlobalDirective(parsed *ParsedConfig, directive string, args []string) {
	switch strings.ToLower(directive) {
	case "user":
		if len(args) > 0 {
			parsed.GlobalConfig["user"] = args[0]
		}
	case "group":
		if len(args) > 0 {
			parsed.GlobalConfig["group"] = args[0]
		}
	case "serveradmin":
		if len(args) > 0 {
			parsed.GlobalConfig["serverAdmin"] = args[0]
		}
	case "listen":
		for _, arg := range args {
			port := p.extractPort(arg)
			if port != "" {
				if ports, ok := parsed.GlobalConfig["listen"].([]string); ok {
					parsed.GlobalConfig["listen"] = append(ports, port)
				} else {
					parsed.GlobalConfig["listen"] = []string{port}
				}
			}
		}
	case "directoryindex":
		if len(args) > 0 {
			parsed.GlobalConfig["directoryIndex"] = strings.Join(args, " ")
		}
	case "typesconfig", "mimetypesfile":
		// MIME types file - could be parsed separately
		if len(args) > 0 {
			parsed.GlobalConfig["mimeTypesFile"] = p.resolvePath(args[0])
		}
	case "addtype":
		if len(args) >= 2 {
			mimeType := args[0]
			for _, ext := range args[1:] {
				ext = strings.TrimPrefix(ext, ".")
				parsed.MimeTypes = append(parsed.MimeTypes, config.MimeType{
					Ext:  ext,
					Type: mimeType,
				})
			}
		}
	}
}

// parseVirtualHostDirective parses directives inside VirtualHost blocks
func (p *ApacheHttpdParser) parseVirtualHostDirective(vhost *config.VirtualHost, directive string, args []string) {
	if vhost == nil {
		return
	}

	switch strings.ToLower(directive) {
	case "servername":
		if len(args) > 0 {
			vhost.ServerName = args[0]
		}
	case "serveralias":
		vhost.ServerAlias = append(vhost.ServerAlias, args...)
	case "documentroot":
		if len(args) > 0 {
			vhost.DocumentRoot = p.resolvePath(args[0])
		}
	case "serveradmin":
		if len(args) > 0 {
			vhost.ServerAdmin = args[0]
		}
	case "errorlog":
		if len(args) > 0 {
			vhost.ErrorLog = p.resolvePath(args[0])
		}
	case "customlog":
		if len(args) > 0 {
			vhost.CustomLog = p.resolvePath(args[0])
		}
	case "directoryindex":
		if len(args) > 0 {
			vhost.DirectoryIndex = strings.Join(args, " ")
		}
	case "phpadminvalue", "phpvalue":
		// PHP configuration - could be used to detect PHP usage
		if len(args) >= 2 && args[0] == "open_basedir" {
			// PHP is likely being used
		}
	}
}

// parseLocationDirective parses directives inside Location/Directory blocks
func (p *ApacheHttpdParser) parseLocationDirective(location *config.Location, directive string, args []string) {
	if location == nil {
		return
	}

	switch strings.ToLower(directive) {
	case "proxy", "proxypass":
		if len(args) >= 2 {
			// ProxyPass /api http://backend:8080
			location.Handler = "proxy"
			location.ProxyUnixSocket = args[1]
			// Try to detect if it's a Unix socket or HTTP
			if strings.HasPrefix(args[1], "unix:") {
				location.ProxyUnixSocket = strings.TrimPrefix(args[1], "unix:")
				location.ProxyType = "http"
			} else if strings.HasPrefix(args[1], "fcgi://") {
				location.ProxyType = "fcgi"
				location.ProxyUnixSocket = strings.TrimPrefix(args[1], "fcgi://")
			} else if strings.HasPrefix(args[1], "http://") || strings.HasPrefix(args[1], "https://") {
				location.ProxyType = "http"
				// For HTTP proxies, we'd need to handle differently
			}
		}
	case "proxypassmatch":
		if len(args) >= 2 {
			location.Handler = "proxy"
			location.MatchType = "regex"
			location.ProxyUnixSocket = args[1]
		}
	case "scriptalias", "scriptaliasmatch":
		if len(args) >= 2 {
			location.Handler = "cgi"
			location.CGIPath = args[0]
			if directive == "scriptaliasmatch" {
				location.MatchType = "regex"
			}
		}
	case "directoryindex":
		if len(args) > 0 {
			location.DirectoryIndex = strings.Join(args, " ")
		}
	case "sethandler":
		if len(args) > 0 {
			handler := strings.ToLower(args[0])
			switch {
			case handler == "proxy:fcgi" || handler == "fcgid-script":
				location.Handler = "proxy"
				location.ProxyType = "fcgi"
			case handler == "proxy" || handler == "proxy-server":
				location.Handler = "proxy"
				location.ProxyType = "http"
			case handler == "cgi-script":
				location.Handler = "cgi"
			case strings.Contains(handler, "php"):
				location.Handler = "php"
			}
		}
	case "phpadminvalue", "phpflag":
		// PHP is being used
		location.Handler = "php"
	}
}

// extractPort extracts port number from Apache Listen/VirtualHost directive
func (p *ApacheHttpdParser) extractPort(arg string) string {
	// Handle formats like: "80", "*:80", "192.168.1.1:443", "[::1]:8080"
	// Extract port number
	re := regexp.MustCompile(`:(\d+)$`)
	matches := re.FindStringSubmatch(arg)
	if len(matches) > 1 {
		return matches[1]
	}
	// If no colon, assume it's just a port number
	if matched, _ := regexp.MatchString(`^\d+$`, arg); matched {
		return arg
	}
	return ""
}

// resolvePath resolves relative paths relative to config file directory
func (p *ApacheHttpdParser) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(p.baseDir, path)
}
