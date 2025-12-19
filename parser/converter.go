package parser

import (
	"fmt"

	"fasthttp/config"
)

// FastHTTPConverter converts ParsedConfig to FastHTTP Config
type FastHTTPConverter struct{}

// NewFastHTTPConverter creates a new converter
func NewFastHTTPConverter() *FastHTTPConverter {
	return &FastHTTPConverter{}
}

// Convert transforms a ParsedConfig to FastHTTP Config
func (c *FastHTTPConverter) Convert(parsed *ParsedConfig, baseConfig *config.Config) (*config.Config, error) {
	if baseConfig == nil {
		baseConfig = &config.Config{}
	}

	// Create new config starting with base
	result := *baseConfig

	// Convert global config
	if user, ok := parsed.GlobalConfig["user"].(string); ok && user != "" {
		result.User = user
	}
	if group, ok := parsed.GlobalConfig["group"].(string); ok && group != "" {
		result.Group = group
	}
	if serverAdmin, ok := parsed.GlobalConfig["serverAdmin"].(string); ok && serverAdmin != "" {
		result.ServerAdmin = serverAdmin
	}
	if listen, ok := parsed.GlobalConfig["listen"].([]string); ok && len(listen) > 0 {
		result.Listen = listen
	}
	if directoryIndex, ok := parsed.GlobalConfig["directoryIndex"].(string); ok && directoryIndex != "" {
		result.DirectoryIndex = directoryIndex
	}

	// Merge MIME types (avoid duplicates)
	mimeMap := make(map[string]bool)
	for _, mt := range result.MimeTypes {
		mimeMap[mt.Ext] = true
	}
	for _, mt := range parsed.MimeTypes {
		if !mimeMap[mt.Ext] {
			result.MimeTypes = append(result.MimeTypes, mt)
			mimeMap[mt.Ext] = true
		}
	}

	// Add virtual hosts
	result.VirtualHosts = append(result.VirtualHosts, parsed.VirtualHosts...)

	// Compile location regexes for all virtual hosts
	for i := range result.VirtualHosts {
		if err := result.VirtualHosts[i].CompileLocationRegexes(); err != nil {
			return nil, fmt.Errorf("error compiling location regexes for %s: %w", result.VirtualHosts[i].ServerName, err)
		}
	}

	return &result, nil
}
