# VPS Deployment Guide

This guide shows how to deploy DevProxy on a remote VPS with a custom domain (e.g., `proxy.soulreturns.com`).

## Prerequisites

- Docker and Docker Compose installed on your VPS
- Domain name pointing to your VPS IP
- Ports 80, 8090, and 9099 accessible

## Quick Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Soul-Returns/proxy.git
   cd proxy
   ```

2. **Create network**
   ```bash
   docker network create dev-proxy
   ```

3. **Configure domain**
   ```bash
   cp .env.example .env
   nano .env
   ```
   
   Update with your domain:
   ```bash
   DOMAIN=proxy.yourdomain.com
   AGENT_PORT=9099
   ```

4. **Start DevProxy**
   ```bash
   docker compose up -d --build
   ```

5. **Access the UI**
   - Web UI: `http://proxy.yourdomain.com:8090`
   - Or configure reverse proxy (nginx/caddy) to forward traffic

## Agent Configuration

### Bind Address for Remote Access

By default, the agent GUI binds to `127.0.0.1` (localhost only) for security. To make the agent accessible from the network (e.g., on your VPS), use the `--bind-addr` flag:

```bash
# Bind to all interfaces (accessible from network)
sudo ./devproxy-agent --bind-addr 0.0.0.0

# Default (localhost only, more secure)
sudo ./devproxy-agent
```

**Security Note:** Only use `--bind-addr 0.0.0.0` on trusted networks or behind a firewall. For VPS deployments, consider using a reverse proxy with authentication instead.

The agent will be accessible at `http://proxy.yourdomain.com:9099` when running on the VPS with `--bind-addr 0.0.0.0`.

### Download Agent on Remote Machine

On the machine where you want to run the agent:

**Windows:**
```powershell
# Download from your VPS
Invoke-WebRequest -Uri "http://proxy.yourdomain.com:8090/api/agent/download/windows" -OutFile "devproxy-agent.exe"

# Run as administrator
```

**Linux:**
```bash
# Download
curl -O http://proxy.yourdomain.com:8090/api/agent/download/linux

# Make executable
chmod +x devproxy-agent

# Run on VPS (bind to all interfaces for remote access)
sudo ./devproxy-agent --api-url http://proxy.yourdomain.com:8090 --bind-addr 0.0.0.0

# Or run on local machine (default localhost binding)
sudo ./devproxy-agent --api-url http://proxy.yourdomain.com:8090
```

## Reverse Proxy Configuration (Optional)

To access DevProxy on port 80/443 without port numbers:

### Nginx Example

```nginx
server {
    listen 80;
    server_name proxy.yourdomain.com;

    location / {
        proxy_pass http://localhost:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Caddy Example

```
proxy.yourdomain.com {
    reverse_proxy localhost:8090
}
```

## Updating

When updates are available:

```bash
cd proxy
docker compose down
git pull origin main
docker compose up -d --build
```

Or download the latest release ZIP from GitHub and extract it.

## Firewall Configuration

Ensure these ports are open:

```bash
# Ubuntu/Debian with UFW
sudo ufw allow 80/tcp
sudo ufw allow 8090/tcp
sudo ufw allow 9099/tcp

# Or with iptables
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8090 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 9099 -j ACCEPT
```

## Troubleshooting

### Agent connection issues

If the agent can't connect to the backend:

1. Verify the domain is correctly set in `.env`
2. Check firewall rules allow port 8090
3. Ensure the agent is using the correct API URL: `--api-url http://proxy.yourdomain.com:8090`

### Check logs

```bash
docker compose logs -f api
docker compose logs -f caddy
```

### Restart services

```bash
docker compose restart
```

## Security Considerations

- Consider using HTTPS with a reverse proxy (nginx/caddy with Let's Encrypt)
- Restrict access to ports 8090/9099 if not needed publicly
- Use firewall rules to limit access to trusted IPs
- Keep DevProxy updated via the Updates tab in the Web UI
