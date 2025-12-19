# FastHTTP Admin Panel

React-based admin panel for managing FastHTTP web server configuration.

## Setup

1. Install dependencies:
```bash
npm install
```

2. Start development server:
```bash
npm run dev
```

3. Build for production:
```bash
npm run build
```

## Configuration

The admin panel connects to the FastHTTP admin API (default: http://localhost:8080/api/v1).

Set `VITE_API_URL` environment variable to change the API URL.

## Features

- Dashboard with server statistics
- Virtual host management (CRUD)
- Location block management
- Configuration editor
- Server control (reload, restart)
