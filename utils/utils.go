package utils

import (
	"net/http"
	"net/url"
	"path"
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
