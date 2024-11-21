package main

import (
// 	"context"
	"fmt"
    "encoding/json"
	"html"
	"log"
	"path"
	"net/http"
	"net/url"
// 	"io/ioutil"
	"regexp"
	"strconv"
	"os"
	"strings"
// 	"time"
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

type FastHTTPMimeType struct {
    Ext string `json:"ext"`
    Type string `json:"type"`
}
type FastHTTPConfig struct {
	User  string `json:"user"`
	Group string `json:"group"`
	ServerAdmin string `json:"serverAdmin"`
	Listen []string `json:"listen"`
	VirtualHosts []FastHTTPVirtualHost `json:"virtualHosts"`
	HttpPort string `json:"httpPort"`
	HttpsPort string `json:"httpsPort"`
	MimeTypes []FastHTTPMimeType `json:"mimeTypes"`
}

func GetFileName(uri string) string {
	// Parse the URI
	parsedURI, err := url.Parse(uri)
	if err != nil {
		// Handle error if the URI is malformed
		fmt.Printf("Error parsing URI: %v\n", err)
		return ""
	}

	// Get the file name from the path
	return path.Base(parsedURI.Path)
}

func isFileRequest(uri string) (bool, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return false, err
	}

	ext := path.Ext(parsedURI.Path)
	return ext != "", nil
}

func main() {

    command := os.Args[1]
    if len(os.Args) < 2 {
        fmt.Println("Usage: fasthttp <command>")
        os.Exit(1)
    }

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
            log.Printf("Method: %s", html.EscapeString(r.Method))

            virtualHost := getVirtualHostByServerName(r.Host)
            if virtualHost != nil {

                currentUri := r.RequestURI

                isPHP := true
//                 files, err := ioutil.ReadDir(virtualHost.DocumentRoot)
//                 if err == nil {
//                     for _, file := range files {
//                         if file.Name() == "index.php" {
//                             if currentUri == "/" {
//                                 isPHP = true
//                                 break
//                             }
//                         }
//                     }
//                 }

                isFile, _ := isFileRequest(currentUri)
                if isFile {
                    isPHP = false
                }
                for _, mimeType := range config.MimeTypes {
                   if strings.HasSuffix(currentUri, mimeType.Ext) {
                      isPHP = false
                      break
                   }
                   if strings.HasSuffix(currentUri, ".php") {
                        isPHP = true
                        break
                   }

                    pattern := `^.*\.php(\?.*)?$`
                    // Compile the regex
                    re := regexp.MustCompile(pattern)
                    // Check if the URI matches
                    if re.MatchString(currentUri) {
                        isPHP = true
                        break
                    }
               }

                log.Printf(currentUri)
                log.Printf("isPHP: %t", isPHP)

                if (isPHP && virtualHost.PHPProxyFCGI != "") {

                    fileName := GetFileName(currentUri)
                    if fileName == "/" || fileName == "" {
                        fileName = "index.php"
                    }

                    log.Printf("Serving PHP file: %s", fileName)

                    connFactory := gofast.SimpleConnFactory("tcp", virtualHost.PHPProxyFCGI)

                    gofastHandler := gofast.NewHandler(
                        gofast.NewFileEndpoint(virtualHost.DocumentRoot + "/" + fileName)(gofast.BasicSession),
                        gofast.SimpleClientFactory(connFactory),
                    )

                    http.HandlerFunc(gofastHandler.ServeHTTP).ServeHTTP(w, r)

                } else {
                    http.FileServer(http.Dir(virtualHost.DocumentRoot)).ServeHTTP(w, r)
                }

            } else {
                log.Printf("Virtual host not found")
                http.FileServer(http.Dir("/var/www/html")).ServeHTTP(w, r)
            }

		}),
	}

	// Run the server in a goroutine
	if command == "start" {

        // Get the current process ID
        pid := os.Getpid()

        // Create or overwrite the PID file
        pidFile := "/var/run/fasthttp.pid"
        file, err := os.Create(pidFile)
        if err != nil {
            log.Fatal("Error creating PID file:", err)
        }
        defer file.Close()

        // Write the process ID to the file
        _, err = file.WriteString(strconv.Itoa(pid))
        if err != nil {
            log.Fatal("Error writing to PID file:", err)
        }

		log.Println("Starting FastHTTP server on port: " + config.HttpPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
			 _, err = file.WriteString("")
             if err != nil {
                log.Fatal("Error writing to PID file:", err)
             }
		}

        select {}

	} else if command == "stop" {
        log.Println("Shutting down server...")

        // Finding process ID of the server
        pidFile, err := os.Open("/var/run/fasthttp.pid")
        if err != nil {
            log.Fatalf("Error opening PID file: %v", err)
        }
        defer pidFile.Close()

        pidBytes, err := os.ReadFile("/var/run/fasthttp.pid")
        if err != nil {
            log.Fatalf("Error reading PID file: %v", err)
        }
        pid, err := strconv.Atoi(string(pidBytes))
        if err != nil {
            log.Fatalf("Error converting PID to integer: %v", err)
        }

        // Kill the server
        process, err := os.FindProcess(pid)
        if err != nil {
            log.Fatalf("Error finding process: %v", err)
        }
        err = process.Kill()
        if err != nil {
            log.Fatalf("Error killing process: %v", err)
        }

        log.Println("Server stopped")
    } else {
        fmt.Println("Unknown command")
    }

}
