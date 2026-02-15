import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import {
  Box,
  Flex,
  VStack,
  HStack,
  Text,
  Button,
  IconButton,
  Tabs,
  TabList,
  Tab,
  Select,
  Input,
  Textarea,
  Checkbox,
  Badge,
  Tooltip,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Icon,
  useColorModeValue,
  Divider,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Tag,
  TagLabel,
  Wrap,
  WrapItem,
  useToast,
} from '@chakra-ui/react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  FiPlay,
  FiCpu,
  FiGrid,
  FiCode,
  FiDatabase,
  FiTable,
  FiColumns,
  FiFilter,
  FiArrowDown,
  FiPlus,
  FiTrash2,
  FiX,
  FiChevronDown,
  FiUpload,
  FiCopy,
  FiDownload,
  FiSearch,
  FiHelpCircle,
  FiZap,
} from 'react-icons/fi'
import { SQLEditor } from './SQLEditor'
import { springs } from '../utils/animations'

const MotionBox = motion(Box)

// Types
interface Column {
  name: string
  alias?: string
  aggregate?: string
}

interface Condition {
  column: string
  operator: string
  value: string
  logic: 'AND' | 'OR'
}

interface OrderBy {
  column: string
  direction: 'ASC' | 'DESC'
}

interface SchemaInfo {
  name: string
  tables: TableInfo[]
}

interface TableInfo {
  name: string
  columns: ColumnInfo[]
  has_geometry?: boolean
  geometry_column?: string
}

interface ColumnInfo {
  name: string
  type: string
  nullable: boolean
}

interface QueryResult {
  columns: ColumnInfo[]
  rows: Record<string, unknown>[]
  row_count: number
  total_count?: number
  duration_ms: number
  sql: string
  has_more?: boolean
}

interface QueryPanelProps {
  serviceName: string
  initialSchema?: string
  initialTable?: string
  initialSQL?: string
  onClose?: () => void
  onPublishToGeoServer?: (sql: string, name: string) => void
}

const OPERATORS = [
  { value: '=', label: 'equals' },
  { value: '!=', label: 'not equals' },
  { value: '<', label: 'less than' },
  { value: '<=', label: 'less or equal' },
  { value: '>', label: 'greater than' },
  { value: '>=', label: 'greater or equal' },
  { value: 'LIKE', label: 'contains' },
  { value: 'ILIKE', label: 'contains (case insensitive)' },
  { value: 'IS NULL', label: 'is null' },
  { value: 'IS NOT NULL', label: 'is not null' },
  { value: 'IN', label: 'in list' },
]

const AGGREGATES = [
  { value: '', label: 'None' },
  { value: 'COUNT', label: 'Count' },
  { value: 'SUM', label: 'Sum' },
  { value: 'AVG', label: 'Average' },
  { value: 'MIN', label: 'Min' },
  { value: 'MAX', label: 'Max' },
  { value: 'ST_Extent', label: 'Extent (geo)' },
  { value: 'ST_Union', label: 'Union (geo)' },
]

const AI_EXAMPLE_QUESTIONS = [
  'Show me all records from the last 30 days',
  'Find the top 10 largest by area',
  'Count records grouped by type',
  'Show records where name contains "park"',
  'Find all polygons that intersect with each other',
]

