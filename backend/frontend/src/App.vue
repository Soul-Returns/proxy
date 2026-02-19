<template>
  <div class="container">
    <div class="header">
      <div>
        <h1>DevProxy</h1>
        <p class="subtitle">Manage your local development proxy routes</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-icon" @click="exportConfig" title="Export">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
        </button>
        <label class="btn btn-icon" title="Import">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          <input type="file" accept=".json" @change="importConfig" style="display: none;">
        </label>
      </div>
    </div>

    <!-- Add Route Form -->
    <div class="card">
      <h2>Add New Route</h2>
      <form @submit.prevent="addRoute">
        <div class="form-row">
          <div class="form-group">
            <label>Name</label>
            <input type="text" v-model="newRoute.name" placeholder="My Project" required>
          </div>
          <div class="form-group">
            <label>Domain</label>
            <input type="text" v-model="newRoute.domain" placeholder="myproject.test" required>
          </div>
          <div class="form-group">
            <label>Target</label>
            <input type="text" v-model="newRoute.target" placeholder="php-container:80" required>
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

    <!-- Routes Table -->
    <div class="card">
      <h2>Proxy Routes</h2>
      <div v-if="routes.length === 0" class="empty-state">
        <p>No routes configured yet. Add your first route above!</p>
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
            <td>
              <a :href="'http://' + route.domain" target="_blank" class="domain-link">
                {{ route.domain }}
              </a>
            </td>
            <td class="target-text">{{ route.target }}</td>
            <td>
              <button 
                :class="['status-badge', getHealthClass(route.id)]" 
                @click="showHealthDetails(route.id)"
                :title="getHealthTooltip(route.id)"
              >
                <span :class="['status-dot', getHealthClass(route.id)]"></span>
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
                <button class="btn btn-icon btn-sm" @click="editRoute(route)" title="Edit">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
                <button class="btn btn-icon btn-sm" @click="deleteRoute(route)" title="Delete">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Edit Modal -->
    <div v-if="editingRoute" class="modal-overlay" @click.self="editingRoute = null">
      <div class="modal">
        <div class="modal-header">
          <h2>Edit Route</h2>
          <button class="btn btn-icon" @click="editingRoute = null">×</button>
        </div>
        <form @submit.prevent="saveRoute">
          <div class="form-group">
            <label>Name</label>
            <input type="text" v-model="editingRoute.name" required>
          </div>
          <div class="form-group">
            <label>Domain</label>
            <input type="text" v-model="editingRoute.domain" required>
          </div>
          <div class="form-group">
            <label>Target</label>
            <input type="text" v-model="editingRoute.target" required>
          </div>
          <div class="checkbox-group">
            <input type="checkbox" id="edit-enabled" v-model="editingRoute.enabled">
            <label for="edit-enabled" style="margin: 0;">Enabled</label>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-icon" @click="editingRoute = null">Cancel</button>
            <button type="submit" class="btn btn-primary">Save Changes</button>
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
          <div class="health-row">
            <span class="health-label">Status</span>
            <span :class="['status-badge', healthDetails.healthy ? 'status-healthy' : 'status-unhealthy']">
              {{ healthDetails.healthy ? 'Healthy' : 'Unhealthy' }}
            </span>
          </div>
          <div class="health-row">
            <span class="health-label">Target</span>
            <code>{{ healthDetails.target }}</code>
          </div>
          <div class="health-row">
            <span class="health-label">DNS Resolved</span>
            <span :class="healthDetails.dns_resolved ? 'text-success' : 'text-danger'">
              {{ healthDetails.dns_resolved ? 'Yes' : 'No' }}
            </span>
          </div>
          <div v-if="healthDetails.resolved_ip" class="health-row">
            <span class="health-label">Resolved IP</span>
            <code>{{ healthDetails.resolved_ip }}</code>
          </div>
          <div v-if="healthDetails.status_code" class="health-row">
            <span class="health-label">HTTP Status</span>
            <span>{{ healthDetails.status_code }}</span>
          </div>
          <div v-if="healthDetails.response_time_ms" class="health-row">
            <span class="health-label">Response Time</span>
            <span>{{ healthDetails.response_time_ms }}ms</span>
          </div>
          <div v-if="healthDetails.error_type" class="health-row">
            <span class="health-label">Error Type</span>
            <span class="error-type">{{ healthDetails.error_type }}</span>
          </div>
          <div v-if="healthDetails.error" class="health-row error-row">
            <span class="health-label">Error</span>
            <code class="error-message">{{ healthDetails.error }}</code>
          </div>
          <div v-if="healthDetails.tip" class="health-tip">
            <strong>💡 Tip:</strong> {{ healthDetails.tip }}
          </div>
          <div class="health-row">
            <span class="health-label">Last Check</span>
            <span>{{ formatTime(healthDetails.last_check) }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="healthDetails = null">Close</button>
        </div>
      </div>
    </div>

    <!-- Toast -->
    <div v-if="toast.show" :class="['toast', 'toast-' + toast.type]">
      {{ toast.message }}
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      routes: [],
      healthStatus: {},
      newRoute: {
        name: '',
        domain: '',
        target: '',
        enabled: true
      },
      editingRoute: null,
      healthDetails: null,
      toast: {
        show: false,
        message: '',
        type: 'success'
      }
    }
  },
  mounted() {
    this.fetchRoutes()
    this.fetchHealth()
    setInterval(() => this.fetchHealth(), 30000)
  },
  methods: {
    async fetchRoutes() {
      try {
        const res = await fetch('/api/routes')
        this.routes = await res.json()
      } catch (err) {
        this.showToast('Failed to fetch routes', 'error')
      }
    },
    async fetchHealth() {
      try {
        const res = await fetch('/api/health')
        const statuses = await res.json()
        this.healthStatus = {}
        statuses.forEach(s => {
          this.healthStatus[s.route_id] = s
        })
      } catch (err) {
        console.error('Failed to fetch health status', err)
      }
    },
    async addRoute() {
      try {
        const res = await fetch('/api/routes', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(this.newRoute)
        })
        if (!res.ok) throw new Error('Failed to add route')
        this.newRoute = { name: '', domain: '', target: '', enabled: true }
        await this.fetchRoutes()
        this.showToast('Route added successfully', 'success')
      } catch (err) {
        this.showToast(err.message, 'error')
      }
    },
    async toggleRoute(route) {
      try {
        await fetch(`/api/routes/${route.id}/toggle`, { method: 'POST' })
        await this.fetchRoutes()
      } catch (err) {
        this.showToast('Failed to toggle route', 'error')
      }
    },
    editRoute(route) {
      this.editingRoute = { ...route }
    },
    async saveRoute() {
      try {
        const res = await fetch(`/api/routes/${this.editingRoute.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(this.editingRoute)
        })
        if (!res.ok) throw new Error('Failed to update route')
        this.editingRoute = null
        await this.fetchRoutes()
        this.showToast('Route updated successfully', 'success')
      } catch (err) {
        this.showToast(err.message, 'error')
      }
    },
    async deleteRoute(route) {
      if (!confirm(`Delete route "${route.name}"?`)) return
      try {
        await fetch(`/api/routes/${route.id}`, { method: 'DELETE' })
        await this.fetchRoutes()
        this.showToast('Route deleted', 'success')
      } catch (err) {
        this.showToast('Failed to delete route', 'error')
      }
    },
    getHealthClass(routeId) {
      const status = this.healthStatus[routeId]
      if (!status) return 'status-unknown'
      return status.healthy ? 'status-healthy' : 'status-unhealthy'
    },
    getHealthText(routeId) {
      const status = this.healthStatus[routeId]
      if (!status) return 'Unknown'
      return status.healthy ? 'Healthy' : 'Unhealthy'
    },
    getHealthTooltip(routeId) {
      const status = this.healthStatus[routeId]
      if (!status) return 'Click for details'
      if (status.healthy) return `Healthy - ${status.response_time_ms || 0}ms`
      return status.error_type ? `${status.error_type}: Click for details` : 'Click for details'
    },
    showHealthDetails(routeId) {
      const status = this.healthStatus[routeId]
      if (status) {
        this.healthDetails = status
      }
    },
    formatTime(isoString) {
      if (!isoString) return 'Never'
      const date = new Date(isoString)
      return date.toLocaleTimeString()
    },
    async exportConfig() {
      try {
        const res = await fetch('/api/export')
        const data = await res.json()
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = 'devproxy-config.json'
        a.click()
        URL.revokeObjectURL(url)
      } catch (err) {
        this.showToast('Failed to export config', 'error')
      }
    },
    async importConfig(event) {
      const file = event.target.files[0]
      if (!file) return
      try {
        const text = await file.text()
        const data = JSON.parse(text)
        const res = await fetch('/api/import', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data)
        })
        if (!res.ok) throw new Error('Import failed')
        await this.fetchRoutes()
        this.showToast('Config imported successfully', 'success')
      } catch (err) {
        this.showToast(err.message, 'error')
      }
      event.target.value = ''
    },
    showToast(message, type = 'success') {
      this.toast = { show: true, message, type }
      setTimeout(() => { this.toast.show = false }, 3000)
    }
  }
}
</script>
