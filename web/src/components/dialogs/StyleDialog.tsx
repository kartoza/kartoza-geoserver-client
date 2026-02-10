import { useState, useEffect, useCallback, useMemo } from 'react'
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Button,
  Box,
  Flex,
  VStack,
  HStack,
  FormControl,
  FormLabel,
  Input,
  Select,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Text,
  Icon,
  Badge,
  useToast,
  useColorModeValue,
  IconButton,
  Divider,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Slider,
  SliderTrack,
  SliderFilledTrack,
  SliderThumb,
  Alert,
  AlertIcon,
  Spinner,
  Collapse,
  useDisclosure,
} from '@chakra-ui/react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import CodeMirror from '@uiw/react-codemirror'
import { xml } from '@codemirror/lang-xml'
import { css } from '@codemirror/lang-css'
import {
  FiDroplet,
  FiCode,
  FiEye,
  FiSave,
  FiSquare,
  FiCircle,
  FiMinus,
  FiGrid,
  FiChevronDown,
  FiChevronUp,
  FiImage,
} from 'react-icons/fi'
import { useUIStore } from '../../stores/uiStore'
import * as api from '../../api/client'

// Default SLD template for new styles
const DEFAULT_SLD = `<?xml version="1.0" encoding="UTF-8"?>
<StyledLayerDescriptor version="1.0.0"
  xsi:schemaLocation="http://www.opengis.net/sld StyledLayerDescriptor.xsd"
  xmlns="http://www.opengis.net/sld"
  xmlns:ogc="http://www.opengis.net/ogc"
  xmlns:xlink="http://www.w3.org/1999/xlink"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <NamedLayer>
    <Name>NewStyle</Name>
    <UserStyle>
      <Title>New Style</Title>
      <FeatureTypeStyle>
        <Rule>
          <Name>Default</Name>
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">#3388ff</CssParameter>
              <CssParameter name="fill-opacity">0.6</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">#2266cc</CssParameter>
              <CssParameter name="stroke-width">1</CssParameter>
            </Stroke>
          </PolygonSymbolizer>
        </Rule>
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`

// Default CSS template for new styles
const DEFAULT_CSS = `/* GeoServer CSS Style */
* {
  fill: #3388ff;
  fill-opacity: 0.6;
  stroke: #2266cc;
  stroke-width: 1;
}`

// Classification methods
type ClassificationMethod = 'equal-interval' | 'quantile' | 'jenks' | 'pretty'

// Color ramps for classification
const COLOR_RAMPS: Record<string, string[]> = {
  'blue-to-red': ['#2166ac', '#67a9cf', '#d1e5f0', '#fddbc7', '#ef8a62', '#b2182b'],
  'green-to-red': ['#1a9850', '#91cf60', '#d9ef8b', '#fee08b', '#fc8d59', '#d73027'],
  'viridis': ['#440154', '#443983', '#31688e', '#21918c', '#35b779', '#fde725'],
  'spectral': ['#9e0142', '#d53e4f', '#f46d43', '#fdae61', '#fee08b', '#e6f598', '#abdda4', '#66c2a5', '#3288bd', '#5e4fa2'],
  'blues': ['#f7fbff', '#deebf7', '#c6dbef', '#9ecae1', '#6baed6', '#4292c6', '#2171b5', '#084594'],
  'reds': ['#fff5f0', '#fee0d2', '#fcbba1', '#fc9272', '#fb6a4a', '#ef3b2c', '#cb181d', '#99000d'],
  'greens': ['#f7fcf5', '#e5f5e0', '#c7e9c0', '#a1d99b', '#74c476', '#41ab5d', '#238b45', '#005a32'],
}

// Raster-specific color ramps with more gradient options
const RASTER_COLOR_RAMPS: Record<string, { name: string; colors: string[]; description: string }> = {
  'rainbow': {
    name: 'Rainbow',
    description: 'Classic rainbow spectrum',
    colors: ['#9400D3', '#4B0082', '#0000FF', '#00FF00', '#FFFF00', '#FF7F00', '#FF0000'],
  },
  'terrain': {
    name: 'Terrain',
    description: 'Natural terrain colors (blue-green-brown)',
    colors: ['#0000FF', '#00FFFF', '#00FF00', '#FFFF00', '#FF8C00', '#8B4513', '#FFFFFF'],
  },
  'elevation': {
    name: 'Elevation',
    description: 'DEM/elevation visualization',
    colors: ['#006400', '#228B22', '#90EE90', '#FFFF00', '#FFA500', '#8B4513', '#FFFFFF'],
  },
  'temperature': {
    name: 'Temperature',
    description: 'Cold to hot temperature scale',
    colors: ['#0000FF', '#00BFFF', '#00FFFF', '#FFFF00', '#FFA500', '#FF0000', '#8B0000'],
  },
  'ndvi': {
    name: 'NDVI',
    description: 'Vegetation index (brown-yellow-green)',
    colors: ['#8B4513', '#D2691E', '#FFD700', '#ADFF2F', '#32CD32', '#006400'],
  },
  'bathymetry': {
    name: 'Bathymetry',
    description: 'Ocean depth visualization',
    colors: ['#000033', '#000066', '#0000CC', '#0066FF', '#00CCFF', '#99FFFF'],
  },
  'grayscale': {
    name: 'Grayscale',
    description: 'Black to white gradient',
    colors: ['#000000', '#333333', '#666666', '#999999', '#CCCCCC', '#FFFFFF'],
  },
  'magma': {
    name: 'Magma',
    description: 'Dark purple to bright yellow',
    colors: ['#000004', '#3B0F70', '#8C2981', '#DE4968', '#FE9F6D', '#FCFDBF'],
  },
  'plasma': {
    name: 'Plasma',
    description: 'Blue-purple to yellow',
    colors: ['#0D0887', '#6A00A8', '#B12A90', '#E16462', '#FCA636', '#F0F921'],
  },
  'inferno': {
    name: 'Inferno',
    description: 'Black through red to bright yellow',
    colors: ['#000004', '#420A68', '#932667', '#DD513A', '#FCA50A', '#FCFFA4'],
  },
}

// Hillshade presets
const HILLSHADE_PRESETS = [
  { name: 'Default', azimuth: 315, altitude: 45, zFactor: 1 },
  { name: 'Morning Light', azimuth: 90, altitude: 30, zFactor: 1 },
  { name: 'Evening Light', azimuth: 270, altitude: 30, zFactor: 1 },
  { name: 'High Contrast', azimuth: 315, altitude: 60, zFactor: 2 },
  { name: 'Subtle', azimuth: 315, altitude: 45, zFactor: 0.5 },
]

