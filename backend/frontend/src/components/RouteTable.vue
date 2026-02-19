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
              <button class="btn btn-icon btn-sm" @click="$emit('edit', route)" title="Edit">✏️</button>
              <button class="btn btn-icon btn-sm" @click="$emit('delete', route)" title="Delete">🗑️</button>
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
