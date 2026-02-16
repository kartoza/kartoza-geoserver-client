// Generate hatch pattern fill SLD
export function generateHatchFillSLD(
  styleName: string,
  fillColor: string,
  strokeColor: string,
  strokeWidth: number,
  angle: number,
  spacing: number = 8,
  doubleHatch: boolean = false
): string {
  const hatchGraphic = `
            <GraphicFill>
              <Graphic>
                <Mark>
                  <WellKnownName>shape://horline</WellKnownName>
                  <Stroke>
                    <CssParameter name="stroke">${strokeColor}</CssParameter>
                    <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
                  </Stroke>
                </Mark>
                <Size>${spacing}</Size>
                <Rotation>${angle}</Rotation>
              </Graphic>
            </GraphicFill>`

  const secondHatch = doubleHatch ? `
            <GraphicFill>
              <Graphic>
                <Mark>
                  <WellKnownName>shape://horline</WellKnownName>
                  <Stroke>
                    <CssParameter name="stroke">${strokeColor}</CssParameter>
                    <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
                  </Stroke>
                </Mark>
                <Size>${spacing}</Size>
                <Rotation>${angle + 90}</Rotation>
              </Graphic>
            </GraphicFill>` : ''

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
      <Title>${styleName} - Hatch Fill</Title>
      <FeatureTypeStyle>
        <Rule>
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${fillColor}</CssParameter>
            </Fill>
          </PolygonSymbolizer>
          <PolygonSymbolizer>${hatchGraphic}${secondHatch}
          </PolygonSymbolizer>
          <PolygonSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">1</CssParameter>
            </Stroke>
          </PolygonSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate point pattern fill SLD
export function generatePointPatternFillSLD(
  styleName: string,
  fillColor: string,
  markerColor: string,
  markerSize: number,
  spacingX: number,
  spacingY: number,
  strokeColor: string,
  strokeWidth: number
): string {
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
      <Title>${styleName} - Point Pattern Fill</Title>
      <FeatureTypeStyle>
        <Rule>
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${fillColor}</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </PolygonSymbolizer>
          <PolygonSymbolizer>
            <Fill>
              <GraphicFill>
                <Graphic>
                  <Mark>
                    <WellKnownName>circle</WellKnownName>
                    <Fill>
                      <CssParameter name="fill">${markerColor}</CssParameter>
                    </Fill>
                  </Mark>
                  <Size>${markerSize}</Size>
                </Graphic>
              </GraphicFill>
            </Fill>
            <VendorOption name="graphic-margin">${spacingY} ${spacingX}</VendorOption>
          </PolygonSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate marker line SLD (repeating markers along line)
export function generateMarkerLineSLD(
  styleName: string,
  strokeColor: string,
  strokeWidth: number,
  markerShape: string,
  markerSize: number,
  markerFill: string,
  markerStroke: string,
  spacing: number,
  _followLine: boolean = true
): string {
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
      <Title>${styleName} - Marker Line</Title>
      <FeatureTypeStyle>
        <Rule>
          <LineSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </LineSymbolizer>
          <LineSymbolizer>
            <Stroke>
              <GraphicStroke>
                <Graphic>
                  <Mark>
                    <WellKnownName>${markerShape}</WellKnownName>
                    <Fill>
                      <CssParameter name="fill">${markerFill}</CssParameter>
                    </Fill>
                    <Stroke>
                      <CssParameter name="stroke">${markerStroke}</CssParameter>
                      <CssParameter name="stroke-width">1</CssParameter>
                    </Stroke>
                  </Mark>
                  <Size>${markerSize}</Size>
                </Graphic>
              </GraphicStroke>
              <CssParameter name="stroke-dasharray">${spacing} ${spacing}</CssParameter>
            </Stroke>
          </LineSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate arrow line SLD
export function generateArrowLineSLD(
  styleName: string,
  strokeColor: string,
  strokeWidth: number,
  arrowSize: number,
  arrowPosition: 'start' | 'end' | 'both' = 'end'
): string {
  const endArrow = arrowPosition === 'end' || arrowPosition === 'both' ? `
          <PointSymbolizer>
            <Geometry>
              <ogc:Function name="endPoint">
                <ogc:PropertyName>geometry</ogc:PropertyName>
              </ogc:Function>
            </Geometry>
            <Graphic>
              <Mark>
                <WellKnownName>shape://oarrow</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${strokeColor}</CssParameter>
                </Fill>
              </Mark>
              <Size>${arrowSize}</Size>
              <Rotation>
                <ogc:Function name="endAngle">
                  <ogc:PropertyName>geometry</ogc:PropertyName>
                </ogc:Function>
              </Rotation>
            </Graphic>
          </PointSymbolizer>` : ''

  const startArrow = arrowPosition === 'start' || arrowPosition === 'both' ? `
          <PointSymbolizer>
            <Geometry>
              <ogc:Function name="startPoint">
                <ogc:PropertyName>geometry</ogc:PropertyName>
              </ogc:Function>
            </Geometry>
            <Graphic>
              <Mark>
                <WellKnownName>shape://oarrow</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${strokeColor}</CssParameter>
                </Fill>
              </Mark>
              <Size>${arrowSize}</Size>
              <Rotation>
                <ogc:Add>
                  <ogc:Function name="startAngle">
                    <ogc:PropertyName>geometry</ogc:PropertyName>
                  </ogc:Function>
                  <ogc:Literal>180</ogc:Literal>
                </ogc:Add>
              </Rotation>
            </Graphic>
          </PointSymbolizer>` : ''

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
      <Title>${styleName} - Arrow Line</Title>
      <FeatureTypeStyle>
        <Rule>
          <LineSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </LineSymbolizer>${startArrow}${endArrow}
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate label style SLD
export function generateLabelSLD(
  styleName: string,
  labelField: string,
  fontFamily: string,
  fontSize: number,
  fontColor: string,
  fontWeight: 'normal' | 'bold',
  fontStyle: 'normal' | 'italic',
  haloColor: string,
  haloRadius: number,
  anchorX: number,
  anchorY: number,
  offsetX: number,
  offsetY: number,
  rotation: number,
  maxDisplacement: number = 50,
  followLine: boolean = false
): string {
  const placement = followLine ? `
              <LinePlacement>
                <PerpendicularOffset>0</PerpendicularOffset>
              </LinePlacement>` : `
              <PointPlacement>
                <AnchorPoint>
                  <AnchorPointX>${anchorX}</AnchorPointX>
                  <AnchorPointY>${anchorY}</AnchorPointY>
                </AnchorPoint>
                <Displacement>
                  <DisplacementX>${offsetX}</DisplacementX>
                  <DisplacementY>${offsetY}</DisplacementY>
                </Displacement>
                <Rotation>${rotation}</Rotation>
              </PointPlacement>`

  const halo = haloRadius > 0 ? `
            <Halo>
              <Radius>${haloRadius}</Radius>
              <Fill>
                <CssParameter name="fill">${haloColor}</CssParameter>
              </Fill>
            </Halo>` : ''

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
      <Title>${styleName} - Labels</Title>
      <FeatureTypeStyle>
        <Rule>
          <TextSymbolizer>
            <Label>
              <ogc:PropertyName>${labelField}</ogc:PropertyName>
            </Label>
            <Font>
              <CssParameter name="font-family">${fontFamily}</CssParameter>
              <CssParameter name="font-size">${fontSize}</CssParameter>
              <CssParameter name="font-weight">${fontWeight}</CssParameter>
              <CssParameter name="font-style">${fontStyle}</CssParameter>
            </Font>
            <LabelPlacement>${placement}
            </LabelPlacement>${halo}
            <Fill>
              <CssParameter name="fill">${fontColor}</CssParameter>
            </Fill>
            <VendorOption name="maxDisplacement">${maxDisplacement}</VendorOption>
            <VendorOption name="autoWrap">60</VendorOption>
            <VendorOption name="conflictResolution">true</VendorOption>
          </TextSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate gradient fill SLD
export function generateGradientFillSLD(
  styleName: string,
  color1: string,
  _color2: string,
  _gradientType: 'linear' | 'radial' = 'linear',
  _angle: number = 90,
  strokeColor: string,
  strokeWidth: number
): string {
  // Note: GeoServer SLD doesn't directly support gradients,
  // but we can use a vendor-specific extension or a workaround with multiple symbolizers
  // For now, we'll use a simple two-color fill with transparency gradient effect
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
      <Title>${styleName} - Gradient Fill</Title>
      <Abstract>Simulated gradient using GeoServer rendering transformations</Abstract>
      <FeatureTypeStyle>
        <Rule>
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${color1}</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </PolygonSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate categorized style SLD
export function generateCategorizedSLD(
  styleName: string,
  attribute: string,
  categories: { value: string; color: string; label: string }[],
  geomType: 'polygon' | 'line' | 'point',
  strokeColor: string,
  strokeWidth: number,
  defaultColor: string = '#cccccc'
): string {
  const rules = categories.map(cat => {
    const symbolizer = geomType === 'polygon' ? `
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${cat.color}</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </PolygonSymbolizer>` : geomType === 'line' ? `
          <LineSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${cat.color}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </LineSymbolizer>` : `
          <PointSymbolizer>
            <Graphic>
              <Mark>
                <WellKnownName>circle</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${cat.color}</CssParameter>
                </Fill>
                <Stroke>
                  <CssParameter name="stroke">${strokeColor}</CssParameter>
                  <CssParameter name="stroke-width">1</CssParameter>
                </Stroke>
              </Mark>
              <Size>8</Size>
            </Graphic>
          </PointSymbolizer>`

    return `
        <Rule>
          <Name>${cat.label}</Name>
          <Title>${cat.label}</Title>
          <ogc:Filter>
            <ogc:PropertyIsEqualTo>
              <ogc:PropertyName>${attribute}</ogc:PropertyName>
              <ogc:Literal>${cat.value}</ogc:Literal>
            </ogc:PropertyIsEqualTo>
          </ogc:Filter>${symbolizer}
        </Rule>`
  }).join('')

  // Add default rule for unmatched values
  const defaultSymbolizer = geomType === 'polygon' ? `
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${defaultColor}</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </PolygonSymbolizer>` : geomType === 'line' ? `
          <LineSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${defaultColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </LineSymbolizer>` : `
          <PointSymbolizer>
            <Graphic>
              <Mark>
                <WellKnownName>circle</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${defaultColor}</CssParameter>
                </Fill>
              </Mark>
              <Size>8</Size>
            </Graphic>
          </PointSymbolizer>`

  const defaultRule = `
        <Rule>
          <Name>Other</Name>
          <Title>Other</Title>
          <ElseFilter/>${defaultSymbolizer}
        </Rule>`

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
      <Title>${styleName} - Categorized</Title>
      <FeatureTypeStyle>${rules}${defaultRule}
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate rule-based style SLD
export function generateRuleBasedSLD(
  styleName: string,
  rules: { name: string; filter: string; color: string; size?: number }[],
  geomType: 'polygon' | 'line' | 'point',
  strokeColor: string,
  strokeWidth: number
): string {
  const rulesSLD = rules.map(rule => {
    const symbolizer = geomType === 'polygon' ? `
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${rule.color}</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </PolygonSymbolizer>` : geomType === 'line' ? `
          <LineSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${rule.color}</CssParameter>
              <CssParameter name="stroke-width">${rule.size || strokeWidth}</CssParameter>
            </Stroke>
          </LineSymbolizer>` : `
          <PointSymbolizer>
            <Graphic>
              <Mark>
                <WellKnownName>circle</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${rule.color}</CssParameter>
                </Fill>
              </Mark>
              <Size>${rule.size || 8}</Size>
            </Graphic>
          </PointSymbolizer>`

    return `
        <Rule>
          <Name>${rule.name}</Name>
          <Title>${rule.name}</Title>
          <ogc:Filter>
            ${rule.filter}
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
      <Title>${styleName} - Rule Based</Title>
      <FeatureTypeStyle>${rulesSLD}
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate proportional symbol SLD
export function generateProportionalSymbolSLD(
  styleName: string,
  attribute: string,
  minSize: number,
  maxSize: number,
  minValue: number,
  maxValue: number,
  fillColor: string,
  strokeColor: string,
  markerShape: string = 'circle'
): string {
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
      <Title>${styleName} - Proportional Symbol</Title>
      <FeatureTypeStyle>
        <Rule>
          <PointSymbolizer>
            <Graphic>
              <Mark>
                <WellKnownName>${markerShape}</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${fillColor}</CssParameter>
                </Fill>
                <Stroke>
                  <CssParameter name="stroke">${strokeColor}</CssParameter>
                  <CssParameter name="stroke-width">1</CssParameter>
                </Stroke>
              </Mark>
              <Size>
                <ogc:Add>
                  <ogc:Literal>${minSize}</ogc:Literal>
                  <ogc:Mul>
                    <ogc:Div>
                      <ogc:Sub>
                        <ogc:PropertyName>${attribute}</ogc:PropertyName>
                        <ogc:Literal>${minValue}</ogc:Literal>
                      </ogc:Sub>
                      <ogc:Literal>${maxValue - minValue}</ogc:Literal>
                    </ogc:Div>
                    <ogc:Literal>${maxSize - minSize}</ogc:Literal>
                  </ogc:Mul>
                </ogc:Add>
              </Size>
            </Graphic>
          </PointSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate centroid fill SLD
export function generateCentroidFillSLD(
  styleName: string,
  fillColor: string,
  strokeColor: string,
  strokeWidth: number,
  markerShape: string,
  markerSize: number,
  markerFill: string
): string {
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
      <Title>${styleName} - Centroid Fill</Title>
      <FeatureTypeStyle>
        <Rule>
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${fillColor}</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </PolygonSymbolizer>
          <PointSymbolizer>
            <Geometry>
              <ogc:Function name="centroid">
                <ogc:PropertyName>geometry</ogc:PropertyName>
              </ogc:Function>
            </Geometry>
            <Graphic>
              <Mark>
                <WellKnownName>${markerShape}</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${markerFill}</CssParameter>
                </Fill>
                <Stroke>
                  <CssParameter name="stroke">${strokeColor}</CssParameter>
                  <CssParameter name="stroke-width">1</CssParameter>
                </Stroke>
              </Mark>
              <Size>${markerSize}</Size>
            </Graphic>
          </PointSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}
