// Classification methods
export type ClassificationMethod = 'equal-interval' | 'quantile' | 'jenks' | 'pretty'

// Style rule interface for visual editor
export interface StyleRule {
  name: string
  filter?: string
  symbolizer: {
    type: 'polygon' | 'line' | 'point'
    fill?: string
    fillOpacity?: number
    stroke?: string
    strokeWidth?: number
    strokeOpacity?: number
    pointShape?: 'circle' | 'square' | 'triangle' | 'star' | 'cross' | 'x'
    pointSize?: number
    haloColor?: string
    haloRadius?: number
    rotation?: number
  }
}

// Point style preset interface (re-exported from constants for convenience)
export interface PointStylePreset {
  name: string
  description: string
  icon: string
  shape: string
  fill: string
  fillOpacity: number
  stroke: string
  strokeWidth: number
  size: number
  haloColor?: string
  haloRadius?: number
  rotation?: number
}
