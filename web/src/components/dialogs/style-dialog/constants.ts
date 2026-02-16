// Default SLD template for new styles
export const DEFAULT_SLD = `<?xml version="1.0" encoding="UTF-8"?>
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
export const DEFAULT_CSS = `/* GeoServer CSS Style */
* {
  fill: #3388ff;
  fill-opacity: 0.6;
  stroke: #2266cc;
  stroke-width: 1;
}`

// Color ramps for classification
export const COLOR_RAMPS: Record<string, string[]> = {
  'blue-to-red': ['#2166ac', '#67a9cf', '#d1e5f0', '#fddbc7', '#ef8a62', '#b2182b'],
  'green-to-red': ['#1a9850', '#91cf60', '#d9ef8b', '#fee08b', '#fc8d59', '#d73027'],
  'viridis': ['#440154', '#443983', '#31688e', '#21918c', '#35b779', '#fde725'],
  'spectral': ['#9e0142', '#d53e4f', '#f46d43', '#fdae61', '#fee08b', '#e6f598', '#abdda4', '#66c2a5', '#3288bd', '#5e4fa2'],
  'blues': ['#f7fbff', '#deebf7', '#c6dbef', '#9ecae1', '#6baed6', '#4292c6', '#2171b5', '#084594'],
  'reds': ['#fff5f0', '#fee0d2', '#fcbba1', '#fc9272', '#fb6a4a', '#ef3b2c', '#cb181d', '#99000d'],
  'greens': ['#f7fcf5', '#e5f5e0', '#c7e9c0', '#a1d99b', '#74c476', '#41ab5d', '#238b45', '#005a32'],
}

// Raster-specific color ramps with more gradient options
export const RASTER_COLOR_RAMPS: Record<string, { name: string; colors: string[]; description: string }> = {
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
export const HILLSHADE_PRESETS = [
  { name: 'Default', azimuth: 315, altitude: 45, zFactor: 1 },
  { name: 'Morning Light', azimuth: 90, altitude: 30, zFactor: 1 },
  { name: 'Evening Light', azimuth: 270, altitude: 30, zFactor: 1 },
  { name: 'High Contrast', azimuth: 315, altitude: 60, zFactor: 2 },
  { name: 'Subtle', azimuth: 315, altitude: 45, zFactor: 0.5 },
]

// QGIS-LIKE MARKER SHAPES
export const MARKER_SHAPES = [
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

// LINE DASH PATTERNS (QGIS-like)
export const LINE_DASH_PATTERNS = [
  { name: 'solid', label: 'Solid', dashArray: '' },
  { name: 'dash', label: 'Dash', dashArray: '10 5' },
  { name: 'dot', label: 'Dot', dashArray: '2 5' },
  { name: 'dash-dot', label: 'Dash Dot', dashArray: '10 5 2 5' },
  { name: 'dash-dot-dot', label: 'Dash Dot Dot', dashArray: '10 5 2 5 2 5' },
  { name: 'long-dash', label: 'Long Dash', dashArray: '20 10' },
  { name: 'short-dash', label: 'Short Dash', dashArray: '5 5' },
  { name: 'dense-dot', label: 'Dense Dot', dashArray: '1 2' },
]

// LINE CAP AND JOIN STYLES
export const LINE_CAP_STYLES = [
  { name: 'butt', label: 'Flat' },
  { name: 'round', label: 'Round' },
  { name: 'square', label: 'Square' },
]

export const LINE_JOIN_STYLES = [
  { name: 'miter', label: 'Miter' },
  { name: 'round', label: 'Round' },
  { name: 'bevel', label: 'Bevel' },
]

// FILL PATTERNS (QGIS-like graphic fills)
export const FILL_PATTERNS = [
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

// BLEND MODES (supported by GeoServer)
export const BLEND_MODES = [
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

// FONT MARKER FONTS
export const FONT_MARKER_FONTS = [
  { name: 'Wingdings', characters: ['✈', '★', '♦', '♣', '♠', '♥', '☎', '✉', '✂', '✓', '✗'] },
  { name: 'Webdings', characters: ['⌂', '⌘', '⚙', '⚡', '⚠', '⚑', '⚐', '☀', '☁', '☂', '☃'] },
  { name: 'Symbol', characters: ['α', 'β', 'γ', 'δ', 'ε', 'π', 'Σ', 'Ω', '∞', '≈', '≠'] },
]

// LABEL PLACEMENT OPTIONS
export const LABEL_PLACEMENT_OPTIONS = [
  { name: 'point', label: 'Point on Point' },
  { name: 'line', label: 'Along Line' },
  { name: 'polygon', label: 'Inside Polygon' },
]

export const LABEL_ANCHOR_POINTS = [
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

// Beautiful point style presets
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

export const POINT_STYLE_PRESETS: PointStylePreset[] = [
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
