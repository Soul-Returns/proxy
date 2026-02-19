<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal health-modal">
      <div class="modal-header">
        <h2>Health Details</h2>
        <button class="btn btn-icon" @click="$emit('close')">×</button>
      </div>
      <div class="health-details">
        <div class="health-row">
          <span class="health-label">Status</span>
          <span :class="['status-badge', details.healthy ? 'status-healthy' : 'status-unhealthy']">
            {{ details.healthy ? 'Healthy' : 'Unhealthy' }}
          </span>
        </div>
        <div class="health-row">
          <span class="health-label">Target</span>
          <code>{{ details.target }}</code>
        </div>
        <div class="health-row">
          <span class="health-label">DNS Resolved</span>
          <span :class="details.dns_resolved ? 'text-success' : 'text-danger'">
            {{ details.dns_resolved ? 'Yes' : 'No' }}
          </span>
        </div>
        <div v-if="details.resolved_ip" class="health-row">
          <span class="health-label">IP</span>
          <code>{{ details.resolved_ip }}</code>
        </div>
        <div v-if="details.status_code" class="health-row">
          <span class="health-label">HTTP</span>
          <span>{{ details.status_code }}</span>
        </div>
        <div v-if="details.response_time_ms" class="health-row">
          <span class="health-label">Response</span>
          <span>{{ details.response_time_ms }}ms</span>
        </div>
        <div v-if="details.error_type" class="health-row">
          <span class="health-label">Error Type</span>
          <span class="error-type">{{ details.error_type }}</span>
        </div>
        <div v-if="details.error" class="health-row error-row">
          <span class="health-label">Error</span>
          <code class="error-message">{{ details.error }}</code>
        </div>
        <div v-if="details.tip" class="health-tip">
          <strong>💡 Tip:</strong> {{ details.tip }}
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary" @click="$emit('close')">Close</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'HealthModal',
  props: {
    details: {
      type: Object,
      required: true,
    },
  },
  emits: ['close'],
}
</script>
