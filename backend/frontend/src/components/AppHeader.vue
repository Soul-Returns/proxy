<template>
  <div class="header">
    <div>
      <h1>DevProxy</h1>
      <p class="subtitle">Manage your local development proxy routes</p>
    </div>
    <div class="header-actions">
      <button
        v-if="hasChanges"
        class="btn btn-apply"
        @click="$emit('reload')"
        :disabled="reloading"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
        </svg>
        Apply Changes
      </button>
      <button
        class="btn btn-icon"
        @click="$emit('reload')"
        title="Reload Caddy proxy"
        :disabled="reloading"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M23 4v6h-6M1 20v-6h6"/>
          <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
        </svg>
      </button>
      <button class="btn btn-icon" @click="$emit('export')" title="Export configuration">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/>
        </svg>
      </button>
      <label class="btn btn-icon" title="Import configuration">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
        </svg>
        <input type="file" accept=".json" @change="handleImport" style="display: none;">
      </label>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AppHeader',
  props: {
    hasChanges: Boolean,
    reloading: Boolean,
  },
  emits: ['reload', 'export', 'import'],
  methods: {
    handleImport(e) {
      const file = e.target.files[0]
      if (file) {
        this.$emit('import', file)
        e.target.value = ''
      }
    },
  },
}
</script>
