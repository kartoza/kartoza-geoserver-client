import type { ClassificationMethod } from '../types'

// Calculate equal interval breaks
export function equalIntervalBreaks(values: number[], numClasses: number): number[] {
  const min = Math.min(...values)
  const max = Math.max(...values)
  const interval = (max - min) / numClasses
  const breaks: number[] = []
  for (let i = 0; i <= numClasses; i++) {
    breaks.push(min + i * interval)
  }
  return breaks
}

// Calculate quantile breaks
export function quantileBreaks(values: number[], numClasses: number): number[] {
  const sorted = [...values].sort((a, b) => a - b)
  const breaks: number[] = []
  for (let i = 0; i <= numClasses; i++) {
    const index = Math.floor((i / numClasses) * (sorted.length - 1))
    breaks.push(sorted[index])
  }
  return breaks
}

// Calculate Jenks natural breaks (simplified Ckmeans approximation)
export function jenksBreaks(values: number[], numClasses: number): number[] {
  // Simplified implementation - uses k-means-like approach
  const sorted = [...values].sort((a, b) => a - b)
  const n = sorted.length

  if (n <= numClasses) {
    return sorted
  }

  // Initialize breaks evenly
  const breaks = equalIntervalBreaks(values, numClasses)

  // Iteratively optimize (simplified version)
  for (let iter = 0; iter < 10; iter++) {
    // Assign each value to a class
    const classMeans: number[] = []
    for (let i = 0; i < numClasses; i++) {
      const lower = breaks[i]
      const upper = breaks[i + 1]
      const classValues = sorted.filter(v => v >= lower && (i === numClasses - 1 ? v <= upper : v < upper))
      if (classValues.length > 0) {
        classMeans.push(classValues.reduce((a, b) => a + b, 0) / classValues.length)
      } else {
        classMeans.push((lower + upper) / 2)
      }
    }

    // Update breaks as midpoints between class means
    for (let i = 1; i < numClasses; i++) {
      breaks[i] = (classMeans[i - 1] + classMeans[i]) / 2
    }
  }

  return breaks
}

// Calculate pretty breaks (nice round numbers)
export function prettyBreaks(values: number[], numClasses: number): number[] {
  const min = Math.min(...values)
  const max = Math.max(...values)
  const range = max - min

  // Calculate a nice interval
  const rawInterval = range / numClasses
  const magnitude = Math.pow(10, Math.floor(Math.log10(rawInterval)))
  const normalized = rawInterval / magnitude

  let niceInterval: number
  if (normalized <= 1) niceInterval = 1 * magnitude
  else if (normalized <= 2) niceInterval = 2 * magnitude
  else if (normalized <= 5) niceInterval = 5 * magnitude
  else niceInterval = 10 * magnitude

  const niceMin = Math.floor(min / niceInterval) * niceInterval
  const breaks: number[] = []
  for (let i = 0; i <= numClasses; i++) {
    breaks.push(niceMin + i * niceInterval)
  }

  return breaks
}

// Calculate breaks based on method
export function calculateBreaks(values: number[], numClasses: number, method: ClassificationMethod): number[] {
  switch (method) {
    case 'equal-interval':
      return equalIntervalBreaks(values, numClasses)
    case 'quantile':
      return quantileBreaks(values, numClasses)
    case 'jenks':
      return jenksBreaks(values, numClasses)
    case 'pretty':
      return prettyBreaks(values, numClasses)
    default:
      return equalIntervalBreaks(values, numClasses)
  }
}

// Interpolate colors from a ramp
export function interpolateColors(ramp: string[], numColors: number): string[] {
  if (numColors <= ramp.length) {
    // Sample from the ramp
    const result: string[] = []
    for (let i = 0; i < numColors; i++) {
      const index = Math.floor((i / (numColors - 1)) * (ramp.length - 1))
      result.push(ramp[index])
    }
    return result
  }
  // Just return the ramp if we need more colors than available
  return ramp.slice(0, numColors)
}

// Generate classified SLD
export function generateClassifiedSLD(
  styleName: string,
  attribute: string,
  breaks: number[],
  colors: string[],
  geometryType: 'polygon' | 'line' | 'point'
): string {
  const rules = breaks.slice(0, -1).map((lower, i) => {
    const upper = breaks[i + 1]
    const color = colors[i] || colors[colors.length - 1]
    const ruleName = `${lower.toFixed(2)} - ${upper.toFixed(2)}`

    let symbolizer = ''
    if (geometryType === 'polygon') {
      symbolizer = `
            <PolygonSymbolizer>
              <Fill>
                <CssParameter name="fill">${color}</CssParameter>
                <CssParameter name="fill-opacity">0.8</CssParameter>
              </Fill>
              <Stroke>
                <CssParameter name="stroke">#333333</CssParameter>
                <CssParameter name="stroke-width">0.5</CssParameter>
              </Stroke>
            </PolygonSymbolizer>`
    } else if (geometryType === 'line') {
      symbolizer = `
            <LineSymbolizer>
              <Stroke>
                <CssParameter name="stroke">${color}</CssParameter>
                <CssParameter name="stroke-width">2</CssParameter>
              </Stroke>
            </LineSymbolizer>`
    } else {
      symbolizer = `
            <PointSymbolizer>
              <Graphic>
                <Mark>
                  <WellKnownName>circle</WellKnownName>
                  <Fill>
                    <CssParameter name="fill">${color}</CssParameter>
                  </Fill>
                  <Stroke>
                    <CssParameter name="stroke">#333333</CssParameter>
                    <CssParameter name="stroke-width">1</CssParameter>
                  </Stroke>
                </Mark>
                <Size>8</Size>
              </Graphic>
            </PointSymbolizer>`
    }

    return `
          <Rule>
            <Name>${ruleName}</Name>
            <Title>${ruleName}</Title>
            <ogc:Filter>
              <ogc:And>
                <ogc:PropertyIsGreaterThanOrEqualTo>
                  <ogc:PropertyName>${attribute}</ogc:PropertyName>
                  <ogc:Literal>${lower}</ogc:Literal>
                </ogc:PropertyIsGreaterThanOrEqualTo>
                <ogc:PropertyIsLessThan>
                  <ogc:PropertyName>${attribute}</ogc:PropertyName>
                  <ogc:Literal>${upper}</ogc:Literal>
                </ogc:PropertyIsLessThan>
              </ogc:And>
            </ogc:Filter>${symbolizer}
          </Rule>`
  }).join('')

  return `<?xml version="1.0" encoding="UTF-8"?>
<StyledLayerDescriptor version="1.0.0"
  xsi:schemaLocation="http://www.opengis.net/sld StyledLayerDescriptor.xsd"
  xmlns="http://www.opengis.net/sld"
  xmlns:ogc="http://www.opengis.net/ogc"
  xmlns:xlink="http://www.w3.org/1999/xlink"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <NamedLayer>
    <Name>${styleName}</Name>
    <UserStyle>
      <Title>${styleName} - Classified</Title>
      <FeatureTypeStyle>${rules}
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}
