/**
 * Workspace API
 */

import { API_BASE, handleResponse } from './common'
import type { Workspace, WorkspaceConfig } from '../types'

export async function getWorkspaces(connId: string): Promise<Workspace[]> {
  const response = await fetch(`${API_BASE}/workspaces/${connId}`)
  return handleResponse<Workspace[]>(response)
}

export async function getWorkspace(connId: string, name: string): Promise<WorkspaceConfig> {
  const response = await fetch(`${API_BASE}/workspaces/${connId}/${name}`)
  return handleResponse<WorkspaceConfig>(response)
}

export async function createWorkspace(connId: string, config: WorkspaceConfig): Promise<Workspace> {
  const response = await fetch(`${API_BASE}/workspaces/${connId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  return handleResponse<Workspace>(response)
}

export async function updateWorkspace(connId: string, name: string, config: WorkspaceConfig): Promise<WorkspaceConfig> {
  const response = await fetch(`${API_BASE}/workspaces/${connId}/${name}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  return handleResponse<WorkspaceConfig>(response)
}

export async function deleteWorkspace(connId: string, name: string, recurse = false): Promise<void> {
  const params = recurse ? '?recurse=true' : ''
  const response = await fetch(`${API_BASE}/workspaces/${connId}/${name}${params}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}
