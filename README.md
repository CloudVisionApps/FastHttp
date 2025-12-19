
# FastHTTP

FastHTTP is a high-performance web server designed to handle multiple virtual hosts, each running PHP-FPM. It is built with speed and efficiency in mind, providing a robust solution for serving dynamic websites and applications.

## Features

### Core Features

- **Multiple Virtual Hosts**: Supports the configuration and management of multiple virtual hosts, making it easy to run multiple websites on the same server.
- **Server Aliases**: Configure multiple domain names for a single virtual host using `serverAlias`.
- **Multiple Ports**: Listen on multiple ports globally or per virtual host. Supports both privileged (< 1024) and non-privileged ports.
- **User/Group Switching**: Run the server as a specific user and group, with automatic privilege dropping for privileged ports.

### Request Handlers

- **PHP-FPM Integration**: Seamlessly integrates with PHP-FPM via FastCGI for optimal performance in serving PHP applications.
- **CGI Support**: Execute CGI scripts (.cgi, .pl, .py, .sh) or any executable file. Supports user/group switching per virtual host.
- **HTTP/FastCGI Proxy**: Proxy requests to Unix sockets or TCP backends. Supports both HTTP and FastCGI proxy types.
- **Static File Serving**: Serve static files with automatic MIME type detection and custom MIME type configuration.
- **Directory Listings**: Automatic directory listings with customizable templates when no index file is found.

### Location Blocks

- **Path-based Routing**: Configure location blocks with different handlers for specific URL paths.
- **Match Types**: Support for prefix matching, regex matching, and case-insensitive regex matching.
- **Per-location Configuration**: Override directory index, handler type, and proxy settings per location.
- **Priority System**: Regex matches take priority over prefix matches, with longest prefix winning for prefix matches.

### SSL/TLS

- **Per-Virtual-Host SSL**: Configure SSL certificates independently for each virtual host.
- **Certificate Management**: Support for custom certificate and key file paths.

### Security

- **Rate Limiting**: Configurable rate limiting per IP address to prevent abuse.
- **Admin Authentication**: Optional HTTP Basic Authentication for admin API access.
- **IP Whitelist**: Restrict admin API access to specific IP addresses.

### Logging

- **Multiple Log Files**: Configure separate log files for web server, admin API, and errors.
- **Custom Log Formats**: Define named log formats (e.g., "combined", "common") and use them in custom log entries.
- **Multiple Log Entries**: Configure multiple custom log and error log entries per virtual host.
- **Flexible Logging**: Log to files or stdout/stderr.

### Configuration

- **JSON Configuration**: Simple, human-readable JSON configuration format.
- **Configuration Includes**: Include other configuration files recursively with circular dependency detection.
- **Directory Index**: Configure default directory index files globally, per virtual host, or per location.
- **MIME Types**: Custom MIME type configuration for file extensions.

### Admin API

- **REST API**: Full REST API for managing configuration, virtual hosts, and locations.
- **React Admin Panel**: Modern web-based admin interface for managing the server.
- **Server Control**: Start, stop, reload, and restart the server via API.
- **Configuration Management**: Get, update, and reload configuration without downtime.
- **Virtual Host Management**: Create, read, update, and delete virtual hosts via API.
- **Location Management**: Manage location blocks for each virtual host via API.
- **Server Statistics**: Get server status and statistics.

### Migration Tools

- **Apache/httpd Converter**: Convert Apache httpd.conf configuration files to FastHTTP JSON format.
- **Command-line Conversion**: Use `fasthttp convert` command to migrate from Apache.

### Server Management

- **Server Commands**: `start`, `stop`, `status`, `convert` commands for server management.
- **PID File Management**: Automatic PID file creation for process management.
- **Graceful Shutdown**: Clean shutdown handling for all server processes.
- **Configuration Testing**: Validate configuration before starting the server.

### Performance

- **High Performance**: Optimized for speed, ensuring minimal resource usage while handling large amounts of traffic.
- **Concurrent Request Handling**: Efficient handling of multiple concurrent requests.
- **Efficient Routing**: Fast request routing with location block matching.

## Installation

To install FastHTTP, follow these steps:

1. Clone the repository:
   ```bash
   git clone https://github.com/CloudVisionApps/FastHTTP.git
   ```

2. Navigate to the project directory:
   ```bash
   cd FastHTTP
   ```

3. Install the necessary dependencies:
   ```bash
   # Install required packages (e.g., PHP-FPM, etc.)
   sudo apt-get install php-fpm
   ```

4. Configure your virtual hosts by editing the `fasthttp.json` file.

