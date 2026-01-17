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

  let images = $state<Image[]>([])
  let loading = $state(true)
  let error = $state('')
  let page = $state(1)
  let total = $state(0)
  let deploying = $state<string | null>(null)
  let deployError = $state('')

  const limit = 20

  $effect(() => {
    page
    loadImages()
  })

  async function loadImages() {
    loading = true
    error = ''
    try {
      const data = await api.getImages(projectId, page, limit)
      images = data.images || []
      total = data.total || 0
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load images'
    } finally {
      loading = false
    }
  }

  async function deploy(imageId: string, env: string) {
    if (!confirm(`Deploy this image to ${env}?`)) return

    deploying = `${imageId}-${env}`
    deployError = ''
    try {
      await api.rollback(projectId, env, imageId)
      // Reload to update status
      await loadImages()
    } catch (err) {
      deployError = err instanceof Error ? err.message : 'Deploy failed'
    } finally {
      deploying = null
    }
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }
</script>

<div>
  <!-- Header -->
  <div class="flex items-center justify-between mb-8">
    <div>
      <div class="flex items-center gap-2 text-sm text-slate-500 mb-1">
        <a href={href(`/project/${projectId}`)} class="hover:text-cyan-600">{projectId}</a>
        <span>/</span>
        <span>Images</span>
      </div>
      <h1 class="text-2xl font-bold text-slate-900">Deployment History</h1>
    </div>
  </div>

  {#if deployError}
    <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
      {deployError}
    </div>
  {/if}

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan-600"></div>
    </div>
  {:else if error}
    <div class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">
      {error}
      <button onclick={loadImages} class="ml-2 underline">Retry</button>
    </div>
  {:else if images.length === 0}
    <div class="text-center py-12 bg-white rounded-md border border-slate-200">
      <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-slate-900">No images yet</h3>
      <p class="text-slate-500 mt-1">Deploy your first version using the CLI</p>
    </div>
  {:else}
    <div class="bg-white rounded-md border border-slate-200 overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Image Hash
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Files
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Size
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Created
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Status
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-slate-500 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          {#each images as image, i}
            <tr class="hover:bg-slate-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <code class="text-sm text-slate-900 font-mono">
                  {image.content_hash.slice(0, 16)}...
                </code>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
                {image.file_count} files
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
                {formatSize(image.total_size)}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
                {dayjs(image.created_at).fromNow()}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center gap-2">
                  {#if image.deployed_to?.includes('prod')}
                    <span class="px-2 py-1 text-xs font-medium bg-cyan-600 text-white rounded">
                      prod
                    </span>
                  {/if}
                  {#if image.deployed_to?.includes('beta')}
                    <span class="px-2 py-1 text-xs font-medium bg-slate-700 text-white rounded">
                      beta
                    </span>
                  {/if}
                  {#if !image.deployed_to?.length}
                    <span class="text-sm text-slate-400">—</span>
                  {/if}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    onclick={() => deploy(image.image_id, 'beta')}
                    disabled={deploying !== null}
                    class="px-3 py-1 text-xs font-medium text-amber-700 bg-amber-50 border border-amber-200 rounded hover:bg-amber-100 disabled:opacity-50 transition"
                  >
                    {deploying === `${image.image_id}-beta` ? '...' : 'Deploy to Beta'}
                  </button>
                  <button
                    onclick={() => deploy(image.image_id, 'prod')}
                    disabled={deploying !== null}
                    class="px-3 py-1 text-xs font-medium text-cyan-700 bg-cyan-50 border border-cyan-200 rounded hover:bg-cyan-100 disabled:opacity-50 transition"
                  >
                    {deploying === `${image.image_id}-prod` ? '...' : 'Deploy to Prod'}
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    {#if total > limit}
      <div class="mt-4 flex items-center justify-between">
        <p class="text-sm text-slate-500">
          Showing {(page - 1) * limit + 1} to {Math.min(page * limit, total)} of {total} images
        </p>
        <div class="flex items-center gap-2">
          <button
            onclick={() => page--}
            disabled={page === 1}
            class="px-3 py-1 text-sm border border-slate-300 rounded hover:bg-slate-50 disabled:opacity-50"
          >
            Previous
          </button>
          <button
            onclick={() => page++}
            disabled={page * limit >= total}
            class="px-3 py-1 text-sm border border-slate-300 rounded hover:bg-slate-50 disabled:opacity-50"
          >
            Next
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>
