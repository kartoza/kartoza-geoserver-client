import { useState, useEffect, useCallback, useRef } from 'react'
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalBody,
  ModalCloseButton,
  ModalHeader,
  Text,
  Box,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Spinner,
  Alert,
  AlertIcon,
  AlertDescription,
  HStack,
  VStack,
  Badge,
  Button,
  IconButton,
  Flex,
  useColorModeValue,
  Tooltip,
  Textarea,
  Select,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Divider,
  Code,
  useToast,
} from '@chakra-ui/react'
import {
  FiPlay,
  FiChevronLeft,
  FiChevronRight,
  FiDownload,
  FiMap,
  FiTable,
  FiCode,
  FiDatabase,
} from 'react-icons/fi'
import { useUIStore } from '../../stores/uiStore'
import * as api from '../../api/client'
import type { DuckDBTableInfo, DuckDBQueryResponse } from '../../types'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'

export default function DuckDBQueryDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const dialogData = useUIStore((state) => state.dialogData)
  const closeDialog = useUIStore((state) => state.closeDialog)

  const isOpen = activeDialog === 'duckdbquery'

  if (!isOpen) {
    return null
  }

  return (
    <Modal isOpen={isOpen} onClose={closeDialog} size="full" scrollBehavior="inside">
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent maxH="95vh" maxW="95vw" m={4} borderRadius="xl" overflow="hidden">
        <DuckDBQueryContent dialogData={dialogData} />
      </ModalContent>
    </Modal>
  )
}

interface DuckDBQueryContentProps {
  dialogData: {
    data?: {
      s3ConnectionId?: string
      s3BucketName?: string
      s3ObjectKey?: string
      displayName?: string
    }
  } | null
}

