// Generate raster color map SLD
export function generateRasterColorMapSLD(
  styleName: string,
  colorRamp: string[],
  minValue: number,
  maxValue: number,
  colorMapType: 'ramp' | 'intervals' | 'values' = 'ramp',
  opacity: number = 1
): string {
  const range = maxValue - minValue
  const colorMapEntries = colorRamp.map((color, i) => {
    const quantity = minValue + (range * i) / (colorRamp.length - 1)
    return `          <ColorMapEntry color="${color}" quantity="${quantity.toFixed(2)}" opacity="${opacity}" />`
  }).join('\n')

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
      <Title>${styleName} - Raster Color Map</Title>
      <FeatureTypeStyle>
        <Rule>
          <RasterSymbolizer>
            <Opacity>${opacity}</Opacity>
            <ColorMap type="${colorMapType}">
${colorMapEntries}
            </ColorMap>
          </RasterSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate hillshade SLD
export function generateHillshadeSLD(
  styleName: string,
  azimuth: number = 315,
  altitude: number = 45,
  zFactor: number = 1,
  opacity: number = 1
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
      <Title>${styleName} - Hillshade</Title>
      <FeatureTypeStyle>
        <Rule>
          <RasterSymbolizer>
            <Opacity>${opacity}</Opacity>
            <ShadedRelief>
              <BrightnessOnly>false</BrightnessOnly>
              <ReliefFactor>${zFactor}</ReliefFactor>
            </ShadedRelief>
            <VendorOption name="algorithm">zevenbergenThorne</VendorOption>
            <VendorOption name="azimuth">${azimuth}</VendorOption>
            <VendorOption name="altitude">${altitude}</VendorOption>
          </RasterSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate combined hillshade + color ramp SLD
export function generateHillshadeWithColorSLD(
  styleName: string,
  colorRamp: string[],
  minValue: number,
  maxValue: number,
  azimuth: number = 315,
  altitude: number = 45,
  zFactor: number = 1,
  opacity: number = 1
): string {
  const range = maxValue - minValue
  const colorMapEntries = colorRamp.map((color, i) => {
    const quantity = minValue + (range * i) / (colorRamp.length - 1)
    return `          <ColorMapEntry color="${color}" quantity="${quantity.toFixed(2)}" />`
  }).join('\n')

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
      <Title>${styleName} - Hillshade with Color</Title>
      <FeatureTypeStyle>
        <Rule>
          <RasterSymbolizer>
            <Opacity>${opacity}</Opacity>
            <ColorMap type="ramp">
${colorMapEntries}
            </ColorMap>
            <ShadedRelief>
              <BrightnessOnly>false</BrightnessOnly>
              <ReliefFactor>${zFactor}</ReliefFactor>
            </ShadedRelief>
            <VendorOption name="algorithm">zevenbergenThorne</VendorOption>
            <VendorOption name="azimuth">${azimuth}</VendorOption>
            <VendorOption name="altitude">${altitude}</VendorOption>
          </RasterSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate contrast enhancement SLD
export function generateContrastEnhancementSLD(
  styleName: string,
  method: 'normalize' | 'histogram' | 'none' = 'normalize',
  gammaValue: number = 1.0,
  opacity: number = 1
): string {
  const contrastEnhancement = method === 'none' ? '' : `
            <ContrastEnhancement>
              <${method === 'normalize' ? 'Normalize' : 'Histogram'} />
              <GammaValue>${gammaValue}</GammaValue>
            </ContrastEnhancement>`

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
      <Title>${styleName} - Enhanced</Title>
      <FeatureTypeStyle>
        <Rule>
          <RasterSymbolizer>
            <Opacity>${opacity}</Opacity>${contrastEnhancement}
          </RasterSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Generate raster contour SLD
export function generateRasterContourSLD(
  styleName: string,
  interval: number,
  strokeColor: string,
  strokeWidth: number,
  _majorInterval: number = 5,
  _majorStrokeWidth: number = 2,
  labelContours: boolean = true
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
      <Title>${styleName} - Contours</Title>
      <FeatureTypeStyle>
        <Transformation>
          <ogc:Function name="ras:Contour">
            <ogc:Function name="parameter">
              <ogc:Literal>data</ogc:Literal>
            </ogc:Function>
            <ogc:Function name="parameter">
              <ogc:Literal>levels</ogc:Literal>
              <ogc:Literal>${interval}</ogc:Literal>
            </ogc:Function>
          </ogc:Function>
        </Transformation>
        <Rule>
          <Name>Contour</Name>
          <LineSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${strokeColor}</CssParameter>
              <CssParameter name="stroke-width">${strokeWidth}</CssParameter>
            </Stroke>
          </LineSymbolizer>${labelContours ? `
          <TextSymbolizer>
            <Label>
              <ogc:PropertyName>value</ogc:PropertyName>
            </Label>
            <Font>
              <CssParameter name="font-family">Arial</CssParameter>
              <CssParameter name="font-size">10</CssParameter>
            </Font>
            <LabelPlacement>
              <LinePlacement/>
            </LabelPlacement>
            <Fill>
              <CssParameter name="fill">${strokeColor}</CssParameter>
            </Fill>
            <VendorOption name="followLine">true</VendorOption>
            <VendorOption name="maxAngleDelta">90</VendorOption>
            <VendorOption name="repeat">200</VendorOption>
          </TextSymbolizer>` : ''}
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}
