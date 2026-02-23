import { useState, useCallback, useEffect } from 'react'
import { parquetReadObjects } from 'hyparquet'
import { compressors } from 'hyparquet-compressors'
import type { FeatureCollection, Feature, Geometry } from 'geojson'
import { wkbToGeoJSON, convertBigInts } from '../utils/wkbParser'
import type { S3PreviewMetadata } from '../../../types'

interface UseGeoParquetResult {
  geoparquetData: FeatureCollection | null
  geoparquetLoading: boolean
  geoparquetError: string | null
}

/**
 * Hook for loading and parsing GeoParquet files client-side using hyparquet
 */
export function useGeoParquet(metadata: S3PreviewMetadata | null): UseGeoParquetResult {
  const [geoparquetData, setGeoparquetData] = useState<FeatureCollection | null>(null)
  const [geoparquetLoading, setGeoparquetLoading] = useState(false)
  const [geoparquetError, setGeoparquetError] = useState<string | null>(null)

  const loadGeoParquet = useCallback(async (proxyUrl: string) => {
    setGeoparquetLoading(true)
    setGeoparquetError(null)

    try {
      // Fetch the parquet file via proxy
      const response = await fetch(proxyUrl)
      if (!response.ok) {
        throw new Error(`Failed to fetch GeoParquet: ${response.status}`)
      }

      const arrayBuffer = await response.arrayBuffer()

      // Use hyparquet to read the parquet file as objects
      // hyparquet automatically decodes GeoParquet geometry columns to GeoJSON
      // Include compressors for zstd, lz4, brotli, gzip support
      const rows = await parquetReadObjects({ file: arrayBuffer, compressors })

      // Find the geometry column (commonly named 'geometry', 'geom', or 'wkb_geometry')
      let geometryColumn = 'geometry'
      if (rows.length > 0) {
        const firstRow = rows[0]
        if ('geometry' in firstRow) {
          geometryColumn = 'geometry'
        } else if ('geom' in firstRow) {
          geometryColumn = 'geom'
        } else if ('wkb_geometry' in firstRow) {
          geometryColumn = 'wkb_geometry'
        }
      }

      // Convert to GeoJSON FeatureCollection
      const features: Feature[] = rows.map((row) => {
        const rawGeom = row[geometryColumn]
        const geometry = wkbToGeoJSON(rawGeom)
        const properties: Record<string, unknown> = {}

        // Copy all non-geometry properties, converting BigInts
        for (const [key, value] of Object.entries(row)) {
          if (key !== geometryColumn) {
            properties[key] = convertBigInts(value)
          }
        }

        return {
          type: 'Feature' as const,
          geometry: geometry as Geometry,
          properties,
        }
      }).filter(f => f.geometry !== null)

      const featureCollection: FeatureCollection = {
        type: 'FeatureCollection',
        features,
      }

      console.log(`Loaded ${features.length} features from GeoParquet client-side`)
      if (features.length > 0) {
        console.log('First feature geometry:', JSON.stringify(features[0].geometry, null, 2))
        console.log('First feature geometry type:', features[0].geometry?.type)
      }
      setGeoparquetData(featureCollection)
    } catch (err) {
      console.error('Failed to load GeoParquet client-side:', err)
      setGeoparquetError(err instanceof Error ? err.message : 'Failed to load GeoParquet')
    } finally {
      setGeoparquetLoading(false)
    }
  }, [])

  // Trigger GeoParquet loading when metadata indicates geoparquet format
  useEffect(() => {
    if (metadata?.format === 'geoparquet' && metadata.proxyUrl && !geoparquetData && !geoparquetLoading) {
      loadGeoParquet(metadata.proxyUrl)
    }
  }, [metadata, geoparquetData, geoparquetLoading, loadGeoParquet])

  return { geoparquetData, geoparquetLoading, geoparquetError }
}
