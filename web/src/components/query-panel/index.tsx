// Query Panel - Main component that orchestrates all sub-components

import { useState, useEffect, useMemo, useCallback } from 'react'
import {
  Box,
  Flex,
  useColorModeValue,
  useToast,
} from '@chakra-ui/react'
import { motion, AnimatePresence } from 'framer-motion'
import { springs } from '../../utils/animations'
import { QueryPanelHeader } from './QueryPanelHeader'
import { QueryModeTabs } from './QueryModeTabs'
import { AIMode } from './AIMode'
import { VisualBuilder } from './VisualBuilder'
import { SQLMode } from './SQLMode'
import { ExecuteButton } from './ExecuteButton'
import { ResultsPanel } from './ResultsPanel'
import { buildSQL, exportToCSV, exportToJSON, modifySQLWithOffset } from './utils'
import {
  useSchemas,
  useAIProvider,
  useQueryExecution,
  useAIQuery,
  useResizableSplitter,
  useInfiniteScroll,
} from './hooks'
import type { QueryPanelProps, Column, Condition, OrderBy, QueryMode } from './types'

const MotionBox = motion(Box)

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

  // Mode state
  const [activeMode, setActiveMode] = useState<QueryMode>('visual')

  // Custom hooks
  const { schemas, selectedSchema, setSelectedSchema } = useSchemas(serviceName, initialSchema)
  const aiProviderAvailable = useAIProvider()
  const { executing, result, error, loadingMore, executeQuery, setError } = useQueryExecution(serviceName, 100)

  // Visual builder state
  const [selectedTable, setSelectedTable] = useState(initialTable)
  const [selectedColumns, setSelectedColumns] = useState<Column[]>([])
  const [conditions, setConditions] = useState<Condition[]>([])
  const [orderBy, setOrderBy] = useState<OrderBy[]>([])
  const [limit, setLimit] = useState(100)
  const [distinct, setDistinct] = useState(false)

  // SQL Editor state
  const [sqlQuery, setSqlQuery] = useState(initialSQL)
  const [generatedSQL, setGeneratedSQL] = useState('')

  // AI state
  const { aiQuestion, setAiQuestion, aiLoading, aiResponse, handleAIQuery } = useAIQuery(
    serviceName,
    selectedSchema,
    limit
  )

  // Resizable splitter
  const { splitPosition, isDragging, containerRef, handleSplitterMouseDown } = useResizableSplitter()

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

  // Generate SQL from visual builder
  useEffect(() => {
    if (activeMode === 'visual') {
      if (selectedTable) {
        const sql = buildSQL(
          selectedSchema,
          selectedTable,
          selectedColumns,
          conditions,
          orderBy,
          limit,
          0,
          distinct
        )
        setGeneratedSQL(sql)
      } else {
        setGeneratedSQL('')
      }
    }
  }, [activeMode, selectedSchema, selectedTable, selectedColumns, conditions, orderBy, limit, distinct])

  // AI query handler
  const handleAIQueryWrapper = useCallback(async () => {
    const sql = await handleAIQuery(setError)
    if (sql) {
      setSqlQuery(sql)
    }
  }, [handleAIQuery, setError])

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

    const currentOffset = result.rows.length
    const modifiedSQL = modifySQLWithOffset(sql, limit, currentOffset)

    executeQuery(modifiedSQL, true)
  }, [result, loadingMore, activeMode, aiResponse, generatedSQL, sqlQuery, limit, executeQuery])

  // Infinite scroll
  const tableContainerRef = useInfiniteScroll(result, loadingMore, loadMore)

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
      content = exportToCSV(result.columns.map(c => c.name), result.rows)
      filename = 'query-results.csv'
      mimeType = 'text/csv'
    } else {
      content = exportToJSON(result.rows)
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
      <QueryPanelHeader
        serviceName={serviceName}
        schemaCount={schemas.length}
        onClose={onClose}
      />

      {/* Mode Tabs */}
      <QueryModeTabs
        activeMode={activeMode}
        onModeChange={setActiveMode}
      />

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
                <AIMode
                  serviceName={serviceName}
                  schemas={schemas}
                  selectedSchema={selectedSchema}
                  aiQuestion={aiQuestion}
                  aiLoading={aiLoading}
                  aiProviderAvailable={aiProviderAvailable}
                  aiResponse={aiResponse}
                  sqlQuery={sqlQuery}
                  onSchemaChange={setSelectedSchema}
                  onQuestionChange={setAiQuestion}
                  onSqlChange={setSqlQuery}
                  onGenerateSQL={handleAIQueryWrapper}
                />
              )}

              {/* Visual Builder Mode */}
              {activeMode === 'visual' && (
                <VisualBuilder
                  schemas={schemas}
                  selectedSchema={selectedSchema}
                  selectedTable={selectedTable}
                  selectedColumns={selectedColumns}
                  conditions={conditions}
                  orderBy={orderBy}
                  limit={limit}
                  distinct={distinct}
                  generatedSQL={generatedSQL}
                  currentSchema={currentSchema}
                  currentTable={currentTable}
                  availableColumns={availableColumns}
                  onSchemaChange={setSelectedSchema}
                  onTableChange={setSelectedTable}
                  onColumnsChange={setSelectedColumns}
                  onConditionsChange={setConditions}
                  onOrderByChange={setOrderBy}
                  onLimitChange={setLimit}
                  onDistinctChange={setDistinct}
                  onCopySQL={copySQL}
                />
              )}

              {/* SQL Editor Mode */}
              {activeMode === 'sql' && (
                <SQLMode
                  serviceName={serviceName}
                  schemas={schemas}
                  sqlQuery={sqlQuery}
                  onSqlChange={setSqlQuery}
                  onCopySQL={copySQL}
                />
              )}
            </AnimatePresence>
          </Box>

          {/* Execute Button - Fixed at bottom */}
          <ExecuteButton
            executing={executing}
            result={result}
            generatedSQL={generatedSQL}
            sqlQuery={sqlQuery}
            activeMode={activeMode}
            borderColor={borderColor}
            bgColor={bgColor}
            onExecute={handleExecute}
            onPublishToGeoServer={onPublishToGeoServer}
          />
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
          <ResultsPanel
            result={result}
            error={error}
            executing={executing}
            loadingMore={loadingMore}
            tableContainerRef={tableContainerRef}
            onExport={exportResults}
          />
        </Box>
      </Flex>
    </MotionBox>
  )
}

export default QueryPanel

// Re-export types for convenience
export type { QueryPanelProps } from './types'
