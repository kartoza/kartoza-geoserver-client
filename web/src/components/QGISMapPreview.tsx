// Re-export the MapLibre-based preview component
// This replaces the qgis-js WASM-based preview which had issues with:
// - Large download size (~56MB)
// - No support for network layers (XYZ, WMS, etc.)
// - Browser compatibility issues

export { default } from './QGISMapLibrePreview'
