/**
 * Apache Iceberg REST Catalog API
 */

import { API_BASE, handleResponse } from './common'
import type {
  IcebergConnection,
  IcebergConnectionCreate,
  IcebergTestResult,
  IcebergNamespace,
  IcebergTable,
  IcebergSchema,
  IcebergSnapshot,
} from '../types'

// Iceberg Connection API
export async function getIcebergConnections(): Promise<IcebergConnection[]> {
  const response = await fetch(`${API_BASE}/iceberg/connections`)
  return handleResponse<IcebergConnection[]>(response)
}

export async function getIcebergConnection(id: string): Promise<IcebergConnection> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${id}`)
  return handleResponse<IcebergConnection>(response)
}

export async function createIcebergConnection(conn: IcebergConnectionCreate): Promise<IcebergConnection> {
  const response = await fetch(`${API_BASE}/iceberg/connections`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(conn),
  })
  return handleResponse<IcebergConnection>(response)
}

export async function updateIcebergConnection(id: string, conn: Partial<IcebergConnectionCreate>): Promise<IcebergConnection> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(conn),
  })
  return handleResponse<IcebergConnection>(response)
}

export async function deleteIcebergConnection(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${id}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}

export async function testIcebergConnection(id: string): Promise<IcebergTestResult> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${id}/test`, {
    method: 'POST',
  })
  return handleResponse<IcebergTestResult>(response)
}

export async function testIcebergConnectionDirect(conn: IcebergConnectionCreate): Promise<IcebergTestResult> {
  const response = await fetch(`${API_BASE}/iceberg/connections/test`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(conn),
  })
  return handleResponse<IcebergTestResult>(response)
}

// Iceberg Namespace API
export async function getIcebergNamespaces(connectionId: string): Promise<IcebergNamespace[]> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces`)
  return handleResponse<IcebergNamespace[]>(response)
}

export async function getIcebergNamespace(connectionId: string, namespace: string): Promise<IcebergNamespace> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}`)
  return handleResponse<IcebergNamespace>(response)
}

export async function createIcebergNamespace(connectionId: string, namespace: string, properties?: Record<string, string>): Promise<IcebergNamespace> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ namespace, properties }),
  })
  return handleResponse<IcebergNamespace>(response)
}

export async function deleteIcebergNamespace(connectionId: string, namespace: string): Promise<void> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}

// Iceberg Table API
export async function getIcebergTables(connectionId: string, namespace: string): Promise<IcebergTable[]> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}/tables`)
  return handleResponse<IcebergTable[]>(response)
}

export async function getIcebergTable(connectionId: string, namespace: string, table: string): Promise<IcebergTable> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}/tables/${encodeURIComponent(table)}`)
  return handleResponse<IcebergTable>(response)
}

export async function getIcebergTableSchema(connectionId: string, namespace: string, table: string): Promise<IcebergSchema> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}/tables/${encodeURIComponent(table)}/schema`)
  return handleResponse<IcebergSchema>(response)
}

export async function getIcebergTableSnapshots(connectionId: string, namespace: string, table: string): Promise<IcebergSnapshot[]> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}/tables/${encodeURIComponent(table)}/snapshots`)
  return handleResponse<IcebergSnapshot[]>(response)
}

// Create Iceberg table from S3 GeoParquet
export async function createIcebergTableFromS3(
  connectionId: string,
  namespace: string,
  tableName: string,
  s3ConnectionId: string,
  s3Bucket: string,
  s3Key: string
): Promise<IcebergTable> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}/tables`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: tableName,
      source: 's3',
      s3ConnectionId,
      s3Bucket,
      s3Key,
    }),
  })
  return handleResponse<IcebergTable>(response)
}

export async function deleteIcebergTable(connectionId: string, namespace: string, table: string, purge = false): Promise<void> {
  const params = purge ? '?purge=true' : ''
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}/tables/${encodeURIComponent(table)}${params}`, {
    method: 'DELETE',
  })
  return handleResponse<void>(response)
}

// Get Jupyter URL for Iceberg connection
export async function getIcebergJupyterUrl(connectionId: string): Promise<{ url: string; available: boolean }> {
  const response = await fetch(`${API_BASE}/iceberg/connections/${connectionId}/jupyter`)
  return handleResponse<{ url: string; available: boolean }>(response)
}

// Create table schema field type
export interface IcebergFieldCreate {
  id: number
  name: string
  type: string // 'string', 'long', 'double', 'boolean', 'timestamp', 'binary', etc.
  required: boolean
  doc?: string
}

// Create table request
export interface CreateIcebergTableRequest {
  name: string
  location?: string
  schema: {
    type: string
    fields: IcebergFieldCreate[]
  }
  properties?: Record<string, string>
}

// Create a new table in a namespace
export async function createIcebergTable(
  connectionId: string,
  namespace: string,
  request: CreateIcebergTableRequest
): Promise<IcebergTable> {
  const response = await fetch(
    `${API_BASE}/iceberg/connections/${connectionId}/namespaces/${encodeURIComponent(namespace)}/tables`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    }
  )
  return handleResponse<IcebergTable>(response)
}

// Backward compatibility alias
export const getIcebergSnapshots = getIcebergTableSnapshots
