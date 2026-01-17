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

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(error.message || `HTTP ${response.status}`)
  }

  return response.json()
}

// Types
export interface Project {
  id: string
  name: string
  owner_id: string
  owner_email?: string // Only visible to admin
  created_at: string
  updated_at: string
}

export interface Image {
  id: string
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
  project: Project
  environments: {
    [env: string]: {
      image_id: string
      content_hash: string
      file_count: number
      updated_at: string
    }
  }
}

// API functions
export const api = {
  // Projects
  async getProjects(): Promise<Project[]> {
    return request('/projects')
  },

  async getProject(id: string): Promise<Project> {
    return request(`/projects/${id}`)
  },

  // Deployments
  async getCurrentDeployment(projectName: string, env: string = 'prod'): Promise<CurrentDeployment> {
    return request(`/current?project=${projectName}&environment=${env}`)
  },

  // Images
  async getImages(projectId: string, page: number = 1, limit: number = 20): Promise<{ images: Image[], total: number }> {
    return request(`/images?project_id=${projectId}&page=${page}&limit=${limit}`)
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
  async getHistory(projectName: string, env: string = 'prod', limit: number = 20): Promise<Image[]> {
    return request(`/history?project=${projectName}&env=${env}&limit=${limit}`)
  },

  // Health
  async getHealth(): Promise<{ status: string }> {
    return request('/health')
  }
}
