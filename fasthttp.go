package main

import (
	"context"
	"fmt"
    "encoding/json"
	"html"
	"log"
	"net/http"
	"io/ioutil"
	"os"
	"os/signal"
// 	"strings"
	"syscall"
	"time"
	"github.com/yookoala/gofast"
)

type FastHTTPVirtualHost struct {
    PortType string `json:"portType"`
    Listen []string `json:"listen"`
    ServerName string `json:"serverName"`
    ServerAlias []string `json:"serverAlias"`
    DocumentRoot string `json:"documentRoot"`
    User string `json:"user"`
    Group string `json:"group"`
    ServerAdmin string `json:"serverAdmin"`
    ErrorLog string `json:"errorLog"`
    CustomLog string `json:"customLog"`
    DirectoryIndex string `json:"directoryIndex"`
    PHPProxyFCGI string `json:"phpProxyFCGI"`
}

type FastHTTPConfig struct {
	User  string `json:"user"`
	Group string `json:"group"`
	ServerAdmin string `json:"serverAdmin"`
	Listen []string `json:"listen"`
	VirtualHosts []FastHTTPVirtualHost `json:"virtualHosts"`
	HttpPort string `json:"httpPort"`
	HttpsPort string `json:"httpsPort"`
}

func main() {

    configFilePath := "/fast-http/fasthttp.json"
    configFile, err := os.Open(configFilePath)
    if err != nil {
        fmt.Println("Error opening FastHTTP JSON file:", err)
        return
    }
    defer configFile.Close()

    var config FastHTTPConfig
    err = json.NewDecoder(configFile).Decode(&config)
    if err != nil {
        fmt.Println("Error parsing FastHTTP JSON configuration:", err)
        return
    }

    getVirtualHostByServerName := func(serverName string) *FastHTTPVirtualHost {
        for i, v := range config.VirtualHosts {
          if v.ServerName == serverName {
              return &config.VirtualHosts[i]
          }
      }
      return nil
    }

	server := &http.Server{
		Addr: ":" + config.HttpPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Printf("Request from %s", r.RemoteAddr)
            log.Printf("Host: %s", html.EscapeString(r.Host))

            currentUri := r.RequestURI
            log.Printf("currentUri: %s", currentUri)

            virtualHost := getVirtualHostByServerName(r.Host)
            if virtualHost != nil {
//                 log.Printf("Virtual host found: %s", virtualHost.ServerName)
//                 log.Printf("Document root: %s", virtualHost.DocumentRoot)

//                 currentUri := r.RequestURI

                isPHP := false
                files, err := ioutil.ReadDir(virtualHost.DocumentRoot)
                if err == nil {
                    for _, file := range files {
                        if file.Name() == "index.php" {
                            isPHP = true
                            break
                        }
                    }
                }

                if (isPHP && virtualHost.PHPProxyFCGI != "") {

                    connFactory := gofast.SimpleConnFactory("tcp", virtualHost.PHPProxyFCGI)

                    gfhandler := gofast.NewHandler(
                        gofast.NewFileEndpoint(virtualHost.DocumentRoot + "/index.php")(gofast.BasicSession),
                        gofast.SimpleClientFactory(connFactory),
                    )

                    http.HandlerFunc(gfhandler.ServeHTTP).ServeHTTP(w, r)

                } else {
                    http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
                }

            } else {
                log.Printf("Virtual host not found")
                http.FileServer(http.Dir("/var/www/html")).ServeHTTP(w, r)
            }

		}),
	}

	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		log.Println("Starting FastHTTP server on port: " + config.HttpPort)
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
