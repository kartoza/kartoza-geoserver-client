import { API_BASE } from './common'
import type { UploadResult } from '../types'

export const CHUNK_SIZE = 5 * 1024 * 1024

export async function initUploadSession(
  connId: string,
  workspace: string,
  filename: string,
  totalSize: number,
  chunkSize = CHUNK_SIZE,
): Promise<{ sessionId: string; chunkSize: number; totalChunks: number }> {
  const res = await fetch(`${API_BASE}/upload/init`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ connId, workspace, filename, totalSize, chunkSize }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function uploadChunk(
  sessionId: string,
  chunkIndex: number,
  chunk: Blob,
  onProgress?: (pct: number) => void,
): Promise<{ received: number; totalChunks: number }> {
  return new Promise((resolve, reject) => {
    const formData = new FormData()
    formData.append('sessionId', sessionId)
    formData.append('chunkIndex', String(chunkIndex))
    formData.append('chunk', chunk)

    const xhr = new XMLHttpRequest()
    xhr.open('POST', `${API_BASE}/upload/chunk`)

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    }

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText))
        } catch {
          reject(new Error('Invalid JSON response'))
        }
      } else {
        try {
          const err = JSON.parse(xhr.responseText)
          reject(new Error(err.error || `HTTP ${xhr.status}`))
        } catch {
          reject(new Error(`HTTP ${xhr.status}`))
        }
      }
    }

    xhr.onerror = () => reject(new Error('Network error during chunk upload'))
    xhr.onabort = () => reject(new Error('Upload aborted'))

    xhr.send(formData)
  })
}

export async function completeUpload(sessionId: string): Promise<UploadResult> {
  const res = await fetch(`${API_BASE}/upload/complete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sessionId }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function cancelUpload(sessionId: string): Promise<void> {
  await fetch(`${API_BASE}/upload/session/${encodeURIComponent(sessionId)}`, {
    method: 'DELETE',
  })
}

export interface GeoServerProgress {
  sent: number
  total: number
  done: boolean
}

export async function getGeoServerProgress(sessionId: string): Promise<GeoServerProgress> {
  const res = await fetch(`${API_BASE}/upload/session/${encodeURIComponent(sessionId)}/progress`)
  if (!res.ok) return { sent: 0, total: 0, done: true }
  return res.json()
}