// Generate raster color map SLD
function generateRasterColorMapSLD(
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
function generateHillshadeSLD(
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
function generateHillshadeWithColorSLD(
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
function generateContrastEnhancementSLD(
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

/* eslint-disable @typescript-eslint/no-unused-vars */
// ============================================
// QGIS-LIKE MARKER SHAPES
// ============================================
const MARKER_SHAPES = [
  { name: 'circle', label: 'Circle', wellKnownName: 'circle' },
  { name: 'square', label: 'Square', wellKnownName: 'square' },
  { name: 'triangle', label: 'Triangle', wellKnownName: 'triangle' },
  { name: 'star', label: 'Star', wellKnownName: 'star' },
  { name: 'cross', label: 'Cross', wellKnownName: 'cross' },
  { name: 'x', label: 'X', wellKnownName: 'x' },
  { name: 'diamond', label: 'Diamond', wellKnownName: 'shape://vertline' },
  { name: 'pentagon', label: 'Pentagon', wellKnownName: 'pentagon' },
  { name: 'hexagon', label: 'Hexagon', wellKnownName: 'hexagon' },
  { name: 'octagon', label: 'Octagon', wellKnownName: 'octagon' },
  { name: 'arrow', label: 'Arrow', wellKnownName: 'shape://oarrow' },
  { name: 'carrow', label: 'Closed Arrow', wellKnownName: 'shape://carrow' },
]

// ============================================
// LINE DASH PATTERNS (QGIS-like)
// ============================================
const LINE_DASH_PATTERNS = [
  { name: 'solid', label: 'Solid', dashArray: '' },
  { name: 'dash', label: 'Dash', dashArray: '10 5' },
  { name: 'dot', label: 'Dot', dashArray: '2 5' },
  { name: 'dash-dot', label: 'Dash Dot', dashArray: '10 5 2 5' },
  { name: 'dash-dot-dot', label: 'Dash Dot Dot', dashArray: '10 5 2 5 2 5' },
  { name: 'long-dash', label: 'Long Dash', dashArray: '20 10' },
  { name: 'short-dash', label: 'Short Dash', dashArray: '5 5' },
  { name: 'dense-dot', label: 'Dense Dot', dashArray: '1 2' },
]

// ============================================
// LINE CAP AND JOIN STYLES
// ============================================
const LINE_CAP_STYLES = [
  { name: 'butt', label: 'Flat' },
  { name: 'round', label: 'Round' },
  { name: 'square', label: 'Square' },
]

const LINE_JOIN_STYLES = [
  { name: 'miter', label: 'Miter' },
  { name: 'round', label: 'Round' },
  { name: 'bevel', label: 'Bevel' },
]

// ============================================
// FILL PATTERNS (QGIS-like graphic fills)
// ============================================
const FILL_PATTERNS = [
  { name: 'solid', label: 'Solid Fill', type: 'solid' },
  { name: 'horizontal', label: 'Horizontal Lines', type: 'hatch', angle: 0 },
  { name: 'vertical', label: 'Vertical Lines', type: 'hatch', angle: 90 },
  { name: 'cross', label: 'Cross Hatch', type: 'hatch', angle: 0, double: true },
  { name: 'forward-diagonal', label: 'Forward Diagonal', type: 'hatch', angle: 45 },
  { name: 'backward-diagonal', label: 'Backward Diagonal', type: 'hatch', angle: 135 },
  { name: 'diagonal-cross', label: 'Diagonal Cross', type: 'hatch', angle: 45, double: true },
  { name: 'dot-grid', label: 'Dot Grid', type: 'point-pattern' },
  { name: 'dense-dot', label: 'Dense Dots', type: 'point-pattern', spacing: 4 },
  { name: 'sparse-dot', label: 'Sparse Dots', type: 'point-pattern', spacing: 12 },
]

// ============================================
// BLEND MODES (supported by GeoServer)
// ============================================
const BLEND_MODES = [
  { name: 'normal', label: 'Normal' },
  { name: 'multiply', label: 'Multiply' },
  { name: 'screen', label: 'Screen' },
  { name: 'overlay', label: 'Overlay' },
  { name: 'darken', label: 'Darken' },
  { name: 'lighten', label: 'Lighten' },
  { name: 'color-dodge', label: 'Color Dodge' },
  { name: 'color-burn', label: 'Color Burn' },
  { name: 'hard-light', label: 'Hard Light' },
  { name: 'soft-light', label: 'Soft Light' },
  { name: 'difference', label: 'Difference' },
  { name: 'exclusion', label: 'Exclusion' },
]

// ============================================
// FONT MARKER FONTS
// ============================================
const _FONT_MARKER_FONTS = [
  { name: 'Wingdings', characters: ['✈', '★', '♦', '♣', '♠', '♥', '☎', '✉', '✂', '✓', '✗'] },
  { name: 'Webdings', characters: ['⌂', '⌘', '⚙', '⚡', '⚠', '⚑', '⚐', '☀', '☁', '☂', '☃'] },
  { name: 'Symbol', characters: ['α', 'β', 'γ', 'δ', 'ε', 'π', 'Σ', 'Ω', '∞', '≈', '≠'] },
]

// ============================================
// LABEL PLACEMENT OPTIONS
// ============================================
const _LABEL_PLACEMENT_OPTIONS = [
  { name: 'point', label: 'Point on Point' },
  { name: 'line', label: 'Along Line' },
  { name: 'polygon', label: 'Inside Polygon' },
]

const _LABEL_ANCHOR_POINTS = [
  { x: 0, y: 0, label: 'Top Left' },
  { x: 0.5, y: 0, label: 'Top Center' },
  { x: 1, y: 0, label: 'Top Right' },
  { x: 0, y: 0.5, label: 'Middle Left' },
  { x: 0.5, y: 0.5, label: 'Center' },
  { x: 1, y: 0.5, label: 'Middle Right' },
  { x: 0, y: 1, label: 'Bottom Left' },
  { x: 0.5, y: 1, label: 'Bottom Center' },
  { x: 1, y: 1, label: 'Bottom Right' },
]
/* eslint-enable @typescript-eslint/no-unused-vars */

// Export for future use
export {
  _FONT_MARKER_FONTS,
  _LABEL_PLACEMENT_OPTIONS,
  _LABEL_ANCHOR_POINTS,
  MARKER_SHAPES,
  LINE_DASH_PATTERNS,
  LINE_CAP_STYLES,
  LINE_JOIN_STYLES,
  FILL_PATTERNS,
  BLEND_MODES,
}

// ============================================
// SLD GENERATORS FOR ADVANCED FEATURES
// These generators are available for use by the UI wizards
// ============================================
/* eslint-disable @typescript-eslint/no-unused-vars */

// Generate hatch pattern fill SLD
function generateHatchFillSLD(
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
function generatePointPatternFillSLD(
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
function generateMarkerLineSLD(
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
function generateArrowLineSLD(
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
function generateLabelSLD(
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
function generateGradientFillSLD(
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
function generateCategorizedSLD(
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
function generateRuleBasedSLD(
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
function generateProportionalSymbolSLD(
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
function generateCentroidFillSLD(
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

// Generate raster contour SLD
function generateRasterContourSLD(
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
// Export SLD generators for external use
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
  generateRasterContourSLD,
}
/* eslint-enable @typescript-eslint/no-unused-vars */

// Calculate equal interval breaks
function equalIntervalBreaks(values: number[], numClasses: number): number[] {
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
function quantileBreaks(values: number[], numClasses: number): number[] {
  const sorted = [...values].sort((a, b) => a - b)
  const breaks: number[] = []
  for (let i = 0; i <= numClasses; i++) {
    const index = Math.floor((i / numClasses) * (sorted.length - 1))
    breaks.push(sorted[index])
  }
  return breaks
}

// Calculate Jenks natural breaks (simplified Ckmeans approximation)
function jenksBreaks(values: number[], numClasses: number): number[] {
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
function prettyBreaks(values: number[], numClasses: number): number[] {
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
function calculateBreaks(values: number[], numClasses: number, method: ClassificationMethod): number[] {
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
function interpolateColors(ramp: string[], numColors: number): string[] {
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
function generateClassifiedSLD(
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

// Beautiful point style presets
interface PointStylePreset {
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

const POINT_STYLE_PRESETS: PointStylePreset[] = [
  // Classic pins
  {
    name: 'Ocean Drop',
    description: 'Deep blue with soft glow',
    icon: '💧',
    shape: 'circle',
    fill: '#0077be',
    fillOpacity: 0.9,
    stroke: '#004d80',
    strokeWidth: 2,
    size: 12,
    haloColor: '#87ceeb',
    haloRadius: 3,
  },
  {
    name: 'Sunset Glow',
    description: 'Warm orange with golden halo',
    icon: '🌅',
    shape: 'circle',
    fill: '#ff6b35',
    fillOpacity: 0.95,
    stroke: '#cc4400',
    strokeWidth: 2,
    size: 14,
    haloColor: '#ffd700',
    haloRadius: 4,
  },
  {
    name: 'Forest Emerald',
    description: 'Rich green nature marker',
    icon: '🌲',
    shape: 'circle',
    fill: '#228b22',
    fillOpacity: 0.9,
    stroke: '#145214',
    strokeWidth: 2,
    size: 12,
    haloColor: '#90ee90',
    haloRadius: 3,
  },
  {
    name: 'Royal Purple',
    description: 'Elegant purple with silver edge',
    icon: '👑',
    shape: 'circle',
    fill: '#8b008b',
    fillOpacity: 0.9,
    stroke: '#c0c0c0',
    strokeWidth: 2,
    size: 12,
    haloColor: '#dda0dd',
    haloRadius: 3,
  },
  {
    name: 'Cherry Blossom',
    description: 'Soft pink with delicate outline',
    icon: '🌸',
    shape: 'circle',
    fill: '#ffb7c5',
    fillOpacity: 0.95,
    stroke: '#ff69b4',
    strokeWidth: 1.5,
    size: 11,
    haloColor: '#fff0f5',
    haloRadius: 2,
  },
  // Geometric shapes
  {
    name: 'Diamond Marker',
    description: 'Classic diamond shape',
    icon: '💎',
    shape: 'square',
    fill: '#00bfff',
    fillOpacity: 0.85,
    stroke: '#0080ff',
    strokeWidth: 2,
    size: 14,
    rotation: 45,
  },
  {
    name: 'Golden Star',
    description: 'Bright star for highlights',
    icon: '⭐',
    shape: 'star',
    fill: '#ffd700',
    fillOpacity: 1,
    stroke: '#b8860b',
    strokeWidth: 1.5,
    size: 16,
  },
  {
    name: 'Ruby Triangle',
    description: 'Bold triangular marker',
    icon: '🔺',
    shape: 'triangle',
    fill: '#dc143c',
    fillOpacity: 0.9,
    stroke: '#8b0000',
    strokeWidth: 2,
    size: 14,
  },
  {
    name: 'Navy Square',
    description: 'Professional square marker',
    icon: '🔷',
    shape: 'square',
    fill: '#1e3a5f',
    fillOpacity: 0.95,
    stroke: '#0d1b2a',
    strokeWidth: 2,
    size: 12,
  },
  // Special effects
  {
    name: 'Neon Pulse',
    description: 'Bright neon with glow effect',
    icon: '💡',
    shape: 'circle',
    fill: '#00ff88',
    fillOpacity: 1,
    stroke: '#00cc6a',
    strokeWidth: 1,
    size: 10,
    haloColor: '#00ff88',
    haloRadius: 6,
  },
  {
    name: 'Fire Ember',
    description: 'Hot red with orange glow',
    icon: '🔥',
    shape: 'circle',
    fill: '#ff4500',
    fillOpacity: 0.95,
    stroke: '#8b0000',
    strokeWidth: 2,
    size: 12,
    haloColor: '#ff8c00',
    haloRadius: 4,
  },
  {
    name: 'Ice Crystal',
    description: 'Cool blue with frost effect',
    icon: '❄️',
    shape: 'star',
    fill: '#e0ffff',
    fillOpacity: 0.9,
    stroke: '#00ced1',
    strokeWidth: 1.5,
    size: 14,
    haloColor: '#b0e0e6',
    haloRadius: 3,
  },
  {
    name: 'Midnight Shadow',
    description: 'Dark with subtle shadow',
    icon: '🌙',
    shape: 'circle',
    fill: '#2c3e50',
    fillOpacity: 0.95,
    stroke: '#1a252f',
    strokeWidth: 2,
    size: 12,
    haloColor: '#34495e',
    haloRadius: 4,
  },
  {
    name: 'Coral Reef',
    description: 'Vibrant coral with aqua edge',
    icon: '🐠',
    shape: 'circle',
    fill: '#ff7f50',
    fillOpacity: 0.9,
    stroke: '#20b2aa',
    strokeWidth: 2,
    size: 12,
    haloColor: '#40e0d0',
    haloRadius: 3,
  },
  // Professional markers
  {
    name: 'Map Pin Red',
    description: 'Classic red location pin',
    icon: '📍',
    shape: 'circle',
    fill: '#e74c3c',
    fillOpacity: 1,
    stroke: '#c0392b',
    strokeWidth: 2,
    size: 14,
    haloColor: '#ffffff',
    haloRadius: 2,
  },
  {
    name: 'Corporate Blue',
    description: 'Clean professional blue',
    icon: '🏢',
    shape: 'circle',
    fill: '#3498db',
    fillOpacity: 0.95,
    stroke: '#2980b9',
    strokeWidth: 2,
    size: 12,
  },
  {
    name: 'Success Green',
    description: 'Positive indicator marker',
    icon: '✅',
    shape: 'circle',
    fill: '#27ae60',
    fillOpacity: 0.95,
    stroke: '#1e8449',
    strokeWidth: 2,
    size: 12,
  },
  {
    name: 'Warning Amber',
    description: 'Attention-grabbing yellow',
    icon: '⚠️',
    shape: 'triangle',
    fill: '#f39c12',
    fillOpacity: 1,
    stroke: '#d68910',
    strokeWidth: 2,
    size: 14,
  },
  {
    name: 'Cross Mark',
    description: 'X marker for exclusions',
    icon: '❌',
    shape: 'cross',
    fill: '#e74c3c',
    fillOpacity: 1,
    stroke: '#c0392b',
    strokeWidth: 2,
    size: 14,
  },
  {
    name: 'Target Bullseye',
    description: 'Precision crosshair marker',
    icon: '🎯',
    shape: 'cross',
    fill: '#2c3e50',
    fillOpacity: 1,
    stroke: '#e74c3c',
    strokeWidth: 3,
    size: 16,
  },
]

// Style rule interface for visual editor
interface StyleRule {
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

// Parse SLD to extract style rules for visual editing
function parseSLDRules(sldContent: string): StyleRule[] {
  const rules: StyleRule[] = []

  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(sldContent, 'text/xml')
    const ruleElements = doc.querySelectorAll('Rule')

    ruleElements.forEach((ruleEl, index) => {
      const nameEl = ruleEl.querySelector('Name')
      const rule: StyleRule = {
        name: nameEl?.textContent || `Rule ${index + 1}`,
        symbolizer: { type: 'polygon' }
      }

      // Parse PolygonSymbolizer
      const polySymb = ruleEl.querySelector('PolygonSymbolizer')
      if (polySymb) {
        rule.symbolizer.type = 'polygon'
        const fillParams = polySymb.querySelectorAll('Fill CssParameter')
        fillParams.forEach(param => {
          const name = param.getAttribute('name')
          if (name === 'fill') rule.symbolizer.fill = param.textContent || '#3388ff'
          if (name === 'fill-opacity') rule.symbolizer.fillOpacity = parseFloat(param.textContent || '1')
        })
        const strokeParams = polySymb.querySelectorAll('Stroke CssParameter')
        strokeParams.forEach(param => {
          const name = param.getAttribute('name')
          if (name === 'stroke') rule.symbolizer.stroke = param.textContent || '#2266cc'
          if (name === 'stroke-width') rule.symbolizer.strokeWidth = parseFloat(param.textContent || '1')
          if (name === 'stroke-opacity') rule.symbolizer.strokeOpacity = parseFloat(param.textContent || '1')
        })
      }

      // Parse LineSymbolizer
      const lineSymb = ruleEl.querySelector('LineSymbolizer')
      if (lineSymb) {
        rule.symbolizer.type = 'line'
        const strokeParams = lineSymb.querySelectorAll('Stroke CssParameter')
        strokeParams.forEach(param => {
          const name = param.getAttribute('name')
          if (name === 'stroke') rule.symbolizer.stroke = param.textContent || '#3388ff'
          if (name === 'stroke-width') rule.symbolizer.strokeWidth = parseFloat(param.textContent || '2')
          if (name === 'stroke-opacity') rule.symbolizer.strokeOpacity = parseFloat(param.textContent || '1')
        })
      }

      // Parse PointSymbolizer
      const pointSymb = ruleEl.querySelector('PointSymbolizer')
      if (pointSymb) {
        rule.symbolizer.type = 'point'
        rule.symbolizer.pointShape = 'circle'
        rule.symbolizer.pointSize = 8

        const fillParams = pointSymb.querySelectorAll('Fill CssParameter')
        fillParams.forEach(param => {
          const name = param.getAttribute('name')
          if (name === 'fill') rule.symbolizer.fill = param.textContent || '#3388ff'
          if (name === 'fill-opacity') rule.symbolizer.fillOpacity = parseFloat(param.textContent || '1')
        })
        const strokeParams = pointSymb.querySelectorAll('Stroke CssParameter')
        strokeParams.forEach(param => {
          const name = param.getAttribute('name')
          if (name === 'stroke') rule.symbolizer.stroke = param.textContent || '#2266cc'
          if (name === 'stroke-width') rule.symbolizer.strokeWidth = parseFloat(param.textContent || '1')
        })
        const sizeEl = pointSymb.querySelector('Size')
        if (sizeEl) rule.symbolizer.pointSize = parseFloat(sizeEl.textContent || '8')
      }

      rules.push(rule)
    })
  } catch (e) {
    console.error('Failed to parse SLD:', e)
  }

  return rules.length > 0 ? rules : [{
    name: 'Default',
    symbolizer: {
      type: 'polygon',
      fill: '#3388ff',
      fillOpacity: 0.6,
      stroke: '#2266cc',
      strokeWidth: 1,
    }
  }]
}

// Generate SLD from style rules
function generateSLD(styleName: string, rules: StyleRule[]): string {
  let rulesXml = ''

  rules.forEach(rule => {
    let symbolizerXml = ''

    if (rule.symbolizer.type === 'polygon') {
      symbolizerXml = `
          <PolygonSymbolizer>
            <Fill>
              <CssParameter name="fill">${rule.symbolizer.fill || '#3388ff'}</CssParameter>
              <CssParameter name="fill-opacity">${rule.symbolizer.fillOpacity ?? 0.6}</CssParameter>
            </Fill>
            <Stroke>
              <CssParameter name="stroke">${rule.symbolizer.stroke || '#2266cc'}</CssParameter>
              <CssParameter name="stroke-width">${rule.symbolizer.strokeWidth || 1}</CssParameter>
            </Stroke>
          </PolygonSymbolizer>`
    } else if (rule.symbolizer.type === 'line') {
      symbolizerXml = `
          <LineSymbolizer>
            <Stroke>
              <CssParameter name="stroke">${rule.symbolizer.stroke || '#3388ff'}</CssParameter>
              <CssParameter name="stroke-width">${rule.symbolizer.strokeWidth || 2}</CssParameter>
              <CssParameter name="stroke-opacity">${rule.symbolizer.strokeOpacity ?? 1}</CssParameter>
            </Stroke>
          </LineSymbolizer>`
    } else if (rule.symbolizer.type === 'point') {
      // Build optional rotation element
      const rotationXml = rule.symbolizer.rotation
        ? `\n              <Rotation>${rule.symbolizer.rotation}</Rotation>`
        : ''

      // Build the main point symbolizer
      let mainSymbolizer = `
          <PointSymbolizer>
            <Graphic>
              <Mark>
                <WellKnownName>${rule.symbolizer.pointShape || 'circle'}</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${rule.symbolizer.fill || '#3388ff'}</CssParameter>
                  <CssParameter name="fill-opacity">${rule.symbolizer.fillOpacity ?? 1}</CssParameter>
                </Fill>
                <Stroke>
                  <CssParameter name="stroke">${rule.symbolizer.stroke || '#2266cc'}</CssParameter>
                  <CssParameter name="stroke-width">${rule.symbolizer.strokeWidth || 1}</CssParameter>
                </Stroke>
              </Mark>
              <Size>${rule.symbolizer.pointSize || 8}</Size>${rotationXml}
            </Graphic>
          </PointSymbolizer>`

      // Add halo effect as an additional symbolizer behind the main one
      if (rule.symbolizer.haloColor && rule.symbolizer.haloRadius) {
        const haloSize = (rule.symbolizer.pointSize || 8) + (rule.symbolizer.haloRadius * 2)
        const haloSymbolizer = `
          <PointSymbolizer>
            <Graphic>
              <Mark>
                <WellKnownName>${rule.symbolizer.pointShape || 'circle'}</WellKnownName>
                <Fill>
                  <CssParameter name="fill">${rule.symbolizer.haloColor}</CssParameter>
                  <CssParameter name="fill-opacity">0.4</CssParameter>
                </Fill>
              </Mark>
              <Size>${haloSize}</Size>${rotationXml}
            </Graphic>
          </PointSymbolizer>`
        symbolizerXml = haloSymbolizer + mainSymbolizer
      } else {
        symbolizerXml = mainSymbolizer
      }
    }

    rulesXml += `
        <Rule>
          <Name>${rule.name}</Name>${symbolizerXml}
        </Rule>`
  })

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
      <Title>${styleName}</Title>
      <FeatureTypeStyle>${rulesXml}
      </FeatureTypeStyle>
    </UserStyle>
  </NamedLayer>
</StyledLayerDescriptor>`
}

// Color picker component
function ColorPicker({
  value,
  onChange,
  label
}: {
  value: string
  onChange: (color: string) => void
  label: string
}) {
  const borderColor = useColorModeValue('gray.200', 'gray.600')

  return (
    <FormControl>
      <FormLabel fontSize="sm">{label}</FormLabel>
      <HStack>
        <Box position="relative">
          <Box
            w="40px"
            h="40px"
            borderRadius="full"
            bg={value}
            border="2px solid"
            borderColor={borderColor}
            cursor="pointer"
            _hover={{ borderColor: 'kartoza.400' }}
            transition="border-color 0.2s"
          />
          <Input
            type="color"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            position="absolute"
            top={0}
            left={0}
            w="40px"
            h="40px"
            opacity={0}
            cursor="pointer"
          />
        </Box>
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          size="sm"
          w="100px"
          fontFamily="mono"
          borderColor={borderColor}
        />
      </HStack>
    </FormControl>
  )
}

// Visual rule editor component
function RuleEditor({
  rule,
  onChange,
  onDelete,
}: {
  rule: StyleRule
  onChange: (rule: StyleRule) => void
  onDelete: () => void
}) {
  const bgColor = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')
  const [showMorePresets, setShowMorePresets] = useState(false)

  const updateSymbolizer = (updates: Partial<StyleRule['symbolizer']>) => {
    onChange({
      ...rule,
      symbolizer: { ...rule.symbolizer, ...updates }
    })
  }

  return (
    <Box
      p={4}
      bg={bgColor}
      borderRadius="lg"
      border="1px solid"
      borderColor={borderColor}
    >
      <VStack spacing={4} align="stretch">
        <HStack justify="space-between">
          <FormControl maxW="200px">
            <FormLabel fontSize="sm">Rule Name</FormLabel>
            <Input
              size="sm"
              value={rule.name}
              onChange={(e) => onChange({ ...rule, name: e.target.value })}
            />
          </FormControl>
          <FormControl maxW="150px">
            <FormLabel fontSize="sm">Geometry Type</FormLabel>
            <Select
              size="sm"
              value={rule.symbolizer.type}
              onChange={(e) => updateSymbolizer({ type: e.target.value as 'polygon' | 'line' | 'point' })}
            >
              <option value="polygon">Polygon</option>
              <option value="line">Line</option>
              <option value="point">Point</option>
            </Select>
          </FormControl>
          <IconButton
            aria-label="Delete rule"
            icon={<FiMinus />}
            size="sm"
            colorScheme="red"
            variant="ghost"
            onClick={onDelete}
          />
        </HStack>

        <Divider />

        {/* Fill settings (for polygon and point) */}
        {(rule.symbolizer.type === 'polygon' || rule.symbolizer.type === 'point') && (
          <Box>
            <Text fontWeight="600" fontSize="sm" mb={2}>Fill</Text>
            <HStack spacing={4} wrap="wrap">
              <ColorPicker
                label="Color"
                value={rule.symbolizer.fill || '#3388ff'}
                onChange={(color) => updateSymbolizer({ fill: color })}
              />
              <FormControl maxW="150px">
                <FormLabel fontSize="sm">Opacity</FormLabel>
                <HStack>
                  <Slider
                    value={rule.symbolizer.fillOpacity ?? 1}
                    min={0}
                    max={1}
                    step={0.1}
                    onChange={(val) => updateSymbolizer({ fillOpacity: val })}
                  >
                    <SliderTrack>
                      <SliderFilledTrack bg="kartoza.500" />
                    </SliderTrack>
                    <SliderThumb />
                  </Slider>
                  <Text fontSize="sm" w="40px">{((rule.symbolizer.fillOpacity ?? 1) * 100).toFixed(0)}%</Text>
                </HStack>
              </FormControl>
            </HStack>
          </Box>
        )}

        {/* Stroke settings */}
        <Box>
          <Text fontWeight="600" fontSize="sm" mb={2}>Stroke</Text>
          <HStack spacing={4} wrap="wrap">
            <ColorPicker
              label="Color"
              value={rule.symbolizer.stroke || '#2266cc'}
              onChange={(color) => updateSymbolizer({ stroke: color })}
            />
            <FormControl maxW="100px">
              <FormLabel fontSize="sm">Width</FormLabel>
              <NumberInput
                size="sm"
                value={rule.symbolizer.strokeWidth || 1}
                min={0}
                max={20}
                step={0.5}
                onChange={(_, val) => updateSymbolizer({ strokeWidth: val })}
              >
                <NumberInputField />
                <NumberInputStepper>
                  <NumberIncrementStepper />
                  <NumberDecrementStepper />
                </NumberInputStepper>
              </NumberInput>
            </FormControl>
            {rule.symbolizer.type === 'line' && (
              <FormControl maxW="150px">
                <FormLabel fontSize="sm">Opacity</FormLabel>
                <HStack>
                  <Slider
                    value={rule.symbolizer.strokeOpacity ?? 1}
                    min={0}
                    max={1}
                    step={0.1}
                    onChange={(val) => updateSymbolizer({ strokeOpacity: val })}
                  >
                    <SliderTrack>
                      <SliderFilledTrack bg="kartoza.500" />
                    </SliderTrack>
                    <SliderThumb />
                  </Slider>
                  <Text fontSize="sm" w="40px">{((rule.symbolizer.strokeOpacity ?? 1) * 100).toFixed(0)}%</Text>
                </HStack>
              </FormControl>
            )}
          </HStack>
        </Box>

        {/* Point-specific settings */}
        {rule.symbolizer.type === 'point' && (
          <Box>
            <Text fontWeight="600" fontSize="sm" mb={2}>Point Symbol</Text>

            {/* Style Presets Gallery */}
            <Box mb={4}>
              <Text fontSize="xs" color="gray.500" mb={2}>Quick Presets</Text>
              <Flex flexWrap="wrap" gap={2}>
                {POINT_STYLE_PRESETS.slice(0, 10).map((preset) => (
                  <Box
                    key={preset.name}
                    title={`${preset.name}: ${preset.description}`}
                    cursor="pointer"
                    p={1}
                    borderRadius="md"
                    border="2px solid"
                    borderColor={
                      rule.symbolizer.fill === preset.fill &&
                      rule.symbolizer.pointShape === preset.shape
                        ? 'kartoza.500'
                        : 'transparent'
                    }
                    _hover={{ borderColor: 'kartoza.300', bg: 'gray.50' }}
                    onClick={() => {
                      updateSymbolizer({
                        pointShape: preset.shape as 'circle' | 'square' | 'triangle' | 'star' | 'cross' | 'x',
                        fill: preset.fill,
                        fillOpacity: preset.fillOpacity,
                        stroke: preset.stroke,
                        strokeWidth: preset.strokeWidth,
                        pointSize: preset.size,
                        haloColor: preset.haloColor,
                        haloRadius: preset.haloRadius,
                        rotation: preset.rotation,
                      })
                    }}
                  >
                    <Box
                      w="28px"
                      h="28px"
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      position="relative"
                    >
                      {/* Halo effect */}
                      {preset.haloColor && (
                        <Box
                          position="absolute"
                          w={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                          h={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                          borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                          bg={preset.haloColor}
                          opacity={0.4}
                          transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                          style={{
                            clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' : undefined,
                          }}
                        />
                      )}
                      {/* Main symbol */}
                      <Box
                        position="relative"
                        w={`${preset.size}px`}
                        h={`${preset.size}px`}
                        borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                        bg={preset.fill}
                        opacity={preset.fillOpacity}
                        border={`${preset.strokeWidth}px solid ${preset.stroke}`}
                        transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                        style={{
                          clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                                   preset.shape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                                   preset.shape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                                   undefined,
                        }}
                      />
                    </Box>
                  </Box>
                ))}
              </Flex>

              {/* Show more presets */}
              <Collapse in={showMorePresets} animateOpacity>
                <Flex flexWrap="wrap" gap={2} mt={2}>
                  {POINT_STYLE_PRESETS.slice(10).map((preset) => (
                    <Box
                      key={preset.name}
                      title={`${preset.name}: ${preset.description}`}
                      cursor="pointer"
                      p={1}
                      borderRadius="md"
                      border="2px solid"
                      borderColor={
                        rule.symbolizer.fill === preset.fill &&
                        rule.symbolizer.pointShape === preset.shape
                          ? 'kartoza.500'
                          : 'transparent'
                      }
                      _hover={{ borderColor: 'kartoza.300', bg: 'gray.50' }}
                      onClick={() => {
                        updateSymbolizer({
                          pointShape: preset.shape as 'circle' | 'square' | 'triangle' | 'star' | 'cross' | 'x',
                          fill: preset.fill,
                          fillOpacity: preset.fillOpacity,
                          stroke: preset.stroke,
                          strokeWidth: preset.strokeWidth,
                          pointSize: preset.size,
                          haloColor: preset.haloColor,
                          haloRadius: preset.haloRadius,
                          rotation: preset.rotation,
                        })
                      }}
                    >
                      <Box
                        w="28px"
                        h="28px"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                        position="relative"
                      >
                        {preset.haloColor && (
                          <Box
                            position="absolute"
                            w={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                            h={`${preset.size + (preset.haloRadius || 0) * 2}px`}
                            borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                            bg={preset.haloColor}
                            opacity={0.4}
                            transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                            style={{
                              clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' : undefined,
                            }}
                          />
                        )}
                        <Box
                          position="relative"
                          w={`${preset.size}px`}
                          h={`${preset.size}px`}
                          borderRadius={preset.shape === 'circle' ? '50%' : preset.shape === 'triangle' ? '0' : 'sm'}
                          bg={preset.fill}
                          opacity={preset.fillOpacity}
                          border={`${preset.strokeWidth}px solid ${preset.stroke}`}
                          transform={preset.rotation ? `rotate(${preset.rotation}deg)` : undefined}
                          style={{
                            clipPath: preset.shape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                                     preset.shape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                                     preset.shape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                                     undefined,
                          }}
                        />
                      </Box>
                    </Box>
                  ))}
                </Flex>
              </Collapse>

              <Button
                size="xs"
                variant="ghost"
                mt={2}
                onClick={() => setShowMorePresets(!showMorePresets)}
                rightIcon={<Icon as={showMorePresets ? FiChevronUp : FiChevronDown} />}
              >
                {showMorePresets ? 'Show Less' : `Show ${POINT_STYLE_PRESETS.length - 10} More`}
              </Button>
            </Box>

            <Divider my={3} />

            {/* Manual controls */}
            <HStack spacing={4} wrap="wrap">
              <FormControl maxW="150px">
                <FormLabel fontSize="sm">Shape</FormLabel>
                <Select
                  size="sm"
                  value={rule.symbolizer.pointShape || 'circle'}
                  onChange={(e) => updateSymbolizer({ pointShape: e.target.value as 'circle' | 'square' | 'triangle' | 'star' | 'cross' | 'x' })}
                >
                  <option value="circle">Circle</option>
                  <option value="square">Square</option>
                  <option value="triangle">Triangle</option>
                  <option value="star">Star</option>
                  <option value="cross">Cross</option>
                  <option value="x">X</option>
                </Select>
              </FormControl>
              <FormControl maxW="100px">
                <FormLabel fontSize="sm">Size</FormLabel>
                <NumberInput
                  size="sm"
                  value={rule.symbolizer.pointSize || 8}
                  min={1}
                  max={50}
                  onChange={(_, val) => updateSymbolizer({ pointSize: val })}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
              <FormControl maxW="100px">
                <FormLabel fontSize="sm">Rotation</FormLabel>
                <NumberInput
                  size="sm"
                  value={rule.symbolizer.rotation || 0}
                  min={0}
                  max={360}
                  step={15}
                  onChange={(_, val) => updateSymbolizer({ rotation: val })}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
            </HStack>

            {/* Halo / Glow effect */}
            <Box mt={4}>
              <Text fontWeight="600" fontSize="sm" mb={2}>Halo / Glow Effect</Text>
              <HStack spacing={4} wrap="wrap">
                <ColorPicker
                  label="Halo Color"
                  value={rule.symbolizer.haloColor || '#ffffff'}
                  onChange={(color) => updateSymbolizer({ haloColor: color })}
                />
                <FormControl maxW="120px">
                  <FormLabel fontSize="sm">Halo Radius</FormLabel>
                  <HStack>
                    <Slider
                      value={rule.symbolizer.haloRadius || 0}
                      min={0}
                      max={10}
                      step={1}
                      onChange={(val) => updateSymbolizer({ haloRadius: val })}
                    >
                      <SliderTrack>
                        <SliderFilledTrack bg="kartoza.500" />
                      </SliderTrack>
                      <SliderThumb />
                    </Slider>
                    <Text fontSize="sm" w="30px">{rule.symbolizer.haloRadius || 0}px</Text>
                  </HStack>
                </FormControl>
                {rule.symbolizer.haloRadius && rule.symbolizer.haloRadius > 0 && (
                  <Button
                    size="xs"
                    variant="ghost"
                    colorScheme="red"
                    onClick={() => updateSymbolizer({ haloRadius: 0, haloColor: undefined })}
                  >
                    Remove Halo
                  </Button>
                )}
              </HStack>
            </Box>
          </Box>
        )}

        {/* Preview swatch */}
        <Box>
          <Text fontWeight="600" fontSize="sm" mb={2}>Preview</Text>
          <Box
            w="100px"
            h="60px"
            borderRadius="md"
            border="1px solid"
            borderColor={borderColor}
            display="flex"
            alignItems="center"
            justifyContent="center"
            bg="gray.100"
          >
            {rule.symbolizer.type === 'polygon' && (
              <Box
                w="60px"
                h="40px"
                borderRadius="sm"
                bg={rule.symbolizer.fill}
                opacity={rule.symbolizer.fillOpacity}
                border={`${rule.symbolizer.strokeWidth}px solid ${rule.symbolizer.stroke}`}
              />
            )}
            {rule.symbolizer.type === 'line' && (
              <Box
                w="60px"
                h={`${Math.max(2, rule.symbolizer.strokeWidth || 2)}px`}
                bg={rule.symbolizer.stroke}
                opacity={rule.symbolizer.strokeOpacity}
              />
            )}
            {rule.symbolizer.type === 'point' && (
              <Box position="relative" display="flex" alignItems="center" justifyContent="center">
                {/* Halo effect */}
                {rule.symbolizer.haloColor && rule.symbolizer.haloRadius && rule.symbolizer.haloRadius > 0 && (
                  <Box
                    position="absolute"
                    w={`${(rule.symbolizer.pointSize || 8) + rule.symbolizer.haloRadius * 2}px`}
                    h={`${(rule.symbolizer.pointSize || 8) + rule.symbolizer.haloRadius * 2}px`}
                    borderRadius={rule.symbolizer.pointShape === 'circle' ? '50%' : rule.symbolizer.pointShape === 'triangle' ? '0' : 'sm'}
                    bg={rule.symbolizer.haloColor}
                    opacity={0.4}
                    transform={rule.symbolizer.rotation ? `rotate(${rule.symbolizer.rotation}deg)` : undefined}
                    style={{
                      clipPath: rule.symbolizer.pointShape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                               rule.symbolizer.pointShape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                               rule.symbolizer.pointShape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                               rule.symbolizer.pointShape === 'x' ? 'polygon(10% 0%, 50% 40%, 90% 0%, 100% 10%, 60% 50%, 100% 90%, 90% 100%, 50% 60%, 10% 100%, 0% 90%, 40% 50%, 0% 10%)' :
                               undefined,
                    }}
                  />
                )}
                {/* Main point symbol */}
                <Box
                  position="relative"
                  w={`${rule.symbolizer.pointSize || 8}px`}
                  h={`${rule.symbolizer.pointSize || 8}px`}
                  borderRadius={rule.symbolizer.pointShape === 'circle' ? '50%' : rule.symbolizer.pointShape === 'triangle' ? '0' : 'sm'}
                  bg={rule.symbolizer.fill}
                  opacity={rule.symbolizer.fillOpacity}
                  border={`${rule.symbolizer.strokeWidth}px solid ${rule.symbolizer.stroke}`}
                  transform={rule.symbolizer.rotation ? `rotate(${rule.symbolizer.rotation}deg)` : undefined}
                  style={{
                    clipPath: rule.symbolizer.pointShape === 'triangle' ? 'polygon(50% 0%, 0% 100%, 100% 100%)' :
                             rule.symbolizer.pointShape === 'star' ? 'polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%)' :
                             rule.symbolizer.pointShape === 'cross' ? 'polygon(35% 0%, 65% 0%, 65% 35%, 100% 35%, 100% 65%, 65% 65%, 65% 100%, 35% 100%, 35% 65%, 0% 65%, 0% 35%, 35% 35%)' :
                             rule.symbolizer.pointShape === 'x' ? 'polygon(10% 0%, 50% 40%, 90% 0%, 100% 10%, 60% 50%, 100% 90%, 90% 100%, 50% 60%, 10% 100%, 0% 90%, 40% 50%, 0% 10%)' :
                             undefined,
                  }}
                />
              </Box>
            )}
          </Box>
        </Box>
      </VStack>
    </Box>
  )
}

export function StyleDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const dialogData = useUIStore((state) => state.dialogData)
  const closeDialog = useUIStore((state) => state.closeDialog)
  const toast = useToast()
  const queryClient = useQueryClient()

  const isOpen = activeDialog === 'style'
  const isEditMode = dialogData?.mode === 'edit'
  const connectionId = dialogData?.data?.connectionId as string
  const workspace = dialogData?.data?.workspace as string
  const styleName = dialogData?.data?.name as string
  const previewLayer = dialogData?.data?.previewLayer as string | undefined

  // State
  const [name, setName] = useState('')
  const [format, setFormat] = useState<'sld' | 'css'>('sld')
  const [content, setContent] = useState('')
  const [rules, setRules] = useState<StyleRule[]>([])
  const [activeTab, setActiveTab] = useState(0)
  const [hasChanges, setHasChanges] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)

  // Classification wizard state
  const { isOpen: classifyPanelOpen, onToggle: toggleClassifyPanel } = useDisclosure()
  const [classifyAttribute, setClassifyAttribute] = useState('')
  const [classifyMethod, setClassifyMethod] = useState<ClassificationMethod>('equal-interval')
  const [classifyClasses, setClassifyClasses] = useState(5)
  const [classifyColorRamp, setClassifyColorRamp] = useState('blue-to-red')
  const [classifyGeomType, setClassifyGeomType] = useState<'polygon' | 'line' | 'point'>('polygon')
  const [classifySampleValues, setClassifySampleValues] = useState('')

  // Raster style wizard state
  const { isOpen: rasterPanelOpen, onToggle: toggleRasterPanel } = useDisclosure()
  const [rasterStyleType, setRasterStyleType] = useState<'colormap' | 'hillshade' | 'hillshade-color' | 'contrast'>('colormap')
  const [rasterColorRamp, setRasterColorRamp] = useState('rainbow')
  const [rasterMinValue, setRasterMinValue] = useState(0)
  const [rasterMaxValue, setRasterMaxValue] = useState(1000)
  const [rasterOpacity, setRasterOpacity] = useState(1)
  const [rasterColorMapType, setRasterColorMapType] = useState<'ramp' | 'intervals' | 'values'>('ramp')
  const [hillshadeAzimuth, setHillshadeAzimuth] = useState(315)
  const [hillshadeAltitude, setHillshadeAltitude] = useState(45)
  const [hillshadeZFactor, setHillshadeZFactor] = useState(1)
  const [contrastMethod, setContrastMethod] = useState<'normalize' | 'histogram' | 'none'>('normalize')
  const [gammaValue, setGammaValue] = useState(1.0)

  const bgColor = useColorModeValue('gray.50', 'gray.900')
  const headerBg = useColorModeValue('linear-gradient(135deg, #3d9970 0%, #2d7a5a 100%)', 'linear-gradient(135deg, #2d7a5a 0%, #1d5a40 100%)')

  // Fetch style content when editing
  const { data: styleData, isLoading } = useQuery({
    queryKey: ['style', connectionId, workspace, styleName],
    queryFn: () => api.getStyleContent(connectionId, workspace, styleName),
    enabled: isOpen && isEditMode && !!styleName,
  })

  // Initialize form when dialog opens or data loads
  useEffect(() => {
    if (!isOpen) return

    if (isEditMode && styleData) {
      setName(styleData.name)
      setFormat(styleData.format as 'sld' | 'css')
      setContent(styleData.content)
      if (styleData.format === 'sld') {
        setRules(parseSLDRules(styleData.content))
      }
    } else if (!isEditMode) {
      // New style
      setName('')
      setFormat('sld')
      setContent(DEFAULT_SLD)
      setRules(parseSLDRules(DEFAULT_SLD))
    }
    setHasChanges(false)
    setValidationError(null)
  }, [isOpen, isEditMode, styleData])

  // Sync visual editor changes to code
  useEffect(() => {
    if (format === 'sld' && activeTab === 0 && rules.length > 0) {
      const newContent = generateSLD(name || 'NewStyle', rules)
      if (newContent !== content) {
        setContent(newContent)
        setHasChanges(true)
      }
    }
  }, [rules, name])

  // Validate SLD content
  const validateContent = useCallback(() => {
    if (format === 'sld') {
      try {
        const parser = new DOMParser()
        const doc = parser.parseFromString(content, 'text/xml')
        const parseError = doc.querySelector('parsererror')
        if (parseError) {
          setValidationError('Invalid XML: ' + parseError.textContent?.slice(0, 100))
          return false
        }
        setValidationError(null)
        return true
      } catch (e) {
        setValidationError('Failed to parse SLD')
        return false
      }
    }
    setValidationError(null)
    return true
  }, [content, format])

  // Mutations
  const updateMutation = useMutation({
    mutationFn: () => api.updateStyleContent(connectionId, workspace, styleName, content, format),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['styles', connectionId, workspace] })
      queryClient.invalidateQueries({ queryKey: ['style', connectionId, workspace, styleName] })
      toast({ title: 'Style updated', status: 'success', duration: 3000 })
      setHasChanges(false)
    },
    onError: (error: Error) => {
      toast({ title: 'Failed to update style', description: error.message, status: 'error', duration: 5000 })
    },
  })

  const createMutation = useMutation({
    mutationFn: () => api.createStyle(connectionId, workspace, name, content, format),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['styles', connectionId, workspace] })
      toast({ title: 'Style created', status: 'success', duration: 3000 })
      closeDialog()
    },
    onError: (error: Error) => {
      toast({ title: 'Failed to create style', description: error.message, status: 'error', duration: 5000 })
    },
  })

  const handleSave = () => {
    if (!validateContent()) {
      toast({ title: 'Please fix validation errors', status: 'warning', duration: 3000 })
      return
    }

    if (!isEditMode && !name.trim()) {
      toast({ title: 'Style name is required', status: 'warning', duration: 3000 })
      return
    }

    if (isEditMode) {
      updateMutation.mutate()
    } else {
      createMutation.mutate()
    }
  }

  const handleCodeChange = (value: string) => {
    setContent(value)
    setHasChanges(true)
    if (format === 'sld') {
      setRules(parseSLDRules(value))
    }
  }

  const handleFormatChange = (newFormat: 'sld' | 'css') => {
    if (newFormat === format) return

    // Convert content or use default
    if (newFormat === 'sld') {
      setContent(DEFAULT_SLD)
      setRules(parseSLDRules(DEFAULT_SLD))
    } else {
      setContent(DEFAULT_CSS)
      setRules([])
    }
    setFormat(newFormat)
    setHasChanges(true)
  }

  const addRule = () => {
    setRules([...rules, {
      name: `Rule ${rules.length + 1}`,
      symbolizer: {
        type: 'polygon',
        fill: '#3388ff',
        fillOpacity: 0.6,
        stroke: '#2266cc',
        strokeWidth: 1,
      }
    }])
  }

  const updateRule = (index: number, rule: StyleRule) => {
    const newRules = [...rules]
    newRules[index] = rule
    setRules(newRules)
    setHasChanges(true)
  }

  const deleteRule = (index: number) => {
    if (rules.length <= 1) {
      toast({ title: 'Cannot delete the last rule', status: 'warning', duration: 3000 })
      return
    }
    setRules(rules.filter((_, i) => i !== index))
    setHasChanges(true)
  }

  // Generate preview URL
  const handlePreview = async () => {
    if (!previewLayer) {
      toast({ title: 'No preview layer specified', status: 'info', duration: 3000 })
      return
    }

    // Start a preview session with the style applied
    try {
      const { url } = await api.startPreview({
        connId: connectionId,
        workspace,
        layerName: previewLayer,
        storeName: '',
        storeType: 'datastore',
        layerType: 'vector',
      })
      setPreviewUrl(url)
    } catch (error) {
      toast({ title: 'Failed to start preview', status: 'error', duration: 3000 })
    }
  }

  // Generate classified style
  const handleGenerateClassifiedStyle = () => {
    // Parse sample values
    const values = classifySampleValues
      .split(',')
      .map(s => parseFloat(s.trim()))
      .filter(n => !isNaN(n))

    if (values.length < 2) {
      toast({
        title: 'Need more sample values',
        description: 'Please enter at least 2 numeric values separated by commas',
        status: 'warning',
        duration: 3000,
      })
      return
    }

    // Calculate breaks
    const breaks = calculateBreaks(values, classifyClasses, classifyMethod)
    const colors = interpolateColors(COLOR_RAMPS[classifyColorRamp], classifyClasses)

    // Generate SLD
    const sld = generateClassifiedSLD(
      name || 'ClassifiedStyle',
      classifyAttribute,
      breaks,
      colors,
      classifyGeomType
    )

    setContent(sld)
    setFormat('sld')
    setRules(parseSLDRules(sld))
    setHasChanges(true)
    setActiveTab(1) // Switch to code editor to show result

    toast({
      title: 'Classified style generated',
      description: `Created ${classifyClasses} classes using ${classifyMethod.replace('-', ' ')}`,
      status: 'success',
      duration: 3000,
    })
  }

  // Generate raster style
  const handleGenerateRasterStyle = () => {
    const styleName_ = name || 'RasterStyle'
    const colorRamp = RASTER_COLOR_RAMPS[rasterColorRamp]?.colors || RASTER_COLOR_RAMPS['rainbow'].colors
    let sld: string

    switch (rasterStyleType) {
      case 'colormap':
        sld = generateRasterColorMapSLD(
          styleName_,
          colorRamp,
          rasterMinValue,
          rasterMaxValue,
          rasterColorMapType,
          rasterOpacity
        )
        break
      case 'hillshade':
        sld = generateHillshadeSLD(
          styleName_,
          hillshadeAzimuth,
          hillshadeAltitude,
          hillshadeZFactor,
          rasterOpacity
        )
        break
      case 'hillshade-color':
        sld = generateHillshadeWithColorSLD(
          styleName_,
          colorRamp,
          rasterMinValue,
          rasterMaxValue,
          hillshadeAzimuth,
          hillshadeAltitude,
          hillshadeZFactor,
          rasterOpacity
        )
        break
      case 'contrast':
        sld = generateContrastEnhancementSLD(
          styleName_,
          contrastMethod,
          gammaValue,
          rasterOpacity
        )
        break
      default:
        sld = generateRasterColorMapSLD(styleName_, colorRamp, rasterMinValue, rasterMaxValue, 'ramp', rasterOpacity)
    }

    setContent(sld)
    setFormat('sld')
    setRules([]) // Clear vector rules
    setHasChanges(true)
    setActiveTab(1) // Switch to code editor to show result

    toast({
      title: 'Raster style generated',
      description: `Created ${rasterStyleType === 'hillshade' ? 'hillshade' : rasterStyleType === 'hillshade-color' ? 'hillshade with colors' : rasterStyleType === 'contrast' ? 'contrast enhanced' : 'color map'} style`,
      status: 'success',
      duration: 3000,
    })
  }

  // Apply hillshade preset
  const applyHillshadePreset = (preset: typeof HILLSHADE_PRESETS[0]) => {
    setHillshadeAzimuth(preset.azimuth)
    setHillshadeAltitude(preset.altitude)
    setHillshadeZFactor(preset.zFactor)
  }

  const extensions = useMemo(() => {
    return format === 'sld' ? [xml()] : [css()]
  }, [format])

  const isLoading_ = isLoading || updateMutation.isPending || createMutation.isPending

  return (
    <Modal isOpen={isOpen} onClose={closeDialog} size="6xl" scrollBehavior="inside">
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent maxH="90vh" borderRadius="xl" overflow="hidden">
        {/* Gradient Header */}
        <Box
          bg={headerBg}
          color="white"
          px={6}
          py={4}
        >
          <Flex align="center" justify="space-between">
            <HStack spacing={3}>
              <Icon as={FiDroplet} boxSize={6} />
              <Box>
                <Text fontSize="lg" fontWeight="600">
                  {isEditMode ? 'Edit Style' : 'Create Style'}
                </Text>
                <Text fontSize="sm" opacity={0.9}>
                  {isEditMode ? `Editing ${styleName}` : 'Create a new map style'}
                </Text>
              </Box>
            </HStack>
            <HStack>
              {hasChanges && (
                <Badge colorScheme="yellow" variant="solid">
                  Unsaved Changes
                </Badge>
              )}
              <Badge colorScheme={format === 'sld' ? 'blue' : 'purple'} variant="solid">
                {format.toUpperCase()}
              </Badge>
            </HStack>
          </Flex>
        </Box>
        <ModalCloseButton color="white" />

        <ModalBody p={0} bg={bgColor}>
          {isLoading ? (
            <Flex h="400px" align="center" justify="center">
              <Spinner size="xl" color="kartoza.500" />
            </Flex>
          ) : (
            <Flex h="70vh">
              {/* Left panel - Style Properties */}
              <Box w="250px" borderRight="1px solid" borderColor="gray.200" p={4} overflowY="auto">
                <VStack spacing={4} align="stretch">
                  {!isEditMode && (
                    <FormControl isRequired>
                      <FormLabel>Style Name</FormLabel>
                      <Input
                        value={name}
                        onChange={(e) => {
                          setName(e.target.value)
                          setHasChanges(true)
                        }}
                        placeholder="my-style"
                      />
                    </FormControl>
                  )}

                  <FormControl>
                    <FormLabel>Format</FormLabel>
                    <Select
                      value={format}
                      onChange={(e) => handleFormatChange(e.target.value as 'sld' | 'css')}
                    >
                      <option value="sld">SLD (Styled Layer Descriptor)</option>
                      <option value="css">CSS (GeoServer CSS)</option>
                    </Select>
                  </FormControl>

                  <Divider />

                  {format === 'sld' && (
                    <>
                      <Text fontWeight="600">Quick Actions</Text>
                      <Button
                        size="sm"
                        leftIcon={<Icon as={FiSquare} />}
                        variant="outline"
                        onClick={() => setRules([{
                          name: 'Polygon',
                          symbolizer: { type: 'polygon', fill: '#3388ff', fillOpacity: 0.6, stroke: '#2266cc', strokeWidth: 1 }
                        }])}
                      >
                        Polygon Style
                      </Button>
                      <Button
                        size="sm"
                        leftIcon={<Icon as={FiMinus} />}
                        variant="outline"
                        onClick={() => setRules([{
                          name: 'Line',
                          symbolizer: { type: 'line', stroke: '#3388ff', strokeWidth: 2, strokeOpacity: 1 }
                        }])}
                      >
                        Line Style
                      </Button>
                      <Button
                        size="sm"
                        leftIcon={<Icon as={FiCircle} />}
                        variant="outline"
                        onClick={() => setRules([{
                          name: 'Point',
                          symbolizer: { type: 'point', fill: '#3388ff', fillOpacity: 1, stroke: '#2266cc', strokeWidth: 1, pointShape: 'circle', pointSize: 8 }
                        }])}
                      >
                        Point Style
                      </Button>

                      <Divider />

                      <Text fontWeight="600">Classified Style</Text>
                      <Button
                        size="sm"
                        leftIcon={<Icon as={FiGrid} />}
                        rightIcon={<Icon as={classifyPanelOpen ? FiChevronUp : FiChevronDown} />}
                        variant="outline"
                        colorScheme="kartoza"
                        onClick={toggleClassifyPanel}
                      >
                        Choropleth Wizard
                      </Button>

                      <Collapse in={classifyPanelOpen} animateOpacity>
                        <VStack spacing={3} p={3} bg="gray.50" borderRadius="md" align="stretch">
                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Attribute</FormLabel>
                            <Input
                              size="sm"
                              value={classifyAttribute}
                              onChange={(e) => setClassifyAttribute(e.target.value)}
                              placeholder="population"
                            />
                          </FormControl>

                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Method</FormLabel>
                            <Select
                              size="sm"
                              value={classifyMethod}
                              onChange={(e) => setClassifyMethod(e.target.value as ClassificationMethod)}
                            >
                              <option value="equal-interval">Equal Interval</option>
                              <option value="quantile">Quantile</option>
                              <option value="jenks">Jenks Natural Breaks</option>
                              <option value="pretty">Pretty Breaks</option>
                            </Select>
                          </FormControl>

                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Classes: {classifyClasses}</FormLabel>
                            <Slider
                              value={classifyClasses}
                              onChange={(v) => setClassifyClasses(v)}
                              min={3}
                              max={10}
                              step={1}
                            >
                              <SliderTrack>
                                <SliderFilledTrack bg="kartoza.500" />
                              </SliderTrack>
                              <SliderThumb />
                            </Slider>
                          </FormControl>

                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Color Ramp</FormLabel>
                            <Select
                              size="sm"
                              value={classifyColorRamp}
                              onChange={(e) => setClassifyColorRamp(e.target.value)}
                            >
                              {Object.keys(COLOR_RAMPS).map((rampName) => (
                                <option key={rampName} value={rampName}>
                                  {rampName.replace(/-/g, ' ')}
                                </option>
                              ))}
                            </Select>
                            <HStack mt={1} spacing={1}>
                              {interpolateColors(COLOR_RAMPS[classifyColorRamp], classifyClasses).map((color, i) => (
                                <Box key={i} w="16px" h="12px" bg={color} borderRadius="sm" />
                              ))}
                            </HStack>
                          </FormControl>

                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Geometry Type</FormLabel>
                            <Select
                              size="sm"
                              value={classifyGeomType}
                              onChange={(e) => setClassifyGeomType(e.target.value as 'polygon' | 'line' | 'point')}
                            >
                              <option value="polygon">Polygon</option>
                              <option value="line">Line</option>
                              <option value="point">Point</option>
                            </Select>
                          </FormControl>

                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Sample Values (comma separated)</FormLabel>
                            <Input
                              size="sm"
                              value={classifySampleValues}
                              onChange={(e) => setClassifySampleValues(e.target.value)}
                              placeholder="10, 25, 50, 100, 200"
                            />
                          </FormControl>

                          <Button
                            size="sm"
                            colorScheme="kartoza"
                            onClick={handleGenerateClassifiedStyle}
                            isDisabled={!classifyAttribute || !classifySampleValues}
                          >
                            Generate Style
                          </Button>
                        </VStack>
                      </Collapse>

                      <Divider />

                      <Text fontWeight="600">Raster Style</Text>
                      <Button
                        size="sm"
                        leftIcon={<Icon as={FiImage} />}
                        rightIcon={<Icon as={rasterPanelOpen ? FiChevronUp : FiChevronDown} />}
                        variant="outline"
                        colorScheme="purple"
                        onClick={toggleRasterPanel}
                      >
                        Raster Wizard
                      </Button>

                      <Collapse in={rasterPanelOpen} animateOpacity>
                        <VStack spacing={3} p={3} bg="purple.50" borderRadius="md" align="stretch">
                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Style Type</FormLabel>
                            <Select
                              size="sm"
                              value={rasterStyleType}
                              onChange={(e) => setRasterStyleType(e.target.value as 'colormap' | 'hillshade' | 'hillshade-color' | 'contrast')}
                            >
                              <option value="colormap">Color Map (Graduated)</option>
                              <option value="hillshade">Hillshade Only</option>
                              <option value="hillshade-color">Hillshade + Colors</option>
                              <option value="contrast">Contrast Enhancement</option>
                            </Select>
                          </FormControl>

                          {(rasterStyleType === 'colormap' || rasterStyleType === 'hillshade-color') && (
                            <>
                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Color Ramp</FormLabel>
                                <Select
                                  size="sm"
                                  value={rasterColorRamp}
                                  onChange={(e) => setRasterColorRamp(e.target.value)}
                                >
                                  {Object.entries(RASTER_COLOR_RAMPS).map(([key, ramp]) => (
                                    <option key={key} value={key}>{ramp.name}</option>
                                  ))}
                                </Select>
                                <HStack mt={1} spacing={0}>
                                  {RASTER_COLOR_RAMPS[rasterColorRamp]?.colors.map((color, i) => (
                                    <Box key={i} flex="1" h="8px" bg={color} borderRadius={i === 0 ? 'sm 0 0 sm' : i === RASTER_COLOR_RAMPS[rasterColorRamp].colors.length - 1 ? '0 sm sm 0' : '0'} />
                                  ))}
                                </HStack>
                                <Text fontSize="xs" color="gray.500" mt={1}>
                                  {RASTER_COLOR_RAMPS[rasterColorRamp]?.description}
                                </Text>
                              </FormControl>

                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Min Value</FormLabel>
                                <Input
                                  size="sm"
                                  type="number"
                                  value={rasterMinValue}
                                  onChange={(e) => setRasterMinValue(parseFloat(e.target.value) || 0)}
                                />
                              </FormControl>

                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Max Value</FormLabel>
                                <Input
                                  size="sm"
                                  type="number"
                                  value={rasterMaxValue}
                                  onChange={(e) => setRasterMaxValue(parseFloat(e.target.value) || 1000)}
                                />
                              </FormControl>

                              {rasterStyleType === 'colormap' && (
                                <FormControl size="sm">
                                  <FormLabel fontSize="xs">Color Map Type</FormLabel>
                                  <Select
                                    size="sm"
                                    value={rasterColorMapType}
                                    onChange={(e) => setRasterColorMapType(e.target.value as 'ramp' | 'intervals' | 'values')}
                                  >
                                    <option value="ramp">Ramp (Smooth gradient)</option>
                                    <option value="intervals">Intervals (Discrete)</option>
                                    <option value="values">Values (Exact match)</option>
                                  </Select>
                                </FormControl>
                              )}
                            </>
                          )}

                          {(rasterStyleType === 'hillshade' || rasterStyleType === 'hillshade-color') && (
                            <>
                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Hillshade Preset</FormLabel>
                                <Select
                                  size="sm"
                                  placeholder="Select preset..."
                                  onChange={(e) => {
                                    const preset = HILLSHADE_PRESETS.find(p => p.name === e.target.value)
                                    if (preset) applyHillshadePreset(preset)
                                  }}
                                >
                                  {HILLSHADE_PRESETS.map((preset) => (
                                    <option key={preset.name} value={preset.name}>{preset.name}</option>
                                  ))}
                                </Select>
                              </FormControl>

                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Sun Azimuth: {hillshadeAzimuth}°</FormLabel>
                                <Slider
                                  value={hillshadeAzimuth}
                                  onChange={(v) => setHillshadeAzimuth(v)}
                                  min={0}
                                  max={360}
                                  step={15}
                                >
                                  <SliderTrack>
                                    <SliderFilledTrack bg="purple.500" />
                                  </SliderTrack>
                                  <SliderThumb />
                                </Slider>
                              </FormControl>

                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Sun Altitude: {hillshadeAltitude}°</FormLabel>
                                <Slider
                                  value={hillshadeAltitude}
                                  onChange={(v) => setHillshadeAltitude(v)}
                                  min={0}
                                  max={90}
                                  step={5}
                                >
                                  <SliderTrack>
                                    <SliderFilledTrack bg="purple.500" />
                                  </SliderTrack>
                                  <SliderThumb />
                                </Slider>
                              </FormControl>

                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Z Factor: {hillshadeZFactor}</FormLabel>
                                <Slider
                                  value={hillshadeZFactor}
                                  onChange={(v) => setHillshadeZFactor(v)}
                                  min={0.1}
                                  max={5}
                                  step={0.1}
                                >
                                  <SliderTrack>
                                    <SliderFilledTrack bg="purple.500" />
                                  </SliderTrack>
                                  <SliderThumb />
                                </Slider>
                              </FormControl>
                            </>
                          )}

                          {rasterStyleType === 'contrast' && (
                            <>
                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Enhancement Method</FormLabel>
                                <Select
                                  size="sm"
                                  value={contrastMethod}
                                  onChange={(e) => setContrastMethod(e.target.value as 'normalize' | 'histogram' | 'none')}
                                >
                                  <option value="normalize">Normalize (Stretch to min/max)</option>
                                  <option value="histogram">Histogram Equalization</option>
                                  <option value="none">None</option>
                                </Select>
                              </FormControl>

                              <FormControl size="sm">
                                <FormLabel fontSize="xs">Gamma: {gammaValue.toFixed(1)}</FormLabel>
                                <Slider
                                  value={gammaValue}
                                  onChange={(v) => setGammaValue(v)}
                                  min={0.1}
                                  max={3}
                                  step={0.1}
                                >
                                  <SliderTrack>
                                    <SliderFilledTrack bg="purple.500" />
                                  </SliderTrack>
                                  <SliderThumb />
                                </Slider>
                              </FormControl>
                            </>
                          )}

                          <FormControl size="sm">
                            <FormLabel fontSize="xs">Opacity: {(rasterOpacity * 100).toFixed(0)}%</FormLabel>
                            <Slider
                              value={rasterOpacity}
                              onChange={(v) => setRasterOpacity(v)}
                              min={0}
                              max={1}
                              step={0.05}
                            >
                              <SliderTrack>
                                <SliderFilledTrack bg="purple.500" />
                              </SliderTrack>
                              <SliderThumb />
                            </Slider>
                          </FormControl>

                          <Button
                            size="sm"
                            colorScheme="purple"
                            onClick={handleGenerateRasterStyle}
                          >
                            Generate Raster Style
                          </Button>
                        </VStack>
                      </Collapse>
                    </>
                  )}

                  {previewLayer && (
                    <>
                      <Divider />
                      <Text fontWeight="600">Preview</Text>
                      <Text fontSize="sm" color="gray.600">Layer: {previewLayer}</Text>
                      <Button
                        size="sm"
                        leftIcon={<Icon as={FiEye} />}
                        colorScheme="kartoza"
                        onClick={handlePreview}
                      >
                        Preview on Map
                      </Button>
                    </>
                  )}
                </VStack>
              </Box>

              {/* Main content area */}
              <Box flex="1" display="flex" flexDirection="column">
                <Tabs index={activeTab} onChange={setActiveTab} flex="1" display="flex" flexDirection="column">
                  <TabList px={4} pt={2}>
                    {format === 'sld' && (
                      <Tab>
                        <Icon as={FiDroplet} mr={2} />
                        Visual Editor
                      </Tab>
                    )}
                    <Tab>
                      <Icon as={FiCode} mr={2} />
                      Code Editor
                    </Tab>
                    {previewUrl && (
                      <Tab>
                        <Icon as={FiEye} mr={2} />
                        Map Preview
                      </Tab>
                    )}
                  </TabList>

                  <TabPanels flex="1" overflow="hidden">
                    {format === 'sld' && (
                      <TabPanel h="100%" overflowY="auto" p={4}>
                        <VStack spacing={4} align="stretch">
                          {validationError && (
                            <Alert status="warning" borderRadius="md">
                              <AlertIcon />
                              {validationError}
                            </Alert>
                          )}

                          {rules.map((rule, index) => (
                            <RuleEditor
                              key={index}
                              rule={rule}
                              onChange={(r) => updateRule(index, r)}
                              onDelete={() => deleteRule(index)}
                            />
                          ))}

                          <Button
                            leftIcon={<Icon as={FiDroplet} />}
                            onClick={addRule}
                            variant="outline"
                            colorScheme="kartoza"
                          >
                            Add Rule
                          </Button>
                        </VStack>
                      </TabPanel>
                    )}

                    <TabPanel h="100%" p={0}>
                      <Box h="100%" position="relative">
                        {validationError && (
                          <Alert status="warning" position="absolute" top={0} left={0} right={0} zIndex={1}>
                            <AlertIcon />
                            {validationError}
                          </Alert>
                        )}
                        <CodeMirror
                          value={content}
                          height="100%"
                          extensions={extensions}
                          onChange={handleCodeChange}
                          onBlur={validateContent}
                          theme="light"
                          style={{ height: '100%' }}
                        />
                      </Box>
                    </TabPanel>

                    {previewUrl && (
                      <TabPanel h="100%" p={0}>
                        <iframe
                          src={previewUrl}
                          style={{ width: '100%', height: '100%', border: 'none' }}
                          title="Style Preview"
                        />
                      </TabPanel>
                    )}
                  </TabPanels>
                </Tabs>
              </Box>
            </Flex>
          )}
        </ModalBody>

        <ModalFooter
          gap={3}
          borderTop="1px solid"
          borderTopColor="gray.200"
          bg="gray.50"
        >
          <Button variant="ghost" onClick={closeDialog}>
            Cancel
          </Button>
          <Button
            colorScheme="kartoza"
            leftIcon={<Icon as={FiSave} />}
            onClick={handleSave}
            isLoading={isLoading_}
            isDisabled={!hasChanges && isEditMode}
          >
            {isEditMode ? 'Save Changes' : 'Create Style'}
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
