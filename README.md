# DevProxy

Local reverse proxy with web UI for Docker Compose projects. No Docker socket required.

## Quick Start

```bash
# 1. Create proxy network (once)
docker network create dev-proxy

# 2. Start DevProxy
cd proxy
docker compose up -d --build

# 3. Open UI
http://localhost:8090
```

## Connect Your Project

Add `docker-compose.override.yaml` to your project:

```yaml
services:
  nginx:  # your web service
    networks:
      - dev-proxy
      - default

networks:
  default:
  dev-proxy:
    external: true
```

Then restart: `docker compose up -d`

## Add a Route

1. Find container name: `docker compose ps`
2. In DevProxy UI (http://localhost:8090):
   - **Domain:** `myapp.test`
   - **Target:** `myproject-nginx-1:80`
3. Add to hosts file:
   ```
   127.0.0.1    myapp.test
   ```
   - Windows: `C:\Windows\System32\drivers\etc\hosts` (run as Admin)
   - Linux/Mac: `/etc/hosts`

4. Access: `http://myapp.test`

## Ports

| Port | Service |
|------|---------|
| 80   | Caddy (proxy) |
| 8090 | Web UI |

## Troubleshooting

```bash
# Check what's on the network
docker network inspect dev-proxy

# View Caddy logs
docker compose logs caddy

# Restart proxy
docker compose restart
```

## Tech Stack

- **Caddy** - Reverse proxy
- **Go** - API backend
- **Vue.js** - Web UI
- **SQLite** - Route storage
