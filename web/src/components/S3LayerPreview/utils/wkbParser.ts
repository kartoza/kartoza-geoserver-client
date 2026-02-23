import type { Geometry } from 'geojson'

/**
 * Pure JavaScript WKB (Well-Known Binary) parser
 * Browser-compatible, no Buffer dependency
 */
export function parseWKB(wkb: Uint8Array): Geometry | null {
  const view = new DataView(wkb.buffer, wkb.byteOffset, wkb.byteLength)
  let offset = 0

  const readByte = () => wkb[offset++]
  const readUInt32 = (littleEndian: boolean) => {
    const val = view.getUint32(offset, littleEndian)
    offset += 4
    return val
  }
  const readDouble = (littleEndian: boolean) => {
    const val = view.getFloat64(offset, littleEndian)
    offset += 8
    return val
  }

  const readPoint = (le: boolean): [number, number] => {
    const x = readDouble(le)
    const y = readDouble(le)
    return [x, y]
  }

  const readLinearRing = (le: boolean): [number, number][] => {
    const numPoints = readUInt32(le)
    const points: [number, number][] = []
    for (let i = 0; i < numPoints; i++) {
      points.push(readPoint(le))
    }
    return points
  }

  const readGeometry = (): Geometry | null => {
    const byteOrder = readByte()
    const littleEndian = byteOrder === 1
    let wkbType = readUInt32(littleEndian)

    // Handle EWKB with SRID (high bits set)
    const hasSRID = (wkbType & 0x20000000) !== 0
    if (hasSRID) {
      readUInt32(littleEndian) // Skip SRID
    }
    // Mask off SRID and Z/M flags to get base type
    wkbType = wkbType & 0xFF

    switch (wkbType) {
      case 1: // Point
        return { type: 'Point', coordinates: readPoint(littleEndian) }

      case 2: { // LineString
        const numPoints = readUInt32(littleEndian)
        const coords: [number, number][] = []
        for (let i = 0; i < numPoints; i++) {
          coords.push(readPoint(littleEndian))
        }
        return { type: 'LineString', coordinates: coords }
      }

      case 3: { // Polygon
        const numRings = readUInt32(littleEndian)
        const rings: [number, number][][] = []
        for (let i = 0; i < numRings; i++) {
          rings.push(readLinearRing(littleEndian))
        }
        return { type: 'Polygon', coordinates: rings }
      }

      case 4: { // MultiPoint
        const numPoints = readUInt32(littleEndian)
        const points: [number, number][] = []
        for (let i = 0; i < numPoints; i++) {
          const geom = readGeometry()
          if (geom && geom.type === 'Point') {
            points.push(geom.coordinates as [number, number])
          }
        }
        return { type: 'MultiPoint', coordinates: points }
      }

      case 5: { // MultiLineString
        const numLines = readUInt32(littleEndian)
        const lines: [number, number][][] = []
        for (let i = 0; i < numLines; i++) {
          const geom = readGeometry()
          if (geom && geom.type === 'LineString') {
            lines.push(geom.coordinates as [number, number][])
          }
        }
        return { type: 'MultiLineString', coordinates: lines }
      }

      case 6: { // MultiPolygon
        const numPolygons = readUInt32(littleEndian)
        const polygons: [number, number][][][] = []
        for (let i = 0; i < numPolygons; i++) {
          const geom = readGeometry()
          if (geom && geom.type === 'Polygon') {
            polygons.push(geom.coordinates as [number, number][][])
          }
        }
        return { type: 'MultiPolygon', coordinates: polygons }
      }

      case 7: { // GeometryCollection
        const numGeoms = readUInt32(littleEndian)
        const geometries: Geometry[] = []
        for (let i = 0; i < numGeoms; i++) {
          const geom = readGeometry()
          if (geom) geometries.push(geom)
        }
        return { type: 'GeometryCollection', geometries }
      }

      default:
        console.warn('Unknown WKB geometry type:', wkbType)
        return null
    }
  }

  try {
    return readGeometry()
  } catch (err) {
    console.error('Failed to parse WKB:', err)
    return null
  }
}

/**
 * Convert WKB data to GeoJSON geometry
 * Handles both raw Uint8Array WKB and already-parsed GeoJSON
 */
export function wkbToGeoJSON(wkbData: unknown): Geometry | null {
  try {
    if (wkbData instanceof Uint8Array) {
      return parseWKB(wkbData)
    } else if (wkbData && typeof wkbData === 'object' && 'type' in wkbData) {
      // Already GeoJSON
      return wkbData as Geometry
    }
    console.warn('Unknown geometry format:', typeof wkbData, wkbData)
    return null
  } catch (err) {
    console.error('Failed to parse WKB geometry:', err)
    return null
  }
}

/**
 * Convert BigInt values to numbers recursively
 * (MapLibre/JSON don't support BigInt)
 */
export function convertBigInts(obj: unknown): unknown {
  if (typeof obj === 'bigint') {
    return Number(obj)
  }
  if (Array.isArray(obj)) {
    return obj.map(convertBigInts)
  }
  if (obj !== null && typeof obj === 'object') {
    const result: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(obj)) {
      result[k] = convertBigInts(v)
    }
    return result
  }
  return obj
}
