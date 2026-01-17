// Authentication state management

interface User {
  id: string
  email: string
  isAnonymous: boolean
  isAdmin: boolean
}

interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
  loading: boolean
}

function createAuth() {
  const state = $state<AuthState>({
    isAuthenticated: false,
    user: null,
    token: null,
    loading: true
  })

  // Check for existing token on load
  const savedToken = localStorage.getItem('sitepod_token')
  if (savedToken) {
    state.token = savedToken
    state.isAuthenticated = true
    // Validate token and fetch user info
    validateToken(savedToken).then(user => {
      if (user) {
        state.user = user
      } else {
        // Token invalid, clear it
        state.token = null
        state.isAuthenticated = false
        localStorage.removeItem('sitepod_token')
      }
      state.loading = false
    })
  } else {
    state.loading = false
  }

  async function validateToken(token: string): Promise<User | null> {
    try {
      const response = await fetch('/api/v1/auth/info', {
        headers: { 'Authorization': `Bearer ${token}` }
      })
      if (!response.ok) return null
      const data = await response.json()
      return {
        id: data.id,
        email: data.email || '',
        isAnonymous: data.is_anonymous || false,
        isAdmin: data.is_admin || false
      }
    } catch {
      return null
    }
  }

  return {
    get isAuthenticated() { return state.isAuthenticated },
    get user() { return state.user },
    get token() { return state.token },
    get loading() { return state.loading },

    async login(email: string, password: string) {
      state.loading = true
      try {
        const response = await fetch('/api/collections/users/auth-with-password', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ identity: email, password })
        })
        if (!response.ok) throw new Error('Login failed')
        const data = await response.json()
        state.token = data.token
        state.user = {
          id: data.record?.id || '',
          email: data.record?.email || email,
          isAnonymous: data.record?.is_anonymous || false,
          isAdmin: false
        }
        state.isAuthenticated = true
        localStorage.setItem('sitepod_token', data.token)
        return { success: true, message: 'Logged in' }
      } finally {
        state.loading = false
      }
    },

    async loginAdmin(email: string, password: string) {
      state.loading = true
      try {
        const response = await fetch('/api/admins/auth-with-password', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ identity: email, password })
        })
        if (!response.ok) throw new Error('Admin login failed')
        const data = await response.json()
        state.token = data.token
        state.user = {
          id: data.admin?.id || '',
          email: data.admin?.email || email,
          isAnonymous: false,
          isAdmin: true
        }
        state.isAuthenticated = true
        localStorage.setItem('sitepod_token', data.token)
        return { success: true, message: 'Admin logged in' }
      } finally {
        state.loading = false
      }
    },

    async loginAnonymous() {
      state.loading = true
      try {
        const response = await fetch('/api/v1/auth/anonymous', {
          method: 'POST'
        })
        if (!response.ok) {
          const error = await response.json().catch(() => ({}))
          throw new Error(error.message || 'Anonymous login failed')
        }
        const data = await response.json()
        state.token = data.token
        state.user = {
          id: data.user_id,
          email: '',
          isAnonymous: true,
          isAdmin: false
        }
        state.isAuthenticated = true
        localStorage.setItem('sitepod_token', data.token)
        return { success: true }
      } finally {
        state.loading = false
      }
    },

    setToken(token: string, user: User) {
      state.token = token
      state.user = user
      state.isAuthenticated = true
      localStorage.setItem('sitepod_token', token)
    },

    logout() {
      state.token = null
      state.user = null
      state.isAuthenticated = false
      localStorage.removeItem('sitepod_token')
    }
  }
}

export const auth = createAuth()
