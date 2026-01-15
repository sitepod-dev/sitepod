// Simple hash-based router for Svelte 5

interface Route {
  path: string
  params: Record<string, string>
}

function parseRoute(hash: string): Route {
  const path = hash.replace(/^#/, '') || '/'
  const params: Record<string, string> = {}

  // Extract project ID from /project/:id patterns
  const projectMatch = path.match(/^\/project\/([^\/]+)/)
  if (projectMatch) {
    params.id = projectMatch[1]
  }

  return { path, params }
}

class Router {
  init() {
    // Set initial route
    this.handleHashChange()

    // Listen for hash changes
    window.addEventListener('hashchange', () => this.handleHashChange())
  }

  private handleHashChange() {
    const route = parseRoute(window.location.hash)
    currentRoute.path = route.path
    currentRoute.params = route.params
  }

  navigate(path: string) {
    window.location.hash = path
  }
}

export const router = new Router()

// Reactive route state using Svelte 5 runes
export const currentRoute = $state<Route>({
  path: '/',
  params: {}
})

// Helper to create links
export function href(path: string): string {
  return `#${path}`
}
