<script lang="ts">
  import { api, type Project } from '../api/client'
  import { href } from '../lib/router.svelte'

  interface Props {
    projectId: string
  }
  let { projectId }: Props = $props()

  let loading = $state(true)
  let error = $state('')
  let saving = $state(false)
  let saveMessage = $state('')

  // Form state - will be populated from API when we have full project data
  let customDomain = $state('')
  let description = $state('')

  $effect(() => {
    loadSettings()
  })

  async function loadSettings() {
    loading = true
    error = ''
    try {
      // TODO: Load project settings from API
      // For now, just show the form
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load settings'
    } finally {
      loading = false
    }
  }

  async function saveSettings() {
    saving = true
    saveMessage = ''
    try {
      // TODO: Save settings via API
      saveMessage = 'Settings saved successfully'
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save settings'
    } finally {
      saving = false
    }
  }
</script>

<div>
  <!-- Header -->
  <div class="flex items-center justify-between mb-8">
    <div>
      <div class="flex items-center gap-2 text-sm text-slate-500 mb-1">
        <a href={href(`/project/${projectId}`)} class="hover:text-cyan-600">{projectId}</a>
        <span>/</span>
        <span>Settings</span>
      </div>
      <h1 class="text-2xl font-bold text-slate-900">Project Settings</h1>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan-600"></div>
    </div>
  {:else}
    <div class="space-y-6">
      <!-- Project Info -->
      <div class="bg-white rounded-md border border-slate-200 p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">Project Information</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Project Name</label>
            <input
              type="text"
              value={projectId}
              disabled
              class="w-full px-4 py-2 border border-slate-300 rounded-md bg-slate-50 text-slate-500"
            />
            <p class="mt-1 text-sm text-slate-500">Project name cannot be changed</p>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Default URL</label>
            <input
              type="text"
              value="{projectId}.sitepod.dev"
              disabled
              class="w-full px-4 py-2 border border-slate-300 rounded-md bg-slate-50 text-slate-500"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Description</label>
            <textarea
              bind:value={description}
              placeholder="Optional project description"
              rows="3"
              class="w-full px-4 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-cyan-500 focus:border-transparent outline-none"
            ></textarea>
          </div>
        </div>
      </div>

      <!-- Custom Domain -->
      <div class="bg-white rounded-md border border-slate-200 p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">Custom Domain</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Domain</label>
            <input
              type="text"
              bind:value={customDomain}
              placeholder="example.com"
              class="w-full px-4 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-cyan-500 focus:border-transparent outline-none"
            />
            <p class="mt-1 text-sm text-slate-500">
              Add a CNAME record pointing to <code class="bg-slate-100 px-1 rounded font-mono">{projectId}.sitepod.dev</code>
            </p>
          </div>

          {#if customDomain}
            <div class="bg-slate-50 rounded-md p-4">
              <p class="text-sm font-medium text-slate-700 mb-2">DNS Configuration</p>
              <div class="bg-white rounded-md border border-slate-200 p-3 font-mono text-sm">
                <div class="grid grid-cols-3 gap-4">
                  <div>
                    <span class="text-slate-500">Type</span>
                    <p class="text-slate-900">CNAME</p>
                  </div>
                  <div>
                    <span class="text-slate-500">Name</span>
                    <p class="text-slate-900">{customDomain}</p>
                  </div>
                  <div>
                    <span class="text-slate-500">Value</span>
                    <p class="text-slate-900">{projectId}.sitepod.dev</p>
                  </div>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Danger Zone -->
      <div class="bg-white rounded-md border border-red-200 p-6">
        <h2 class="text-lg font-semibold text-red-600 mb-4">Danger Zone</h2>
        <div class="space-y-4">
          <div class="flex items-center justify-between p-4 bg-red-50 rounded-md">
            <div>
              <p class="font-medium text-slate-900">Delete Project</p>
              <p class="text-sm text-slate-500">
                Permanently delete this project and all deployments
              </p>
            </div>
            <button
              class="px-4 py-2 text-sm font-medium text-red-600 border border-red-300 rounded-md hover:bg-red-100 transition"
            >
              Delete Project
            </button>
          </div>
        </div>
      </div>

      <!-- Save Button -->
      {#if saveMessage}
        <div class="p-3 bg-green-50 border border-green-200 rounded-lg text-green-600 text-sm">
          {saveMessage}
        </div>
      {/if}

      {#if error}
        <div class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {error}
        </div>
      {/if}

      <div class="flex justify-end">
        <button
          onclick={saveSettings}
          disabled={saving}
          class="px-6 py-2 bg-cyan-600 text-white font-medium rounded-md hover:bg-cyan-700 disabled:opacity-50 transition"
        >
          {saving ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  {/if}
</div>
