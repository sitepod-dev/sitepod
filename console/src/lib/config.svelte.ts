// Server configuration (loaded from /api/v1/config)

interface Config {
  domain: string
  isDemo: boolean
  loaded: boolean
}

function createConfig() {
  const state = $state<Config>({
    domain: '',
    isDemo: false,
    loaded: false
  })

  // Load config on startup
  loadConfig()

  async function loadConfig() {
    try {
      const res = await fetch('/api/v1/config')
      if (res.ok) {
        const data = await res.json()
        state.domain = data.domain || ''
        state.isDemo = data.is_demo === true
      }
    } catch {
      // Ignore errors, use defaults
    }
    state.loaded = true
  }

  return {
    get domain() { return state.domain },
    get isDemo() { return state.isDemo },
    get loaded() { return state.loaded },

    // Get project URL
    getProjectUrl(projectName: string, env: 'prod' | 'beta' = 'prod', subdomain?: string): string {
      const protocol = typeof window !== 'undefined' && window.location.protocol === 'http:' ? 'http' : 'https'
      const domain = state.domain || 'sitepod.dev'
      const name = subdomain || projectName
      if (env === 'prod') {
        return `${protocol}://${name}.${domain}`
      }
      return `${protocol}://${name}-beta.${domain}`
    }
  }
}

export const config = createConfig()
