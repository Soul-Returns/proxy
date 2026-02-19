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
        ⚡ Apply Changes
      </button>
      <button
        class="btn btn-icon"
        @click="$emit('reload')"
        title="Reload Caddy proxy"
        :disabled="reloading"
      >
        🔄
      </button>
      <button class="btn btn-icon" @click="$emit('export')" title="Export configuration">
        ⬇️
      </button>
      <label class="btn btn-icon" title="Import configuration">
        ⬆️
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