export const QueryPanel: React.FC<QueryPanelProps> = ({
  serviceName,
  initialSchema = 'public',
  initialTable = '',
  initialSQL = '',
  onClose,
  onPublishToGeoServer,
}) => {
  const toast = useToast()

  // Colors
  const bgColor = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')
  const headerBg = useColorModeValue('gray.50', 'gray.700')
  const tableBg = useColorModeValue('gray.50', 'gray.900')
  const hoverBg = useColorModeValue('blue.50', 'blue.900')

  // Mode state
  const [activeMode, setActiveMode] = useState<'ai' | 'visual' | 'sql'>('visual')

  // Schema state
  const [schemas, setSchemas] = useState<SchemaInfo[]>([])
  const [, setLoadingSchemas] = useState(true)

  // Visual builder state
  const [selectedSchema, setSelectedSchema] = useState(initialSchema)
  const [selectedTable, setSelectedTable] = useState(initialTable)
  const [selectedColumns, setSelectedColumns] = useState<Column[]>([])
  const [conditions, setConditions] = useState<Condition[]>([])
  const [orderBy, setOrderBy] = useState<OrderBy[]>([])
  const [limit, setLimit] = useState(100)
  const [offset, setOffset] = useState(0)
  const [distinct, setDistinct] = useState(false)

  // SQL Editor state
  const [sqlQuery, setSqlQuery] = useState(initialSQL)
  const [generatedSQL, setGeneratedSQL] = useState('')

  // AI state
  const [aiQuestion, setAiQuestion] = useState('')
  const [aiLoading, setAiLoading] = useState(false)
  const [aiResponse, setAiResponse] = useState<{
    sql?: string
    explanation?: string
    confidence?: number
    warnings?: string[]
  } | null>(null)
  const [aiProviderAvailable, setAiProviderAvailable] = useState(true)

  // Execution state
  const [executing, setExecuting] = useState(false)
  const [result, setResult] = useState<QueryResult | null>(null)
  const [error, setError] = useState('')
  const [loadingMore, setLoadingMore] = useState(false)

  // Infinite scroll ref
  const tableContainerRef = useRef<HTMLDivElement>(null)

  // Resizable splitter state
  const [splitPosition, setSplitPosition] = useState(() => {
    const saved = localStorage.getItem('query-panel-split')
    return saved ? parseInt(saved, 10) : 50
  })
  const [isDragging, setIsDragging] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const splitPositionRef = useRef(splitPosition)

  // Current schema/table info
  const currentSchema = useMemo(() =>
    schemas.find(s => s.name === selectedSchema),
    [schemas, selectedSchema]
  )
  const currentTable = useMemo(() =>
    currentSchema?.tables.find(t => t.name === selectedTable),
    [currentSchema, selectedTable]
  )
  const availableColumns = useMemo(() =>
    currentTable?.columns || [],
    [currentTable]
  )

  // Load schemas on mount
  useEffect(() => {
    setLoadingSchemas(true)
    fetch(`/api/pg/services/${encodeURIComponent(serviceName)}/schemas`)
      .then(res => res.json())
      .then(data => {
        if (data.schemas) {
          setSchemas(data.schemas)
          // Auto-select first schema if needed
          if (data.schemas.length > 0 && !selectedSchema) {
            setSelectedSchema(data.schemas[0].name)
          }
        }
      })
      .catch(err => {
        console.error('Failed to load schemas:', err)
        setError('Failed to load database schemas')
      })
      .finally(() => setLoadingSchemas(false))
  }, [serviceName])

  // Check AI provider availability
  useEffect(() => {
    fetch('/api/ai/providers')
      .then(res => res.json())
      .then(data => {
        const active = data.providers?.find((p: { active: boolean }) => p.active)
        setAiProviderAvailable(active?.available ?? false)
      })
      .catch(() => setAiProviderAvailable(false))
  }, [])

  // Generate SQL from visual builder
  useEffect(() => {
    if (activeMode === 'visual') {
      if (selectedTable) {
        const sql = buildSQL()
        setGeneratedSQL(sql)
      } else {
        setGeneratedSQL('')
      }
    }
  }, [activeMode, selectedSchema, selectedTable, selectedColumns, conditions, orderBy, limit, distinct])

  const buildSQL = useCallback(() => {
    if (!selectedTable) return ''

    const cols = selectedColumns.length > 0
      ? selectedColumns.map(c => {
          if (c.aggregate) {
            return `${c.aggregate}(${c.name})${c.alias ? ` AS ${c.alias}` : ''}`
          }
          return c.alias ? `${c.name} AS ${c.alias}` : c.name
        }).join(', ')
      : '*'

    let sql = `SELECT ${distinct ? 'DISTINCT ' : ''}${cols}\nFROM ${selectedSchema}.${selectedTable}`

    if (conditions.length > 0) {
      const whereClause = conditions.map((c, i) => {
        let clause = ''
        if (i > 0) clause += ` ${c.logic} `
        if (['IS NULL', 'IS NOT NULL'].includes(c.operator)) {
          clause += `${c.column} ${c.operator}`
        } else if (c.operator === 'IN') {
          clause += `${c.column} IN (${c.value})`
        } else if (['LIKE', 'ILIKE'].includes(c.operator)) {
          clause += `${c.column} ${c.operator} '%${c.value}%'`
        } else {
          clause += `${c.column} ${c.operator} '${c.value}'`
        }
        return clause
      }).join('')
      sql += `\nWHERE ${whereClause}`
    }

    // Group by if aggregates are used
    const hasAggregates = selectedColumns.some(c => c.aggregate)
    const nonAggColumns = selectedColumns.filter(c => !c.aggregate)
    if (hasAggregates && nonAggColumns.length > 0) {
      sql += `\nGROUP BY ${nonAggColumns.map(c => c.name).join(', ')}`
    }

    if (orderBy.length > 0) {
      sql += `\nORDER BY ${orderBy.map(o => `${o.column} ${o.direction}`).join(', ')}`
    }

    sql += `\nLIMIT ${limit}`
    if (offset > 0) {
      sql += ` OFFSET ${offset}`
    }

    return sql
  }, [selectedSchema, selectedTable, selectedColumns, conditions, orderBy, limit, offset, distinct])

  // Execute query
  const executeQuery = useCallback(async (sql: string, appendResults = false) => {
    if (!sql.trim()) {
      setError('Please enter a query')
      return
    }

    if (appendResults) {
      setLoadingMore(true)
    } else {
      setExecuting(true)
      setError('')
      setResult(null)
      setOffset(0)
    }

    try {
      const response = await fetch('/api/query/execute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sql,
          service_name: serviceName,
          max_rows: limit,
          offset: appendResults ? (result?.rows.length || 0) : 0,
        }),
      })

      const data = await response.json()

      if (data.success) {
        // Ensure rows is always an array
        const rows = data.result?.rows || []
        const columns = data.result?.columns || []

        if (appendResults && result) {
          setResult({
            ...data.result,
            columns,
            rows: [...(result.rows || []), ...rows],
            has_more: rows.length === limit,
          })
        } else {
          setResult({
            ...data.result,
            columns,
            rows,
            has_more: rows.length === limit,
          })
        }
        setGeneratedSQL(data.sql || sql)
      } else {
        setError(data.error || 'Query execution failed')
      }
    } catch (err) {
      setError('Failed to execute query: ' + (err as Error).message)
    } finally {
      setExecuting(false)
      setLoadingMore(false)
    }
  }, [serviceName, limit, result])

  // AI query handler
  const handleAIQuery = useCallback(async () => {
    if (!aiQuestion.trim() || aiLoading) return

    setAiLoading(true)
    setAiResponse(null)
    setError('')

    try {
      const response = await fetch('/api/ai/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          question: aiQuestion.trim(),
          service_name: serviceName,
          schema_name: selectedSchema,
          max_rows: limit,
          execute: false,
        }),
      })

      const data = await response.json()

      if (data.success && data.sql) {
        setAiResponse({
          sql: data.sql,
          explanation: data.explanation,
          confidence: data.confidence,
          warnings: data.warnings,
        })
        setSqlQuery(data.sql)
      } else {
        setError(data.error || 'AI query generation failed')
      }
    } catch (err) {
      setError('AI query failed: ' + (err as Error).message)
    } finally {
      setAiLoading(false)
    }
  }, [aiQuestion, aiLoading, serviceName, selectedSchema, limit])

  // Handle execute based on mode
  const handleExecute = useCallback(() => {
    let sql = ''
    switch (activeMode) {
      case 'ai':
        sql = aiResponse?.sql || sqlQuery
        break
      case 'visual':
        sql = generatedSQL
        break
      case 'sql':
        sql = sqlQuery
        break
    }
    executeQuery(sql)
  }, [activeMode, aiResponse, generatedSQL, sqlQuery, executeQuery])

  // Load more (infinite scroll)
  const loadMore = useCallback(() => {
    if (!result?.has_more || loadingMore) return

    let sql = ''
    switch (activeMode) {
      case 'ai':
        sql = aiResponse?.sql || sqlQuery
        break
      case 'visual':
        sql = generatedSQL
        break
      case 'sql':
        sql = sqlQuery
        break
    }

    // Modify SQL to add offset
    const currentOffset = result.rows.length
    const modifiedSQL = sql.replace(/LIMIT \d+(\s+OFFSET \d+)?/i, `LIMIT ${limit} OFFSET ${currentOffset}`)

    executeQuery(modifiedSQL.includes('OFFSET') ? modifiedSQL : `${sql} OFFSET ${currentOffset}`, true)
  }, [result, loadingMore, activeMode, aiResponse, generatedSQL, sqlQuery, limit, executeQuery])

  // Infinite scroll handler
  useEffect(() => {
    const container = tableContainerRef.current
    if (!container) return

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container
      if (scrollHeight - scrollTop - clientHeight < 100 && result?.has_more && !loadingMore) {
        loadMore()
      }
    }

    container.addEventListener('scroll', handleScroll)
    return () => container.removeEventListener('scroll', handleScroll)
  }, [result, loadingMore, loadMore])

  // Splitter drag handlers
  const handleSplitterMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  // Keep ref in sync with state
  useEffect(() => {
    splitPositionRef.current = splitPosition
  }, [splitPosition])

  useEffect(() => {
    if (!isDragging) return

    const handleMouseMove = (e: MouseEvent) => {
      if (!containerRef.current) return
      const rect = containerRef.current.getBoundingClientRect()
      const newPosition = ((e.clientX - rect.left) / rect.width) * 100
      const clamped = Math.max(25, Math.min(75, newPosition))
      setSplitPosition(clamped)
    }

    const handleMouseUp = () => {
      setIsDragging(false)
      localStorage.setItem('query-panel-split', splitPositionRef.current.toString())
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging])

  // Column handlers
  const addColumn = () => setSelectedColumns([...selectedColumns, { name: '' }])
  const removeColumn = (index: number) => setSelectedColumns(selectedColumns.filter((_, i) => i !== index))
  const updateColumn = (index: number, updates: Partial<Column>) => {
    const newColumns = [...selectedColumns]
    newColumns[index] = { ...newColumns[index], ...updates }
    setSelectedColumns(newColumns)
  }

  // Condition handlers
  const addCondition = () => setConditions([...conditions, { column: '', operator: '=', value: '', logic: 'AND' }])
  const removeCondition = (index: number) => setConditions(conditions.filter((_, i) => i !== index))
  const updateCondition = (index: number, updates: Partial<Condition>) => {
    const newConditions = [...conditions]
    newConditions[index] = { ...newConditions[index], ...updates }
    setConditions(newConditions)
  }

  // Order by handlers
  const addOrderBy = () => setOrderBy([...orderBy, { column: '', direction: 'ASC' }])
  const removeOrderBy = (index: number) => setOrderBy(orderBy.filter((_, i) => i !== index))
  const updateOrderBy = (index: number, updates: Partial<OrderBy>) => {
    const newOrderBy = [...orderBy]
    newOrderBy[index] = { ...newOrderBy[index], ...updates }
    setOrderBy(newOrderBy)
  }

  // Copy SQL to clipboard
  const copySQL = () => {
    const sql = activeMode === 'visual' ? generatedSQL : sqlQuery
    navigator.clipboard.writeText(sql)
    toast({
      title: 'Copied',
      description: 'SQL copied to clipboard',
      status: 'success',
      duration: 2000,
    })
  }

  // Export results
  const exportResults = (format: 'csv' | 'json') => {
    if (!result) return

    let content = ''
    let filename = ''
    let mimeType = ''

    if (format === 'csv') {
      const headers = result.columns.map(c => c.name).join(',')
      const rows = result.rows.map(row =>
        result.columns.map(c => {
          const val = row[c.name]
          if (val === null) return ''
          if (typeof val === 'string' && val.includes(',')) return `"${val}"`
          return String(val)
        }).join(',')
      ).join('\n')
      content = `${headers}\n${rows}`
      filename = 'query-results.csv'
      mimeType = 'text/csv'
    } else {
      content = JSON.stringify(result.rows, null, 2)
      filename = 'query-results.json'
      mimeType = 'application/json'
    }

    const blob = new Blob([content], { type: mimeType })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <MotionBox
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      transition={springs.default}
      bg={bgColor}
      borderRadius="2xl"
      boxShadow="2xl"
      overflow="hidden"
      h="100%"
      display="flex"
      flexDirection="column"
    >
      {/* Header */}
      <Flex
        bg="linear-gradient(135deg, #3182CE 0%, #2B6CB0 100%)"
        px={6}
        py={4}
        align="center"
        justify="space-between"
      >
        <HStack spacing={3}>
          <Box bg="whiteAlpha.200" p={2} borderRadius="lg">
            <Icon as={FiDatabase} boxSize={5} color="white" />
          </Box>
          <Box>
            <Text color="white" fontWeight="600" fontSize="lg">
              Query Designer
            </Text>
            <Text color="whiteAlpha.800" fontSize="sm">
              {serviceName}
            </Text>
          </Box>
        </HStack>
        <HStack spacing={2}>
          <Badge colorScheme="whiteAlpha" variant="subtle" px={3} py={1} borderRadius="full">
            {schemas.length} schemas
          </Badge>
          {onClose && (
            <IconButton
              aria-label="Close"
              icon={<FiX />}
              variant="ghost"
              color="white"
              _hover={{ bg: 'whiteAlpha.200' }}
              onClick={onClose}
            />
          )}
        </HStack>
      </Flex>

      {/* Mode Tabs */}
      <Tabs
        index={activeMode === 'ai' ? 0 : activeMode === 'visual' ? 1 : 2}
        onChange={(i) => setActiveMode(i === 0 ? 'ai' : i === 1 ? 'visual' : 'sql')}
        variant="soft-rounded"
        colorScheme="blue"
        px={4}
        py={3}
        bg={headerBg}
      >
        <TabList>
          <Tab gap={2}>
            <Icon as={FiCpu} />
            AI Natural Language
          </Tab>
          <Tab gap={2}>
            <Icon as={FiGrid} />
            Visual Builder
          </Tab>
          <Tab gap={2}>
            <Icon as={FiCode} />
            SQL Editor
          </Tab>
        </TabList>
      </Tabs>

      {/* Content Area */}
      <Flex ref={containerRef} flex={1} overflow="hidden" position="relative">
        {/* Left Panel - Query Builder */}
        <Flex
          w={`${splitPosition}%`}
          flexDirection="column"
          overflow="hidden"
        >
          {/* Scrollable content area */}
          <Box flex={1} overflow="auto" p={4}>
          <AnimatePresence mode="wait">
            {/* AI Mode */}
            {activeMode === 'ai' && (
              <MotionBox
                key="ai"
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                transition={springs.snappy}
              >
                <VStack spacing={4} align="stretch">
                  {/* AI Provider Status */}
                  {!aiProviderAvailable && (
                    <Alert status="warning" borderRadius="lg">
                      <AlertIcon />
                      <Box>
                        <AlertTitle>Ollama not available</AlertTitle>
                        <AlertDescription fontSize="sm">
                          Start Ollama with: <code>ollama serve</code>
                        </AlertDescription>
                      </Box>
                    </Alert>
                  )}

                  {/* Question Input */}
                  <Box>
                    <Text fontWeight="600" mb={2} fontSize="sm" color="gray.600">
                      Ask a question about your data
                    </Text>
                    <Textarea
                      value={aiQuestion}
                      onChange={(e) => setAiQuestion(e.target.value)}
                      placeholder="e.g., Show me all records where status is active..."
                      rows={4}
                      borderRadius="xl"
                      _focus={{
                        borderColor: 'purple.400',
                        boxShadow: '0 0 0 3px rgba(159, 122, 234, 0.2)',
                      }}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' && !e.shiftKey) {
                          e.preventDefault()
                          handleAIQuery()
                        }
                      }}
                    />
                  </Box>

                  {/* Schema Selection */}
                  <HStack>
                    <Select
                      value={selectedSchema}
                      onChange={(e) => setSelectedSchema(e.target.value)}
                      size="sm"
                      borderRadius="lg"
                      flex={1}
                    >
                      {schemas.map(s => (
                        <option key={s.name} value={s.name}>{s.name}</option>
                      ))}
                    </Select>
                    <Button
                      colorScheme="purple"
                      leftIcon={aiLoading ? <Spinner size="sm" /> : <FiZap />}
                      onClick={handleAIQuery}
                      isDisabled={!aiQuestion.trim() || aiLoading || !aiProviderAvailable}
                      borderRadius="xl"
                    >
                      Generate SQL
                    </Button>
                  </HStack>

                  {/* Example Questions */}
                  <Box>
                    <HStack mb={2}>
                      <Icon as={FiHelpCircle} color="gray.400" />
                      <Text fontSize="xs" color="gray.500">Examples</Text>
                    </HStack>
                    <Wrap spacing={2}>
                      {AI_EXAMPLE_QUESTIONS.map((q, i) => (
                        <WrapItem key={i}>
                          <Tag
                            size="sm"
                            variant="subtle"
                            colorScheme="purple"
                            cursor="pointer"
                            onClick={() => setAiQuestion(q)}
                            _hover={{ bg: 'purple.100' }}
                            borderRadius="full"
                          >
                            <TagLabel>{q}</TagLabel>
                          </Tag>
                        </WrapItem>
                      ))}
                    </Wrap>
                  </Box>

                  {/* AI Response */}
                  {aiResponse && (
                    <MotionBox
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      bg="purple.50"
                      p={4}
                      borderRadius="xl"
                      borderLeft="4px solid"
                      borderLeftColor="purple.400"
                    >
                      {aiResponse.confidence !== undefined && (
                        <HStack mb={2}>
                          <Badge
                            colorScheme={
                              aiResponse.confidence >= 0.8 ? 'green' :
                              aiResponse.confidence >= 0.5 ? 'yellow' : 'red'
                            }
                          >
                            {Math.round(aiResponse.confidence * 100)}% confidence
                          </Badge>
                        </HStack>
                      )}
                      {aiResponse.explanation && (
                        <Text fontSize="sm" color="purple.700" mb={2}>
                          {aiResponse.explanation}
                        </Text>
                      )}
                      {aiResponse.warnings && aiResponse.warnings.length > 0 && (
                        <Alert status="warning" size="sm" borderRadius="md" mb={2}>
                          <AlertIcon />
                          <VStack align="start" spacing={0}>
                            {aiResponse.warnings.map((w, i) => (
                              <Text key={i} fontSize="xs">{w}</Text>
                            ))}
                          </VStack>
                        </Alert>
                      )}
                    </MotionBox>
                  )}

                  {/* Generated SQL Preview */}
                  {(aiResponse?.sql || sqlQuery) && activeMode === 'ai' && (
                    <Box>
                      <Text fontWeight="600" mb={2} fontSize="sm" color="gray.600">
                        Generated SQL
                      </Text>
                      <SQLEditor
                        value={aiResponse?.sql || sqlQuery}
                        onChange={setSqlQuery}
                        height="150px"
                        serviceName={serviceName}
                        schemas={schemas}
                      />
                    </Box>
                  )}
                </VStack>
              </MotionBox>
            )}

            {/* Visual Builder Mode */}
            {activeMode === 'visual' && (
              <MotionBox
                key="visual"
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                transition={springs.snappy}
              >
                <VStack spacing={4} align="stretch">
                  {/* Table Selection */}
                  <HStack>
                    <Box flex={1}>
                      <Text fontSize="xs" fontWeight="600" color="gray.500" mb={1}>
                        Schema
                      </Text>
                      <Select
                        value={selectedSchema}
                        onChange={(e) => {
                          setSelectedSchema(e.target.value)
                          setSelectedTable('')
                          setSelectedColumns([])
                        }}
                        size="sm"
                        borderRadius="lg"
                      >
                        {schemas.map(s => (
                          <option key={s.name} value={s.name}>{s.name}</option>
                        ))}
                      </Select>
                    </Box>
                    <Box flex={2}>
                      <Text fontSize="xs" fontWeight="600" color="gray.500" mb={1}>
                        Table
                      </Text>
                      <Select
                        value={selectedTable}
                        onChange={(e) => {
                          setSelectedTable(e.target.value)
                          setSelectedColumns([])
                        }}
                        size="sm"
                        borderRadius="lg"
                        placeholder="Select table..."
                      >
                        {currentSchema?.tables.map(t => (
                          <option key={t.name} value={t.name}>
                            {t.name} {t.has_geometry && '(geo)'}
                          </option>
                        ))}
                      </Select>
                    </Box>
                  </HStack>

                  {/* Available Columns */}
                  {currentTable && (
                    <Box
                      bg={tableBg}
                      p={3}
                      borderRadius="lg"
                      maxH="120px"
                      overflowY="auto"
                    >
                      <HStack mb={2}>
                        <Icon as={FiColumns} color="gray.500" size="sm" />
                        <Text fontSize="xs" fontWeight="600" color="gray.500">
                          Available Columns ({availableColumns.length})
                        </Text>
                      </HStack>
                      <Wrap spacing={1}>
                        {availableColumns.map(col => (
                          <WrapItem key={col.name}>
                            <Tag
                              size="sm"
                              variant="subtle"
                              colorScheme="gray"
                              cursor="pointer"
                              onClick={() => {
                                if (!selectedColumns.find(c => c.name === col.name)) {
                                  setSelectedColumns([...selectedColumns, { name: col.name }])
                                }
                              }}
                            >
                              <TagLabel>{col.name}</TagLabel>
                              <Text fontSize="2xs" color="gray.400" ml={1}>
                                {col.type}
                              </Text>
                            </Tag>
                          </WrapItem>
                        ))}
                      </Wrap>
                    </Box>
                  )}

                  {/* Selected Columns */}
                  <Box>
                    <Flex justify="space-between" align="center" mb={2}>
                      <HStack>
                        <Icon as={FiTable} color="blue.500" />
                        <Text fontSize="sm" fontWeight="600">
                          SELECT Columns
                        </Text>
                      </HStack>
                      <Button
                        size="xs"
                        leftIcon={<FiPlus />}
                        onClick={addColumn}
                        variant="ghost"
                        colorScheme="blue"
                      >
                        Add
                      </Button>
                    </Flex>
                    {selectedColumns.length === 0 ? (
                      <Text fontSize="sm" color="gray.500" fontStyle="italic">
                        All columns (*)
                      </Text>
                    ) : (
                      <VStack spacing={2} align="stretch">
                        {selectedColumns.map((col, i) => (
                          <HStack key={i}>
                            <Select
                              value={col.name}
                              onChange={(e) => updateColumn(i, { name: e.target.value })}
                              size="sm"
                              flex={2}
                              borderRadius="lg"
                            >
                              <option value="">Select...</option>
                              {availableColumns.map(c => (
                                <option key={c.name} value={c.name}>{c.name}</option>
                              ))}
                            </Select>
                            <Select
                              value={col.aggregate || ''}
                              onChange={(e) => updateColumn(i, { aggregate: e.target.value })}
                              size="sm"
                              flex={1}
                              borderRadius="lg"
                            >
                              {AGGREGATES.map(a => (
                                <option key={a.value} value={a.value}>{a.label}</option>
                              ))}
                            </Select>
                            <Input
                              value={col.alias || ''}
                              onChange={(e) => updateColumn(i, { alias: e.target.value })}
                              placeholder="alias"
                              size="sm"
                              flex={1}
                              borderRadius="lg"
                            />
                            <IconButton
                              aria-label="Remove"
                              icon={<FiTrash2 />}
                              size="sm"
                              variant="ghost"
                              colorScheme="red"
                              onClick={() => removeColumn(i)}
                            />
                          </HStack>
                        ))}
                      </VStack>
                    )}
                  </Box>

                  <Divider />

                  {/* WHERE Conditions */}
                  <Box>
                    <Flex justify="space-between" align="center" mb={2}>
                      <HStack>
                        <Icon as={FiFilter} color="orange.500" />
                        <Text fontSize="sm" fontWeight="600">
                          WHERE Conditions
                        </Text>
                      </HStack>
                      <Button
                        size="xs"
                        leftIcon={<FiPlus />}
                        onClick={addCondition}
                        variant="ghost"
                        colorScheme="orange"
                      >
                        Add
                      </Button>
                    </Flex>
                    {conditions.length === 0 ? (
                      <Text fontSize="sm" color="gray.500" fontStyle="italic">
                        No conditions
                      </Text>
                    ) : (
                      <VStack spacing={2} align="stretch">
                        {conditions.map((cond, i) => (
                          <HStack key={i} flexWrap="wrap">
                            {i > 0 && (
                              <Select
                                value={cond.logic}
                                onChange={(e) => updateCondition(i, { logic: e.target.value as 'AND' | 'OR' })}
                                size="sm"
                                w="80px"
                                borderRadius="lg"
                              >
                                <option value="AND">AND</option>
                                <option value="OR">OR</option>
                              </Select>
                            )}
                            <Select
                              value={cond.column}
                              onChange={(e) => updateCondition(i, { column: e.target.value })}
                              size="sm"
                              flex={1}
                              minW="100px"
                              borderRadius="lg"
                            >
                              <option value="">Column...</option>
                              {availableColumns.map(c => (
                                <option key={c.name} value={c.name}>{c.name}</option>
                              ))}
                            </Select>
                            <Select
                              value={cond.operator}
                              onChange={(e) => updateCondition(i, { operator: e.target.value })}
                              size="sm"
                              w="150px"
                              borderRadius="lg"
                            >
                              {OPERATORS.map(op => (
                                <option key={op.value} value={op.value}>{op.label}</option>
                              ))}
                            </Select>
                            {!['IS NULL', 'IS NOT NULL'].includes(cond.operator) && (
                              <Input
                                value={cond.value}
                                onChange={(e) => updateCondition(i, { value: e.target.value })}
                                placeholder="value"
                                size="sm"
                                flex={1}
                                minW="80px"
                                borderRadius="lg"
                              />
                            )}
                            <IconButton
                              aria-label="Remove"
                              icon={<FiTrash2 />}
                              size="sm"
                              variant="ghost"
                              colorScheme="red"
                              onClick={() => removeCondition(i)}
                            />
                          </HStack>
                        ))}
                      </VStack>
                    )}
                  </Box>

                  <Divider />

                  {/* ORDER BY */}
                  <Box>
                    <Flex justify="space-between" align="center" mb={2}>
                      <HStack>
                        <Icon as={FiArrowDown} color="green.500" />
                        <Text fontSize="sm" fontWeight="600">
                          ORDER BY
                        </Text>
                      </HStack>
                      <Button
                        size="xs"
                        leftIcon={<FiPlus />}
                        onClick={addOrderBy}
                        variant="ghost"
                        colorScheme="green"
                      >
                        Add
                      </Button>
                    </Flex>
                    {orderBy.length === 0 ? (
                      <Text fontSize="sm" color="gray.500" fontStyle="italic">
                        No sorting
                      </Text>
                    ) : (
                      <VStack spacing={2} align="stretch">
                        {orderBy.map((ob, i) => (
                          <HStack key={i}>
                            <Select
                              value={ob.column}
                              onChange={(e) => updateOrderBy(i, { column: e.target.value })}
                              size="sm"
                              flex={2}
                              borderRadius="lg"
                            >
                              <option value="">Column...</option>
                              {availableColumns.map(c => (
                                <option key={c.name} value={c.name}>{c.name}</option>
                              ))}
                            </Select>
                            <Select
                              value={ob.direction}
                              onChange={(e) => updateOrderBy(i, { direction: e.target.value as 'ASC' | 'DESC' })}
                              size="sm"
                              w="100px"
                              borderRadius="lg"
                            >
                              <option value="ASC">ASC</option>
                              <option value="DESC">DESC</option>
                            </Select>
                            <IconButton
                              aria-label="Remove"
                              icon={<FiTrash2 />}
                              size="sm"
                              variant="ghost"
                              colorScheme="red"
                              onClick={() => removeOrderBy(i)}
                            />
                          </HStack>
                        ))}
                      </VStack>
                    )}
                  </Box>

                  <Divider />

                  {/* Options */}
                  <HStack spacing={4}>
                    <Checkbox
                      isChecked={distinct}
                      onChange={(e) => setDistinct(e.target.checked)}
                      size="sm"
                    >
                      DISTINCT
                    </Checkbox>
                    <HStack>
                      <Text fontSize="sm">LIMIT:</Text>
                      <Input
                        type="number"
                        value={limit}
                        onChange={(e) => setLimit(parseInt(e.target.value) || 100)}
                        size="sm"
                        w="80px"
                        borderRadius="lg"
                        min={1}
                        max={10000}
                      />
                    </HStack>
                  </HStack>

                  {/* Generated SQL Preview */}
                  {generatedSQL && (
                    <Box>
                      <HStack justify="space-between" mb={2}>
                        <Text fontSize="sm" fontWeight="600" color="gray.600">
                          Generated SQL
                        </Text>
                        <IconButton
                          aria-label="Copy"
                          icon={<FiCopy />}
                          size="xs"
                          variant="ghost"
                          onClick={copySQL}
                        />
                      </HStack>
                      <Box
                        bg="gray.900"
                        color="gray.100"
                        p={3}
                        borderRadius="lg"
                        fontSize="xs"
                        fontFamily="mono"
                        whiteSpace="pre-wrap"
                        maxH="150px"
                        overflow="auto"
                      >
                        {generatedSQL}
                      </Box>
                    </Box>
                  )}
                </VStack>
              </MotionBox>
            )}

            {/* SQL Editor Mode */}
            {activeMode === 'sql' && (
              <MotionBox
                key="sql"
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                transition={springs.snappy}
                h="100%"
              >
                <VStack spacing={4} align="stretch" h="100%">
                  <HStack justify="space-between">
                    <Text fontWeight="600" fontSize="sm" color="gray.600">
                      Write your SQL query
                    </Text>
                    <HStack>
                      <IconButton
                        aria-label="Copy"
                        icon={<FiCopy />}
                        size="sm"
                        variant="ghost"
                        onClick={copySQL}
                      />
                    </HStack>
                  </HStack>
                  <Box flex={1} minH="300px">
                    <SQLEditor
                      value={sqlQuery}
                      onChange={setSqlQuery}
                      height="100%"
                      serviceName={serviceName}
                      schemas={schemas}
                      placeholder="SELECT * FROM schema.table WHERE ..."
                    />
                  </Box>
                </VStack>
              </MotionBox>
            )}
          </AnimatePresence>
          </Box>

          {/* Execute Button - Fixed at bottom */}
          <Flex
            p={4}
            gap={2}
            borderTop="1px solid"
            borderColor={borderColor}
            bg={bgColor}
            flexShrink={0}
          >
            <Button
              colorScheme="blue"
              leftIcon={executing ? <Spinner size="sm" /> : <FiPlay />}
              onClick={handleExecute}
              isDisabled={executing}
              flex={1}
              borderRadius="xl"
              size="lg"
            >
              {executing ? 'Executing...' : 'Execute Query'}
            </Button>
            {onPublishToGeoServer && result && (
              <Tooltip label="Publish as SQL View Layer to GeoServer">
                <Button
                  colorScheme="green"
                  leftIcon={<FiUpload />}
                  onClick={() => {
                    const sql = activeMode === 'visual' ? generatedSQL : sqlQuery
                    onPublishToGeoServer(sql, 'sql_view_' + Date.now())
                  }}
                  borderRadius="xl"
                  size="lg"
                >
                  Publish
                </Button>
              </Tooltip>
            )}
          </Flex>
        </Flex>

        {/* Resizable Splitter */}
        <Box
          w="4px"
          cursor="col-resize"
          bg={isDragging ? 'blue.400' : borderColor}
          _hover={{ bg: 'blue.400' }}
          transition="background 0.2s"
          onMouseDown={handleSplitterMouseDown}
          flexShrink={0}
        />

        {/* Right Panel - Results */}
        <Box
          w={`calc(${100 - splitPosition}% - 4px)`}
          display="flex"
          flexDirection="column"
          overflow="hidden"
        >
          {/* Results Header */}
          <Flex
            px={4}
            py={3}
            bg={headerBg}
            borderBottom="1px solid"
            borderColor={borderColor}
            align="center"
            justify="space-between"
          >
            <HStack>
              <Icon as={FiTable} color="blue.500" />
              <Text fontWeight="600">Results</Text>
              {result && (
                <Badge colorScheme="blue" borderRadius="full">
                  {result.rows.length}
                  {result.total_count ? ` / ${result.total_count}` : ''} rows
                </Badge>
              )}
              {result && (
                <Badge colorScheme="gray" borderRadius="full">
                  {result.duration_ms.toFixed(2)}ms
                </Badge>
              )}
            </HStack>
            {result && (
              <Menu>
                <MenuButton
                  as={Button}
                  size="sm"
                  variant="ghost"
                  rightIcon={<FiChevronDown />}
                >
                  <Icon as={FiDownload} />
                </MenuButton>
                <MenuList>
                  <MenuItem icon={<FiDownload />} onClick={() => exportResults('csv')}>
                    Export as CSV
                  </MenuItem>
                  <MenuItem icon={<FiDownload />} onClick={() => exportResults('json')}>
                    Export as JSON
                  </MenuItem>
                </MenuList>
              </Menu>
            )}
          </Flex>

          {/* Error Display */}
          {error && (
            <Alert status="error" borderRadius="none">
              <AlertIcon />
              <AlertDescription fontSize="sm">{error}</AlertDescription>
            </Alert>
          )}

          {/* Results Table with Infinite Scroll */}
          <Box
            ref={tableContainerRef}
            flex={1}
            overflow="auto"
            position="relative"
          >
            {executing && !result ? (
              <Flex h="100%" align="center" justify="center">
                <VStack spacing={4}>
                  <Spinner size="xl" color="blue.500" thickness="4px" />
                  <Text color="gray.500">Executing query...</Text>
                </VStack>
              </Flex>
            ) : result ? (
              <>
                <Table size="sm" variant="simple">
                  <Thead position="sticky" top={0} bg={headerBg} zIndex={1}>
                    <Tr>
                      <Th
                        w="50px"
                        textAlign="center"
                        borderBottomWidth="2px"
                        borderBottomColor={borderColor}
                      >
                        #
                      </Th>
                      {result.columns.map((col, i) => (
                        <Th
                          key={i}
                          borderBottomWidth="2px"
                          borderBottomColor={borderColor}
                          whiteSpace="nowrap"
                        >
                          <VStack align="start" spacing={0}>
                            <Text>{col.name}</Text>
                            <Text fontSize="2xs" color="gray.400" fontWeight="normal">
                              {col.type}
                            </Text>
                          </VStack>
                        </Th>
                      ))}
                    </Tr>
                  </Thead>
                  <Tbody>
                    <AnimatePresence>
                      {result.rows.map((row, rowIndex) => (
                        <MotionBox
                          as={Tr}
                          key={rowIndex}
                          initial={{ opacity: 0 }}
                          animate={{ opacity: 1 }}
                          exit={{ opacity: 0 }}
                          transition={{ delay: rowIndex * 0.01 }}
                          _hover={{ bg: hoverBg }}
                        >
                          <Td
                            textAlign="center"
                            color="gray.400"
                            fontSize="xs"
                            fontFamily="mono"
                          >
                            {rowIndex + 1}
                          </Td>
                          {result.columns.map((col, colIndex) => {
                            const value = row[col.name]
                            const isNull = value === null
                            const displayValue = isNull
                              ? 'NULL'
                              : typeof value === 'object'
                              ? JSON.stringify(value)
                              : String(value)

                            return (
                              <Td
                                key={colIndex}
                                maxW="300px"
                                overflow="hidden"
                                textOverflow="ellipsis"
                                whiteSpace="nowrap"
                                color={isNull ? 'gray.400' : undefined}
                                fontStyle={isNull ? 'italic' : undefined}
                                fontSize="sm"
                                title={displayValue}
                              >
                                {displayValue.length > 100
                                  ? displayValue.substring(0, 100) + '...'
                                  : displayValue}
                              </Td>
                            )
                          })}
                        </MotionBox>
                      ))}
                    </AnimatePresence>
                  </Tbody>
                </Table>

                {/* Loading More Indicator */}
                {loadingMore && (
                  <Flex py={4} justify="center">
                    <HStack>
                      <Spinner size="sm" />
                      <Text fontSize="sm" color="gray.500">Loading more...</Text>
                    </HStack>
                  </Flex>
                )}

                {/* End of Results */}
                {!result.has_more && result.rows.length > 0 && (
                  <Flex py={4} justify="center">
                    <Text fontSize="sm" color="gray.400">
                      End of results ({result.rows.length} rows)
                    </Text>
                  </Flex>
                )}
              </>
            ) : (
              <Flex h="100%" align="center" justify="center">
                <VStack spacing={4} color="gray.400">
                  <Icon as={FiSearch} boxSize={12} />
                  <Text>Execute a query to see results</Text>
                </VStack>
              </Flex>
            )}
          </Box>
        </Box>
      </Flex>
    </MotionBox>
  )
}

export default QueryPanel
