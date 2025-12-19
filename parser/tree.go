package parser

import (
	"path/filepath"
	"strings"

	"fasthttp/config"
)

// ConfigNode represents a node in the Apache configuration tree
type ConfigNode struct {
	Type       string                 // Node type: "root", "VirtualHost", "Directory", "Location", "FilesMatch", "IfModule", etc.
	Directive  string                 // Directive name (e.g., "VirtualHost", "Directory")
	Arguments  []string                // Directive arguments
	Directives map[string][]string    // Simple directives (name -> values)
	Children   []*ConfigNode          // Child nodes (nested blocks)
	Parent     *ConfigNode            // Parent node
}

// NewConfigNode creates a new configuration node
func NewConfigNode(nodeType, directive string, args []string) *ConfigNode {
	return &ConfigNode{
		Type:       nodeType,
		Directive:  directive,
		Arguments:  args,
		Directives: make(map[string][]string),
		Children:   []*ConfigNode{},
		Parent:     nil,
	}
}

// AddChild adds a child node to this node
func (n *ConfigNode) AddChild(child *ConfigNode) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// AddDirective adds a simple directive to this node
func (n *ConfigNode) AddDirective(name string, values []string) {
	if n.Directives == nil {
		n.Directives = make(map[string][]string)
	}
	n.Directives[name] = values
}

