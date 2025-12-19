# FastHTTP

> **Lightning-fast web server for PHP applications** ⚡

FastHTTP is a high-performance, modern web server built in Go, designed to handle multiple virtual hosts with PHP-FPM integration. Perfect for developers who want speed, simplicity, and powerful features without the complexity.

## Why FastHTTP?

🚀 **Blazing Fast** - Built with Go for exceptional performance  
🔧 **Easy Configuration** - Simple JSON-based configuration  
🌐 **Multi-Site Ready** - Run multiple virtual hosts effortlessly  
🔒 **Secure by Default** - Built-in rate limiting and security features  
📊 **Admin Panel** - Modern web interface for server management  
🔄 **Migration Tools** - Easy migration from Apache/httpd  

## Quick Start

### Installation

```bash
git clone https://github.com/CloudVisionApps/FastHTTP.git
cd FastHTTP
```

### Basic Configuration

Create a `fasthttp.json` file:

```json
{
  "listen": ["80"],
  "virtualHosts": [
    {
      "serverName": "example.com",
      "documentRoot": "/var/www/example.com",
      "phpProxyFCGI": "127.0.0.1:9000"
    }
  ]
}
```

### Start the Server

```bash
./fasthttp start
```

That's it! Your server is now running. 🎉

## Key Features

- ✅ **PHP-FPM Integration** - Seamless FastCGI support
- ✅ **Multiple Virtual Hosts** - Host unlimited websites
- ✅ **SSL/TLS Support** - Per-virtual-host SSL configuration
- ✅ **Location Blocks** - Advanced path-based routing
- ✅ **CGI & Proxy Support** - Handle any application type
- ✅ **Rate Limiting** - Built-in DDoS protection
- ✅ **Admin API** - REST API + React admin panel
- ✅ **Apache Converter** - Migrate from Apache in seconds

## Basic Usage

```bash
# Start server
./fasthttp start

# Stop server
./fasthttp stop

# Check status
./fasthttp status

# Convert Apache config
./fasthttp convert --from apache --input httpd.conf --output fasthttp.json
```

## Example Configuration

Here's a real-world example showing multiple sites with different features:

```json
{
  "listen": ["80", "443"],
  "rateLimitRequests": 100,
  "rateLimitWindowSeconds": 60,
  "adminEnabled": true,
  "adminPort": "8080",
  "virtualHosts": [
    {
      "serverName": "mysite.com",
      "serverAlias": ["www.mysite.com"],
      "documentRoot": "/var/www/mysite.com",
      "phpProxyFCGI": "127.0.0.1:9000",
      "locations": [
        {
          "path": "/api",
          "matchType": "prefix",
          "handler": "proxy",
          "proxyUnixSocket": "/var/run/api.sock",
          "proxyType": "http"
        }
      ],
      "ssl": {
        "enabled": true,
        "certificateFile": "/etc/ssl/certs/mysite.com.crt",
        "certificateKeyFile": "/etc/ssl/private/mysite.com.key"
      }
    },
    {
      "serverName": "blog.mysite.com",
      "documentRoot": "/var/www/blog",
      "phpProxyFCGI": "127.0.0.1:9001"
    }
  ]
}
```

This configuration demonstrates:
- Multiple virtual hosts on the same server
- SSL/TLS for secure connections
- Location blocks for API proxying
- Rate limiting protection
- Admin panel access

## Admin Panel

Enable the admin panel by adding to your config:

```json
{
  "adminEnabled": true,
  "adminPort": "8080"
}
```

Then access it at `http://localhost:8080` for a beautiful web interface to manage your server.

## Documentation

- 📖 **[Advanced Documentation](ADVANCED.md)** - Complete technical reference
- 🔌 **[Admin API Docs](admin/README.md)** - REST API documentation
- 🔄 **[Parser Docs](parser/README.md)** - Configuration parser details

## Contributing

Contributions are welcome! Feel free to fork, create a branch, and submit a pull request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Built with ❤️ for developers who value speed and simplicity**
