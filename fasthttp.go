package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"fasthttp/admin"
	"fasthttp/config"
	"fasthttp/handlers"
	"fasthttp/parser"
	"fasthttp/process"
	"fasthttp/ratelimit"
	"fasthttp/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fasthttp <command>")
		os.Exit(1)
	}

	command := os.Args[1]
	configFilePath := "fasthttp.json"

	cfg, err := config.Load(configFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize loggers
	if err := utils.InitLoggers(cfg.LogFile, cfg.AdminLogFile, cfg.ErrorLogFile); err != nil {
		fmt.Printf("Error initializing loggers: %v\n", err)
		os.Exit(1)
	}

	utils.WebServerLog("Configuration loaded successfully from %s", configFilePath)

	switch command {
	case "start":
		startServer(cfg)
	case "stop":
		if err := process.Stop(); err != nil {
			utils.ErrorLog("Error stopping server: %v", err)
			os.Exit(1)
		}
	case "status":
		ports := cfg.GetAllListenPorts()
		portStr := "80"
		if len(ports) > 0 {
			portStr = ports[0]
		}
		if err := process.Status(portStr); err != nil {
			utils.ErrorLog("Error getting status: %v", err)
			os.Exit(1)
		}
	case "convert":
		handleConvert()
	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}
}

