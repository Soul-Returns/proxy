import { ref, onMounted, onUnmounted } from 'vue'
import { healthApi } from '../api'

const healthStatus = ref({})
let intervalId = null

export function useHealth() {
  const fetchHealth = async () => {
    try {
      const statuses = await healthApi.getStatuses()
      healthStatus.value = {}
      statuses.forEach((s) => {
        healthStatus.value[s.route_id] = s
      })
    } catch (e) {
      console.error('Failed to fetch health status:', e)
    }
  }

  const getHealthClass = (id) => {
    const s = healthStatus.value[id]
    return s ? (s.healthy ? 'status-healthy' : 'status-unhealthy') : 'status-unknown'
  }

  const getHealthText = (id) => {
    const s = healthStatus.value[id]
    return s ? (s.healthy ? 'Healthy' : 'Unhealthy') : 'Checking...'
  }

  const getHealthTooltip = (id) => {
    const s = healthStatus.value[id]
    return s ? (s.healthy ? `OK - ${s.response_time_ms}ms` : s.error_type) : 'Click for details'
  }

  const getHealthDetails = (id) => {
    return healthStatus.value[id] || null
  }

  const startPolling = () => {
    fetchHealth()
    intervalId = setInterval(fetchHealth, 30000)
  }

  const stopPolling = () => {
    if (intervalId) {
      clearInterval(intervalId)
      intervalId = null
    }
  }

  return {
    healthStatus,
    fetchHealth,
    getHealthClass,
    getHealthText,
    getHealthTooltip,
    getHealthDetails,
    startPolling,
    stopPolling,
  }
}
