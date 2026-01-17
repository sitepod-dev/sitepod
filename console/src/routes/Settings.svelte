<script lang="ts">
  import { api, type Project, type Domain } from '../api/client'
  import { config } from '../lib/config.svelte'
  import { href, router } from '../lib/router.svelte'

  interface Props {
    projectId: string
  }
  let { projectId }: Props = $props()

  let project = $state<Project | null>(null)
  let domains = $state<Domain[]>([])
  let loading = $state(true)
  let error = $state('')

  let addDomainInput = $state('')
  let addingDomain = $state(false)
  let addMessage = $state('')
  let addError = $state('')
  let verificationTxt = $state('')

  let verifyingDomain = $state<string | null>(null)
  let verifyMessage = $state('')
  let verifyError = $state('')

  let removingDomain = $state<string | null>(null)

  let deletingProject = $state(false)
  let deleteError = $state('')

  $effect(() => {
    loadSettings()
  })

  async function loadSettings() {
    loading = true
    error = ''
    try {
      project = await api.getProject(projectId)
      const data = await api.listDomains(projectId)
      domains = data.domains || []
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load settings'
    } finally {
      loading = false
    }
  }

  async function addDomain() {
    const domain = addDomainInput.trim().toLowerCase()
    if (!domain) return

    addingDomain = true
    addMessage = ''
    addError = ''
    verificationTxt = ''

    try {
      const resp = await api.addDomain(projectId, domain, '/')
      addMessage = resp.status === 'active' ? 'Domain active' : 'Domain added'
      if (resp.verification_txt) {
        verificationTxt = resp.verification_txt
      }
      addDomainInput = ''
      const data = await api.listDomains(projectId)
      domains = data.domains || []
    } catch (err) {
      addError = err instanceof Error ? err.message : 'Failed to add domain'
    } finally {
      addingDomain = false
    }
  }

  async function verifyDomain(domain: string) {
    verifyingDomain = domain
    verifyMessage = ''
    verifyError = ''
    try {
      const resp = await api.verifyDomain(domain)
      verifyMessage = resp.message || (resp.verified ? 'Domain verified' : 'Domain not verified')
      const data = await api.listDomains(projectId)
      domains = data.domains || []
    } catch (err) {
      verifyError = err instanceof Error ? err.message : 'Failed to verify domain'
    } finally {
      verifyingDomain = null
    }
  }

  async function removeDomain(domain: string) {
    if (!confirm(`Remove domain ${domain}?`)) return

    removingDomain = domain
    try {
      await api.removeDomain(domain)
      const data = await api.listDomains(projectId)
      domains = data.domains || []
    } catch (err) {
      verifyError = err instanceof Error ? err.message : 'Failed to remove domain'
    } finally {
      removingDomain = null
    }
  }

  async function deleteProject() {
    if (!confirm(`Delete project ${projectId}? This cannot be undone.`)) return

    deletingProject = true
    deleteError = ''
    try {
      await api.deleteProject(projectId)
      router.navigate('/')
    } catch (err) {
      deleteError = err instanceof Error ? err.message : 'Failed to delete project'
    } finally {
      deletingProject = false
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
              value={project?.name || projectId}
              disabled
              class="w-full px-4 py-2 border border-slate-300 rounded-md bg-slate-50 text-slate-500"
            />
            <p class="mt-1 text-sm text-slate-500">Project name cannot be changed</p>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Default URL</label>
            <input
              type="text"
              value="{project?.subdomain || projectId}.{config.domain}"
              disabled
              class="w-full px-4 py-2 border border-slate-300 rounded-md bg-slate-50 text-slate-500"
            />
          </div>
        </div>
      </div>

      <!-- Custom Domain -->
      <div class="bg-white rounded-md border border-slate-200 p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">Custom Domains</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Domain</label>
            <div class="flex flex-col gap-2 sm:flex-row">
              <input
                type="text"
                bind:value={addDomainInput}
                placeholder="example.com"
                class="w-full px-4 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-cyan-500 focus:border-transparent outline-none"
                disabled={addingDomain}
              />
              <button
                type="button"
                onclick={addDomain}
                disabled={addingDomain || !addDomainInput}
                class="px-4 py-2 bg-cyan-600 text-white font-medium rounded-md hover:bg-cyan-700 disabled:opacity-50"
              >
                {addingDomain ? 'Adding...' : 'Add Domain'}
              </button>
            </div>
            <p class="mt-1 text-sm text-slate-500">
              Point DNS to <code class="bg-slate-100 px-1 rounded font-mono">{project?.subdomain || projectId}.{config.domain}</code>
            </p>
          </div>

          {#if addMessage}
            <div class="p-3 bg-green-50 border border-green-200 rounded-lg text-green-600 text-sm">
              {addMessage}
            </div>
          {/if}

          {#if addError}
            <div class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
              {addError}
            </div>
          {/if}

          {#if verificationTxt}
            <div class="bg-amber-50 border border-amber-200 rounded-lg p-4">
              <p class="text-sm font-medium text-amber-800 mb-2">DNS verification required</p>
              <code class="text-sm text-amber-700 font-mono">{verificationTxt}</code>
            </div>
          {/if}

          {#if verifyMessage}
            <div class="p-3 bg-slate-50 border border-slate-200 rounded-lg text-slate-600 text-sm">
              {verifyMessage}
            </div>
          {/if}

          {#if verifyError}
            <div class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
              {verifyError}
            </div>
          {/if}

          <div class="space-y-2">
            {#if domains.length === 0}
              <p class="text-sm text-slate-500">No domains configured yet.</p>
            {:else}
              {#each domains as domain}
                <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between border border-slate-200 rounded-md p-3">
                  <div>
                    <div class="flex items-center gap-2">
                      <span class="font-medium text-slate-900">{domain.domain}</span>
                      {#if domain.is_primary}
                        <span class="text-xs px-2 py-0.5 rounded bg-slate-100 text-slate-600">primary</span>
                      {/if}
                      <span class="text-xs px-2 py-0.5 rounded {domain.status === 'active' ? 'bg-green-100 text-green-700' : domain.status === 'pending' ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-600'}">
                        {domain.status}
                      </span>
                      <span class="text-xs text-slate-400">{domain.type}</span>
                    </div>
                    <p class="text-xs text-slate-500 mt-1">Path: {domain.slug}</p>
                  </div>
                  <div class="flex items-center gap-2">
                    {#if domain.type === 'custom' && domain.status !== 'active'}
                      <button
                        type="button"
                        onclick={() => verifyDomain(domain.domain)}
                        disabled={verifyingDomain === domain.domain}
                        class="px-3 py-1 text-xs font-medium text-amber-700 bg-amber-50 border border-amber-200 rounded hover:bg-amber-100 disabled:opacity-50"
                      >
                        {verifyingDomain === domain.domain ? 'Verifying...' : 'Verify'}
                      </button>
                    {/if}
                    <button
                      type="button"
                      onclick={() => removeDomain(domain.domain)}
                      disabled={removingDomain === domain.domain || (domain.is_primary && domain.type === 'system')}
                      class="px-3 py-1 text-xs font-medium text-red-700 bg-red-50 border border-red-200 rounded hover:bg-red-100 disabled:opacity-50"
                    >
                      {removingDomain === domain.domain ? 'Removing...' : 'Remove'}
                    </button>
                  </div>
                </div>
              {/each}
            {/if}
          </div>
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
              onclick={deleteProject}
              disabled={deletingProject}
              class="px-4 py-2 text-sm font-medium text-red-600 border border-red-300 rounded-md hover:bg-red-100 disabled:opacity-50 transition"
            >
              {deletingProject ? 'Deleting...' : 'Delete Project'}
            </button>
          </div>
        </div>
      </div>

      {#if deleteError}
        <div class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {deleteError}
        </div>
      {/if}

      {#if error}
        <div class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {error}
        </div>
      {/if}
    </div>
  {/if}
</div>
