<script lang="ts">
  import { href, currentRoute } from './router.svelte'
  import { auth } from './auth.svelte'
  import type { Snippet } from 'svelte'

  interface Props {
    children?: Snippet
  }
  let { children }: Props = $props()

  function isActive(path: string): boolean {
    if (path === '/') return currentRoute.path === '/'
    return currentRoute.path.startsWith(path)
  }
</script>

<div class="min-h-screen bg-slate-50">
  <!-- Header -->
  <header class="bg-white border-b border-slate-200 sticky top-0 z-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center h-16">
        <!-- Logo -->
        <a href={href('/')} class="flex items-center gap-2">
          <img src="/logo-icon.svg" alt="SitePod" class="h-7 w-7" />
          <span class="font-semibold text-slate-900">SitePod</span>
        </a>

        <!-- Nav -->
        <nav class="flex items-center gap-6">
          <a
            href={href('/')}
            class="text-sm font-medium transition-colors {isActive('/') && currentRoute.path === '/' ? 'text-cyan-600' : 'text-slate-600 hover:text-slate-900'}"
          >
            Projects
          </a>
          {#if currentRoute.params.id}
            <span class="text-slate-300">/</span>
            <a
              href={href(`/project/${currentRoute.params.id}`)}
              class="text-sm font-medium transition-colors {isActive('/project/') ? 'text-cyan-600' : 'text-slate-600 hover:text-slate-900'}"
            >
              {currentRoute.params.id}
            </a>
          {/if}
        </nav>

        <!-- User menu -->
        <div class="flex items-center gap-4">
          {#if auth.user?.isAnonymous}
            <span class="text-xs text-slate-500 bg-slate-100 px-2 py-1 rounded">Anonymous</span>
          {:else if auth.user?.email}
            <span class="text-sm text-slate-600">{auth.user.email}</span>
          {/if}
          <button
            onclick={() => auth.logout()}
            class="text-sm text-slate-500 hover:text-slate-700"
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  </header>

  <!-- Main content -->
  <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    {#if children}
      {@render children()}
    {/if}
  </main>
</div>
