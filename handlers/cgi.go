package handlers

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fasthttp/config"
)

// CGIHandler handles CGI program execution
type CGIHandler struct{}

// NewCGIHandler creates a new CGIHandler
func NewCGIHandler() *CGIHandler {
	return &CGIHandler{}
}

// CanHandle returns true if this is a CGI request
func (h *CGIHandler) CanHandle(r *http.Request, virtualHost *config.VirtualHost) bool {
	urlPath := r.URL.Path
	fullPath := filepath.Join(virtualHost.DocumentRoot, filepath.Clean(urlPath))

	// Check if the file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}

	// Check if it's in the CGI path (if configured)
	if virtualHost.CGIPath != "" && strings.HasPrefix(urlPath, virtualHost.CGIPath) {
		return true
	}

	// Check for common CGI extensions
	cgiExtensions := []string{".cgi", ".pl", ".py", ".sh"}
	for _, ext := range cgiExtensions {
		if strings.HasSuffix(fullPath, ext) {
			return true
		}
	}

	// Check if file is executable (Unix)
	if info.Mode()&0111 != 0 {
		return true
	}

	return false
}

// Handle executes CGI programs
func (h *CGIHandler) Handle(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost) error {
	urlPath := r.URL.Path
	fullPath := filepath.Join(virtualHost.DocumentRoot, filepath.Clean(urlPath))

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return nil
	}

	log.Printf("Executing CGI: %s", fullPath)

	// Set up environment variables for CGI
	env := os.Environ()
	env = append(env, "REQUEST_METHOD="+r.Method)
	env = append(env, "REQUEST_URI="+r.RequestURI)
	env = append(env, "QUERY_STRING="+r.URL.RawQuery)
	env = append(env, "SCRIPT_NAME="+urlPath)
	env = append(env, "SCRIPT_FILENAME="+fullPath)
	env = append(env, "DOCUMENT_ROOT="+virtualHost.DocumentRoot)
	env = append(env, "SERVER_NAME="+r.Host)
	env = append(env, "SERVER_PORT="+r.URL.Port())
	env = append(env, "SERVER_PROTOCOL=HTTP/1.1")
	env = append(env, "HTTP_HOST="+r.Host)

	// Add HTTP headers as environment variables
	for key, values := range r.Header {
		envKey := "HTTP_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		env = append(env, envKey+"="+strings.Join(values, ", "))
	}

	// Add client information
	if r.RemoteAddr != "" {
		env = append(env, "REMOTE_ADDR="+r.RemoteAddr)
	}

	// Create command
	cmd := exec.Command(fullPath)
	cmd.Env = env
	cmd.Dir = filepath.Dir(fullPath)

	// Set up stdin from request body
	cmd.Stdin = r.Body

	// Capture stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("CGI execution error: %v", err)
		http.Error(w, "CGI execution failed", http.StatusInternalServerError)
		return err
	}

	// Write output to response
	// Note: In a production system, you'd want to parse headers from output
	// For now, we'll write the output directly
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(output)

	return nil
}
