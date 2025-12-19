package handlers

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fasthttp/config"
	"fasthttp/utils"

	"github.com/yookoala/gofast"
)

type Handler struct {
	config *config.Config
}

func New(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	virtualHost := h.config.GetVirtualHostByServerName(r.Host)
	if virtualHost != nil {

	    log.Printf("Request from %s", r.RemoteAddr)
        log.Printf("Host: %s", html.EscapeString(r.Host))
        log.Printf("Method: %s", html.EscapeString(r.Method))

		h.handleVirtualHost(w, r, virtualHost)
	} else {
// 		log.Printf("Virtual host not found")
		http.FileServer(http.Dir("/var/www/html")).ServeHTTP(w, r)
	}
}

func (h *Handler) handleVirtualHost(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost) {
	currentUri := r.RequestURI

	isPHP := h.isPHPRequest(currentUri)

	log.Printf("URI: %s", currentUri)
	log.Printf("isPHP: %t", isPHP)

	if isPHP && virtualHost.PHPProxyFCGI != "" {
		h.handlePHP(w, r, virtualHost, currentUri)
	} else {
		h.handleFileServer(w, r, virtualHost)
	}
}

func (h *Handler) isPHPRequest(uri string) bool {
	isFile, _ := utils.IsFileRequest(uri)
	if isFile {
		return false
	}

	for _, mimeType := range h.config.MimeTypes {
		if strings.HasSuffix(uri, mimeType.Ext) {
			return false
		}
		if strings.HasSuffix(uri, ".php") {
			return true
		}

		pattern := `^.*\.php(\?.*)?$`
		re := regexp.MustCompile(pattern)
		if re.MatchString(uri) {
			return true
		}
	}

	return true
}

func (h *Handler) handlePHP(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, uri string) {
	fileName := utils.GetFileName(uri)
	if fileName == "/" || fileName == "" {
		fileName = "index.php"
	}

	log.Printf("Serving PHP file: %s", fileName)

	connFactory := gofast.SimpleConnFactory("tcp", virtualHost.PHPProxyFCGI)

	gofastHandler := gofast.NewHandler(
		gofast.NewFileEndpoint(virtualHost.DocumentRoot+"/"+fileName)(gofast.BasicSession),
		gofast.SimpleClientFactory(connFactory),
	)

	http.HandlerFunc(gofastHandler.ServeHTTP).ServeHTTP(w, r)
}

func (h *Handler) handleFileServer(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost) {
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
		return
	}

	// If it's a directory, check for index files or show directory listing
	if info.IsDir() {
		// Check for index files
		indexFiles := []string{"index.html", "index.htm", "index.php"}
		if virtualHost.DirectoryIndex != "" {
			indexFiles = append([]string{virtualHost.DirectoryIndex}, indexFiles...)
		}

		for _, indexFile := range indexFiles {
			indexPath := filepath.Join(fullPath, indexFile)
			if _, err := os.Stat(indexPath); err == nil {
				// Index file exists, serve it
				r.URL.Path = filepath.Join(urlPath, indexFile)
				http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
				return
			}
		}

		// No index file found, show directory listing with template
		h.handleDirectoryListing(w, r, virtualHost, fullPath, urlPath)
		return
	}

	// It's a file, serve it normally
	http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
}

type DirectoryEntry struct {
	Name    string
	URL     string
	Size    string
	ModTime string
	Icon    string
}

type DirectoryListingData struct {
	Path      string
	Breadcrumb template.HTML
	Entries   []DirectoryEntry
}

func (h *Handler) handleDirectoryListing(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, dirPath, urlPath string) {
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
		log.Printf("Template not found, using default listing: %v", err)
		http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
		return
	}

	data := DirectoryListingData{
		Path:      urlPath,
		Breadcrumb: template.HTML(breadcrumb),
		Entries:   dirEntries,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error generating directory listing", http.StatusInternalServerError)
		return
	}
}
