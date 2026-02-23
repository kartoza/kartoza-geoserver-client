/**
 * Style API
 */

import { API_BASE, handleResponse } from './common'
import type { Style } from '../types'

export async function getStyles(connId: string, workspace: string): Promise<Style[]> {
  const response = await fetch(`${API_BASE}/styles/${connId}/${workspace}`)
  return handleResponse<Style[]>(response)
}

export async function deleteStyle(connId: string, workspace: string, name: string, purge = false): Promise<void> {
  const params = purge ? '?purge=true' : ''
  const response = await fetch(`${API_BASE}/styles/${connId}/${workspace}/${name}${params}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}

// Layer styles association
export interface LayerStyles {
  defaultStyle: string
  additionalStyles: string[]
}

export async function getLayerStyles(connId: string, workspace: string, layer: string): Promise<LayerStyles> {
  const response = await fetch(`${API_BASE}/layerstyles/${connId}/${workspace}/${layer}`)
  return handleResponse<LayerStyles>(response)
}

export async function updateLayerStyles(
  connId: string,
  workspace: string,
  layer: string,
  defaultStyle: string,
  additionalStyles: string[]
): Promise<LayerStyles> {
  const response = await fetch(`${API_BASE}/layerstyles/${connId}/${workspace}/${layer}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ defaultStyle, additionalStyles }),
  })
  return handleResponse<LayerStyles>(response)
}

// Style content for editor
export interface StyleContent {
  name: string
  workspace: string
  format: 'sld' | 'css' | 'mbstyle'
  content: string
}

export async function getStyleContent(connId: string, workspace: string, name: string): Promise<StyleContent> {
  const response = await fetch(`${API_BASE}/styles/${connId}/${workspace}/${name}`)
  return handleResponse<StyleContent>(response)
}

export async function updateStyleContent(
  connId: string,
  workspace: string,
  name: string,
  content: string,
  format: 'sld' | 'css' | 'mbstyle' = 'sld'
): Promise<StyleContent> {
  const response = await fetch(`${API_BASE}/styles/${connId}/${workspace}/${name}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content, format }),
  })
  return handleResponse<StyleContent>(response)
}

export async function createStyle(
  connId: string,
  workspace: string,
  name: string,
  content: string,
  format: 'sld' | 'css' | 'mbstyle' = 'sld'
): Promise<StyleContent> {
  const response = await fetch(`${API_BASE}/styles/${connId}/${workspace}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, content, format }),
  })
  return handleResponse<StyleContent>(response)
}
