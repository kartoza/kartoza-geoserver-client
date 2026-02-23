/**
 * Layer API
 */

import { API_BASE, handleResponse } from './common'
import type {
  Layer,
  LayerUpdate,
  LayerMetadata,
  LayerMetadataUpdate,
  FeatureType,
  Coverage,
} from '../types'

// Layer API
export async function getLayers(connId: string, workspace: string): Promise<Layer[]> {
  const response = await fetch(`${API_BASE}/layers/${connId}/${workspace}`)
  return handleResponse<Layer[]>(response)
}

export async function getLayer(connId: string, workspace: string, name: string): Promise<Layer> {
  const response = await fetch(`${API_BASE}/layers/${connId}/${workspace}/${name}`)
  return handleResponse<Layer>(response)
}

export async function updateLayer(connId: string, workspace: string, name: string, update: LayerUpdate): Promise<Layer> {
  const response = await fetch(`${API_BASE}/layers/${connId}/${workspace}/${name}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(update),
  })
  return handleResponse<Layer>(response)
}

export async function deleteLayer(connId: string, workspace: string, name: string): Promise<void> {
  const response = await fetch(`${API_BASE}/layers/${connId}/${workspace}/${name}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}

// Layer Metadata API
export async function getLayerFullMetadata(connId: string, workspace: string, name: string): Promise<LayerMetadata> {
  const response = await fetch(`${API_BASE}/layermetadata/${connId}/${workspace}/${name}`)
  return handleResponse<LayerMetadata>(response)
}

export async function getLayerFeatureCount(connId: string, workspace: string, name: string): Promise<number> {
  const response = await fetch(`${API_BASE}/layers/${connId}/${workspace}/${name}/count`)
  const data = await handleResponse<{ count: number }>(response)
  return data.count
}

export async function updateLayerMetadata(
  connId: string,
  workspace: string,
  name: string,
  update: LayerMetadataUpdate
): Promise<LayerMetadata> {
  const response = await fetch(`${API_BASE}/layermetadata/${connId}/${workspace}/${name}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(update),
  })
  return handleResponse<LayerMetadata>(response)
}

// Feature Type API
export async function getFeatureTypes(connId: string, workspace: string, store: string): Promise<FeatureType[]> {
  const response = await fetch(`${API_BASE}/featuretypes/${connId}/${workspace}/${store}`)
  return handleResponse<FeatureType[]>(response)
}

export async function publishFeatureType(connId: string, workspace: string, store: string, name: string): Promise<FeatureType> {
  const response = await fetch(`${API_BASE}/featuretypes/${connId}/${workspace}/${store}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
  return handleResponse<FeatureType>(response)
}

// Coverage API
export async function getCoverages(connId: string, workspace: string, store: string): Promise<Coverage[]> {
  const response = await fetch(`${API_BASE}/coverages/${connId}/${workspace}/${store}`)
  return handleResponse<Coverage[]>(response)
}

export async function publishCoverage(connId: string, workspace: string, store: string, name: string): Promise<Coverage> {
  const response = await fetch(`${API_BASE}/coverages/${connId}/${workspace}/${store}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
  return handleResponse<Coverage>(response)
}
