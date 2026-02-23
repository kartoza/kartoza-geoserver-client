/**
 * S3LayerPreview - Component for previewing S3 geospatial files
 *
 * This module has been refactored for maintainability:
 * - hooks/useS3Preview.ts - Metadata loading hook
 * - hooks/useGeoParquet.ts - GeoParquet parsing hook with WKB parser
 * - components/MetadataPanel.tsx - Metadata display component
 * - components/TableView.tsx - Table view component
 * - utils/wkbParser.ts - WKB parsing utility
 * - utils/formatters.ts - Formatting utilities
 *
 * The main component remains here due to complex state interactions
 * between MapLibre, Cesium, and point cloud viewers.
 */

export { default } from '../S3LayerPreview.tsx'
export { useS3Preview } from './hooks/useS3Preview'
export { useGeoParquet } from './hooks/useGeoParquet'
export { parseWKB, wkbToGeoJSON, convertBigInts } from './utils/wkbParser'
export { formatSize, getFormatBadgeColor, getPreviewTypeBadgeColor } from './utils/formatters'
