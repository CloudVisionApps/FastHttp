# FastHTTP Admin API

REST API for managing FastHTTP web server configuration using Fiber framework.

## Setup

1. Install dependencies:
```bash
go mod tidy
```

2. Enable admin API in `fasthttp.json`:
```json
{
  "adminEnabled": true,
  "adminPort": "8080"
}
```

3. Start the server - admin API will start automatically on the configured port.

## API Endpoints

### Health Check
- `GET /api/v1/health` - Check if API is running

### Configuration
- `GET /api/v1/config` - Get current configuration
- `PUT /api/v1/config` - Update entire configuration
- `POST /api/v1/config/reload` - Reload configuration from file

### Virtual Hosts
- `GET /api/v1/virtualhosts` - List all virtual hosts
- `GET /api/v1/virtualhosts/:serverName` - Get specific virtual host
- `POST /api/v1/virtualhosts` - Create new virtual host
- `PUT /api/v1/virtualhosts/:serverName` - Update virtual host
- `DELETE /api/v1/virtualhosts/:serverName` - Delete virtual host

### Locations
- `GET /api/v1/virtualhosts/:serverName/locations` - Get all locations for a virtual host
- `POST /api/v1/virtualhosts/:serverName/locations` - Add location to virtual host
- `PUT /api/v1/virtualhosts/:serverName/locations/:index` - Update location
- `DELETE /api/v1/virtualhosts/:serverName/locations/:index` - Delete location

### Server Control
- `GET /api/v1/server/status` - Get server status
- `POST /api/v1/server/reload` - Reload server configuration
- `POST /api/v1/server/restart` - Restart server

### Statistics
- `GET /api/v1/stats` - Get server statistics

## React Admin Panel

The React admin panel is located in `admin-ui/` directory.

To start the admin panel:
```bash
cd admin-ui
npm install
npm run dev
```

The admin panel will be available at `http://localhost:3000` and will connect to the API at `http://localhost:8080/api/v1` by default.