func startServer(cfg *config.Config) {
	// Start admin API if enabled
	if cfg.AdminEnabled {
		adminPort := cfg.AdminPort
		if adminPort == "" {
			adminPort = "8080"
		}
		go func() {
			admin.StartAdminPanel(cfg, "fasthttp.json", adminPort)
		}()
		utils.WebServerLog("[Web Server] Admin API enabled on port: %s", adminPort)
	}

	// Initialize rate limiter
	maxRequests, windowSeconds := cfg.GetRateLimitConfig()
	rateLimiter := ratelimit.New(maxRequests, windowSeconds)

	// Create request handler with rate limiting middleware
	handler := handlers.New(cfg)
	rateLimitHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check rate limit
		clientIP := utils.GetClientIP(r)
		if !rateLimiter.Allow(clientIP) {
// 			log.Printf("Rate limit exceeded for IP: %s", clientIP)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		handler.ServeHTTP(w, r)
	})

	// Get all ports to listen on
	listenPorts := cfg.GetAllListenPorts()
	
	// If no ports configured, use default port 80
	if len(listenPorts) == 0 {
		listenPorts = []string{"80"}
	}

	// Write PID file (before switching user, in case we need root permissions)
	if err := process.WritePID(); err != nil {
		utils.ErrorLog("Error writing PID file: %v", err)
		os.Exit(1)
	}

	// Start listening on all ports
	// Note: For privileged ports (< 1024), we need root to bind
	// After binding, we'll drop privileges to the configured user/group
	if len(listenPorts) == 1 {
		// Single port - bind and drop privileges if needed
		server := &http.Server{
			Addr:    ":" + listenPorts[0],
			Handler: rateLimitHandler,
		}
		
		port, _ := strconv.Atoi(listenPorts[0])
		isPrivileged := port < 1024
		
		if isPrivileged && (cfg.User != "" || cfg.Group != "") && os.Geteuid() == 0 {
			// Create listener manually, then drop privileges
			listener, err := net.Listen("tcp", ":"+listenPorts[0])
			if err != nil {
				utils.ErrorLog("Error creating listener: %v", err)
				os.Exit(1)
	}

			// Drop privileges after binding
			if err := utils.SwitchUserGroup(cfg.User, cfg.Group); err != nil {
				utils.ErrorLog("Error dropping privileges: %v", err)
				listener.Close()
				os.Exit(1)
			}
			
			currentUser, currentGroup, _ := utils.GetCurrentUser()
			utils.WebServerLog("[Web Server] Dropped privileges, running as user: %s, group: %s", currentUser, currentGroup)
			utils.WebServerLog("[Web Server] Starting FastHTTP server on port: %s", listenPorts[0])
			
			if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
				utils.ErrorLog("Server failed: %v", err)
				os.Exit(1)
			}
		} else {
			// Normal binding (non-privileged or no user switching)
			utils.WebServerLog("[Web Server] Starting FastHTTP server on port: %s", listenPorts[0])
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				utils.ErrorLog("Server failed: %v", err)
				os.Exit(1)
			}
		}
	} else {
		// Multiple ports - need to handle privileged ports correctly
		// Check if we have any privileged ports and if we're running as root
		hasPrivilegedPorts := false
		for _, port := range listenPorts {
			portNum, _ := strconv.Atoi(port)
			if portNum < 1024 {
				hasPrivilegedPorts = true
				break
			}
		}

		// If we have privileged ports and want to drop privileges, we need to:
		// 1. Create all listeners as root
		// 2. Drop privileges
		// 3. Serve on all listeners
		if hasPrivilegedPorts && (cfg.User != "" || cfg.Group != "") && os.Geteuid() == 0 {
			// Create all listeners first
			listeners := make(map[string]net.Listener)
			for _, port := range listenPorts {
				listener, err := net.Listen("tcp", ":"+port)
				if err != nil {
					utils.ErrorLog("Error creating listener for port %s: %v", port, err)
					// Close already created listeners
					for _, l := range listeners {
						l.Close()
					}
					os.Exit(1)
				}
				listeners[port] = listener
			}

			// Drop privileges after binding
			if err := utils.SwitchUserGroup(cfg.User, cfg.Group); err != nil {
				utils.ErrorLog("Error dropping privileges: %v", err)
				// Close listeners
				for _, l := range listeners {
					l.Close()
				}
				os.Exit(1)
			}

			currentUser, currentGroup, _ := utils.GetCurrentUser()
			utils.WebServerLog("[Web Server] Dropped privileges, running as user: %s, group: %s", currentUser, currentGroup)

			// Start serving on all listeners
			for port, listener := range listeners {
				go func(p string, l net.Listener) {
					server := &http.Server{
						Addr:    ":" + p,
						Handler: rateLimitHandler,
					}
					utils.WebServerLog("[Web Server] Starting FastHTTP server on port: %s", p)
					if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
						utils.ErrorLog("Server failed on port %s: %v", p, err)
					}
				}(port, listener)
			}
			// Keep main goroutine alive
			select {}
		} else {
			// No privileged ports or not running as root - can switch user before binding
			if cfg.User != "" || cfg.Group != "" {
				// Try to switch before binding (will only work if not binding to privileged ports)
				if err := utils.SwitchUserGroup(cfg.User, cfg.Group); err != nil {
					utils.ErrorLog("Warning: Could not switch to user/group before binding: %v", err)
					utils.ErrorLog("If binding to privileged ports (< 1024), start as root")
				} else {
					currentUser, currentGroup, _ := utils.GetCurrentUser()
					utils.WebServerLog("[Web Server] Running as user: %s, group: %s", currentUser, currentGroup)
				}
			}

			for _, port := range listenPorts {
				go func(p string) {
					server := &http.Server{
						Addr:    ":" + p,
						Handler: rateLimitHandler,
					}
					utils.WebServerLog("[Web Server] Starting FastHTTP server on port: %s", p)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						utils.ErrorLog("Server failed on port %s: %v", p, err)
					}
				}(port)
			}
			// Keep main goroutine alive
			select {}
		}
	}
}

