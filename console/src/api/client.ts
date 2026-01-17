// API client for SitePod backend

import { auth } from '../lib/auth.svelte'

const BASE_URL = '/api/v1'

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: unknown
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  if (auth.token) {
    headers['Authorization'] = `Bearer ${auth.token}`
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    method: options.method || 'GET',
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined
  })

  const text = await response.text()
  let data: unknown = null
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = text
    }
  }

  if (!response.ok) {
    const message =
      (typeof data === 'object' && data !== null && ('error' in data || 'message' in data)
        ? ((data as { error?: string, message?: string }).error || (data as { message?: string }).message)
        : typeof data === 'string'
          ? data
          : null) || `HTTP ${response.status}`
    throw new Error(message)
  }

  return (data ?? {}) as T
}

// Types
export interface Project {
  id: string
  name: string
  subdomain?: string
  owner_id: string
  owner_email?: string // Only visible to admin
  created_at: string
  updated_at: string
}

export interface Image {
  id: string
  image_id: string
  project_id: string
  content_hash: string
  file_count: number
  total_size: number
  created_at: string
  deployed_to?: string[]
}

export interface Release {
  id: string
  project_id: string
  image_id: string
  environment: string
  created_at: string
}

export interface CurrentDeployment {
  image_id: string
  content_hash: string
  deployed_at: string
  file_count?: number
}

export interface Domain {
  domain: string
  slug: string
  type: string
  status: string
  is_primary: boolean
  created_at: string
}

export interface AddDomainResponse {
  domain: string
  slug: string
  status: string
  verification_token?: string
  verification_txt?: string
}

export interface VerifyDomainResponse {
  domain: string
  status: string
  verified: boolean
  message?: string
}

// API functions
export const api = {
  // Projects
  async getProjects(): Promise<Project[]> {
    return request('/projects')
  },

  async getProject(id: string): Promise<Project> {
    return request(`/projects/${encodeURIComponent(id)}`)
  },

  // Deployments
  async getCurrentDeployment(projectName: string, env: string = 'prod'): Promise<CurrentDeployment> {
    return request(`/current?project=${encodeURIComponent(projectName)}&environment=${env}`)
  },

  // Images
  async getImages(projectName: string, page: number = 1, limit: number = 20): Promise<{ images: Image[], total: number }> {
    return request(`/images?project=${encodeURIComponent(projectName)}&page=${page}&limit=${limit}`)
  },

  // Releases
  async createRelease(projectId: string, imageId: string, environment: string): Promise<Release> {
    return request('/release', {
      method: 'POST',
      body: { project_id: projectId, image_id: imageId, environment }
    })
  },

  async rollback(projectName: string, environment: string, imageId: string): Promise<void> {
    return request('/rollback', {
      method: 'POST',
      body: { project: projectName, environment, image_id: imageId }
    })
  },

  // History
  async getHistory(projectName: string, env: string = 'prod', limit: number = 20): Promise<{ items: { image_id: string, content_hash: string, created_at: string, git_commit?: string }[] }> {
    return request(`/history?project=${encodeURIComponent(projectName)}&env=${env}&limit=${limit}`)
  },

  // Domains
  async listDomains(projectName: string): Promise<{ domains: Domain[] }> {
    return request(`/domains?project=${encodeURIComponent(projectName)}`)
  },

  async addDomain(projectName: string, domain: string, slug: string = '/'): Promise<AddDomainResponse> {
    return request('/domains', {
      method: 'POST',
      body: { project: projectName, domain, slug }
    })
  },

  async verifyDomain(domain: string): Promise<VerifyDomainResponse> {
    return request(`/domains/${encodeURIComponent(domain)}/verify`, { method: 'POST' })
  },

  async removeDomain(domain: string): Promise<void> {
    await request(`/domains/${encodeURIComponent(domain)}`, { method: 'DELETE' })
  },

  // Projects
  async deleteProject(projectName: string): Promise<{ message?: string }> {
    return request(`/projects/${encodeURIComponent(projectName)}`, { method: 'DELETE' })
  },

  // Health
  async getHealth(): Promise<{ status: string }> {
    return request('/health')
  }
}
