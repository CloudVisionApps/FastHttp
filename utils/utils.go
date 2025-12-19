package utils

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func GetFileName(uri string) string {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	return path.Base(parsedURI.Path)
}

func IsFileRequest(uri string) (bool, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return false, err
	}

	ext := path.Ext(parsedURI.Path)
	return ext != "", nil
}

func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// GetIndexFiles returns the list of index files to check, in Apache-style order
// If directoryIndex is set, it's parsed as space-separated list and used first
// Then defaults are appended: index.html, index.htm, index.php
func GetIndexFiles(directoryIndex string) []string {
	var indexFiles []string

	// Parse DirectoryIndex if set (Apache allows space-separated list)
	if directoryIndex != "" {
		parts := strings.Fields(directoryIndex)
		indexFiles = append(indexFiles, parts...)
	}

	// Add default index files in Apache order (if not already in list)
	defaults := []string{"index.html", "index.htm", "index.php"}
	for _, def := range defaults {
		found := false
		for _, existing := range indexFiles {
			if existing == def {
				found = true
				break
			}
		}
		if !found {
			indexFiles = append(indexFiles, def)
		}
	}

	return indexFiles
}

// FindIndexFile checks a directory for index files in order and returns the first found
// Returns the index file path if found, empty string otherwise
func FindIndexFile(dirPath, directoryIndex string) string {
	indexFiles := GetIndexFiles(directoryIndex)

	for _, indexFile := range indexFiles {
		indexPath := filepath.Join(dirPath, indexFile)
		if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
			return indexFile
		}
	}

	return ""
}
