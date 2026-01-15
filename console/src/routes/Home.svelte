<script lang="ts">
  import { api, type Project } from '../api/client'
  import { href } from '../lib/router.svelte'
  import dayjs from 'dayjs'
  import relativeTime from 'dayjs/plugin/relativeTime'

  dayjs.extend(relativeTime)

  let projects = $state<Project[]>([])
  let loading = $state(true)
  let error = $state('')
  let search = $state('')

  $effect(() => {
    loadProjects()
  })

  async function loadProjects() {
    loading = true
    error = ''
    try {
      projects = await api.getProjects()
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load projects'
    } finally {
      loading = false
    }
  }

  let filteredProjects = $derived(
    projects.filter(p =>
      p.name.toLowerCase().includes(search.toLowerCase())
    )
  )
</script>

<div>
  <!-- Header -->
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="text-2xl font-bold text-slate-900">Projects</h1>
      <p class="text-slate-500 mt-1">Manage your deployed sites</p>
    </div>
    <div class="flex items-center gap-4">
      <input
        type="text"
        bind:value={search}
        placeholder="Search projects..."
        class="px-4 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-cyan-500 focus:border-transparent outline-none w-64"
      />
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan-600"></div>
    </div>
  {:else if error}
    <div class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">
      {error}
      <button onclick={loadProjects} class="ml-2 underline">Retry</button>
    </div>
  {:else if filteredProjects.length === 0}
    <div class="text-center py-12">
      <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-slate-900">No projects yet</h3>
      <p class="text-slate-500 mt-1">Deploy your first site using the CLI</p>
      <div class="mt-4 bg-slate-100 rounded-md p-4 inline-block text-left">
        <code class="text-sm text-slate-700 font-mono">
          sitepod login<br>
          sitepod deploy
        </code>
      </div>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each filteredProjects as project}
        <a
          href={href(`/project/${project.name}`)}
          class="bg-white rounded-md border border-slate-200 p-6 hover:border-cyan-300 hover:shadow-sm transition group"
        >
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-cyan-600/10 border border-cyan-600/20 rounded-md flex items-center justify-center">
                <span class="text-cyan-700 font-semibold text-sm">
                  {project.name.charAt(0).toUpperCase()}
                </span>
              </div>
              <div>
                <h3 class="font-semibold text-slate-900 group-hover:text-cyan-600 transition">
                  {project.name}
                </h3>
                <p class="text-sm text-slate-500">
                  {project.name}.sitepod.dev
                </p>
              </div>
            </div>
          </div>
          <div class="mt-4 flex items-center gap-4 text-sm text-slate-500">
            <span>Updated {dayjs(project.updated_at).fromNow()}</span>
          </div>
        </a>
      {/each}
    </div>
  {/if}
</div>