function DuckDBQueryContent({ dialogData }: DuckDBQueryContentProps) {
  const [tableInfo, setTableInfo] = useState<DuckDBTableInfo | null>(null)
  const [queryResult, setQueryResult] = useState<DuckDBQueryResponse | null>(null)
  const [geojsonData, setGeojsonData] = useState<GeoJSON.FeatureCollection | null>(null)
  const [sql, setSql] = useState('')
  const [loading, setLoading] = useState(true)
  const [queryLoading, setQueryLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(100)
  const [activeTab, setActiveTab] = useState(0) // 0=table, 1=map
  const toast = useToast()
  const mapContainerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)

  const tableBg = useColorModeValue('white', 'gray.800')
  const headerBg = useColorModeValue('gray.50', 'gray.700')
  const borderColor = useColorModeValue('gray.200', 'gray.600')
  const hoverBg = useColorModeValue('gray.50', 'gray.700')
  const codeBg = useColorModeValue('gray.100', 'gray.700')

  const connectionId = dialogData?.data?.s3ConnectionId ?? ''
  const bucketName = dialogData?.data?.s3BucketName ?? ''
  const objectKey = dialogData?.data?.s3ObjectKey ?? ''
  const displayName = dialogData?.data?.displayName ?? objectKey.split('/').pop() ?? ''

  // Fetch table info on mount
  useEffect(() => {
    async function fetchTableInfo() {
      if (!connectionId || !bucketName || !objectKey) {
        setError('Missing file information')
        setLoading(false)
        return
      }

      setLoading(true)
      setError(null)

      try {
        const info = await api.getDuckDBTableInfo(connectionId, bucketName, objectKey)
        setTableInfo(info)
        // Set default query
        setSql(info.sampleQueries?.[0] || "SELECT * FROM 'data' LIMIT 100")
      } catch (err) {
        setError((err as Error).message)
      } finally {
        setLoading(false)
      }
    }

    fetchTableInfo()
  }, [connectionId, bucketName, objectKey])

  // Execute SQL query
  const executeQuery = useCallback(async () => {
    if (!connectionId || !bucketName || !objectKey || !sql.trim()) {
      return
    }

    setQueryLoading(true)
    setError(null)

    try {
      const result = await api.executeDuckDBQuery(connectionId, bucketName, objectKey, {
        sql,
        limit: pageSize,
        offset: page * pageSize,
      })

      if (result.error) {
        setError(result.error)
        setQueryResult(null)
      } else {
        setQueryResult(result)
        setError(null)
      }

      // If the result has geometry, also fetch GeoJSON for map
      if (result.geometryColumn && !result.error) {
        try {
          const geojson = await api.executeDuckDBQueryAsGeoJSON(connectionId, bucketName, objectKey, {
            sql,
            limit: 1000, // Limit for map performance
          })
          setGeojsonData(geojson)
        } catch {
          // GeoJSON fetch can fail silently, table view still works
        }
      }
    } catch (err) {
      setError((err as Error).message)
      setQueryResult(null)
    } finally {
      setQueryLoading(false)
    }
  }, [connectionId, bucketName, objectKey, sql, page, pageSize])

  // Initialize map when switching to map tab
  useEffect(() => {
    if (activeTab !== 1 || !mapContainerRef.current || !geojsonData) return

    // Clean up existing map
    if (mapRef.current) {
      mapRef.current.remove()
      mapRef.current = null
    }

    const map = new maplibregl.Map({
      container: mapContainerRef.current,
      style: {
        version: 8,
        sources: {
          osm: {
            type: 'raster',
            tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
            tileSize: 256,
            attribution: '&copy; OpenStreetMap Contributors',
          },
        },
        layers: [
          {
            id: 'osm',
            type: 'raster',
            source: 'osm',
          },
        ],
      },
      center: [0, 0],
      zoom: 2,
    })

    map.addControl(new maplibregl.NavigationControl(), 'top-right')

    map.on('load', () => {
      // Add GeoJSON source
      map.addSource('query-results', {
        type: 'geojson',
        data: geojsonData,
      })

      // Detect geometry type and add appropriate layer
      const features = geojsonData.features
      if (features.length > 0) {
        const firstGeom = features[0].geometry
        if (firstGeom) {
          const geomType = firstGeom.type

          if (geomType === 'Point' || geomType === 'MultiPoint') {
            map.addLayer({
              id: 'query-results-layer',
              type: 'circle',
              source: 'query-results',
              paint: {
                'circle-radius': 6,
                'circle-color': '#3182ce',
                'circle-stroke-color': '#ffffff',
                'circle-stroke-width': 2,
              },
            })
          } else if (geomType === 'LineString' || geomType === 'MultiLineString') {
            map.addLayer({
              id: 'query-results-layer',
              type: 'line',
              source: 'query-results',
              paint: {
                'line-color': '#3182ce',
                'line-width': 3,
              },
            })
          } else {
            // Polygon or other
            map.addLayer({
              id: 'query-results-fill',
              type: 'fill',
              source: 'query-results',
              paint: {
                'fill-color': '#3182ce',
                'fill-opacity': 0.3,
              },
            })
            map.addLayer({
              id: 'query-results-layer',
              type: 'line',
              source: 'query-results',
              paint: {
                'line-color': '#3182ce',
                'line-width': 2,
              },
            })
          }
        }

        // Fit bounds to data
        const bounds = new maplibregl.LngLatBounds()
        features.forEach((feature) => {
          if (feature.geometry) {
            const coords = getCoordinates(feature.geometry)
            coords.forEach((coord) => bounds.extend(coord as [number, number]))
          }
        })
        if (!bounds.isEmpty()) {
          map.fitBounds(bounds, { padding: 50, maxZoom: 15 })
        }
      }

      // Add click popup
      map.on('click', 'query-results-layer', (e) => {
        if (e.features && e.features[0]) {
          const props = e.features[0].properties
          const html = Object.entries(props || {})
            .map(([k, v]) => `<strong>${k}:</strong> ${v}`)
            .join('<br>')

          new maplibregl.Popup()
            .setLngLat(e.lngLat)
            .setHTML(`<div style="max-height: 200px; overflow-y: auto;">${html}</div>`)
            .addTo(map)
        }
      })

      map.on('mouseenter', 'query-results-layer', () => {
        map.getCanvas().style.cursor = 'pointer'
      })
      map.on('mouseleave', 'query-results-layer', () => {
        map.getCanvas().style.cursor = ''
      })
    })

    mapRef.current = map

    return () => {
      if (mapRef.current) {
        mapRef.current.remove()
        mapRef.current = null
      }
    }
  }, [activeTab, geojsonData])

  // Helper to extract coordinates from geometry
  function getCoordinates(geometry: GeoJSON.Geometry): [number, number][] {
    const coords: [number, number][] = []
    const extract = (c: GeoJSON.Position | GeoJSON.Position[] | GeoJSON.Position[][] | GeoJSON.Position[][][]) => {
      if (typeof c[0] === 'number') {
        coords.push([c[0] as number, c[1] as number])
      } else {
        (c as unknown[]).forEach((inner) => extract(inner as GeoJSON.Position))
      }
    }
    if ('coordinates' in geometry) {
      extract(geometry.coordinates)
    }
    return coords
  }

  const handleExportCSV = () => {
    if (!queryResult?.rows) return

    const headers = queryResult.columns.join(',')
    const rows = queryResult.rows.map((row) =>
      queryResult.columns.map((col) => {
        const cell = row[col]
        if (cell === null || cell === undefined) return ''
        const str = String(cell)
        if (str.includes(',') || str.includes('"') || str.includes('\n')) {
          return `"${str.replace(/"/g, '""')}"`
        }
        return str
      }).join(',')
    ).join('\n')

    const csv = `${headers}\n${rows}`
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${displayName}_query_result.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    toast({
      title: 'Export Complete',
      description: `Exported ${queryResult.rowCount} rows to CSV`,
      status: 'success',
      duration: 3000,
    })
  }

  const handleSampleQuery = (query: string) => {
    setSql(query)
    setPage(0)
  }

  // Loading state
  if (loading) {
    return (
      <>
        <ModalHeader borderBottomWidth="1px">
          <HStack spacing={3}>
            <FiDatabase />
            <Text>Loading {displayName}...</Text>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Flex justify="center" align="center" h="400px">
            <Spinner size="xl" />
          </Flex>
        </ModalBody>
      </>
    )
  }

  // Error loading table info
  if (error && !tableInfo) {
    return (
      <>
        <ModalHeader borderBottomWidth="1px">
          <HStack spacing={3}>
            <FiDatabase />
            <Text>Query {displayName}</Text>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Alert status="error" borderRadius="md">
            <AlertIcon />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        </ModalBody>
      </>
    )
  }

  return (
    <>
      <ModalHeader borderBottomWidth="1px" py={3}>
        <HStack spacing={3} justify="space-between">
          <HStack spacing={3}>
            <FiDatabase />
            <Text>DuckDB Query: {displayName}</Text>
            {tableInfo && (
              <>
                <Badge colorScheme="blue">{tableInfo.rowCount.toLocaleString()} rows</Badge>
                <Badge colorScheme="green">{tableInfo.columns.length} columns</Badge>
                {tableInfo.geometryColumn && (
                  <Badge colorScheme="purple">Spatial: {tableInfo.geometryColumn}</Badge>
                )}
              </>
            )}
          </HStack>
          <ModalCloseButton position="static" />
        </HStack>
      </ModalHeader>

      <ModalBody p={0}>
        <Flex h="calc(95vh - 120px)" direction="column">
          {/* SQL Editor Section */}
          <Box p={4} borderBottomWidth="1px">
            <VStack spacing={3} align="stretch">
              {/* Sample queries */}
              {tableInfo?.sampleQueries && tableInfo.sampleQueries.length > 0 && (
                <HStack spacing={2} flexWrap="wrap">
                  <Text fontSize="sm" fontWeight="medium">Sample Queries:</Text>
                  {tableInfo.sampleQueries.map((query, i) => (
                    <Button
                      key={i}
                      size="xs"
                      variant="outline"
                      onClick={() => handleSampleQuery(query)}
                      fontFamily="mono"
                      maxW="300px"
                      isTruncated
                    >
                      {query.substring(0, 40)}...
                    </Button>
                  ))}
                </HStack>
              )}

              {/* SQL Input */}
              <Textarea
                value={sql}
                onChange={(e) => setSql(e.target.value)}
                placeholder="Enter SQL query... Use 'data' as the table name (e.g., SELECT * FROM data LIMIT 10)"
                fontFamily="mono"
                fontSize="sm"
                minH="100px"
                bg={codeBg}
                borderRadius="md"
              />

              <HStack justify="space-between">
                <HStack spacing={2}>
                  <Button
                    leftIcon={<FiPlay />}
                    colorScheme="blue"
                    onClick={executeQuery}
                    isLoading={queryLoading}
                    loadingText="Executing..."
                  >
                    Execute Query
                  </Button>
                  <Select
                    value={pageSize}
                    onChange={(e) => {
                      setPageSize(Number(e.target.value))
                      setPage(0)
                    }}
                    w="120px"
                  >
                    <option value={50}>50 rows</option>
                    <option value={100}>100 rows</option>
                    <option value={500}>500 rows</option>
                    <option value={1000}>1000 rows</option>
                  </Select>
                </HStack>

                {queryResult && (
                  <HStack spacing={2}>
                    <Tooltip label="Export to CSV">
                      <IconButton
                        aria-label="Export CSV"
                        icon={<FiDownload />}
                        size="sm"
                        onClick={handleExportCSV}
                      />
                    </Tooltip>
                    <Badge colorScheme="green">
                      {queryResult.rowCount} results
                      {queryResult.hasMore && ' (more available)'}
                    </Badge>
                  </HStack>
                )}
              </HStack>
            </VStack>
          </Box>

          {/* Query Error */}
          {error && (
            <Box px={4} pt={2}>
              <Alert status="error" borderRadius="md">
                <AlertIcon />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            </Box>
          )}

          {/* Results Section */}
          <Box flex={1} overflow="hidden">
            <Tabs
              index={activeTab}
              onChange={setActiveTab}
              h="100%"
              display="flex"
              flexDirection="column"
            >
              <TabList px={4}>
                <Tab>
                  <HStack spacing={2}>
                    <FiTable />
                    <Text>Table View</Text>
                  </HStack>
                </Tab>
                {(queryResult?.geometryColumn || geojsonData) && (
                  <Tab>
                    <HStack spacing={2}>
                      <FiMap />
                      <Text>Map View</Text>
                    </HStack>
                  </Tab>
                )}
                <Tab>
                  <HStack spacing={2}>
                    <FiCode />
                    <Text>Schema</Text>
                  </HStack>
                </Tab>
              </TabList>

              <TabPanels flex={1} overflow="hidden">
                {/* Table View */}
                <TabPanel h="100%" p={0} overflow="auto">
                  {queryLoading ? (
                    <Flex justify="center" align="center" h="200px">
                      <Spinner size="xl" />
                    </Flex>
                  ) : queryResult && queryResult.rows.length > 0 ? (
                    <Box overflow="auto" h="100%">
                      <Table size="sm" variant="striped">
                        <Thead position="sticky" top={0} bg={headerBg} zIndex={1}>
                          <Tr>
                            {queryResult.columns.map((col) => (
                              <Th key={col} borderColor={borderColor}>
                                {col}
                              </Th>
                            ))}
                          </Tr>
                        </Thead>
                        <Tbody>
                          {queryResult.rows.map((row, i) => (
                            <Tr key={i} _hover={{ bg: hoverBg }}>
                              {queryResult.columns.map((col) => (
                                <Td key={col} borderColor={borderColor} maxW="300px" isTruncated>
                                  {formatCellValue(row[col])}
                                </Td>
                              ))}
                            </Tr>
                          ))}
                        </Tbody>
                      </Table>

                      {/* Pagination */}
                      <Flex justify="center" py={4} borderTopWidth="1px" bg={tableBg}>
                        <HStack spacing={4}>
                          <IconButton
                            aria-label="Previous page"
                            icon={<FiChevronLeft />}
                            size="sm"
                            onClick={() => {
                              setPage((p) => Math.max(0, p - 1))
                              executeQuery()
                            }}
                            isDisabled={page === 0}
                          />
                          <Text fontSize="sm">
                            Page {page + 1}
                            {queryResult.hasMore && ' (more available)'}
                          </Text>
                          <IconButton
                            aria-label="Next page"
                            icon={<FiChevronRight />}
                            size="sm"
                            onClick={() => {
                              setPage((p) => p + 1)
                              executeQuery()
                            }}
                            isDisabled={!queryResult.hasMore}
                          />
                        </HStack>
                      </Flex>
                    </Box>
                  ) : (
                    <Flex justify="center" align="center" h="200px">
                      <Text color="gray.500">
                        {queryResult ? 'No results' : 'Execute a query to see results'}
                      </Text>
                    </Flex>
                  )}
                </TabPanel>

                {/* Map View */}
                {(queryResult?.geometryColumn || geojsonData) && (
                  <TabPanel h="100%" p={0}>
                    {geojsonData ? (
                      <Box ref={mapContainerRef} w="100%" h="100%" />
                    ) : (
                      <Flex justify="center" align="center" h="100%">
                        <VStack>
                          <Text color="gray.500">Execute a query to see results on the map</Text>
                          {queryResult?.geometryColumn && (
                            <Text fontSize="sm" color="gray.400">
                              Geometry column detected: {queryResult.geometryColumn}
                            </Text>
                          )}
                        </VStack>
                      </Flex>
                    )}
                  </TabPanel>
                )}

                {/* Schema View */}
                <TabPanel h="100%" overflow="auto">
                  {tableInfo && (
                    <Box p={4}>
                      <VStack align="stretch" spacing={4}>
                        <Box>
                          <Text fontWeight="bold" mb={2}>File Information</Text>
                          <HStack spacing={4} flexWrap="wrap">
                            <Badge>Rows: {tableInfo.rowCount.toLocaleString()}</Badge>
                            <Badge>Columns: {tableInfo.columns.length}</Badge>
                            {tableInfo.geometryColumn && (
                              <Badge colorScheme="purple">
                                Geometry: {tableInfo.geometryColumn}
                              </Badge>
                            )}
                            {tableInfo.bbox && (
                              <Badge colorScheme="green">
                                Bbox: [{tableInfo.bbox.map((v) => v.toFixed(4)).join(', ')}]
                              </Badge>
                            )}
                          </HStack>
                        </Box>

                        <Divider />

                        <Box>
                          <Text fontWeight="bold" mb={2}>Columns</Text>
                          <Table size="sm" variant="simple">
                            <Thead>
                              <Tr>
                                <Th>Name</Th>
                                <Th>Type</Th>
                              </Tr>
                            </Thead>
                            <Tbody>
                              {tableInfo.columns.map((col) => (
                                <Tr key={col.name}>
                                  <Td fontFamily="mono">{col.name}</Td>
                                  <Td>
                                    <Code fontSize="xs">{col.type}</Code>
                                  </Td>
                                </Tr>
                              ))}
                            </Tbody>
                          </Table>
                        </Box>
                      </VStack>
                    </Box>
                  )}
                </TabPanel>
              </TabPanels>
            </Tabs>
          </Box>
        </Flex>
      </ModalBody>
    </>
  )
}

// Helper to format cell values
function formatCellValue(value: unknown): string {
  if (value === null || value === undefined) return ''
  if (typeof value === 'object') {
    // Handle geometry or JSON objects
    const str = JSON.stringify(value)
    return str.length > 100 ? str.substring(0, 97) + '...' : str
  }
  const str = String(value)
  return str.length > 100 ? str.substring(0, 97) + '...' : str
}
