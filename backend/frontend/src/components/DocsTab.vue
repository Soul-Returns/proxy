<template>
  <div class="docs">
    <div class="card">
      <h2>🏗️ How DevProxy Works</h2>
      <div class="docs-section">
        <h3>Architecture</h3>
        <pre class="architecture-diagram">┌──────────────────────────────────────────────┐
│                  DevProxy                    │
│  ┌────────────┐      ┌────────────┐          │
│  │   Caddy    │      │  Go API    │          │
│  │ (port 80)  │◄─────│ (port 8090)│          │
│  │  reverse   │      │  config +  │          │
│  │  proxy     │      │  web UI    │          │
│  └─────┬──────┘      └─────┬──────┘          │
│        │                   │                 │
│        │             ┌─────┴─────┐           │
│        │             │  SQLite   │           │
│        │             │ (routes)  │           │
│        │             └───────────┘           │
└────────┼─────────────────────────────────────┘
         │ dev-proxy network
         ▼
┌────────────────┐  ┌────────────────┐
│ your-app:80    │  │ other-app:80   │
│ (nginx/apache) │  │ (nginx/apache) │
└────────────────┘  └────────────────┘</pre>
      </div>
      <div class="docs-section">
        <h3>Key Components</h3>
        <dl>
          <dt>🔷 Caddy</dt>
          <dd>Lightweight reverse proxy. Routes requests based on Host header to your containers.</dd>
          <dt>🔷 Go API</dt>
          <dd>Manages routes, generates Caddyfile, serves this UI. Stores data in SQLite.</dd>
          <dt>🔷 dev-proxy Network</dt>
          <dd>Docker bridge network allowing Caddy to reach your containers by name.</dd>
        </dl>
      </div>
      <div class="docs-section">
        <h3>Request Flow</h3>
        <ol>
          <li>Browser → <code>http://myapp.test</code></li>
          <li>Hosts file → <code>127.0.0.1</code></li>
          <li>Caddy receives request on port 80</li>
          <li>Matches Host header → finds route</li>
          <li>Proxies to <code>container:port</code></li>
        </ol>
      </div>
    </div>

    <div class="card">
      <h2>📋 Setup Guide</h2>
      <div class="docs-section">
        <h3>Step 1: Connect Your Project</h3>
        <p>Create <code>docker-compose.override.yaml</code>:</p>
        <pre>services:
  nginx:  # your web service
    networks:
      - dev-proxy
      - default

networks:
  default:
  dev-proxy:
    external: true</pre>
        <p class="hint">💡 Keep <code>default</code> network for internal communication (nginx↔php).</p>
      </div>
      <div class="docs-section">
        <h3>Step 2: Find Container Name</h3>
        <pre>cd your-project
docker compose ps</pre>
        <p>Use the NAME column (e.g., <code>myproject-nginx-1</code>).</p>
      </div>
      <div class="docs-section">
        <h3>Step 3: Add Route</h3>
        <ul>
          <li><strong>Domain:</strong> <code>myproject.test</code></li>
          <li><strong>Target:</strong> <code>myproject-nginx-1:80</code></li>
        </ul>
      </div>
      <div class="docs-section">
        <h3>Step 4: Update Hosts File</h3>
        <p><strong>Windows:</strong> <code>C:\Windows\System32\drivers\etc\hosts</code> (run as Admin)</p>
        <p><strong>Linux/Mac:</strong> <code>/etc/hosts</code></p>
        <pre>127.0.0.1    myproject.test</pre>
      </div>
    </div>

    <div class="card">
      <h2>🔧 Troubleshooting</h2>
      <div class="docs-section">
        <h3>❌ DNS Failure / Container Not Found</h3>
        <p><strong>Causes:</strong> Container not on dev-proxy network, not running, or wrong name.</p>
        <pre>docker network inspect dev-proxy</pre>
      </div>
      <div class="docs-section">
        <h3>❌ Connection Refused</h3>
        <p><strong>Causes:</strong> Wrong port, or web server not running inside container.</p>
        <pre>docker compose logs nginx</pre>
      </div>
      <div class="docs-section">
        <h3>❌ ERR_EMPTY_RESPONSE</h3>
        <p><strong>Causes:</strong> Domain not in hosts file, or Caddy not running.</p>
        <pre>docker compose -f proxy/docker-compose.yaml logs caddy</pre>
      </div>
      <div class="docs-section">
        <h3>Useful Commands</h3>
        <pre># Check proxy network
docker network inspect dev-proxy

# Restart proxy
cd proxy && docker compose restart

# View Caddy config
cat proxy/data/Caddyfile</pre>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'DocsTab',
}
</script>