// GetDirective gets a directive value (returns first value if multiple)
func (n *ConfigNode) GetDirective(name string) string {
	if values, ok := n.Directives[name]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// GetDirectiveAll gets all values for a directive
func (n *ConfigNode) GetDirectiveAll(name string) []string {
	if values, ok := n.Directives[name]; ok {
		return values
	}
	return []string{}
}

// FindChildren finds all child nodes of a specific type
func (n *ConfigNode) FindChildren(nodeType string) []*ConfigNode {
	var result []*ConfigNode
	for _, child := range n.Children {
		if child.Type == nodeType {
			result = append(result, child)
		}
	}
	return result
}

// FindVirtualHosts finds all VirtualHost nodes
func (n *ConfigNode) FindVirtualHosts() []*ConfigNode {
	return n.FindChildren("VirtualHost")
}

// FindDirectories finds all Directory nodes
func (n *ConfigNode) FindDirectories() []*ConfigNode {
	return n.FindChildren("Directory")
}

// FindLocations finds all Location nodes
func (n *ConfigNode) FindLocations() []*ConfigNode {
	return n.FindChildren("Location")
}

// FindFilesMatch finds all FilesMatch nodes
func (n *ConfigNode) FindFilesMatch() []*ConfigNode {
	return n.FindChildren("FilesMatch")
}

// IsInVirtualHost checks if this node is inside a VirtualHost block
func (n *ConfigNode) IsInVirtualHost() bool {
	current := n
	for current != nil {
		if current.Type == "VirtualHost" {
			return true
		}
		current = current.Parent
	}
	return false
}

// GetVirtualHostParent gets the parent VirtualHost node if this node is inside one
func (n *ConfigNode) GetVirtualHostParent() *ConfigNode {
	current := n
	for current != nil {
		if current.Type == "VirtualHost" {
			return current
		}
		current = current.Parent
	}
	return nil
}

// ConvertToParsedConfig converts the tree to ParsedConfig
func (n *ConfigNode) ConvertToParsedConfig() *ParsedConfig {
	parsed := &ParsedConfig{
		VirtualHosts: []config.VirtualHost{},
		GlobalConfig: make(map[string]interface{}),
		MimeTypes:    []config.MimeType{},
		Includes:     []string{},
	}

	// Extract global directives
	if user := n.GetDirective("User"); user != "" {
		parsed.GlobalConfig["user"] = user
	}
	if group := n.GetDirective("Group"); group != "" {
		parsed.GlobalConfig["group"] = group
	}
	if serverAdmin := n.GetDirective("ServerAdmin"); serverAdmin != "" {
		parsed.GlobalConfig["serverAdmin"] = serverAdmin
	}
	if listen := n.GetDirectiveAll("Listen"); len(listen) > 0 {
		parsed.GlobalConfig["listen"] = listen
	}
	if directoryIndex := n.GetDirective("DirectoryIndex"); directoryIndex != "" {
		parsed.GlobalConfig["directoryIndex"] = directoryIndex
	}

	// Extract MIME types
	if addTypes := n.GetDirectiveAll("AddType"); len(addTypes) > 0 {
		for _, addType := range addTypes {
			parts := strings.Fields(addType)
			if len(parts) >= 2 {
				mimeType := parts[0]
				for _, ext := range parts[1:] {
					parsed.MimeTypes = append(parsed.MimeTypes, config.MimeType{
						Ext:  strings.TrimPrefix(ext, "."),
						Type: mimeType,
					})
				}
			}
		}
	}

	// Extract global Directory blocks (not inside VirtualHost)
	var globalLocations []config.Location
	for _, dirNode := range n.FindDirectories() {
		if !dirNode.IsInVirtualHost() {
			locations := dirNode.convertToLocations()
			globalLocations = append(globalLocations, locations...)
		}
	}
	if len(globalLocations) > 0 {
		parsed.GlobalConfig["globalLocations"] = globalLocations
	}

	// Extract VirtualHosts
	for _, vhostNode := range n.FindVirtualHosts() {
		vhost := vhostNode.convertToVirtualHost()
		if vhost != nil {
			parsed.VirtualHosts = append(parsed.VirtualHosts, *vhost)
		}
	}

	return parsed
}

// convertToVirtualHost converts a VirtualHost node to config.VirtualHost
func (n *ConfigNode) convertToVirtualHost() *config.VirtualHost {
	vhost := &config.VirtualHost{
		Locations: []config.Location{},
	}

	// Extract listen ports from VirtualHost arguments (e.g., <VirtualHost *:80>)
	if len(n.Arguments) > 0 {
		port := extractPort(n.Arguments[0])
		if port != "" {
			vhost.Listen = []string{port}
		}
	}

	// Extract directives
	vhost.ServerName = n.GetDirective("ServerName")
	vhost.ServerAlias = n.GetDirectiveAll("ServerAlias")
	vhost.DocumentRoot = n.GetDirective("DocumentRoot")
	vhost.User = n.GetDirective("User")
	if vhost.User == "" {
		vhost.User = n.GetDirective("suPHP_UserGroup")
		if vhost.User == "" {
			vhost.User = n.GetDirective("SuexecUserGroup")
			if vhost.User == "" {
				vhost.User = n.GetDirective("AssignUserID")
				if vhost.User == "" {
					vhost.User = n.GetDirective("PassengerUser")
				}
			}
		}
	}
	vhost.Group = n.GetDirective("Group")
	if vhost.Group == "" {
		// Try to get group from combined directives
		if suphp := n.GetDirective("suPHP_UserGroup"); suphp != "" {
			parts := strings.Fields(suphp)
			if len(parts) > 1 {
				vhost.Group = parts[1]
			}
		}
		if vhost.Group == "" {
			if suexec := n.GetDirective("SuexecUserGroup"); suexec != "" {
				parts := strings.Fields(suexec)
				if len(parts) > 1 {
					vhost.Group = parts[1]
				}
			}
		}
		if vhost.Group == "" {
			if assign := n.GetDirective("AssignUserID"); assign != "" {
				parts := strings.Fields(assign)
				if len(parts) > 1 {
					vhost.Group = parts[1]
				}
			}
		}
		if vhost.Group == "" {
			vhost.Group = n.GetDirective("PassengerGroup")
		}
	}
	vhost.ServerAdmin = n.GetDirective("ServerAdmin")
	vhost.ErrorLog = n.GetDirective("ErrorLog")
	vhost.CustomLog = n.GetDirective("CustomLog")
	vhost.DirectoryIndex = n.GetDirective("DirectoryIndex")
	vhost.PHPProxyFCGI = n.GetDirective("PHPProxyFCGI")
	vhost.CGIPath = n.GetDirective("CGIPath")
	vhost.ProxyUnixSocket = n.GetDirective("ProxyUnixSocket")
	vhost.ProxyPath = n.GetDirective("ProxyPath")
	vhost.ProxyType = n.GetDirective("ProxyType")

	// Extract locations from Directory and Location blocks
	for _, dirNode := range n.FindDirectories() {
		locations := dirNode.convertToLocations()
		vhost.Locations = append(vhost.Locations, locations...)
	}
	for _, locNode := range n.FindLocations() {
		locations := locNode.convertToLocations()
		vhost.Locations = append(vhost.Locations, locations...)
	}

	// Set default ServerName if empty
	if vhost.ServerName == "" {
		if vhost.DocumentRoot != "" {
			vhost.ServerName = filepath.Base(vhost.DocumentRoot)
		} else {
			vhost.ServerName = "_default_"
		}
	}

	return vhost
}

// convertToLocations converts a Directory/Location node to one or more config.Location
// If there are FilesMatch blocks inside, each FilesMatch creates a separate location
func (n *ConfigNode) convertToLocations() []config.Location {
	var locations []config.Location

	// If there are FilesMatch blocks, create a location for each
	filesMatchNodes := n.FindFilesMatch()
	if len(filesMatchNodes) > 0 {
		for _, filesMatchNode := range filesMatchNodes {
			location := &config.Location{
				Handler:   "static",
				MatchType: "regexCaseInsensitive",
			}

			// Get path from FilesMatch arguments
			if len(filesMatchNode.Arguments) > 0 {
				location.Path = strings.Trim(filesMatchNode.Arguments[0], "\"'")
			}

			// Extract SetHandler from FilesMatch if present
			if handler := filesMatchNode.GetDirective("SetHandler"); handler != "" {
				location = parseSetHandler(handler, location)
			}

			// Also inherit directives from parent Directory
			if proxyPass := n.GetDirective("ProxyPass"); proxyPass != "" {
				location.Handler = "proxy"
				// For ProxyPass, the path is the target URL, not the location path
				// We'll use the location's Path for matching, and ProxyUnixSocket/ProxyType for proxy config
				location.ProxyType = "http"
			}
			if directoryIndex := n.GetDirective("DirectoryIndex"); directoryIndex != "" {
				location.DirectoryIndex = directoryIndex
			}

			locations = append(locations, *location)
		}
	} else {
		// No FilesMatch blocks, create location from Directory/Location itself
		location := &config.Location{
			Handler:   "static",
			MatchType: "prefix",
		}

		// Get path from arguments
		if len(n.Arguments) > 0 {
			location.Path = strings.Trim(n.Arguments[0], "\"'")
		}

		// Set match type based on node type
		switch n.Type {
		case "DirectoryMatch", "LocationMatch":
			location.MatchType = "regex"
		case "Files":
			location.MatchType = "regex"
		default:
			location.MatchType = "prefix"
		}

		// Extract directives from this location block
		if handler := n.GetDirective("SetHandler"); handler != "" {
			location = parseSetHandler(handler, location)
		}
		if proxyPass := n.GetDirective("ProxyPass"); proxyPass != "" {
			location.Handler = "proxy"
			// ProxyPass target URL - for now we'll just set handler and type
			// The actual proxy target would need to be stored elsewhere or parsed differently
			location.ProxyType = "http"
		}
		if proxyPassMatch := n.GetDirective("ProxyPassMatch"); proxyPassMatch != "" {
			location.Handler = "proxy"
			// ProxyPassMatch target URL
			location.ProxyType = "http"
			location.MatchType = "regex"
		}
		if scriptAlias := n.GetDirective("ScriptAlias"); scriptAlias != "" {
			location.Handler = "cgi"
			location.CGIPath = scriptAlias
		}
		if directoryIndex := n.GetDirective("DirectoryIndex"); directoryIndex != "" {
			location.DirectoryIndex = directoryIndex
		}

		locations = append(locations, *location)
	}

	return locations
}

// Helper functions (will be moved from httpd.go)
func extractPort(listenStr string) string {
	// Extract port from "80", "*:80", "0.0.0.0:80", etc.
	parts := strings.Split(listenStr, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return listenStr
}

func parseSetHandler(handler string, location *config.Location) *config.Location {
	// Parse SetHandler format: "proxy:unix:/path|fcgi://localhost/"
	handler = strings.Trim(handler, "\"'")
	
	if strings.HasPrefix(handler, "proxy:unix:") {
		// Format: proxy:unix:/path|fcgi://localhost/
		parts := strings.Split(handler, "|")
		if len(parts) >= 1 {
			unixPart := strings.TrimPrefix(parts[0], "proxy:unix:")
			location.ProxyUnixSocket = unixPart
		}
		if len(parts) >= 2 {
			proxyPart := parts[1]
			if strings.Contains(proxyPart, "fcgi://") {
				location.ProxyType = "fcgi"
			} else {
				location.ProxyType = "http"
			}
		}
		location.Handler = "proxy"
	} else if strings.Contains(handler, "php") {
		location.Handler = "php"
	} else if strings.Contains(handler, "cgi") {
		location.Handler = "cgi"
	}
	
	return location
}
