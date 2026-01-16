// Authentication state management

interface User {
  id: string
  email: string
  isAnonymous: boolean
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
    // TODO: Validate token and fetch user info
    state.loading = false
  } else {
    state.loading = false
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
          isAnonymous: data.record?.is_anonymous || false
        }
        state.isAuthenticated = true
        localStorage.setItem('sitepod_token', data.token)
        return { success: true, message: 'Logged in' }
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
        if (!response.ok) throw new Error('Anonymous login failed')
        const data = await response.json()
        state.token = data.token
        state.user = {
          id: data.user_id,
          email: '',
          isAnonymous: true
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
