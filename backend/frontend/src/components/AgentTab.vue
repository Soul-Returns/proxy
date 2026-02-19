<template>
  <div class="docs">
    <!-- Update Banner -->
    <UpdateBanner :updateInfo="updateInfo" />

    <!-- Version Status Card -->
    <div class="card" v-if="agentVersion || updateInfo">
      <h2>📦 Version Status</h2>
      <div class="docs-section">
        <div class="version-info">
          <div class="version-row" v-if="agentVersion">
            <span class="label">Current Version:</span>
            <span class="value"><code>v{{ agentVersion }}</code></span>
          </div>
          <div class="version-row" v-if="!agentVersion && !checkingVersion">
            <span class="label">Agent Status:</span>
            <span class="value text-muted">Not running or not accessible</span>
          </div>
          <div class="version-row" v-if="updateInfo && updateInfo.latest_version">
            <span class="label">Latest Version:</span>
            <span class="value"><code>v{{ updateInfo.latest_version }}</code></span>
          </div>
          <div class="version-row" v-if="updateInfo && updateInfo.checked_at">
            <span class="label">Last Checked:</span>
            <span class="value text-muted">{{ formatDate(updateInfo.checked_at) }}</span>
          </div>
          <div class="version-row" v-if="updateInfo && updateInfo.error">
            <span class="label">Check Error:</span>
            <span class="value error-text">{{ updateInfo.error }}</span>
          </div>
        </div>
        <button 
          @click="checkForUpdates" 
          :disabled="checkingUpdates"
          class="btn-check-updates"
        >
          {{ checkingUpdates ? 'Checking...' : '🔄 Check for Updates' }}
        </button>
      </div>
    </div>

    <!-- Overview Card -->
    <div class="card">
      <h2>🖥️ Host Agent</h2>
      <div class="docs-section">
        <p>The <strong>DevProxy Host Agent</strong> automatically manages your system's hosts file,
           eliminating the manual step of adding <code>127.0.0.1 myapp.test</code> entries.</p>
        <div class="agent-features">
          <div><strong>✅ Automatic sync</strong> — Routes are synced to your hosts file in real-time</div>
          <div><strong>✅ Safe backups</strong> — A backup is created before every hosts file change</div>
          <div><strong>✅ System tray</strong> — Runs silently in the background (Windows)</div>
          <div><strong>✅ Autostart</strong> — Optionally starts with your system</div>
          <div><strong>✅ Built locally</strong> — The binary is compiled during the Docker build, no external downloads</div>
        </div>
      </div>
    </div>

    <!-- Download Card -->
    <div class="card">
      <h2>📥 Download</h2>
      <div class="docs-section">
        <p class="hint" style="margin-bottom: 1rem;">
          🔒 These binaries are built locally from source during <code>docker compose build</code>.
          No external downloads are involved — the agent is compiled inside the Docker build pipeline
          from the source code in the <code>agent/</code> directory.
        </p>
        <div class="download-buttons">
          <a :href="'/api/agent/download/windows'" class="download-btn" download>
            <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
              <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.4H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801"/>
            </svg>
            <div>
              <strong>Windows</strong>
              <span>devproxy-agent.exe</span>
            </div>
          </a>
          <a :href="'/api/agent/download/linux'" class="download-btn" download>
            <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
              <path d="M12.504 0c-.155 0-.315.008-.48.021-4.226.333-3.105 4.807-3.17 6.298-.076 1.092-.3 1.953-1.05 3.02-.885 1.051-2.127 2.75-2.716 4.521-.278.832-.41 1.684-.287 2.489a.424.424 0 00-.11.135c-.26.268-.45.6-.663.839-.199.199-.485.267-.797.4-.313.136-.658.269-.864.68-.09.189-.136.394-.132.602 0 .199.027.4.055.536.058.399.116.728.04.97-.249.68-.28 1.145-.106 1.484.174.334.535.47.94.601.81.2 1.91.135 2.774.6.926.466 1.866.67 2.616.47.526-.116.97-.464 1.208-.946.587-.003 1.23-.269 2.26-.334.699-.058 1.574.267 2.577.2.025.134.063.198.114.333l.003.003c.391.778 1.113 1.368 1.884 1.43.585.047 1.042-.245 1.248-.868.025-.078.036-.214.044-.351.258-.09.543-.534 1.134-.946.81-.59 1.77-2.956 1.165-6.186-.024-.21-.039-.421-.075-.63-.046-.27-.113-.543-.2-.82-.085-.28-.19-.571-.3-.862-.11-.29-.207-.602-.322-.917-.346-1.012-.797-2.172-1.5-3.084-.787-1.082-1.085-2.118-1.103-3.188-.003-.254-.002-.493-.022-.71-.017-.21-.05-.386-.107-.512-.07-.15-.135-.25-.255-.296 0-1.478.589-4.89-3.056-6.199A4.63 4.63 0 0012.504 0z"/>
            </svg>
            <div>
              <strong>Linux</strong>
              <span>devproxy-agent</span>
            </div>
          </a>
        </div>
      </div>
    </div>

    <!-- Installation Guide Card -->
    <div class="card">
      <h2>📋 Installation Guide</h2>
      <div class="docs-section">
        <h3>Windows</h3>
        <ol>
          <li>Download <code>devproxy-agent.exe</code> above</li>
          <li>Place it in a permanent location (e.g., <code>C:\Tools\DevProxy\</code>)</li>
          <li><strong>Right-click → Run as administrator</strong> (required for hosts file access)</li>
          <li>The agent appears in your system tray — right-click for options</li>
          <li>Double-click the tray icon to open the configuration panel</li>
        </ol>
        <p class="hint">💡 Enable "Autostart" in the agent config to start automatically on login.</p>
      </div>
      <div class="docs-section">
        <h3>Linux</h3>
        <ol>
          <li>Download <code>devproxy-agent</code> above</li>
          <li>Make it executable and move to a permanent location:</li>
        </ol>
        <pre>chmod +x devproxy-agent
