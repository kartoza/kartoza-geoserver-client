/**
 * Layer Group API
 */

import { API_BASE, handleResponse } from './common'
import type {
  LayerGroup,
  LayerGroupCreate,
  LayerGroupDetails,
  LayerGroupUpdate,
} from '../types'

export async function getLayerGroups(connId: string, workspace: string): Promise<LayerGroup[]> {
  const response = await fetch(`${API_BASE}/layergroups/${connId}/${workspace}`)
  return handleResponse<LayerGroup[]>(response)
}

export async function createLayerGroup(
  connId: string,
  workspace: string,
  config: LayerGroupCreate
): Promise<LayerGroup> {
  const response = await fetch(`${API_BASE}/layergroups/${connId}/${workspace}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  return handleResponse<LayerGroup>(response)
}

export async function getLayerGroup(
  connId: string,
  workspace: string,
  name: string
): Promise<LayerGroupDetails> {
  const response = await fetch(`${API_BASE}/layergroups/${connId}/${workspace}/${name}`)
  return handleResponse<LayerGroupDetails>(response)
}

export async function updateLayerGroup(
  connId: string,
  workspace: string,
  name: string,
  update: LayerGroupUpdate
): Promise<LayerGroupDetails> {
  const response = await fetch(`${API_BASE}/layergroups/${connId}/${workspace}/${name}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(update),
  })
  return handleResponse<LayerGroupDetails>(response)
}

export async function deleteLayerGroup(connId: string, workspace: string, name: string): Promise<void> {
  const response = await fetch(`${API_BASE}/layergroups/${connId}/${workspace}/${name}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}
