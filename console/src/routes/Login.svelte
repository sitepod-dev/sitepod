<script lang="ts">
  import { auth } from '../lib/auth.svelte'

  let email = $state('')
  let error = $state('')
  let message = $state('')
  let loading = $state(false)

  async function handleEmailLogin(e: Event) {
    e.preventDefault()
    if (!email) return

    loading = true
    error = ''
    try {
      const result = await auth.login(email)
      if (result.success) {
        message = result.message || 'Check your email'
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Login failed'
    } finally {
      loading = false
    }
  }

  async function handleAnonymousLogin() {
    loading = true
    error = ''
    try {
      await auth.loginAnonymous()
    } catch (err) {
      error = err instanceof Error ? err.message : 'Anonymous login failed'
    } finally {
      loading = false
    }
  }
</script>

<div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
  <div class="bg-white rounded-md border border-slate-200 shadow-sm p-8 w-full max-w-md">
    <!-- Logo -->
    <div class="text-center mb-8">
      <img src="/logo-icon.svg" alt="SitePod" class="h-10 w-10 mx-auto mb-4" />
      <h1 class="text-2xl font-bold text-slate-900">SitePod Console</h1>
      <p class="text-slate-500 mt-2">Sign in to manage your deployments</p>
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

    <!-- Email login -->
    <form onsubmit={handleEmailLogin} class="space-y-4">
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
      <button
        type="submit"
        disabled={loading || !email}
        class="w-full py-2 px-4 bg-cyan-600 text-white font-medium rounded-md hover:bg-cyan-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
      >
        {loading ? 'Sending...' : 'Continue with Email'}
      </button>
    </form>

    <div class="relative my-6">
      <div class="absolute inset-0 flex items-center">
        <div class="w-full border-t border-slate-200"></div>
      </div>
      <div class="relative flex justify-center text-sm">
        <span class="px-2 bg-white text-slate-500">or</span>
      </div>
    </div>

    <!-- Anonymous login -->
    <button
      onclick={handleAnonymousLogin}
      disabled={loading}
      class="w-full py-2 px-4 bg-white text-slate-700 font-medium rounded-md border border-slate-300 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
    >
      Continue as Anonymous (24h limit)
    </button>

    <p class="mt-6 text-center text-xs text-slate-500">
      Anonymous sessions are limited to 24 hours.<br>
      Sign in with email to keep your deployments.
    </p>
  </div>
</div>