sudo mv devproxy-agent /usr/local/bin/</pre>
        <ol start="3">
          <li>Run with sudo (required for hosts file access):</li>
        </ol>
        <pre>sudo devproxy-agent</pre>
        <p class="hint">💡 For a permanent setup, create a systemd service or enable autostart in the config GUI.</p>
      </div>
      <div class="docs-section">
        <h3>Command Line Options</h3>
        <pre>devproxy-agent [flags]

  --api-url string    DevProxy API URL (default "http://localhost:8090")
  --config-dir string Config directory (default: platform-specific)
  --no-tray           Disable system tray icon
  --version           Print version and exit</pre>
      </div>
    </div>

    <!-- How It Works Card -->
    <div class="card">
      <h2>🔧 How It Works</h2>
      <div class="docs-section">
        <pre class="architecture-diagram">┌─────────────────────┐        ┌──────────────────────┐
│  DevProxy (Docker)  │        │  Your Machine        │
│                     │        │                      │
│  Go API :8090       │◄───────│  devproxy-agent      │
│  Routes database    │ polls  │  ├── Sync engine     │
│  Caddy proxy        │ every  │  ├── Hosts manager   │
│  Web UI             │ 5 sec  │  ├── Config GUI :9099│
│                     │        │  └── System tray     │
└─────────────────────┘        │                      │
                               │  Hosts file:         │
                               │  127.0.0.1 app.test  │
                               └──────────────────────┘</pre>
      </div>
    </div>
  </div>
</template>

<script>
import { agentApi } from '../api'
import UpdateBanner from './UpdateBanner.vue'

export default {
  name: 'AgentTab',
  components: {
    UpdateBanner,
  },
  data() {
    return {
      agentVersion: null,
      updateInfo: null,
      checkingVersion: false,
      checkingUpdates: false,
    }
  },
  mounted() {
    this.fetchVersion()
    this.checkForUpdates()
    // Check for updates every 30 minutes
    this._updateInterval = setInterval(() => this.checkForUpdates(), 30 * 60 * 1000)
  },
  beforeUnmount() {
    if (this._updateInterval) clearInterval(this._updateInterval)
  },
  methods: {
    async fetchVersion() {
      this.checkingVersion = true
      try {
        const data = await agentApi.getVersion()
        this.agentVersion = data.version
      } catch (error) {
        console.warn('Failed to fetch agent version:', error.message)
      } finally {
        this.checkingVersion = false
      }
    },
    async checkForUpdates() {
      this.checkingUpdates = true
      try {
        const data = await agentApi.checkUpdates()
        this.updateInfo = data
      } catch (error) {
        console.warn('Failed to check for updates:', error.message)
      } finally {
        this.checkingUpdates = false
      }
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
  },
}
</script>

<style scoped>
.agent-features {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 1rem;
  font-size: 0.875rem;
}

.download-buttons {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.download-btn {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 0.5rem;
  color: var(--text);
  text-decoration: none;
  transition: all 0.15s;
  min-width: 200px;
}

.download-btn:hover {
  border-color: var(--primary);
  background: rgba(59, 130, 246, 0.1);
  text-decoration: none;
}

.download-btn svg {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  opacity: 0.8;
}

.download-btn strong {
  display: block;
  font-size: 1rem;
}

.download-btn span {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: monospace;
}

.version-info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.version-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.version-row .label {
  font-weight: 500;
  color: var(--text-muted);
  min-width: 120px;
}

.version-row .value {
  color: var(--text);
}

.version-row .value code {
  padding: 0.125rem 0.375rem;
  background: var(--bg-input);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
}

.text-muted {
  color: var(--text-muted);
}

.error-text {
  color: #ef4444;
  font-size: 0.875rem;
}

.btn-check-updates {
  padding: 0.5rem 1rem;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-check-updates:hover:not(:disabled) {
  background: #2563eb;
}

.btn-check-updates:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

</style>
