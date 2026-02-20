/**
 * Configuration service for managing runtime config.
 */

let config = null

/**
 * Fetch and cache frontend configuration from backend.
 */
export async function fetchConfig() {
  if (config) {
    return config
  }

  try {
    const response = await fetch('/api/config')
    if (!response.ok) {
      throw new Error('Failed to fetch config')
    }
    config = await response.json()
    console.log('Frontend config loaded:', config)
    return config
  } catch (error) {
    console.error('Failed to fetch config:', error)
    // Fallback to localhost for local development
    config = {
      agentUrl: 'http://localhost:9099',
      domain: 'localhost:8090',
      isRemote: false
    }
    return config
  }
}

/**
 * Get the agent URL (fetches config if not cached).
 */
export async function getAgentUrl() {
  const cfg = await fetchConfig()
  return cfg.agentUrl
}

/**
 * Get the full config (fetches if not cached).
 */
export async function getConfig() {
  return await fetchConfig()
}
