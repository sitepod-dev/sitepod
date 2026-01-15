<script lang="ts">
  import { router, currentRoute } from './lib/router.svelte'
  import Layout from './lib/Layout.svelte'
  import Home from './routes/Home.svelte'
  import Project from './routes/Project.svelte'
  import Images from './routes/Images.svelte'
  import Settings from './routes/Settings.svelte'
  import Login from './routes/Login.svelte'
  import { auth } from './lib/auth.svelte'

  // Initialize router
  router.init()
</script>

{#if !auth.isAuthenticated}
  <Login />
{:else}
  <Layout>
    {#if currentRoute.path === '/'}
      <Home />
    {:else if currentRoute.path.startsWith('/project/') && currentRoute.path.endsWith('/images')}
      <Images projectId={currentRoute.params.id} />
    {:else if currentRoute.path.startsWith('/project/') && currentRoute.path.endsWith('/settings')}
      <Settings projectId={currentRoute.params.id} />
    {:else if currentRoute.path.startsWith('/project/')}
      <Project projectId={currentRoute.params.id} />
    {:else}
      <div class="p-8 text-center text-slate-500">Page not found</div>
    {/if}
  </Layout>
{/if}
