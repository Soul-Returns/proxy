# DevProxy

Local reverse proxy with web UI for Docker Compose projects. Automatically manages your hosts file.

## Features

- 🌐 **Reverse Proxy** — Route custom domains to Docker containers
- 🖥️ **Web UI** — Manage routes visually at `localhost:8090`
- ⚡ **Automatic Hosts File** — Optional agent syncs routes to your system's hosts file
- 🔄 **Live Reload** — Changes apply instantly without restarting containers
- 💾 **Persistent Routes** — Stored in SQLite, survives restarts
- 🔍 **Health Checks** — Monitor upstream container status

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

Add to your project's `docker-compose.override.yaml`:

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

Restart: `docker compose up -d`

## Add a Route

1. Find container name: `docker compose ps`
2. In DevProxy UI → Add route:
   - **Domain:** `myapp.test`
   - **Target:** `myproject-nginx-1:80`
3. Access: `http://myapp.test`

### Manual Hosts File (without agent)

If not using the Host Agent, add to your hosts file:
```
127.0.0.1    myapp.test
```
- **Windows:** `C:\Windows\System32\drivers\etc\hosts` (requires admin)
- **Linux/Mac:** `/etc/hosts` (requires sudo)

## Host Agent (Optional)

Automatically syncs routes to your system's hosts file — no manual editing required.

1. In DevProxy UI, go to **Host Agent** tab
2. Download the binary for your OS (built locally, no external downloads)
3. Run it:
   - **Windows:** Right-click → Run as administrator
   - **Linux:** `sudo devproxy-agent`
4. Configure via `localhost:9099` or system tray (Windows)

**Features:**
- ✅ Automatic sync every 5 seconds
- ✅ Safe backups before changes
- ✅ System tray icon (Windows)
- ✅ Autostart on login (optional)

## Ports

| Port | Service |
|------|---------|
| 80   | Caddy proxy |
| 8090 | DevProxy Web UI |
| 9099 | Host Agent config (when running) |

## Troubleshooting

```bash
# Check containers on network
docker network inspect dev-proxy

# View logs
docker compose logs -f caddy
docker compose logs -f api

# Restart
docker compose restart

# Rebuild after updates
docker compose up -d --build
```

## Tech Stack

- **Caddy** — Reverse proxy with automatic reload
- **Go (Gin)** — API backend
- **Vue.js 3** — Web UI
- **SQLite** — Route storage
- **Host Agent** — Cross-platform hosts file manager (Go)
