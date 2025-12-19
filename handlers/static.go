package handlers

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fasthttp/config"
	"fasthttp/utils"
)

// StaticFileHandler handles static file serving and directory listings
type StaticFileHandler struct{}

// NewStaticFileHandler creates a new StaticFileHandler
func NewStaticFileHandler() *StaticFileHandler {
	return &StaticFileHandler{}
}

// CanHandle returns true if this is a static file request
func (h *StaticFileHandler) CanHandle(r *http.Request, virtualHost *config.VirtualHost) bool {
	// Static handler is the default fallback, so it always returns true
	// The router will check other handlers first
	return true
}

// Handle serves static files or directory listings
func (h *StaticFileHandler) Handle(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, effectiveDirectoryIndex string) error {
	// Parse the request path
	urlPath := r.URL.Path
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
		r.URL.Path = urlPath
	}

	// Build the full file path
	fullPath := filepath.Join(virtualHost.DocumentRoot, filepath.Clean(urlPath))

	// Check if the path exists
	info, err := os.Stat(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return nil
	}

	// If it's a directory, check for index files or show directory listing
	if info.IsDir() {
		// Check for index files using Apache-style order
		indexFile := utils.FindIndexFile(fullPath, effectiveDirectoryIndex)
		if indexFile != "" {
			// Index file exists, serve it
			r.URL.Path = filepath.Join(urlPath, indexFile)
			http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
			return nil
		}

		// No index file found, show directory listing with template
		h.handleDirectoryListing(w, r, virtualHost, fullPath, urlPath)
		return nil
	}

	// It's a file, serve it normally
	http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
	return nil
}

type DirectoryEntry struct {
	Name    string
	URL     string
	Size    string
	ModTime string
	Icon    string
}

type DirectoryListingData struct {
	Path       string
	Breadcrumb template.HTML
	Entries    []DirectoryEntry
}

func (h *StaticFileHandler) handleDirectoryListing(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, dirPath, urlPath string) {
	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}

	// Build breadcrumb
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	breadcrumbParts := []string{"<a href=\"/\">Home</a>"}
	currentPath := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		currentPath += "/" + part
		breadcrumbParts = append(breadcrumbParts, "<a href=\""+currentPath+"/\">"+html.EscapeString(part)+"</a>")
	}
	breadcrumb := strings.Join(breadcrumbParts, " / ")

	// Prepare directory entries
	dirEntries := make([]DirectoryEntry, 0, len(entries)+1)

	// Add parent directory link if not at root
	if urlPath != "/" {
		parentPath := filepath.Dir(urlPath)
		if parentPath == "." {
			parentPath = "/"
		}
		if !strings.HasSuffix(parentPath, "/") && parentPath != "/" {
			parentPath += "/"
		}
		dirEntries = append(dirEntries, DirectoryEntry{
			Name:    "..",
			URL:     parentPath,
			Size:    "-",
			ModTime: "-",
			Icon:    "📁",
		})
	}

	// Process directory entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		entryURL := urlPath
		if !strings.HasSuffix(entryURL, "/") {
			entryURL += "/"
		}
		entryURL += entry.Name()
		if entry.IsDir() {
			entryURL += "/"
		}

		var sizeStr string
		if entry.IsDir() {
			sizeStr = "-"
		} else {
			size := info.Size()
			if size < 1024 {
				sizeStr = fmt.Sprintf("%d B", size)
			} else if size < 1024*1024 {
				sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
			} else {
				sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
			}
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
		}

		dirEntries = append(dirEntries, DirectoryEntry{
			Name:    entry.Name(),
			URL:     entryURL,
			Size:    sizeStr,
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			Icon:    icon,
		})
	}

	// Load and execute template
	templatePath := "directory-index.html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		// If template not found, fall back to default directory listing
		utils.ErrorLog("Template not found, using default listing: %v", err)
		http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
		return
	}

	data := DirectoryListingData{
		Path:       urlPath,
		Breadcrumb: template.HTML(breadcrumb),
		Entries:    dirEntries,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		utils.ErrorLog("Error executing template: %v", err)
		http.Error(w, "Error generating directory listing", http.StatusInternalServerError)
		return
	}
}
