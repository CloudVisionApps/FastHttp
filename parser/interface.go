package parser

import (
	"fasthttp/config"
)

// Parser is the interface for configuration parsers
type Parser interface {
	// Parse reads and parses a configuration file
	Parse(filePath string) (*ParsedConfig, error)
	
	// CanParse checks if this parser can handle the given file
	CanParse(filePath string) bool
}

// ParsedConfig represents a parsed configuration that can be converted to FastHTTP config
type ParsedConfig struct {
	VirtualHosts []config.VirtualHost
	GlobalConfig map[string]interface{}
	MimeTypes    []config.MimeType
	Includes     []string
}

// Converter converts parsed config to FastHTTP Config
type Converter interface {
	// Convert transforms a ParsedConfig to FastHTTP Config
	Convert(parsed *ParsedConfig, baseConfig *config.Config) (*config.Config, error)
}
