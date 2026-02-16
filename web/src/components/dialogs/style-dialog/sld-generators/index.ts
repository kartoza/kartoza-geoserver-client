// Re-export all SLD generators for convenient imports
export {
  generateRasterColorMapSLD,
  generateHillshadeSLD,
  generateHillshadeWithColorSLD,
  generateContrastEnhancementSLD,
  generateRasterContourSLD,
} from './raster'

export {
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
} from './vector'

export {
  equalIntervalBreaks,
  quantileBreaks,
  jenksBreaks,
  prettyBreaks,
  calculateBreaks,
  interpolateColors,
  generateClassifiedSLD,
} from './classification'