// handleConvert handles the convert command for migrating from other web servers
func handleConvert() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: fasthttp convert --from <format> --input <input-file> [--output <output-file>]")
		fmt.Println("Formats: apache, httpd")
		fmt.Println("Example: fasthttp convert --from apache --input /etc/httpd/conf/httpd.conf --output fasthttp.json")
		os.Exit(1)
	}

	var fromFormat, inputFile, outputFile string
	
	// Parse arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--from":
			if i+1 < len(os.Args) {
				fromFormat = os.Args[i+1]
				i++
			}
		case "--input":
			if i+1 < len(os.Args) {
				inputFile = os.Args[i+1]
				i++
			}
		case "--output":
			if i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++
			}
		}
	}

	if fromFormat == "" || inputFile == "" {
		fmt.Println("Error: --from and --input are required")
		os.Exit(1)
	}

	if outputFile == "" {
		outputFile = "fasthttp.json"
	}

	// Create parser registry
	registry := parser.NewRegistry()

	// Parse the input file
	fmt.Printf("Parsing %s configuration from %s...\n", fromFormat, inputFile)
	parsed, err := registry.ParseFile(inputFile)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Found %d include(s), %d virtual host(s) in main file\n", len(parsed.Includes), len(parsed.VirtualHosts))
	if len(parsed.Includes) > 0 {
		fmt.Printf("Processing includes: %v\n", parsed.Includes)
	}

	// Convert to FastHTTP config
	converter := parser.NewFastHTTPConverter()
	baseConfig := &config.Config{} // Start with empty config
	fastHTTPConfig, err := converter.Convert(parsed, baseConfig)
	if err != nil {
		fmt.Printf("Error converting config: %v\n", err)
		os.Exit(1)
	}

	// Clean empty values before marshaling
	cleanConfig := removeEmptyValues(fastHTTPConfig)

	// Save to JSON file
	configJSON, err := json.MarshalIndent(cleanConfig, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, configJSON, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted configuration to %s\n", outputFile)
	fmt.Printf("Converted %d virtual host(s)\n", len(fastHTTPConfig.VirtualHosts))
	fmt.Printf("Note: Please review and adjust the configuration as needed.\n")
}

// removeEmptyValues removes empty/null/zero values from the config using reflection
func removeEmptyValues(cfg *config.Config) interface{} {
	return cleanValue(reflect.ValueOf(cfg).Elem())
}

// cleanValue recursively removes empty values from a reflect.Value
func cleanValue(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.Struct:
		result := make(map[string]interface{})
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)
			
			// Skip unexported fields
			if !field.CanInterface() {
				continue
			}
			
			// Get JSON tag name
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			// Remove omitempty and other options
			jsonName := strings.Split(jsonTag, ",")[0]
			if jsonName == "" {
				jsonName = fieldType.Name
			}
			
			// Recursively clean nested values first
			cleaned := cleanValue(field)
			if cleaned == nil {
				continue
			}
			
			// Check if cleaned value is actually non-empty
			cleanedVal := reflect.ValueOf(cleaned)
			if isEmptyValue(cleanedVal) {
				continue
			}
			
			// For maps, check if they have any entries
			if cleanedVal.Kind() == reflect.Map {
				if cleanedVal.Len() == 0 {
					continue
				}
			}
			
			// For slices, check if they have any entries
			if cleanedVal.Kind() == reflect.Slice {
				if cleanedVal.Len() == 0 {
					continue
				}
			}
			
			// Add the cleaned value
			result[jsonName] = cleaned
		}
		if len(result) == 0 {
			return nil
		}
		return result
		
	case reflect.Slice:
		if v.IsNil() || v.Len() == 0 {
			return nil
		}
		result := make([]interface{}, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			cleaned := cleanValue(v.Index(i))
			if cleaned == nil {
				continue
			}
			
			// Check if cleaned value is actually non-empty
			cleanedVal := reflect.ValueOf(cleaned)
			if isEmptyValue(cleanedVal) {
				continue
			}
			
			// For maps, only include if they have entries
			if cleanedVal.Kind() == reflect.Map {
				if cleanedVal.Len() > 0 {
					result = append(result, cleaned)
				}
			} else if cleanedVal.Kind() == reflect.Slice {
				// For slices, only include if they have entries
				if cleanedVal.Len() > 0 {
					result = append(result, cleaned)
				}
			} else {
				// For other types, include if not empty
				result = append(result, cleaned)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
		
	case reflect.Map:
		if v.IsNil() || v.Len() == 0 {
			return nil
		}
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			cleaned := cleanValue(val)
			if cleaned != nil && !isEmptyValue(reflect.ValueOf(cleaned)) {
				result[fmt.Sprintf("%v", key)] = cleaned
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
		
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return cleanValue(v.Elem())
		
	default:
		if isEmptyValue(v) {
			return nil
		}
		return v.Interface()
	}
}

// isEmptyValue checks if a value is empty/zero/null
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool() // Skip false booleans
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
