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
// If called recursively for includes, it will parse the file without processing includes again
func (p *ApacheHttpdParser) Parse(filePath string) (*ParsedConfig, error) {
	return p.parseFile(filePath, false)
}

// parseFile is the internal parsing method that can skip include processing
func (p *ApacheHttpdParser) parseFile(filePath string, skipIncludes bool) (*ParsedConfig, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}
	originalBaseDir := p.baseDir
	p.baseDir = filepath.Dir(absPath)
	defer func() { p.baseDir = originalBaseDir }()

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
	var inFilesMatch bool
	var inIfModule bool
	var ifModuleDepth int
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		originalLine := strings.TrimSpace(scanner.Text())
		line := originalLine

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

		// Handle IfModule blocks (we'll parse them but note they're conditional)
		if directive == "<IfModule" {
			inIfModule = true
			ifModuleDepth = 1
			continue
		} else if directive == "</IfModule>" {
			ifModuleDepth--
			if ifModuleDepth == 0 {
				inIfModule = false
			}
			continue
		} else if strings.HasPrefix(directive, "<If") {
			// Other conditional blocks
			inIfModule = true
			ifModuleDepth = 1
			continue
		} else if strings.HasPrefix(directive, "</If") {
			ifModuleDepth--
			if ifModuleDepth == 0 {
				inIfModule = false
			}
			continue
		}

		// Skip processing directives inside IfModule blocks (we can't know if module is loaded)
		if inIfModule {
			continue
		}

		// Handle includes (with glob pattern expansion)
		if directive == "Include" || directive == "IncludeOptional" {
			for _, arg := range args {
				includePaths := p.expandIncludePath(arg)
				if len(includePaths) == 0 {
					fmt.Printf("Warning: No files matched include pattern: %s\n", arg)
				} else {
					fmt.Printf("Expanded include pattern '%s' to %d file(s)\n", arg, len(includePaths))
				}
				parsed.Includes = append(parsed.Includes, includePaths...)
			}
			continue
		}

		// Handle global directives (including Directory blocks at global level)
		if !inVHost {
			// Handle global Directory blocks
			if directive == "<Directory" || directive == "<DirectoryMatch" {
				if len(args) > 0 {
					currentLocation = &config.Location{
						Path:      args[0],
						MatchType: "prefix",
						Handler:   "static",
					}
					if directive == "<DirectoryMatch" {
						currentLocation.MatchType = "regex"
					}
					inDirectory = true
					continue
				}
			} else if directive == "</Directory>" || directive == "</DirectoryMatch>" {
				if currentLocation != nil {
					// Store global location
					var locations []config.Location
					if existing, ok := parsed.GlobalConfig["globalLocations"].([]config.Location); ok {
						locations = existing
					} else {
						locations = []config.Location{}
					}
					locations = append(locations, *currentLocation)
					parsed.GlobalConfig["globalLocations"] = locations
				}
				currentLocation = nil
				inDirectory = false
				continue
			} else if inDirectory && currentLocation != nil {
				// Parse directives inside global Directory block
				p.parseLocationDirective(currentLocation, directive, args)
				continue
			}
			
			p.parseGlobalDirective(parsed, directive, args)
		}

		// Handle VirtualHost blocks
		// Note: parseDirective trims <>, so <VirtualHost> becomes "VirtualHost"
		if directive == "VirtualHost" && strings.HasPrefix(originalLine, "<") && !strings.HasPrefix(originalLine, "</") {
			if len(args) > 0 {
				// Extract port from VirtualHost directive (e.g., "*:80" or "192.168.1.1:443")
				port := p.extractPort(args[0])
				currentVHost = &config.VirtualHost{
					Listen:    []string{},
					Locations: []config.Location{},
				}
				if port != "" {
					currentVHost.Listen = []string{port}
				}
				inVHost = true
			}
		} else if directive == "VirtualHost" && strings.HasPrefix(originalLine, "</") {
			if currentVHost != nil {
				// Add virtual host even if ServerName is empty (it might be set later or be a default vhost)
				if currentVHost.ServerName == "" {
					// Try to use DocumentRoot as identifier if no ServerName
					if currentVHost.DocumentRoot != "" {
						currentVHost.ServerName = filepath.Base(currentVHost.DocumentRoot)
					} else {
						currentVHost.ServerName = "_default_"
					}
				}
				fmt.Printf("  [DEBUG] Closing VirtualHost block, ServerName=%s, DocumentRoot=%s\n", currentVHost.ServerName, currentVHost.DocumentRoot)
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
				inFilesMatch = false // Reset FilesMatch state
			}
		} else if directive == "</Directory>" || directive == "</DirectoryMatch>" {
			if currentLocation != nil && currentVHost != nil {
				fmt.Printf("  [DEBUG] Closing Directory block, adding location: path=%s, handler=%s, proxySocket=%s\n", currentLocation.Path, currentLocation.Handler, currentLocation.ProxyUnixSocket)
				currentVHost.Locations = append(currentVHost.Locations, *currentLocation)
			}
			currentLocation = nil
			inDirectory = false
			inFilesMatch = false
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
		} else if directive == "FilesMatch" || directive == "Files" {
			// Check if this is an opening tag (parseDirective strips < >)
			if strings.HasPrefix(originalLine, "<") && !strings.HasPrefix(originalLine, "</") {
				if currentLocation != nil && len(args) > 0 {
					// FilesMatch is inside a Directory block, update the location path to match the pattern
					// The pattern is the regex/file pattern (e.g., "\.php$")
					pattern := args[0]
					// Remove quotes if present
					pattern = strings.Trim(pattern, "\"'")
					currentLocation.Path = pattern
					if directive == "FilesMatch" {
						currentLocation.MatchType = "regexCaseInsensitive"
					} else {
						currentLocation.MatchType = "regex"
					}
					inFilesMatch = true
					fmt.Printf("  [DEBUG] Started FilesMatch block, pattern=%s, path=%s, inDirectory=%v\n", pattern, currentLocation.Path, inDirectory)
				} else {
					fmt.Printf("  [DEBUG] FilesMatch found but currentLocation is nil or no args. inDirectory=%v, currentLocation=%v, args=%v\n", inDirectory, currentLocation != nil, args)
				}
			} else if strings.HasPrefix(originalLine, "</") {
				// Closing tag
				inFilesMatch = false
				fmt.Printf("  [DEBUG] Closed FilesMatch block\n")
			}
		} else if inVHost {
			// Parse VirtualHost directives
			p.parseVirtualHostDirective(currentVHost, directive, args)
			
			// Parse Location/Directory/FilesMatch directives
			if inLocation || inDirectory || inFilesMatch {
				if currentLocation != nil {
					if directive == "sethandler" {
						fmt.Printf("  [DEBUG] Parsing SetHandler in context: inDirectory=%v, inFilesMatch=%v, directive=%s, args=%v\n", inDirectory, inFilesMatch, directive, args)
					}
					p.parseLocationDirective(currentLocation, directive, args)
				} else {
					if directive == "sethandler" {
						fmt.Printf("  [DEBUG] SetHandler found but currentLocation is nil. inDirectory=%v, inFilesMatch=%v\n", inDirectory, inFilesMatch)
					}
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

	// Recursively parse included files (only for top-level parse, not recursive calls)
	if !skipIncludes && len(parsed.Includes) > 0 {
		includedVHosts, err := p.parseIncludes(parsed.Includes, make(map[string]bool), 0)
		if err != nil {
			return nil, fmt.Errorf("error parsing includes: %w", err)
		}
		// Merge virtual hosts from included files
		parsed.VirtualHosts = append(parsed.VirtualHosts, includedVHosts...)
	}

	return parsed, nil
}

// parseIncludes recursively parses included files and returns all virtual hosts
func (p *ApacheHttpdParser) parseIncludes(includePaths []string, visited map[string]bool, depth int) ([]config.VirtualHost, error) {
	// Prevent infinite recursion (max depth: 10)
	if depth > 10 {
		return nil, fmt.Errorf("maximum include depth exceeded (circular include?)")
	}

	var allVHosts []config.VirtualHost

	for _, includePath := range includePaths {
		// Resolve absolute path
		absPath, err := filepath.Abs(includePath)
		if err != nil {
			// For IncludeOptional, skip if file doesn't exist
			continue
		}

		// Check for circular includes
		if visited[absPath] {
			continue // Skip already visited files
		}
		visited[absPath] = true

		// Check if file exists (for IncludeOptional)
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			continue // Skip non-existent files (IncludeOptional behavior)
		}

		// Parse the included file (skip includes to avoid double-processing, we handle them here)
		fmt.Printf("Parsing included file: %s\n", absPath)
		includedParsed, err := p.parseFile(absPath, true)
		if err != nil {
			// For IncludeOptional, continue on error; for Include, return error
			// We'll treat all as optional for now to be safe
			fmt.Printf("Warning: Error parsing included file %s: %v\n", absPath, err)
			continue
		}

		// Add virtual hosts from this file
		if len(includedParsed.VirtualHosts) > 0 {
			fmt.Printf("Found %d virtual host(s) in %s\n", len(includedParsed.VirtualHosts), absPath)
			for i, vhost := range includedParsed.VirtualHosts {
				fmt.Printf("  VirtualHost %d: ServerName=%s, DocumentRoot=%s\n", i+1, vhost.ServerName, vhost.DocumentRoot)
			}
		} else {
			fmt.Printf("No virtual hosts found in %s (file parsed but no VirtualHost blocks detected)\n", absPath)
		}
		allVHosts = append(allVHosts, includedParsed.VirtualHosts...)

		// Recursively parse nested includes from this file
		if len(includedParsed.Includes) > 0 {
			nestedVHosts, err := p.parseIncludes(includedParsed.Includes, visited, depth+1)
			if err != nil {
				return nil, err
			}
			allVHosts = append(allVHosts, nestedVHosts...)
		}
	}

	return allVHosts, nil
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
	case "servername":
		if len(args) > 0 {
			parsed.GlobalConfig["serverName"] = args[0]
		}
	case "serverroot":
		if len(args) > 0 {
			parsed.GlobalConfig["serverRoot"] = p.resolvePath(args[0])
		}
	case "pidfile":
		// Ignore, not relevant for FastHTTP
	case "scriptalias":
		// Global ScriptAlias - note it for later use
		if len(args) >= 2 {
			if parsed.GlobalConfig["scriptAliases"] == nil {
				parsed.GlobalConfig["scriptAliases"] = map[string]string{}
			}
			if aliases, ok := parsed.GlobalConfig["scriptAliases"].(map[string]string); ok {
				aliases[args[0]] = p.resolvePath(args[1])
			}
		}
	case "action":
		// Action directives for PHP handlers
		// Action application/x-httpd-remi-php84 /cgi-sys/remi-php84
		if len(args) >= 2 {
			mimeType := args[0]
			handler := args[1]
			if strings.Contains(mimeType, "php") {
				// Note PHP handler
				if parsed.GlobalConfig["phpHandlers"] == nil {
					parsed.GlobalConfig["phpHandlers"] = map[string]string{}
				}
				if handlers, ok := parsed.GlobalConfig["phpHandlers"].(map[string]string); ok {
					handlers[mimeType] = handler
				}
			}
		}
	case "addhandler":
		// AddHandler application/x-httpd-remi-php84 .php .php8 .phtml
		if len(args) >= 2 {
			handler := args[0]
			if strings.Contains(handler, "php") {
				// This indicates PHP is being used
				parsed.GlobalConfig["phpHandler"] = handler
				// Map extensions to PHP
				for _, ext := range args[1:] {
					ext = strings.TrimPrefix(ext, ".")
					// Note these extensions should use PHP handler
					if parsed.GlobalConfig["phpExtensions"] == nil {
						parsed.GlobalConfig["phpExtensions"] = []string{}
					}
					if exts, ok := parsed.GlobalConfig["phpExtensions"].([]string); ok {
						parsed.GlobalConfig["phpExtensions"] = append(exts, ext)
					}
				}
			}
		}
	case "setenv":
		// SetEnv - environment variables (not directly used in FastHTTP, but noted)
		// Could be stored for reference if needed
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
	case "addhandler":
		// AddHandler application/x-httpd-remi-php .php .php8 .phtml
		// Note: AddHandler in Apache doesn't create location blocks, it just sets the handler for file extensions
		// We'll note this for reference but not automatically create location blocks
		// Location blocks should only come from explicit <Location> or <Directory> blocks
	case "phpadminvalue", "phpvalue":
		// PHP configuration - could be used to detect PHP usage
		if len(args) >= 2 && args[0] == "open_basedir" {
			// PHP is likely being used
		}
	case "suphp_usergroup":
		// suPHP_UserGroup user group
		if len(args) >= 2 {
			vhost.User = args[0]
			vhost.Group = args[1]
		}
	case "suexecusergroup":
		// SuexecUserGroup user group
		if len(args) >= 2 {
			vhost.User = args[0]
			vhost.Group = args[1]
		}
	case "assignuserid":
		// AssignUserID user group (mpm_itk_module)
		if len(args) >= 2 {
			vhost.User = args[0]
			vhost.Group = args[1]
		}
	case "passengeruser":
		// PassengerUser user
		if len(args) > 0 {
			vhost.User = args[0]
		}
	case "passengergroup":
		// PassengerGroup group
		if len(args) > 0 {
			vhost.Group = args[0]
		}
	case "setenv":
		// SetEnv - environment variables (we can note these but not directly use them)
		// Could be stored for reference if needed
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
			handler := args[0]
			// Remove quotes if present
			handler = strings.Trim(handler, "\"'")
			handlerLower := strings.ToLower(handler)
			
			// Handle proxy:unix:/path/to/sock|fcgi://localhost/ format
			if strings.HasPrefix(handlerLower, "proxy:unix:") {
				// Extract Unix socket path
				// Format: proxy:unix:/path/to/sock|fcgi://localhost/
				parts := strings.Split(handler, "|")
				if len(parts) > 0 {
					unixPart := strings.TrimPrefix(parts[0], "proxy:unix:")
					unixPart = strings.TrimPrefix(unixPart, "Proxy:unix:")
					location.Handler = "proxy"
					location.ProxyUnixSocket = unixPart
					
					// Check if it's FCGI
					if len(parts) > 1 && strings.Contains(strings.ToLower(parts[1]), "fcgi") {
						location.ProxyType = "fcgi"
					} else {
						location.ProxyType = "http"
					}
					fmt.Printf("  [DEBUG] SetHandler parsed: handler=%s, socket=%s, type=%s\n", location.Handler, location.ProxyUnixSocket, location.ProxyType)
				}
			} else {
				// Handle other handler types
				switch {
				case handlerLower == "proxy:fcgi" || handlerLower == "fcgid-script":
					location.Handler = "proxy"
					location.ProxyType = "fcgi"
				case handlerLower == "proxy" || handlerLower == "proxy-server":
					location.Handler = "proxy"
					location.ProxyType = "http"
				case handlerLower == "cgi-script":
					location.Handler = "cgi"
				case strings.Contains(handlerLower, "php"):
					location.Handler = "php"
				}
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

// expandIncludePath expands glob patterns in include paths
func (p *ApacheHttpdParser) expandIncludePath(pattern string) []string {
	// Check if pattern contains glob characters
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		// No glob, return single resolved path
		return []string{p.resolvePath(pattern)}
	}

	// Resolve base directory for glob
	var searchDir, globPattern string
	if filepath.IsAbs(pattern) {
		searchDir = filepath.Dir(pattern)
		globPattern = filepath.Base(pattern)
	} else {
		// For relative paths, try to resolve relative to baseDir
		// But Apache also supports ServerRoot-relative paths
		// Try baseDir first, then try as-is if it looks like a directory path
		fullPath := filepath.Join(p.baseDir, pattern)
		searchDir = filepath.Dir(fullPath)
		globPattern = filepath.Base(pattern)
		
		// If the resolved directory doesn't exist, try resolving the pattern differently
		// Apache includes are often relative to ServerRoot, not the config file directory
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			// Try treating pattern as relative to parent of baseDir (common for vhosts.d)
			parentDir := filepath.Dir(p.baseDir)
			altPath := filepath.Join(parentDir, pattern)
			altSearchDir := filepath.Dir(altPath)
			if _, err := os.Stat(altSearchDir); err == nil {
				searchDir = altSearchDir
				globPattern = filepath.Base(pattern)
			} else {
				// Try as absolute path from /etc/httpd (common Apache location)
				etcHttpdPath := filepath.Join("/etc/httpd", pattern)
				etcSearchDir := filepath.Dir(etcHttpdPath)
				if _, err := os.Stat(etcSearchDir); err == nil {
					searchDir = etcSearchDir
					globPattern = filepath.Base(pattern)
				}
			}
		}
	}

	// Find matching files
	var matches []string
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		// If directory doesn't exist or can't be read, return empty (IncludeOptional behavior)
		fmt.Printf("Warning: Cannot read directory %s: %v\n", searchDir, err)
		return []string{}
	}

	// Simple glob matching (supports * and ?)
	globRegex := regexp.MustCompile("^" + strings.ReplaceAll(strings.ReplaceAll(globPattern, "*", ".*"), "?", ".") + "$")
	
	for _, entry := range entries {
		if !entry.IsDir() && globRegex.MatchString(entry.Name()) {
			matches = append(matches, filepath.Join(searchDir, entry.Name()))
		}
	}

	return matches
}
