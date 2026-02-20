<template>
  <div class="updates">
    <!-- Backend Update Card -->
    <div class="card">
      <h2>🚀 DevProxy Backend</h2>
      <div class="docs-section">
        <div class="version-display">
          <div class="version-row">
            <span class="label">Current Version:</span>
            <span class="value">
              <code v-if="backendVersion">v{{ backendVersion }}</code>
              <span v-else class="loading">Loading...</span>
            </span>
          </div>
          <div class="version-row" v-if="backendUpdateInfo && backendUpdateInfo.latest_version">
            <span class="label">Latest from GitHub:</span>
            <span class="value"><code>v{{ backendUpdateInfo.latest_version }}</code></span>
          </div>
          <div class="version-row" v-if="backendUpdateInfo && backendUpdateInfo.checked_at">
            <span class="label">Last Checked:</span>
            <span class="value text-muted">{{ formatDate(backendUpdateInfo.checked_at) }}</span>
          </div>
          <div class="version-row">
            <span class="label">Update Channel:</span>
            <span class="value">
              <select v-model="backendUpdateChannel" @change="onBackendChannelChange" class="channel-select">
                <option value="release">Release (Stable)</option>
                <option value="pre-release">Pre-Release (Beta)</option>
              </select>
            </span>
          </div>
        </div>

        <div class="update-status" v-if="backendUpdateInfo">
          <div v-if="backendUpdateInfo.available" class="status-badge update-available">
            ⚠️ Update Available
          </div>
          <div v-else class="status-badge up-to-date">
            ✅ Up to Date
          </div>
        </div>

        <div class="button-group">
          <button 
            @click="checkBackendUpdates" 
            :disabled="checkingBackend"
            class="btn-secondary"
          >
            {{ checkingBackend ? 'Checking...' : '🔄 Check for Updates' }}
          </button>
          <a
            v-if="backendUpdateInfo && backendUpdateInfo.available && backendUpdateInfo.release"
            :href="backendUpdateInfo.release.html_url"
            target="_blank"
            rel="noopener noreferrer"
            class="btn-primary"
          >
            📦 View Release on GitHub
          </a>
          <button
            v-else-if="backendUpdateInfo && !backendUpdateInfo.available"
            disabled
            class="btn-primary btn-disabled"
            title="You are on the latest version"
          >
            ✅ You're Up to Date
          </button>
        </div>

        <!-- Update Instructions (Always Visible, Collapsible) -->
        <div class="collapsible-section">
          <button @click="showBackendInstructions = !showBackendInstructions" class="collapsible-header">
            <span class="collapsible-icon">{{ showBackendInstructions ? '▼' : '▶' }}</span>
            <span>Update Instructions</span>
          </button>
          <div v-show="showBackendInstructions" class="update-instructions">
          <h3>📋 Update Instructions</h3>
          <div class="instructions-box">
            <p><strong>To update the DevProxy backend, choose one of the following methods:</strong></p>
            
            <h4>Method 1: Git Pull (Recommended)</h4>
            <ol>
              <li>Stop the running containers: <code>docker compose down</code></li>
              <li>Pull the latest changes: <code>git pull origin main</code></li>
              <li>Rebuild and start: <code>docker compose up -d --build</code></li>
            </ol>
            
            <h4>Method 2: Download Release</h4>
            <ol>
              <li>Stop the running containers: <code>docker compose down</code></li>
              <li>Backup your data directory: <code>cp -r data data.backup</code></li>
              <li>Download the release ZIP from GitHub and extract it</li>
              <li>Copy your persistent data back: <code>cp -r data.backup/* data/</code></li>
              <li>Rebuild and start: <code>docker compose up -d --build</code></li>
            </ol>
            
            <p class="note"><strong>Note:</strong> Your routes and settings are stored in the <code>data/</code> directory and will be preserved.</p>
          </div>
            <div class="release-notes" v-if="backendUpdateInfo && backendUpdateInfo.available && backendUpdateInfo.release && backendUpdateInfo.release.body">
              <h4>Release Notes:</h4>
              <div class="notes-content" v-html="formatReleaseNotes(backendUpdateInfo.release.body)"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Agent Update Card -->
    <div class="card">
      <h2>🖥️ Host Agent</h2>
      <div class="docs-section">
        <div class="version-display">
          <div class="version-row">
            <span class="label">Running Version:</span>
            <span class="value">
              <span v-if="agentRunningVersion">
                <code>v{{ agentRunningVersion }}</code>
                <span class="version-hint">(currently running on your machine)</span>
              </span>
              <span v-else class="text-muted">Agent not running</span>
            </span>
          </div>
          <div class="version-row">
            <span class="label">Newest Built:</span>
            <span class="value">
              <code v-if="agentInfo && agentInfo.agent_version">v{{ agentInfo.agent_version }}</code>
              <span v-else class="loading">Loading...</span>
              <span class="version-hint" v-if="agentInfo && agentInfo.agent_version">built with docker compose build</span>
            </span>
          </div>
          <div class="version-row" v-if="agentRunningVersion && agentInfo && agentInfo.agent_version && agentRunningVersion !== agentInfo.agent_version">
            <span class="label"></span>
            <span class="value">
              <span class="update-hint">💡 A newer version is available. Follow update instructions below.</span>
            </span>
          </div>
        </div>

        <div class="update-status" v-if="agentRunningVersion && agentInfo && agentInfo.agent_version">
          <div v-if="agentRunningVersion !== agentInfo.agent_version" class="status-badge update-available">
            ⚠️ Update Available
          </div>
          <div v-else class="status-badge up-to-date">
            ✅ Up to Date
          </div>
        </div>

        <div class="info-box">
          <p><strong>ℹ️ About Agent Updates:</strong></p>
          <p>The Host Agent is built during <code>docker compose build</code> and embedded in the backend container. When you update the backend, a new agent binary is automatically built.</p>
          <p v-if="agentRunningVersion && agentInfo && agentInfo.agent_version && agentRunningVersion !== agentInfo.agent_version">
            You're running <code>v{{ agentRunningVersion }}</code>, but <code>v{{ agentInfo.agent_version }}</code> is available. Download and install the newer version to stay up to date.
          </p>
        </div>

        <!-- Agent Update Instructions (Always Visible, Collapsible) -->
        <div class="collapsible-section">
          <button @click="showAgentInstructions = !showAgentInstructions" class="collapsible-header">
            <span class="collapsible-icon">{{ showAgentInstructions ? '▼' : '▶' }}</span>
            <span>Update Instructions</span>
          </button>
          <div v-show="showAgentInstructions" class="update-instructions">
          <h3>📋 Agent Update Instructions</h3>
          <div class="instructions-box">
            <h4>Windows</h4>
            <ol>
              <li>Stop the running agent (right-click tray icon → Exit)</li>
              <li>Download the latest version from the <a href="#" @click.prevent="goToAgentTab">Host Agent tab</a></li>
              <li>Replace the old executable with the new one</li>
              <li>Run as administrator</li>
            </ol>
            <h4>Linux</h4>
            <ol>
              <li>Stop the agent: <code>sudo killall devproxy-agent</code></li>
              <li>Download latest version: <code>curl -O {{ window.location.origin }}/api/agent/download/linux</code></li>
              <li>Replace: <code>sudo mv devproxy-agent /usr/local/bin/devproxy-agent</code></li>
              <li>Set permissions: <code>sudo chmod +x /usr/local/bin/devproxy-agent</code></li>
              <li>Restart: <code>sudo devproxy-agent</code></li>
            </ol>
          </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { agentApi, backendApi } from '../api'
import { getAgentUrl } from '../services/config'

export default {
  name: 'UpdateTab',
  data() {
    return {
      backendVersion: null,
      agentInfo: null,
      agentRunningVersion: null,
      backendUpdateInfo: null,
      checkingBackend: false,
      showBackendInstructions: false,
      showAgentInstructions: false,
      backendUpdateChannel: 'release',
      agentUrl: '',
    }
  },
  async mounted() {
    this.agentUrl = await getAgentUrl()
    this.fetchVersions()
    this.checkBackendUpdates()
  },
  methods: {
    async fetchVersions() {
      try {
        // Fetch backend version
        const backendResp = await fetch('/api/version')
        const backendData = await backendResp.json()
        this.backendVersion = backendData.version

        // Fetch agent info
        const agentResp = await fetch('/api/agent/info')
        const agentData = await agentResp.json()
        this.agentInfo = agentData

        // Try to get running agent version directly from agent (not through backend)
        try {
          const agentUrl = await getAgentUrl()
          const controller = new AbortController()
          const timeout = setTimeout(() => controller.abort(), 2000)
          const resp = await fetch(`${agentUrl}/api/version`, { signal: controller.signal })
          clearTimeout(timeout)
          if (resp.ok) {
            const runningAgentData = await resp.json()
            this.agentRunningVersion = runningAgentData.version
          }
        } catch {
          // Agent not running
          console.debug('Agent not running or not accessible')
        }
      } catch (error) {
        console.error('Failed to fetch versions:', error)
      }
    },
    async checkBackendUpdates() {
      this.checkingBackend = true
      try {
        // Check updates directly from backend (which checks GitHub) with selected channel
        const data = await backendApi.checkUpdates(this.backendUpdateChannel)
        this.backendUpdateInfo = data
        
        // Show toast if up to date
        if (data && !data.available && !data.error) {
          this.showToast('✅ You\'re using the newest version', 'success')
        }
      } catch (error) {
        console.warn('Failed to check backend updates:', error)
        this.backendUpdateInfo = {
          current_version: this.backendVersion,
          error: 'Failed to check for updates: ' + error.message,
          checked_at: new Date().toISOString(),
        }
      } finally {
        this.checkingBackend = false
      }
    },
    onBackendChannelChange() {
      // Re-check updates with new channel
      this.checkBackendUpdates()
    },
    showToast(message, type = 'info') {
      this.$emit('show-toast', { message, type })
    },
    formatDate(dateString) {
      const date = new Date(dateString)
      const now = new Date()
      const diff = now - date
      const minutes = Math.floor(diff / 60000)
      const hours = Math.floor(diff / 3600000)
      const days = Math.floor(diff / 86400000)

      if (minutes < 1) return 'Just now'
      if (minutes < 60) return `${minutes} minute${minutes !== 1 ? 's' : ''} ago`
      if (hours < 24) return `${hours} hour${hours !== 1 ? 's' : ''} ago`
      if (days < 7) return `${days} day${days !== 1 ? 's' : ''} ago`
      return date.toLocaleDateString()
    },
    formatReleaseNotes(body) {
      if (!body) return ''
      
      // Simple markdown conversion
      return body
        .replace(/^### (.+)$/gm, '<h5>$1</h5>')
        .replace(/^## (.+)$/gm, '<h4>$1</h4>')
        .replace(/^# (.+)$/gm, '<h3>$1</h3>')
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
        .replace(/\n\n/g, '<br><br>')
        .replace(/^(\d+)\.\s(.+)$/gm, '<div class="note-step"><strong>$1.</strong> $2</div>')
    },
    goToAgentTab() {
      this.$emit('switch-tab', 'agent')
    },
  },
}
</script>

<style scoped>
.updates {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.version-display {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding: 1rem;
  background: var(--bg-input);
  border-radius: 0.5rem;
}

.version-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.version-row .label {
  font-weight: 500;
  color: var(--text-muted);
  min-width: 140px;
}

.version-row .value {
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.version-row .value code {
  padding: 0.125rem 0.5rem;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
  color: var(--primary);
}

.version-hint {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-style: italic;
  margin-left: 0.5rem;
}

.channel-select {
  padding: 0.375rem 0.75rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 0.375rem;
  color: var(--text);
  font-size: 0.875rem;
  cursor: pointer;
  transition: border-color 0.15s;
}

.channel-select:hover {
  border-color: var(--primary);
}

.channel-select:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.loading {
  color: var(--text-muted);
  font-style: italic;
}

.text-muted {
  color: var(--text-muted);
}

.update-status {
  margin-bottom: 1rem;
}

.status-badge {
  display: inline-block;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-weight: 500;
  font-size: 0.875rem;
}

.update-available {
  background: linear-gradient(135deg, rgba(251, 146, 60, 0.2) 0%, rgba(249, 115, 22, 0.2) 100%);
  border: 1px solid rgba(251, 146, 60, 0.4);
  color: rgb(249, 115, 22);
}

.up-to-date {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.2) 0%, rgba(22, 163, 74, 0.2) 100%);
  border: 1px solid rgba(34, 197, 94, 0.4);
  color: rgb(22, 163, 74);
}

.button-group {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.btn-primary, .btn-secondary {
  padding: 0.625rem 1.25rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  text-decoration: none;
  display: inline-block;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:hover:not(:disabled):not(.btn-disabled) {
  background: #2563eb;
}

.btn-secondary {
  background: var(--bg-input);
  color: var(--text);
  border: 1px solid var(--border);
}

.btn-secondary:hover:not(:disabled) {
  border-color: var(--primary);
  background: rgba(59, 130, 246, 0.1);
}

.btn-primary:disabled,
.btn-secondary:disabled,
.btn-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.collapsible-section {
  margin-top: 1rem;
}

.collapsible-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.75rem 1rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 0.375rem;
  color: var(--text);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  text-align: left;
}

.collapsible-header:hover {
  border-color: var(--primary);
  background: rgba(59, 130, 246, 0.05);
}

.collapsible-icon {
  font-size: 0.75rem;
  color: var(--text-muted);
  transition: transform 0.15s;
}

.info-box {
  padding: 1rem;
  background: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.5rem;
  margin-top: 1rem;
  color: var(--text);
  line-height: 1.6;
}

.info-box p {
  margin: 0.5rem 0;
}

.info-box code {
  padding: 0.125rem 0.375rem;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
}

.update-hint {
  color: rgb(251, 146, 60);
  font-weight: 500;
  font-size: 0.875rem;
}

.update-instructions {
  margin-top: 1.5rem;
  padding: 1.5rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 0.5rem;
}

.update-instructions h3,
.update-instructions h4 {
  margin-top: 0;
  margin-bottom: 1rem;
  color: var(--text);
}

.update-instructions h4 {
  font-size: 1rem;
  margin-top: 1.5rem;
}

.instructions-box {
  color: var(--text);
  line-height: 1.6;
}

.instructions-box ol {
  margin: 0.5rem 0;
  padding-left: 1.5rem;
}

.instructions-box li {
  margin: 0.5rem 0;
}

.instructions-box .note {
  margin-top: 1rem;
  padding: 0.75rem;
  background: rgba(59, 130, 246, 0.05);
  border-left: 3px solid var(--primary);
  border-radius: 0.25rem;
  font-size: 0.875rem;
}

.instructions-box code {
  padding: 0.125rem 0.375rem;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
}

.release-notes {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border);
}

.notes-content {
  color: var(--text);
  line-height: 1.6;
}

.notes-content code {
  padding: 0.125rem 0.375rem;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
}

.notes-content h3, .notes-content h4, .notes-content h5 {
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}

.note-step {
  padding: 0.5rem 0;
}

.release-link-box {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}

.release-link {
  color: var(--primary);
  text-decoration: none;
  font-weight: 500;
}

.release-link:hover {
  text-decoration: underline;
}
</style>