## Usage

### Starting the Server

To start the FastHTTP web server, run the following command:

```bash
./fasthttp start
```

This will start the server and begin serving requests for your configured virtual hosts.

### Stopping the Server

To stop the FastHTTP web server, run the following command:

```bash
./fasthttp stop
```

This will gracefully shut down the server.

### Restarting the Server

To restart the FastHTTP web server (useful after configuration changes), run:

```bash
./fasthttp restart
```

This will stop and then immediately start the server again.

### Reloading the Configuration

To reload the configuration without restarting the entire server, use the following command:

```bash
./fasthttp reload
```

This will apply any changes made to the configuration files without downtime.

### Example Virtual Host Configuration

FastHTTP supports virtual hosts, allowing you to serve multiple websites from a single server. Here’s an example configuration for a virtual host:

```json
{
  "user": "fasthttp",
  "group": "fasthttp",
  "serverAdmin": "root@localhost",
  "listen": ["80", "443"],
  "directoryIndex": "index.html index.php",
  "rateLimitRequests": 100,
  "rateLimitWindowSeconds": 60,
  "adminEnabled": true,
  "adminPort": "8080",
  "adminAuthEnabled": true,
  "adminUsername": "admin",
  "adminPassword": "changeme",
  "adminIPWhitelist": ["127.0.0.1"],
  "logFile": "/var/log/fasthttp/access.log",
  "errorLogFile": "/var/log/fasthttp/error.log",
  "logFormats": [
    {
      "name": "combined",
      "format": "%h %l %u %t \"%r\" %>s %b \"%{Referer}i\" \"%{User-Agent}i\""
    }
  ],
  "mimeTypes": [
    {
      "ext": ".json",
      "type": "application/json"
    }
  ],
  "virtualHosts": [
    {
      "listen": ["80"],
      "serverName": "example.com",
      "serverAlias": ["www.example.com"],
      "documentRoot": "/var/www/example.com",
      "user": "www-data",
      "group": "www-data",
      "directoryIndex": "index.php index.html",
      "phpProxyFCGI": "127.0.0.1:9000",
      "customLog": [
    {
          "path": "/var/log/fasthttp/example.com-access.log",
          "format": "combined"
        }
      ],
      "errorLog": [
        {
          "path": "/var/log/fasthttp/example.com-error.log"
        }
      ],
      "locations": [
        {
          "path": "/api",
          "matchType": "prefix",
          "handler": "proxy",
          "proxyUnixSocket": "/var/run/api.sock",
          "proxyType": "http"
        },
        {
          "path": "/cgi-bin",
          "matchType": "prefix",
          "handler": "cgi",
          "cgiPath": "/cgi-bin"
        },
        {
          "path": "\\.php$",
          "matchType": "regex",
          "handler": "php",
          "phpProxyFCGI": "127.0.0.1:9000"
        }
      ]
    },
    {
      "listen": ["443"],
      "serverName": "example.com",
      "documentRoot": "/var/www/example.com",
      "phpProxyFCGI": "127.0.0.1:9000",
      "ssl": {
        "enabled": true,
        "certificateFile": "/etc/ssl/certs/example.com.crt",
        "certificateKeyFile": "/etc/ssl/private/example.com.key"
      }
    }
  ]
}
```

#### Configuration Options

**Global Settings:**
- **user/group**: User and group to run the server process
- **listen**: Global ports to listen on (applies to all virtual hosts)
- **directoryIndex**: Default directory index files
- **rateLimitRequests/rateLimitWindowSeconds**: Rate limiting configuration
- **adminEnabled/adminPort**: Enable and configure admin API
- **adminAuthEnabled/adminUsername/adminPassword**: Admin authentication
- **adminIPWhitelist**: IP addresses allowed to access admin API
- **logFile/adminLogFile/errorLogFile**: Log file paths
- **logFormats**: Named log format definitions
- **mimeTypes**: Custom MIME type mappings
- **include/includes**: Include other configuration files

**Virtual Host Settings:**
- **listen**: Ports this virtual host listens on (empty = all global ports)
- **serverName**: Primary domain name
- **serverAlias**: Additional domain names
- **documentRoot**: Document root directory
- **user/group**: User and group for this virtual host (for CGI execution)
- **directoryIndex**: Directory index files for this virtual host
- **phpProxyFCGI**: PHP-FPM FastCGI address (TCP)
- **proxyUnixSocket**: Unix socket path for proxy
- **proxyPath**: URL path prefix to proxy
- **proxyType**: Proxy type ("http" or "fcgi")
- **cgiPath**: Path prefix for CGI scripts
- **customLog**: Array of custom log entries with formats
- **errorLog**: Array of error log entries
- **locations**: Array of location blocks
- **ssl**: SSL/TLS configuration
  - **enabled**: Enable SSL
  - **certificateFile**: SSL certificate file path
  - **certificateKeyFile**: SSL private key file path

