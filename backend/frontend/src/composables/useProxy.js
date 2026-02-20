import { ref, computed } from 'vue'
import { proxyApi, configApi } from '../api'
import { useRoutes } from './useRoutes'
import { useHealth } from './useHealth'
import { useToast } from './useToast'

const appliedState = ref([])
const reloading = ref(false)
const { showToast } = useToast()

export function useProxy() {
  const { routes, fetchRoutes } = useRoutes()
  const { fetchHealth } = useHealth()

  const fetchAppliedState = async () => {
    try {
      appliedState.value = await proxyApi.getAppliedState()
    } catch (e) {
      console.error('Failed to fetch applied state:', e)
    }
  }

  const changedRouteIds = computed(() => {
    const changed = new Set()
    const appliedMap = new Map(appliedState.value.map((r) => [r.id, r]))

    // Check for modified or new routes
    for (const route of routes.value) {
      const applied = appliedMap.get(route.id)
      if (!applied) {
        changed.add(route.id)
      } else if (
        applied.name !== route.name ||
        applied.domain !== route.domain ||
        applied.target !== route.target ||
        applied.enabled !== route.enabled
      ) {
        changed.add(route.id)
      }
    }

    // Check for deleted routes
    const currentIds = new Set(routes.value.map((r) => r.id))
    for (const applied of appliedState.value) {
      if (!currentIds.has(applied.id)) {
        changed.add(applied.id)
      }
    }

    return changed
  })

  const hasUnappliedChanges = computed(() => changedRouteIds.value.size > 0)

  const isRouteChanged = (id) => changedRouteIds.value.has(id)

  const reloadProxy = async () => {
    reloading.value = true
    try {
      const data = await proxyApi.reload()
      // Small delay to ensure backend has updated applied state
      await new Promise(resolve => setTimeout(resolve, 100))
      await fetchAppliedState()
      await fetchRoutes()
      showToast(data.message || 'Proxy reloaded', 'success')
      await fetchHealth()
    } catch (e) {
      showToast('Failed to reload proxy', 'error')
    }
    reloading.value = false
  }

  const exportConfig = async () => {
    try {
      const data = await configApi.export()
      const a = document.createElement('a')
      a.href = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }))
      a.download = 'devproxy-config.json'
      a.click()
    } catch (e) {
      showToast('Failed to export config', 'error')
    }
  }

  const importConfig = async (file) => {
    try {
      const text = await file.text()
      const routes = JSON.parse(text)
      await configApi.import(routes)
      await fetchRoutes()
      showToast('Imported successfully', 'success')
    } catch (e) {
      showToast('Failed to import config', 'error')
    }
  }

  return {
    appliedState,
    reloading,
    hasUnappliedChanges,
    isRouteChanged,
    fetchAppliedState,
    reloadProxy,
    exportConfig,
    importConfig,
  }
}
