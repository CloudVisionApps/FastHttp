package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"github.com/yookoala/gofast"
)

func main() {
	server := &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Printf("Request from %s", r.RemoteAddr)
            log.Printf("Host: %s", html.EscapeString(r.Host))

            if (r.Host == "adminbolt.com") {

                documentRoot := "/var/www/html"
                http.FileServer(http.Dir(documentRoot)).ServeHTTP(w, r)

                log.Printf("adminbolt.com")
          } else if (r.Host == "vasil-levski.demo.adminbolt.com") {

                documentRoot2 := "/home/vasi96970cxn/public_html"
                http.FileServer(http.Dir(documentRoot2)).ServeHTTP(w, r)

          } else if (r.Host == "wordpress.demo.adminbolt.com") {

                currentUri := r.RequestURI
                log.Printf("currentUri: %s", currentUri)

//                 isPHP := false
                isFile := false
                fileExtensionList := []string{".php", ".html", ".htm", ".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg", ".xml", ".json", ".txt", ".pdf", ".zip", ".gz", ".tar", ".rar", ".mp3", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".ogg", ".ogv", ".webp", ".woff", ".woff2", ".ttf", ".eot", ".otf", ".swf", ".fla", ".psd", ".ai", ".eps", ".indd", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp", ".md", ".csv", ".sql", ".json", ".xml", ".yml", ".yaml", ".log", ".conf", ".ini", ".htaccess", ".htpasswd", ".bak", ".tmp", ".temp", ".swp", ".swo", ".swn"}
//                 if currentUri == "/" {
//                     isPHP = true
//                 }
                for _, ext := range fileExtensionList {
                    if strings.HasSuffix(currentUri, ext) {
                        isFile = true
                        break
                    }
                }


                documentRoot2 := "/home/word2442we7v/public_html"

                if (isFile == false) {
                    connFactory := gofast.SimpleConnFactory("tcp", "127.0.0.1:9076")

                   gfhandler := gofast.NewHandler(
                        gofast.NewFileEndpoint(documentRoot2 + "/index.php")(gofast.BasicSession),
                        gofast.SimpleClientFactory(connFactory),
                    )

                    http.HandlerFunc(gfhandler.ServeHTTP).ServeHTTP(w, r)
                } else {
                    http.FileServer(http.Dir(documentRoot2)).ServeHTTP(w, r)
                }

          } else {
              fmt.Fprintf(w, "Hello, %q! Host: %s", html.EscapeString(r.URL.Path))
          }

		}),
	}

	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		log.Println("Starting server on :80")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Block until a signal is received
	<-stop
	log.Println("Shutting down server...")

	// Create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
