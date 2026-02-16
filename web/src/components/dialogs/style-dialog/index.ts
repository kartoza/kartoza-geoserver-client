// Main entry point for style dialog
export { StyleDialog } from './StyleDialog'

// Re-export types for external use
export type { ClassificationMethod, StyleRule, PointStylePreset } from './types'

// Re-export constants that might be useful externally
export {
  DEFAULT_SLD,
  DEFAULT_CSS,
  COLOR_RAMPS,
  RASTER_COLOR_RAMPS,
  HILLSHADE_PRESETS,
  MARKER_SHAPES,
  LINE_DASH_PATTERNS,
  LINE_CAP_STYLES,
  LINE_JOIN_STYLES,
  FILL_PATTERNS,
  BLEND_MODES,
  FONT_MARKER_FONTS,
  LABEL_PLACEMENT_OPTIONS,
  LABEL_ANCHOR_POINTS,
  POINT_STYLE_PRESETS,
} from './constants'

// Re-export SLD generators for external use
export * from './sld-generators'

// Re-export utilities
export { parseSLDRules, generateSLD } from './sld-utils'
