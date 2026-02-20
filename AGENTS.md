# AGENTS.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

DevProxy is a local reverse proxy with web UI for Docker Compose projects. It automatically manages routing custom domains to Docker containers and optionally syncs routes to the system hosts file via an agent.

**Tech Stack:**
- **Caddy** — Reverse proxy with automatic reload via API
- **Go (Gin)** — API backend (backend/)
- **Vue.js 3** — Web UI (backend/frontend/)
- **SQLite** — Route storage
- **Host Agent** — Cross-platform hosts file manager (agent/)

**Ports:**
- 80: Caddy reverse proxy
- 8090: DevProxy Web UI
- 9099: Host Agent config UI (when agent is running)

## Architecture

### Multi-Stage Build Process
The project uses a multi-stage Docker build (backend/Dockerfile) that:
1. Builds the Vue.js frontend and embeds it into the Go binary using `//go:embed`
2. Cross-compiles the Host Agent for Windows and Linux (stored in container for user download)
3. Builds the Go API backend with CGO enabled for SQLite support

### Backend Structure (backend/)
- **main.go** — Entry point; initializes DB, Caddy service, health checker, and serves embedded frontend
- **internal/database/** — SQLite operations (routes CRUD)
- **internal/services/** — Business logic:
  - `caddy.go`: Generates Caddyfile from enabled routes, reloads Caddy via HTTP API (port 2019)
  - `health.go`: Background health checker for upstream containers
- **internal/handlers/** — Gin HTTP handlers for API endpoints
- **internal/models/** — Data structures

### Host Agent Structure (agent/)
Standalone Go application that runs on the host machine to automatically sync DevProxy routes to the system hosts file.
- **main.go** — Entry point; starts sync engine, GUI server, and system tray (Windows)
- **sync/sync.go** — Polls DevProxy API every 5 seconds, writes enabled routes to hosts file
- **hosts/** — Cross-platform hosts file manipulation with backups
- **config/** — Configuration management
- **gui/** — Local web UI for agent configuration (port 9099)
- **tray/** — System tray icon (Windows only, tray_windows.go vs tray_other.go)
- **autostart/** — Platform-specific autostart functionality

### Data Flow
1. User adds routes via Web UI → API stores in SQLite
2. API generates Caddyfile from enabled routes → Reloads Caddy via HTTP API
3. Host Agent (optional) fetches routes from API → Updates hosts file with backups

## Development Commands

**IMPORTANT**: This project is developed and run within the Docker Compose stack. All Go, Node.js, and other dependencies are available inside the containers. You should NOT need to install Go, Node, or other tools on your host machine.

### Main Project
```bash
# Start DevProxy (builds all components)
docker compose up -d --build

# View logs
docker compose logs -f api
docker compose logs -f caddy

# Restart services
docker compose restart

# Stop and remove
docker compose down

# Rebuild after code changes
docker compose up -d --build

# Execute commands inside the API container
docker compose exec api sh
```

### Backend Development (Go)
Backend development happens inside the Docker container, but you can also run locally if needed.

```bash
# Run commands inside the container (recommended)
docker compose exec api go test ./...
docker compose exec api go build -o devproxy-api

# OR run locally from backend/ directory (requires Go 1.22+ installed)
cd backend
go run main.go
go build -o devproxy-api
go test ./...
go mod tidy
```

### Host Agent Development (Go)
The agent is built inside the Docker container during the build process, but runs on the host machine.

```bash
# Build agent binaries using Docker (recommended - no Go installation required)
# This builds both Windows and Linux binaries inside the container
docker compose build --build-arg VERSION=1.0.0

# Agent binaries will be available at:
# - Download from UI: http://localhost:8090 → Host Agent tab
# - Inside container: docker compose exec api ls /app/agent

# For local development (requires Go 1.22+ installed)
cd agent

# Run locally (requires admin/sudo for hosts file access)
# Windows:
go run . # (Right-click IDE and run as administrator)

# Linux:
sudo go run .

# Build for current platform
go build -o devproxy-agent

# Build with version
go build -ldflags="-X devproxy-agent/version.Current=1.0.0" -o devproxy-agent

# Cross-compile for Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui -X devproxy-agent/version.Current=1.0.0" -o devproxy-agent.exe

# Cross-compile for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X devproxy-agent/version.Current=1.0.0" -o devproxy-agent

# Run tests
go test ./...

# Dependencies
go mod tidy
```

### Frontend Development (Vue.js)
Frontend is built inside the Docker container, but you can also run a dev server locally.

```bash
# Build frontend inside container (recommended)
docker compose build

# For local development with hot reload (requires Node.js installed)
cd backend/frontend

# Install dependencies
npm install

# Dev server with hot reload (accessible at http://localhost:5173)
npm run dev

# Build for production
npm run build

# Output goes to backend/frontend/dist/ (embedded by Go during Docker build)
```

## Docker Network Setup

DevProxy requires the `dev-proxy` Docker network to route to other containers:

```bash
# Create network (one-time setup)
docker network create dev-proxy

# Connect your containers by adding to docker-compose.override.yaml:
services:
  your-service:
    networks:
      - dev-proxy
      - default

networks:
  default:
  dev-proxy:
    external: true
```

## Domain Configuration

DevProxy supports configurable domains for both local and remote deployments:

### Environment Variables
```bash
DOMAIN=localhost:8090  # Base domain (default)
AGENT_PORT=9099        # Agent port (default)
```

### Configuration Flow
1. Set environment variables in `.env` file (see `.env.example`)
2. Backend reads config on startup via `config.Init()` in `internal/config/config.go`
3. Frontend fetches config from `/api/config` endpoint
4. Agent URL is dynamically constructed based on domain setting

### Deployment Scenarios

**Local Development:**
```bash
DOMAIN=localhost:8090
AGENT_PORT=9099
# Agent accessible at: http://localhost:9099
```

**Remote VPS:**
```bash
DOMAIN=proxy.yourdomain.com
AGENT_PORT=9099
# Agent accessible at: http://proxy.yourdomain.com:9099
```

**Custom Local Domain:**
```bash
DOMAIN=proxy.test
AGENT_PORT=9099
# Requires hosts file entry: 127.0.0.1 proxy.test
# Agent accessible at: http://proxy.test:9099
```

### Implementation Details
- **Config Package**: `backend/internal/config/config.go` manages domain configuration
- **Config Service**: `backend/frontend/src/services/config.js` fetches and caches config
- **Dynamic URLs**: All agent connections use `getAgentUrl()` from config service
- **Frontend Components**: UpdateTab.vue and App.vue use dynamic URLs for agent communication
- **IsRemote Detection**: Automatically detects if deployment is local or remote based on domain

## Key Implementation Details

- **Caddyfile Generation**: Routes are dynamically generated by `services/caddy.go` from enabled routes in SQLite. Caddy is reloaded via its HTTP API at `:2019`.
- **Health Checks**: Background goroutine in `services/health.go` pings upstream targets to verify container availability.
- **Applied State**: The API tracks the last successfully applied Caddyfile configuration to detect pending changes in the UI.
- **Hosts File Sync**: The agent uses a marker system to identify its managed entries, creates timestamped backups before each write, and prunes old backups.
- **Embedded Frontend**: The Vue.js frontend is compiled and embedded into the Go binary using `//go:embed`, so the API serves a fully self-contained SPA.
- **Versioning & Updates**: Built-in Updates tab in the UI checks GitHub releases (https://github.com/Soul-Returns/proxy). Version is set at build time via ldflags from the centralized `VERSION` file. Users can select between "release" and "pre-release" update channels.
- **Domain Configuration**: Configurable domains for local/remote deployments. Agent URLs are dynamically constructed from backend config via `/api/config` endpoint.

## Versioning and Updates

### Version Management
Version numbers for both backend and agent are centrally managed in the `VERSION` file at the project root:
```
BACKEND_VERSION=1.0.0
AGENT_VERSION=1.0.0
```

**To update versions:**
1. Edit the `VERSION` file with new version numbers
2. Rebuild: `docker compose build`
3. The build process automatically:
   - Embeds versions into binaries via ldflags
   - Creates versioned agent binaries: `devproxy-agent-v1.0.0.exe` and `devproxy-agent-v1.0.0`
   - Creates non-versioned copies for backwards compatibility

### Update System Architecture
1. **Version Package** (`agent/version/`): Manages version info and GitHub API integration
2. **Update Channels**: Users can choose between "release" (stable) or "pre-release" (beta) updates
3. **Update Check Flow**:
   - Agent GUI endpoint: `localhost:9099/api/updates/check` (POST)
   - Backend API proxies to agent: `/api/agent/updates/check` (POST)
   - Frontend checks for updates every 30 minutes automatically
   - Manual check via "Check for Updates" button in UI
4. **GitHub Integration**: Queries `https://api.github.com/repos/Soul-Returns/proxy/releases` for latest version
5. **Update Instructions**: Release descriptions on GitHub should contain update instructions (markdown formatted)

### Creating Releases
When creating a new release on GitHub:
1. Update the `VERSION` file with new version numbers
2. Create git tag: `v1.0.0` (semantic versioning)
3. Mark as pre-release if applicable (affects update channel filtering)
4. Use `RELEASE_TEMPLATE.md` for release description
5. The Updates tab in the UI will automatically detect the new release

### Updates Tab Features
The Web UI includes a dedicated **Updates** tab with:
- **Backend Updates**: Check for updates from GitHub, select release/pre-release channel
- **Agent Updates**: Compare running version with newest built version
- **Collapsible Instructions**: Always-visible update guides for both components
- **Version Display**: Current version, latest from GitHub, last checked timestamp
- **Status Badges**: Visual indicators for "Update Available" or "Up to Date"
- **Toast Notifications**: "You're using the newest version" when up to date

### Version Detection Flow
- **Backend Version**: Read from `internal/version.Current` (set via ldflags)
- **Agent Built Version**: Read from `/app/VERSION` file inside container
- **Agent Running Version**: Direct fetch from `localhost:9099/api/version` (browser to host)
- **GitHub Releases**: Backend queries GitHub API directly, filters by channel
