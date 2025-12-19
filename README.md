
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
  "httpPort": "80",
  "httpsPort": "443",
  "virtualHosts": [
    {
      "portType": "http",
      "serverName": "yourdomain.com",
      "documentRoot": "/home/yourdomain/public_html",
      "user": "yourdomain",
      "group": "yourdomain",
      "phpProxyFCGI": "127.0.0.1:9094"
    },
    {
      "portType": "https",
      "serverName": "yourdomain.com",
      "documentRoot": "/home/yourdomain/public_html",
      "user": "yourdomain",
      "group": "yourdomain",
      "phpProxyFCGI": "127.0.0.1:9094",
      "ssl": {
        "enabled": true,
        "certificateFile": "/etc/ssl/certs/yourdomain.crt",
        "certificateKeyFile": "/etc/ssl/private/yourdomain.key"
      }
    }
  ]
}

```

- **serverName**: The domain or IP address for the virtual host.
- **documentRoot**: The directory where the website files are stored.
- **user**: The user under which the website should run.
- **group**: The group under which the website should run.
- **phpProxyFCGI**: The PHP-FPM server address to handle PHP requests.
- **ssl**: Configuration for enabling SSL.
  - **enabled**: Whether SSL is enabled for the virtual host.
  - **certificateFile**: The path to the SSL certificate file.
  - **certificateKeyFile**: The path to the SSL certificate key file.
 
### Access Logs

FastHTTP logs all requests to a log file, which you can find in the `/var/log/fasthttp/` directory by default. You can monitor these logs to troubleshoot and analyze traffic.

```bash
tail -f /var/log/fasthttp/access.log
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
```

This version includes more specific usage instructions for starting, stopping, restarting, and reloading the server, as well as configuration testing. The example virtual host configuration and log monitoring tips should help users understand how to configure and monitor the server effectively.
