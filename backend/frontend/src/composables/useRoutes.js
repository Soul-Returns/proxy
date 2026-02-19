import { ref } from 'vue'
import { routesApi } from '../api'
import { useToast } from './useToast'

const routes = ref([])
const { showToast } = useToast()

export function useRoutes() {
  const fetchRoutes = async () => {
    try {
      routes.value = await routesApi.getAll()
    } catch (e) {
      showToast('Failed to fetch routes', 'error')
    }
  }

  const createRoute = async (route) => {
    try {
      await routesApi.create(route)
      await fetchRoutes()
      showToast('Route added - click Apply Changes to update proxy', 'success')
      return true
    } catch (e) {
      showToast(e.message, 'error')
      return false
    }
  }

  const updateRoute = async (id, route) => {
    try {
      await routesApi.update(id, route)
      await fetchRoutes()
      showToast('Route updated - click Apply Changes to update proxy', 'success')
      return true
    } catch (e) {
      showToast(e.message, 'error')
      return false
    }
  }

  const deleteRoute = async (id, name) => {
    if (!confirm(`Delete "${name}"?`)) return false

    try {
      await routesApi.delete(id)
      await fetchRoutes()
      showToast('Route deleted - click Apply Changes to update proxy', 'success')
      return true
    } catch (e) {
      showToast(e.message, 'error')
      return false
    }
  }

  const toggleRoute = async (id) => {
    try {
      await routesApi.toggle(id)
      await fetchRoutes()
    } catch (e) {
      showToast(e.message, 'error')
    }
  }

  return {
    routes,
    fetchRoutes,
    createRoute,
    updateRoute,
    deleteRoute,
    toggleRoute,
  }
}
