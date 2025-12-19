package handlers

import (
	"html"
	"log"
	"net/http"
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
	log.Printf("Request from %s", r.RemoteAddr)
	log.Printf("Host: %s", html.EscapeString(r.Host))
	log.Printf("Method: %s", html.EscapeString(r.Method))

	virtualHost := h.config.GetVirtualHostByServerName(r.Host)
	if virtualHost != nil {
		h.handleVirtualHost(w, r, virtualHost)
	} else {
		log.Printf("Virtual host not found")
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
		http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
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
