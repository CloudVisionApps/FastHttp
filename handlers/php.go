package handlers

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"fasthttp/config"
	"fasthttp/utils"

	"github.com/yookoala/gofast"
)

// PHPHandler handles PHP requests via FastCGI
type PHPHandler struct{}

// NewPHPHandler creates a new PHPHandler
func NewPHPHandler() *PHPHandler {
	return &PHPHandler{}
}

// CanHandle returns true if this is a PHP request
func (h *PHPHandler) CanHandle(r *http.Request, virtualHost *config.VirtualHost) bool {
	if virtualHost.PHPProxyFCGI == "" {
		return false
	}

	currentUri := r.RequestURI
	return h.isPHPRequest(currentUri)
}

// Handle processes PHP requests via FastCGI
func (h *PHPHandler) Handle(w http.ResponseWriter, r *http.Request, virtualHost *config.VirtualHost, effectiveDirectoryIndex string) error {
	currentUri := r.RequestURI
	fileName := utils.GetFileName(currentUri)
	
	// If root or empty, try to find index file (Apache-style)
	if fileName == "/" || fileName == "" {
		indexFile := utils.FindIndexFile(virtualHost.DocumentRoot, effectiveDirectoryIndex)
		if indexFile != "" {
			fileName = indexFile
		} else {
			fileName = "index.php" // Fallback
		}
	}

	log.Printf("Serving PHP file: %s", fileName)

	connFactory := gofast.SimpleConnFactory("tcp", virtualHost.PHPProxyFCGI)

	gofastHandler := gofast.NewHandler(
		gofast.NewFileEndpoint(virtualHost.DocumentRoot+"/"+fileName)(gofast.BasicSession),
		gofast.SimpleClientFactory(connFactory),
	)

	http.HandlerFunc(gofastHandler.ServeHTTP).ServeHTTP(w, r)
	return nil
}

func (h *PHPHandler) isPHPRequest(uri string) bool {
	isFile, _ := utils.IsFileRequest(uri)
	if isFile {
		return false
	}

	if strings.HasSuffix(uri, ".php") {
		return true
	}

	// Check for PHP in query string or path
	pattern := `^.*\.php(\?.*)?$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(uri)
}
