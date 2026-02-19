<template>
  <div v-if="updateInfo && updateInfo.available" class="update-banner">
    <div class="update-icon">⚠️</div>
    <div class="update-content">
      <div class="update-title">
        <strong>Agent Update Available</strong>
        <span class="version-badge">v{{ updateInfo.current_version }} → v{{ updateInfo.latest_version }}</span>
      </div>
      <div class="update-description">
        A new {{ updateInfo.update_channel === 'pre-release' ? 'pre-release' : 'release' }} version is available.
      </div>
    </div>
    <button @click="showInstructions = !showInstructions" class="btn-toggle">
      {{ showInstructions ? 'Hide' : 'Update Now' }}
    </button>
  </div>

  <!-- Expanded Instructions -->
  <div v-if="updateInfo && updateInfo.available && showInstructions" class="update-instructions">
    <h3>📋 Update Instructions</h3>
    <div class="instructions-content" v-html="formattedInstructions"></div>
    <div class="release-info">
      <a :href="updateInfo.release.html_url" target="_blank" rel="noopener noreferrer" class="release-link">
        View Release Notes →
      </a>
    </div>
  </div>
</template>

<script>
export default {
  name: 'UpdateBanner',
  props: {
    updateInfo: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      showInstructions: false,
    }
  },
  computed: {
    formattedInstructions() {
      if (!this.updateInfo || !this.updateInfo.release) return ''
      
      // Convert markdown to basic HTML for instructions
      const body = this.updateInfo.release.body || 'No update instructions provided.'
      
      // Simple markdown conversion
      return body
        .replace(/^### (.+)$/gm, '<h4>$1</h4>')
        .replace(/^## (.+)$/gm, '<h3>$1</h3>')
        .replace(/^# (.+)$/gm, '<h2>$1</h2>')
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
        .replace(/\n\n/g, '<br><br>')
        .replace(/^(\d+)\.\s(.+)$/gm, '<div class="instruction-step"><strong>$1.</strong> $2</div>')
    },
  },
}
</script>

<style scoped>
.update-banner {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, rgba(251, 146, 60, 0.1) 0%, rgba(249, 115, 22, 0.1) 100%);
  border: 1px solid rgba(251, 146, 60, 0.3);
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
}

.update-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.update-content {
  flex: 1;
}

.update-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.25rem;
}

.update-title strong {
  color: var(--text);
}

.version-badge {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  background: rgba(251, 146, 60, 0.2);
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-family: monospace;
  color: rgb(251, 146, 60);
}

.update-description {
  font-size: 0.875rem;
  color: var(--text-muted);
}

.btn-toggle {
  padding: 0.5rem 1rem;
  background: rgb(251, 146, 60);
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}

.btn-toggle:hover {
  background: rgb(249, 115, 22);
}

.update-instructions {
  padding: 1.5rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
}

.update-instructions h3 {
  margin-top: 0;
  margin-bottom: 1rem;
  color: var(--text);
}

.instructions-content {
  color: var(--text);
  line-height: 1.6;
}

.instructions-content h2,
.instructions-content h3,
.instructions-content h4 {
  margin-top: 1rem;
  margin-bottom: 0.5rem;
  color: var(--text);
}

.instructions-content code {
  padding: 0.125rem 0.375rem;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
}

.instructions-content strong {
  color: var(--text);
}

.instruction-step {
  padding: 0.5rem 0;
}

.release-info {
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
