/**
 * GeoServer Connection API
 */

import { API_BASE, handleResponse } from './common'
import type {
  Connection,
  ConnectionCreate,
  TestConnectionResult,
  ServerInfo,
} from '../types'

export async function getConnections(): Promise<Connection[]> {
  const response = await fetch(`${API_BASE}/connections`)
  return handleResponse<Connection[]>(response)
}

export async function getConnection(id: string): Promise<Connection> {
  const response = await fetch(`${API_BASE}/connections/${id}`)
  return handleResponse<Connection>(response)
}

export async function createConnection(conn: ConnectionCreate): Promise<Connection> {
  const response = await fetch(`${API_BASE}/connections`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(conn),
  })
  return handleResponse<Connection>(response)
}

export async function updateConnection(id: string, conn: Partial<ConnectionCreate>): Promise<Connection> {
  const response = await fetch(`${API_BASE}/connections/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(conn),
  })
  return handleResponse<Connection>(response)
}

export async function deleteConnection(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/connections/${id}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}

export async function testConnection(id: string): Promise<TestConnectionResult> {
  const response = await fetch(`${API_BASE}/connections/${id}/test`, {
    method: 'POST',
  })
  return handleResponse<TestConnectionResult>(response)
}

export async function testConnectionDirect(conn: ConnectionCreate): Promise<TestConnectionResult> {
  const response = await fetch(`${API_BASE}/connections/test`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(conn),
  })
  return handleResponse<TestConnectionResult>(response)
}

export async function getServerInfo(id: string): Promise<ServerInfo> {
  const response = await fetch(`${API_BASE}/connections/${id}/info`)
  return handleResponse<ServerInfo>(response)
}
