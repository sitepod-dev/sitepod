<script lang="ts">
  import { auth } from '../lib/auth.svelte'
  import { onMount } from 'svelte'

  let email = $state('')
  let password = $state('')
  let confirmPassword = $state('')
  let error = $state('')
  let message = $state('')
  let loading = $state(false)
  let isDemo = $state(false)
  let configLoaded = $state(false)
  let mode = $state<'login' | 'register'>('login')

  onMount(async () => {
    try {
      const res = await fetch('/api/v1/config')
      if (res.ok) {
        const data = await res.json()
        isDemo = data.is_demo === true
      }
    } catch {
      // Ignore errors
    }
    configLoaded = true
  })

  function fillDemoCredentials() {
    email = 'demo@sitepod.dev'
    password = 'demo123'
  }

  function fillAdminCredentials() {
    email = 'admin@sitepod.local'
    password = 'sitepod123'
  }

  function switchMode(newMode: 'login' | 'register') {
    mode = newMode
    error = ''
    message = ''
    confirmPassword = ''
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault()
    if (!email || !password) return

    if (mode === 'register' && password !== confirmPassword) {
      error = 'Passwords do not match'
      return
    }

    loading = true
    error = ''
    try {
      const result = await auth.login(email, password, mode === 'register' ? 'register' : 'login')
      if (result.success) {
        message = result.message || (mode === 'register' ? 'Account created' : 'Logged in')
      }
    } catch (err) {
      error = err instanceof Error ? err.message : (mode === 'register' ? 'Sign up failed' : 'Sign in failed')
    } finally {
      loading = false
    }
  }
</script>

<div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
  <div class="bg-white rounded-md border border-slate-200 shadow-sm p-8 w-full max-w-md">
    <!-- Logo -->
    <div class="text-center mb-6">
      <img src="/logo-icon.svg" alt="SitePod" class="h-10 w-10 mx-auto mb-4" />
      <h1 class="text-2xl font-bold text-slate-900">SitePod Console</h1>
      <p class="text-slate-500 mt-2">
        {mode === 'login' ? 'Sign in to manage your deployments' : 'Create a new account'}
      </p>
    </div>

    <!-- Mode Toggle -->
    <div class="flex mb-6 bg-slate-100 rounded-lg p-1">
      <button
        type="button"
        onclick={() => switchMode('login')}
        class="flex-1 py-2 px-4 text-sm font-medium rounded-md transition {mode === 'login' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'}"
      >
        Sign In
      </button>
      <button
        type="button"
        onclick={() => switchMode('register')}
        class="flex-1 py-2 px-4 text-sm font-medium rounded-md transition {mode === 'register' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'}"
      >
        Sign Up
      </button>
    </div>

    {#if error}
      <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
        {error}
      </div>
    {/if}

    {#if message}
      <div class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-green-600 text-sm">
        {message}
      </div>
    {/if}

    <!-- Demo credentials hint -->
    {#if configLoaded && isDemo}
      <div class="mb-4 p-4 bg-amber-50 border border-amber-200 rounded-lg">
        <p class="text-sm font-medium text-amber-800 mb-2">Demo Mode</p>
        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            onclick={fillDemoCredentials}
            class="text-xs px-2 py-1 bg-amber-100 text-amber-700 rounded hover:bg-amber-200 transition"
          >
            Fill Demo User
          </button>
          <button
            type="button"
            onclick={fillAdminCredentials}
            class="text-xs px-2 py-1 bg-orange-100 text-orange-700 rounded hover:bg-orange-200 transition"
          >
            Fill Admin
          </button>
        </div>
        <p class="text-xs text-amber-600 mt-2">
          Demo: demo@sitepod.dev / demo123<br>
          Admin: admin@sitepod.local / sitepod123
        </p>
      </div>
    {/if}

    <!-- Email login/register form -->
    <form onsubmit={handleSubmit} class="space-y-4">
      <div>
        <label for="email" class="block text-sm font-medium text-slate-700 mb-1">Email</label>
        <input
          type="email"
          id="email"
          bind:value={email}
          placeholder="you@example.com"
          class="w-full px-4 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-cyan-500 focus:border-transparent outline-none transition"
          disabled={loading}
        />
      </div>
      <div>
        <label for="password" class="block text-sm font-medium text-slate-700 mb-1">Password</label>
        <input
          type="password"
          id="password"
          bind:value={password}
          placeholder="********"
          class="w-full px-4 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-cyan-500 focus:border-transparent outline-none transition"
          disabled={loading}
        />
      </div>

      {#if mode === 'register'}
        <div>
          <label for="confirmPassword" class="block text-sm font-medium text-slate-700 mb-1">Confirm Password</label>
          <input
            type="password"
            id="confirmPassword"
            bind:value={confirmPassword}
            placeholder="********"
            class="w-full px-4 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-cyan-500 focus:border-transparent outline-none transition"
            disabled={loading}
          />
        </div>
      {/if}

      <button
        type="submit"
        disabled={loading || !email || !password || (mode === 'register' && !confirmPassword)}
        class="w-full py-2 px-4 font-medium rounded-md disabled:opacity-50 disabled:cursor-not-allowed transition bg-cyan-600 text-white hover:bg-cyan-700"
      >
        {#if loading}
          {mode === 'register' ? 'Creating account...' : 'Signing in...'}
        {:else}
          {mode === 'register' ? 'Create Account' : 'Sign In'}
        {/if}
      </button>
    </form>

    <p class="mt-6 text-center text-xs text-slate-500">
      {#if mode === 'login'}
        Don't have an account? <button type="button" onclick={() => switchMode('register')} class="text-cyan-600 hover:underline">Sign Up</button>
      {:else}
        Already have an account? <button type="button" onclick={() => switchMode('login')} class="text-cyan-600 hover:underline">Sign in</button>
      {/if}
    </p>
  </div>
</div>
