<template>
  <div class="container">
    <div class="header">
      <div>
        <h1>DevProxy</h1>
        <p class="subtitle">Manage your local development proxy routes</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-icon" @click="reloadProxy" title="Reload Caddy proxy" :disabled="reloading">🔄</button>
        <button class="btn btn-icon" @click="exportConfig" title="Export configuration">⬇️</button>
        <label class="btn btn-icon" title="Import configuration">
          ⬆️
          <input type="file" accept=".json" @change="importConfig" style="display: none;">
        </label>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="tabs">
      <button :class="['tab', { active: currentTab === 'routes' }]" @click="currentTab = 'routes'">Routes</button>
      <button :class="['tab', { active: currentTab === 'docs' }]" @click="currentTab = 'docs'">Documentation</button>
    </div>

    <!-- Routes Tab -->
    <div v-show="currentTab === 'routes'">
      <div class="card">
        <h2>Add New Route</h2>
        <form @submit.prevent="addRoute">
          <div class="form-row">
            <div class="form-group">
              <label>Name <span class="hint-inline">Display name</span></label>
              <input type="text" v-model="newRoute.name" placeholder="My Symfony App" required>
            </div>
            <div class="form-group">
              <label>Domain <span class="hint-inline">Add to hosts file</span></label>
              <input type="text" v-model="newRoute.domain" placeholder="myapp.test" required>
            </div>
            <div class="form-group">
              <label>Target <span class="hint-inline">container:port</span></label>
              <input type="text" v-model="newRoute.target" placeholder="myapp-nginx-1:80" required>
              <span class="field-hint">💡 Run <code>docker compose ps</code> to find container names</span>
            </div>
          </div>
          <div class="form-row" style="align-items: flex-end;">
            <div class="checkbox-group">
              <input type="checkbox" id="enabled" v-model="newRoute.enabled">
              <label for="enabled" style="margin: 0;">Enable immediately</label>
            </div>
            <button type="submit" class="btn btn-primary">Add Route</button>
          </div>
        </form>
      </div>

      <div class="card">
        <h2>Proxy Routes</h2>
        <div v-if="routes.length === 0" class="empty-state">
          <p>No routes configured yet.</p>
          <p class="hint">Check the <a href="#" @click.prevent="currentTab = 'docs'">Documentation</a> tab for setup instructions.</p>
        </div>
        <table v-else class="routes-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Domain</th>
              <th>Target</th>
              <th>Status</th>
              <th>Enabled</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="route in routes" :key="route.id">
              <td>{{ route.name }}</td>
              <td><a :href="'http://' + route.domain" target="_blank" class="domain-link">{{ route.domain }}</a></td>
              <td class="target-text">{{ route.target }}</td>
              <td>
                <button :class="['status-badge', getHealthClass(route.id)]" @click="showHealthDetails(route.id)" :title="getHealthTooltip(route.id)">
                  {{ getHealthText(route.id) }}
                </button>
              </td>
              <td>
                <label class="toggle">
                  <input type="checkbox" :checked="route.enabled" @change="toggleRoute(route)">
                  <span class="toggle-slider"></span>
                </label>
              </td>
              <td>
                <div class="actions">
                  <button class="btn btn-icon btn-sm" @click="editRoute(route)" title="Edit">✏️</button>
                  <button class="btn btn-icon btn-sm" @click="deleteRoute(route)" title="Delete">🗑️</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Documentation Tab -->
    <div v-show="currentTab === 'docs'" class="docs">
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

    <!-- Edit Modal -->
    <div v-if="editingRoute" class="modal-overlay" @click.self="editingRoute = null">
      <div class="modal">
        <div class="modal-header">
          <h2>Edit Route</h2>
          <button class="btn btn-icon" @click="editingRoute = null">×</button>
        </div>
        <form @submit.prevent="saveRoute">
          <div class="form-group"><label>Name</label><input type="text" v-model="editingRoute.name" required></div>
          <div class="form-group"><label>Domain</label><input type="text" v-model="editingRoute.domain" required></div>
          <div class="form-group"><label>Target</label><input type="text" v-model="editingRoute.target" required></div>
          <div class="checkbox-group">
            <input type="checkbox" id="edit-enabled" v-model="editingRoute.enabled">
            <label for="edit-enabled" style="margin: 0;">Enabled</label>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-icon" @click="editingRoute = null">Cancel</button>
            <button type="submit" class="btn btn-primary">Save</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Health Details Modal -->
    <div v-if="healthDetails" class="modal-overlay" @click.self="healthDetails = null">
      <div class="modal health-modal">
        <div class="modal-header">
          <h2>Health Details</h2>
          <button class="btn btn-icon" @click="healthDetails = null">×</button>
        </div>
        <div class="health-details">
          <div class="health-row"><span class="health-label">Status</span>
            <span :class="['status-badge', healthDetails.healthy ? 'status-healthy' : 'status-unhealthy']">{{ healthDetails.healthy ? 'Healthy' : 'Unhealthy' }}</span>
          </div>
          <div class="health-row"><span class="health-label">Target</span><code>{{ healthDetails.target }}</code></div>
          <div class="health-row"><span class="health-label">DNS Resolved</span>
            <span :class="healthDetails.dns_resolved ? 'text-success' : 'text-danger'">{{ healthDetails.dns_resolved ? 'Yes' : 'No' }}</span>
          </div>
          <div v-if="healthDetails.resolved_ip" class="health-row"><span class="health-label">IP</span><code>{{ healthDetails.resolved_ip }}</code></div>
          <div v-if="healthDetails.status_code" class="health-row"><span class="health-label">HTTP</span><span>{{ healthDetails.status_code }}</span></div>
          <div v-if="healthDetails.response_time_ms" class="health-row"><span class="health-label">Response</span><span>{{ healthDetails.response_time_ms }}ms</span></div>
          <div v-if="healthDetails.error_type" class="health-row"><span class="health-label">Error Type</span><span class="error-type">{{ healthDetails.error_type }}</span></div>
          <div v-if="healthDetails.error" class="health-row error-row"><span class="health-label">Error</span><code class="error-message">{{ healthDetails.error }}</code></div>
          <div v-if="healthDetails.tip" class="health-tip"><strong>💡 Tip:</strong> {{ healthDetails.tip }}</div>
        </div>
        <div class="modal-footer"><button class="btn btn-primary" @click="healthDetails = null">Close</button></div>
      </div>
    </div>

    <!-- Loading Modal -->
    <div v-if="reloading" class="modal-overlay">
      <div class="modal loading-modal">
        <div class="loading-spinner"></div>
        <p>Reloading proxy configuration...</p>
      </div>
    </div>

    <div v-if="toast.show" :class="['toast', 'toast-' + toast.type]">{{ toast.message }}</div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      currentTab: 'routes',
      routes: [],
      healthStatus: {},
      newRoute: { name: '', domain: '', target: '', enabled: true },
      editingRoute: null,
      healthDetails: null,
      reloading: false,
      toast: { show: false, message: '', type: 'success' }
    }
  },
  mounted() {
    this.fetchRoutes()
    this.fetchHealth()
    setInterval(() => this.fetchHealth(), 30000)
  },
  methods: {
    async fetchRoutes() {
      try { this.routes = await (await fetch('/api/routes')).json() }
      catch (e) { this.showToast('Failed to fetch routes', 'error') }
    },
    async fetchHealth() {
      try {
        const statuses = await (await fetch('/api/health')).json()
        this.healthStatus = {}
        statuses.forEach(s => { this.healthStatus[s.route_id] = s })
      } catch (e) { console.error(e) }
    },
    async addRoute() {
      try {
        const res = await fetch('/api/routes', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.newRoute) })
        if (!res.ok) throw new Error('Failed')
        this.newRoute = { name: '', domain: '', target: '', enabled: true }
        await this.fetchRoutes()
        this.showToast('Route added', 'success')
      } catch (e) { this.showToast(e.message, 'error') }
    },
    async toggleRoute(route) { await fetch(`/api/routes/${route.id}/toggle`, { method: 'POST' }); await this.fetchRoutes() },
    editRoute(route) { this.editingRoute = { ...route } },
    async saveRoute() {
      await fetch(`/api/routes/${this.editingRoute.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.editingRoute) })
      this.editingRoute = null
      await this.fetchRoutes()
      this.showToast('Route updated', 'success')
    },
    async deleteRoute(route) {
      if (!confirm(`Delete "${route.name}"?`)) return
      await fetch(`/api/routes/${route.id}`, { method: 'DELETE' })
      await this.fetchRoutes()
      this.showToast('Route deleted', 'success')
    },
    getHealthClass(id) { const s = this.healthStatus[id]; return s ? (s.healthy ? 'status-healthy' : 'status-unhealthy') : 'status-unknown' },
    getHealthText(id) { const s = this.healthStatus[id]; return s ? (s.healthy ? 'Healthy' : 'Unhealthy') : 'Checking...' },
    getHealthTooltip(id) { const s = this.healthStatus[id]; return s ? (s.healthy ? `OK - ${s.response_time_ms}ms` : `${s.error_type}`) : 'Click for details' },
    showHealthDetails(id) { if (this.healthStatus[id]) this.healthDetails = this.healthStatus[id] },
    async exportConfig() {
      const data = await (await fetch('/api/export')).json()
      const a = document.createElement('a')
      a.href = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }))
      a.download = 'devproxy-config.json'
      a.click()
    },
    async importConfig(e) {
      const file = e.target.files[0]; if (!file) return
      await fetch('/api/import', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: await file.text() })
      await this.fetchRoutes()
      this.showToast('Imported', 'success')
      e.target.value = ''
    },
    showToast(msg, type = 'success') { this.toast = { show: true, message: msg, type }; setTimeout(() => { this.toast.show = false }, 3000) },
    async reloadProxy() {
      this.reloading = true
      try {
        const res = await fetch('/api/reload', { method: 'POST' })
        const data = await res.json()
        this.showToast(data.message || 'Proxy reloaded', 'success')
        await this.fetchHealth()
      } catch (e) {
        this.showToast('Failed to reload proxy', 'error')
      }
      this.reloading = false
    }
  }
}
</script>
