<script lang="ts">
  import { api, type Image } from '../api/client'
  import { href } from '../lib/router.svelte'
  import dayjs from 'dayjs'
  import relativeTime from 'dayjs/plugin/relativeTime'

  dayjs.extend(relativeTime)

  interface Props {
    projectId: string
  }
  let { projectId }: Props = $props()

  interface EnvInfo {
    image_id: string
    content_hash: string
    file_count: number
    updated_at: string
  }

  let environments = $state<Record<string, EnvInfo>>({})
  let loading = $state(true)
  let error = $state('')

  $effect(() => {
    loadProject()
  })

  async function loadProject() {
    loading = true
    error = ''
    try {
      // Load both prod and beta
      const [prodData, betaData] = await Promise.all([
        api.getCurrentDeployment(projectId, 'prod').catch(() => null),
        api.getCurrentDeployment(projectId, 'beta').catch(() => null)
      ])

      if (prodData?.environments?.prod) {
        environments.prod = prodData.environments.prod
      }
      if (betaData?.environments?.beta) {
        environments.beta = betaData.environments.beta
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load project'
    } finally {
      loading = false
    }
  }

  function getEnvUrl(env: string): string {
    if (env === 'prod') return `https://${projectId}.sitepod.dev`
    return `https://${projectId}-beta.sitepod.dev`
  }
</script>

<div>
  <!-- Header -->
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="text-2xl font-bold text-slate-900">{projectId}</h1>
      <p class="text-slate-500 mt-1">
        <a href={getEnvUrl('prod')} target="_blank" class="hover:text-cyan-600">
          {projectId}.sitepod.dev
        </a>
      </p>
    </div>
    <div class="flex items-center gap-3">
      <a
        href={href(`/project/${projectId}/images`)}
        class="px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition"
      >
        View Images
      </a>
      <a
        href={href(`/project/${projectId}/settings`)}
        class="px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition"
      >
        Settings
      </a>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan-600"></div>
    </div>
  {:else if error}
    <div class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">
      {error}
      <button onclick={loadProject} class="ml-2 underline">Retry</button>
    </div>
  {:else}
    <!-- Environment cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      {#each ['prod', 'beta'] as env}
        <div class="bg-white rounded-md border border-slate-200 p-6">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full {env === 'prod' ? 'bg-green-500' : 'bg-amber-500'}"></span>
              <h3 class="font-semibold text-slate-900 capitalize">{env}</h3>
            </div>
            <a
              href={getEnvUrl(env)}
              target="_blank"
              class="text-sm text-cyan-600 hover:text-cyan-700"
            >
              Visit →
            </a>
          </div>

          {#if environments[env]}
            <div class="space-y-3">
              <div>
                <span class="text-sm text-slate-500">Image</span>
                <p class="font-mono text-sm text-slate-900 truncate">
                  {environments[env].content_hash.slice(0, 16)}...
                </p>
              </div>
              <div>
                <span class="text-sm text-slate-500">Files</span>
                <p class="text-sm text-slate-900">{environments[env].file_count} files</p>
              </div>
              <div>
                <span class="text-sm text-slate-500">Updated</span>
                <p class="text-sm text-slate-900">
                  {dayjs(environments[env].updated_at).fromNow()}
                </p>
              </div>
            </div>
          {:else}
            <div class="text-center py-6 text-slate-500">
              <p>No deployment yet</p>
              <p class="text-sm mt-1">
                Run <code class="bg-slate-100 px-1 rounded font-mono">sitepod deploy{env === 'prod' ? ' --prod' : ''}</code>
              </p>
            </div>
          {/if}
        </div>
      {/each}
    </div>

    <!-- Quick actions -->
    <div class="mt-8">
      <h3 class="text-lg font-semibold text-slate-900 mb-4">Quick Actions</h3>
      <div class="bg-white rounded-md border border-slate-200 p-6">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <a
            href={href(`/project/${projectId}/images`)}
            class="flex items-center gap-3 p-4 rounded-md border border-slate-200 hover:border-cyan-300 hover:bg-cyan-50 transition"
          >
            <div class="w-10 h-10 bg-cyan-100 rounded-md flex items-center justify-center">
              <svg class="w-5 h-5 text-cyan-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <div>
              <p class="font-medium text-slate-900">Deploy Image</p>
              <p class="text-sm text-slate-500">Deploy a specific version</p>
            </div>
          </a>

          <button
            onclick={() => {/* TODO: rollback modal */}}
            class="flex items-center gap-3 p-4 rounded-md border border-slate-200 hover:border-amber-300 hover:bg-amber-50 transition text-left"
          >
            <div class="w-10 h-10 bg-amber-100 rounded-md flex items-center justify-center">
              <svg class="w-5 h-5 text-amber-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
              </svg>
            </div>
            <div>
              <p class="font-medium text-slate-900">Rollback</p>
              <p class="text-sm text-slate-500">Revert to previous version</p>
            </div>
          </button>

          <a
            href={href(`/project/${projectId}/settings`)}
            class="flex items-center gap-3 p-4 rounded-md border border-slate-200 hover:border-slate-300 hover:bg-slate-50 transition"
          >
            <div class="w-10 h-10 bg-slate-100 rounded-md flex items-center justify-center">
              <svg class="w-5 h-5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <div>
              <p class="font-medium text-slate-900">Settings</p>
              <p class="text-sm text-slate-500">Configure project</p>
            </div>
          </a>
        </div>
      </div>
    </div>
  {/if}
</div>
