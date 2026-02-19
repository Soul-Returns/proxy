# DevProxy

A lightweight reverse proxy with a web GUI for managing Docker Compose development environments.

## Features

- **Web GUI** at `http://localhost:8080` for managing proxy routes
- **No Docker socket required** - routes are configured manually via the GUI
- **Health monitoring** - see the status of your backend services
- **Import/Export** - backup and restore your route configuration
- **Hot reload** - changes take effect immediately without restart

## Quick Start

1. **Create the proxy network** (one-time setup):
   ```bash
   docker network create -d bridge dev-proxy
   ```

2. **Start DevProxy**:
   ```bash
   docker compose up -d --build
   ```

3. **Open the GUI** at [http://localhost:8080](http://localhost:8080)

4. **Add your routes** via the web interface

## Configuring Your Projects

Each project needs to:
1. Join the `dev-proxy` network
2. Be registered in the DevProxy GUI

### Example docker-compose.override.yaml

```yaml
services:
  php:
    networks:
      - dev-proxy
      - default

networks:
  default:
  dev-proxy:
    external: true
```

### Adding a Route

In the DevProxy GUI:
- **Name**: My Project
- **Domain**: myproject.test
- **Target**: myproject-php-1:80 (container name and port)

### Finding the Container Name

```bash
docker compose ps
```
Use the container name shown (e.g., `myproject-php-1`) as the target.

## Hosts File Configuration

Add entries to your hosts file for each domain:

**Windows** (`C:\Windows\System32\drivers\etc\hosts`):
```
127.0.0.1    myproject.test
127.0.0.1    another-project.test
```

**Linux/Mac** (`/etc/hosts`):
```
127.0.0.1    myproject.test
127.0.0.1    another-project.test
```

## Ports

| Port | Service |
|------|---------|
| 80   | Caddy reverse proxy |
| 443  | Caddy HTTPS (self-signed) |
| 8080 | DevProxy management GUI |

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   DevProxy                       │
│  ┌─────────────┐  ┌─────────────┐               │
│  │   Caddy     │  │   Go API    │               │
│  │  (port 80)  │←─│  (port 8080)│               │
│  └──────┬──────┘  └──────┬──────┘               │
│         │                │                       │
│         │         ┌──────┴──────┐               │
│         │         │   SQLite    │               │
│         │         └─────────────┘               │
└─────────┼───────────────────────────────────────┘
          │
          ▼
┌─────────────────┐  ┌─────────────────┐
│ project-a:80    │  │ project-b:80    │
│ (dev-proxy net) │  │ (dev-proxy net) │
└─────────────────┘  └─────────────────┘
```

## Data Persistence

Route configurations are stored in `./data/devproxy.db` (SQLite).

The Caddyfile is auto-generated at `./data/Caddyfile` whenever routes change.

## Troubleshooting

### Route not working

1. Ensure the target container is running
2. Verify the container is on the `dev-proxy` network:
   ```bash
   docker network inspect dev-proxy
   ```
3. Check the health status in the DevProxy GUI

### Container not reachable

Make sure the target container exposes the correct port and is on the `dev-proxy` network.

### Caddy not picking up changes

The Caddyfile is regenerated automatically. If changes aren't reflected, try:
```bash
docker compose restart caddy
```

## License

MIT
