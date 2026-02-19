<template>
  <div class="card">
    <h2>Proxy Routes</h2>
    <div v-if="routes.length === 0" class="empty-state">
      <p>No routes configured yet.</p>
      <p class="hint">Check the <a href="#" @click.prevent="$emit('showDocs')">Documentation</a> tab for setup instructions.</p>
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
          <td class="name-cell">
            <span class="change-indicator" v-if="isChanged(route.id)"></span>
            {{ route.name }}
          </td>
          <td>
            <a :href="'http://' + route.domain" target="_blank" class="domain-link">
              {{ route.domain }}
            </a>
          </td>
          <td class="target-text">{{ route.target }}</td>
          <td>
            <button
              :class="['status-badge', getHealthClass(route.id)]"
              @click="$emit('showHealth', route.id)"
              :title="getHealthTooltip(route.id)"
            >
              {{ getHealthText(route.id) }}
            </button>
          </td>
          <td>
            <label class="toggle">
              <input type="checkbox" :checked="route.enabled" @change="$emit('toggle', route.id)">
              <span class="toggle-slider"></span>
            </label>
          </td>
          <td>
            <div class="actions">
              <button class="btn btn-icon btn-sm" @click="$emit('edit', route)" title="Edit">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
              </button>
              <button class="btn btn-icon btn-sm" @click="$emit('delete', route)" title="Delete">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
                </svg>
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
export default {
  name: 'RouteTable',
  props: {
    routes: {
      type: Array,
      required: true,
    },
    getHealthClass: {
      type: Function,
      required: true,
    },
    getHealthText: {
      type: Function,
      required: true,
    },
    getHealthTooltip: {
      type: Function,
      required: true,
    },
    isChanged: {
      type: Function,
      required: true,
    },
  },
  emits: ['toggle', 'edit', 'delete', 'showHealth', 'showDocs'],
}
</script>