**Location Block Settings:**
- **path**: Path pattern to match
- **matchType**: Match type ("prefix", "regex", or "regexCaseInsensitive")
- **handler**: Handler type ("proxy", "cgi", "php", or "static")
- **proxyUnixSocket**: Unix socket for proxy handler
- **proxyType**: Proxy type for proxy handler
- **cgiPath**: CGI path for CGI handler
- **phpProxyFCGI**: PHP-FPM address for PHP handler
- **directoryIndex**: Directory index for this location

### Logging

FastHTTP supports flexible logging configuration:

**Log Files:**
- **logFile**: Web server access log (default: stdout)
- **adminLogFile**: Admin API access log (default: stdout)
- **errorLogFile**: Error log (default: stderr)

**Per-Virtual-Host Logging:**
- **customLog**: Multiple custom log entries with named formats
- **errorLog**: Multiple error log entries

**Log Formats:**
Define named log formats and use them in custom log entries:

```json
{
  "logFormats": [
    {
      "name": "combined",
      "format": "%h %l %u %t \"%r\" %>s %b \"%{Referer}i\" \"%{User-Agent}i\""
    },
    {
      "name": "common",
      "format": "%h %l %u %t \"%r\" %>s %b"
    }
  ],
  "virtualHosts": [
    {
      "customLog": [
        {
          "path": "/var/log/fasthttp/access.log",
          "format": "combined"
        }
      ]
    }
  ]
}
```

**Monitor Logs:**
```bash
# Web server access log
tail -f /var/log/fasthttp/access.log

# Error log
tail -f /var/log/fasthttp/error.log

# Admin API log
tail -f /var/log/fasthttp/admin.log
```

### Checking Server Status

To check if the server is running:

```bash
./fasthttp status
```

This will display the current server status and process information.

### Converting Apache Configuration

FastHTTP includes a converter tool to migrate from Apache httpd configuration:

```bash
./fasthttp convert --from apache --input /etc/httpd/conf/httpd.conf --output fasthttp.json
```

Or convert from httpd format:

```bash
./fasthttp convert --from httpd --input /etc/httpd/conf/httpd.conf --output fasthttp.json
```

The converter supports:
- Virtual host blocks
- Location and Directory blocks
- Proxy configuration
- CGI configuration
- Log formats and log entries
- Includes and IncludeOptional directives
- MIME types
- Directory index configuration

Note: After conversion, review and adjust the configuration as needed, especially PHP-FPM socket paths and SSL certificates.

### Admin API

FastHTTP includes a REST API and web-based admin panel for managing the server. To enable it, set `adminEnabled: true` in your configuration.

**Access the Admin Panel:**
- Web UI: `http://localhost:8080` (or your configured admin port)
- API Base URL: `http://localhost:8080/api/v1`

**API Endpoints:**
- `GET /api/v1/health` - Health check
- `GET /api/v1/config` - Get current configuration
- `PUT /api/v1/config` - Update entire configuration
- `POST /api/v1/config/reload` - Reload configuration from file
- `GET /api/v1/virtualhosts` - List all virtual hosts
- `GET /api/v1/virtualhosts/:serverName` - Get specific virtual host
- `POST /api/v1/virtualhosts` - Create new virtual host
- `PUT /api/v1/virtualhosts/:serverName` - Update virtual host
- `DELETE /api/v1/virtualhosts/:serverName` - Delete virtual host
- `GET /api/v1/virtualhosts/:serverName/locations` - Get locations for a virtual host
- `POST /api/v1/virtualhosts/:serverName/locations` - Add location
- `PUT /api/v1/virtualhosts/:serverName/locations/:index` - Update location
- `DELETE /api/v1/virtualhosts/:serverName/locations/:index` - Delete location
- `GET /api/v1/server/status` - Get server status
- `POST /api/v1/server/reload` - Reload server configuration
- `POST /api/v1/server/restart` - Restart server
- `GET /api/v1/stats` - Get server statistics

See `admin/README.md` for detailed API documentation.

## Contributing

Contributions are welcome! If you would like to improve or add features to FastHTTP, feel free to fork the repository, create a branch, and submit a pull request.

## License

FastHTTP is open-source and released under the [MIT License](LICENSE).
