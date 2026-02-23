/**
 * Format bytes to human-readable size
 */
export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

/**
 * Get badge color for format type
 */
export function getFormatBadgeColor(format: string): string {
  switch (format) {
    case 'cog': return 'green'
    case 'copc': return 'purple'
    case 'geoparquet': return 'blue'
    case 'geojson': return 'cyan'
    case 'geotiff': return 'orange'
    case 'parquet': return 'teal'
    default: return 'gray'
  }
}

/**
 * Get badge color for preview type
 */
export function getPreviewTypeBadgeColor(type: string): string {
  switch (type) {
    case 'raster': return 'orange'
    case 'vector': return 'blue'
    case 'pointcloud': return 'purple'
    case 'table': return 'teal'
    default: return 'gray'
  }
}
