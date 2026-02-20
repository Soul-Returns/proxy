/**
 * API client for DevProxy backend.
 */

const BASE_URL = '/api'

/**
 * Helper for making API requests.
 */
async function request(endpoint, options = {}) {
  const response = await fetch(`${BASE_URL}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || 'Request failed')
  }

  return response.json()
}

// Routes API
export const routesApi = {
  getAll: () => request('/routes'),
  getById: (id) => request(`/routes/${id}`),
  create: (route) => request('/routes', { method: 'POST', body: JSON.stringify(route) }),
  update: (id, route) => request(`/routes/${id}`, { method: 'PUT', body: JSON.stringify(route) }),
  delete: (id) => request(`/routes/${id}`, { method: 'DELETE' }),
  toggle: (id) => request(`/routes/${id}/toggle`, { method: 'POST' }),
}

// Health API
export const healthApi = {
  getStatuses: () => request('/health'),
}

// Proxy API
export const proxyApi = {
  reload: () => request('/reload', { method: 'POST' }),
  getAppliedState: () => request('/applied-state'),
}

// Config API
export const configApi = {
  export: () => request('/export'),
  import: (routes) => request('/import', { method: 'POST', body: JSON.stringify(routes) }),
}

// Agent API
export const agentApi = {
  getVersion: () => request('/agent/version'),
  checkUpdates: (channel = null) => {
    const body = channel ? JSON.stringify({ channel }) : undefined
    return request('/agent/updates/check', { method: 'POST', body })
  },
}

// Backend API
export const backendApi = {
  checkUpdates: (channel = null) => {
    const body = channel ? JSON.stringify({ channel }) : undefined
    return request('/updates/check', { method: 'POST', body })
  },
}

export default {
  routes: routesApi,
  health: healthApi,
  proxy: proxyApi,
  config: configApi,
  agent: agentApi,
  backend: backendApi,
}
