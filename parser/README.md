# Configuration Parsers

This package provides parsers for reading configuration files from other web servers (Apache/httpd, Nginx, etc.) and converting them to FastHTTP JSON format.

## Architecture

The parser system is designed to be modular and extensible:

- **Parser Interface**: Defines how parsers work
- **Registry**: Manages available parsers
- **Converter**: Transforms parsed configs to FastHTTP format
- **Separate from main app**: All parsing logic is isolated here

## Supported Formats

### Apache httpd.conf

Parses Apache httpd.conf configuration files and converts them to FastHTTP format.

**Supported Directives:**
- `Listen` - Port configuration
- `User` / `Group` - Process user/group
- `ServerName` / `ServerAlias` - Virtual host names
- `DocumentRoot` - Document root directory
- `<VirtualHost>` - Virtual host blocks
- `<Directory>` / `<Location>` - Location blocks
- `ProxyPass` / `ProxyPassMatch` - Proxy configuration
- `ScriptAlias` - CGI configuration
- `DirectoryIndex` - Index files
- `AddType` - MIME types
- `Include` / `IncludeOptional` - Include files

## Usage

### Command Line Tool

```bash
# Convert Apache httpd.conf to FastHTTP JSON
./fasthttp convert --from apache --input /etc/httpd/conf/httpd.conf --output fasthttp.json

# Or use the API
./fasthttp convert --from apache --input /etc/httpd/conf/httpd.conf
```

### Programmatic Usage

```go
import "fasthttp/parser"

// Create registry
registry := parser.NewRegistry()

// Parse Apache config
parsed, err := registry.ParseFile("/etc/httpd/conf/httpd.conf")
if err != nil {
    log.Fatal(err)
}

// Convert to FastHTTP config
converter := parser.NewFastHTTPConverter()
baseConfig := &config.Config{} // or load existing
fastHTTPConfig, err := converter.Convert(parsed, baseConfig)
if err != nil {
    log.Fatal(err)
}

// Save to JSON
configJSON, _ := json.MarshalIndent(fastHTTPConfig, "", "  ")
os.WriteFile("fasthttp.json", configJSON, 0644)
```

## Adding New Parsers

To add support for other web servers (e.g., Nginx):

1. Implement the `Parser` interface:
```go
type MyParser struct{}

func (p *MyParser) CanParse(filePath string) bool {
    // Check if this parser can handle the file
}

func (p *MyParser) Parse(filePath string) (*ParsedConfig, error) {
    // Parse the file and return ParsedConfig
}
```

2. Register it in `registry.go`:
```go
registry.Register(NewMyParser())
```

## Limitations

- Not all Apache directives are supported
- Some complex configurations may need manual adjustment
- PHP-FPM socket paths may need to be configured separately
- SSL/TLS configuration is not converted (configure separately)
