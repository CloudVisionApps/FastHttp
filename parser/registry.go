package parser

import (
	"fmt"
	"os"
)

// Registry manages available parsers
type Registry struct {
	parsers []Parser
}

// NewRegistry creates a new parser registry
func NewRegistry() *Registry {
	registry := &Registry{
		parsers: []Parser{},
	}
	
	// Register built-in parsers
	registry.Register(NewApacheHttpdParser())
	
	return registry
}

// Register adds a parser to the registry
func (r *Registry) Register(parser Parser) {
	r.parsers = append(r.parsers, parser)
}

// FindParser finds a parser that can handle the given file
func (r *Registry) FindParser(filePath string) (Parser, error) {
	for _, parser := range r.parsers {
		if parser.CanParse(filePath) {
			return parser, nil
		}
	}
	return nil, fmt.Errorf("no parser found for file: %s", filePath)
}

// ParseFile parses a configuration file using the appropriate parser
func (r *Registry) ParseFile(filePath string) (*ParsedConfig, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	parser, err := r.FindParser(filePath)
	if err != nil {
		return nil, err
	}

	return parser.Parse(filePath)
}
