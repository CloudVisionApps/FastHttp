
# FastHTTP

FastHTTP is a high-performance web server designed to handle multiple virtual hosts, each running PHP-FPM. It is built with speed and efficiency in mind, providing a robust solution for serving dynamic websites and applications.

## Features

- **Multiple Virtual Hosts**: Supports the configuration and management of multiple virtual hosts, making it easy to run multiple websites on the same server.
- **PHP-FPM Integration**: Seamlessly integrates with PHP-FPM for optimal performance in serving PHP applications.
- **High Performance**: Optimized for speed, ensuring minimal resource usage while handling large amounts of traffic.
- **Easy Configuration**: Simplified configuration process with intuitive syntax for managing server settings and virtual hosts.

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

### Testing Configuration

Before restarting the server after making changes to the configuration file, you can test if your configuration is valid with:

```bash
./fasthttp configtest
```

This will check the configuration syntax and output any errors or warnings.

## Contributing

Contributions are welcome! If you would like to improve or add features to FastHTTP, feel free to fork the repository, create a branch, and submit a pull request.

## License

FastHTTP is open-source and released under the [MIT License](LICENSE).
```

This version includes more specific usage instructions for starting, stopping, restarting, and reloading the server, as well as configuration testing. The example virtual host configuration and log monitoring tips should help users understand how to configure and monitor the server effectively.
