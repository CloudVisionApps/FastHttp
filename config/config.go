package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type VirtualHost struct {
	PortType       string   `json:"portType"`
	Listen         []string `json:"listen"`
	ServerName     string   `json:"serverName"`
	ServerAlias    []string `json:"serverAlias"`
	DocumentRoot   string   `json:"documentRoot"`
	User           string   `json:"user"`
	Group          string   `json:"group"`
	ServerAdmin    string   `json:"serverAdmin"`
	ErrorLog       string   `json:"errorLog"`
	CustomLog      string   `json:"customLog"`
	DirectoryIndex string   `json:"directoryIndex"`
	PHPProxyFCGI   string   `json:"phpProxyFCGI"`
}

type MimeType struct {
	Ext  string `json:"ext"`
	Type string `json:"type"`
}

type Config struct {
	User                  string        `json:"user"`
	Group                 string        `json:"group"`
	ServerAdmin           string        `json:"serverAdmin"`
	Listen                []string      `json:"listen"`
	VirtualHosts          []VirtualHost `json:"virtualHosts"`
	HttpPort              string        `json:"httpPort"`
	HttpsPort             string        `json:"httpsPort"`
	MimeTypes             []MimeType    `json:"mimeTypes"`
	RateLimitRequests     int           `json:"rateLimitRequests"`
	RateLimitWindowSeconds int          `json:"rateLimitWindowSeconds"`
}

func Load(configFilePath string) (*Config, error) {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening FastHTTP JSON file: %w", err)
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error parsing FastHTTP JSON configuration: %w", err)
	}

	return &config, nil
}

func (c *Config) GetVirtualHostByServerName(serverName string) *VirtualHost {
	for i, v := range c.VirtualHosts {
		if v.ServerName == serverName {
			return &c.VirtualHosts[i]
		}
	}
	return nil
}

func (c *Config) GetRateLimitConfig() (maxRequests, windowSeconds int) {
	maxRequests = c.RateLimitRequests
	if maxRequests <= 0 {
		maxRequests = 100 // Default: 100 requests per window
	}
	windowSeconds = c.RateLimitWindowSeconds
	if windowSeconds <= 0 {
		windowSeconds = 60 // Default: 60 seconds window
	}
	return maxRequests, windowSeconds
}
