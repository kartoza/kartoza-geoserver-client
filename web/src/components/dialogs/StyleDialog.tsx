// This file is kept for backwards compatibility
// The StyleDialog has been refactored into style-dialog/ directory
// All new code should import from './style-dialog'

export {
  StyleDialog,
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
  parseSLDRules,
  generateSLD,
  equalIntervalBreaks,
  quantileBreaks,
  jenksBreaks,
  prettyBreaks,
  calculateBreaks,
  interpolateColors,
  generateClassifiedSLD,
  generateRasterColorMapSLD,
  generateHillshadeSLD,
  generateHillshadeWithColorSLD,
  generateContrastEnhancementSLD,
  generateRasterContourSLD,
  generateHatchFillSLD,
  generatePointPatternFillSLD,
  generateMarkerLineSLD,
  generateArrowLineSLD,
  generateLabelSLD,
  generateGradientFillSLD,
  generateCategorizedSLD,
  generateRuleBasedSLD,
  generateProportionalSymbolSLD,
  generateCentroidFillSLD,
} from './style-dialog'

export type {
  ClassificationMethod,
  StyleRule,
  PointStylePreset,
} from './style-dialog'

// Legacy named exports that were previously available
import {
  FONT_MARKER_FONTS as _FONT_MARKER_FONTS,
  LABEL_PLACEMENT_OPTIONS as _LABEL_PLACEMENT_OPTIONS,
  LABEL_ANCHOR_POINTS as _LABEL_ANCHOR_POINTS,
} from './style-dialog'

export {
  _FONT_MARKER_FONTS,
  _LABEL_PLACEMENT_OPTIONS,
  _LABEL_ANCHOR_POINTS,
}
