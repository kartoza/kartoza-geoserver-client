import { Box, Button, Text, useToast } from '@chakra-ui/react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { FiDatabase } from 'react-icons/fi'
import { useTreeStore, generateNodeId } from '../../../stores/treeStore'
import { useUIStore } from '../../../stores/uiStore'
import type { TreeNode } from '../../../types'
import * as api from '../../../api/client'
import { TreeNodeRow } from '../TreeNodeRow'
import { PGSchemaNode } from './PGSchemaNode'
import type { PGServiceNodeProps } from '../types'

export function PGServiceNode({ service }: PGServiceNodeProps) {
  const nodeId = generateNodeId('pgservice', service.name)
  const isExpanded = useTreeStore((state) => state.isExpanded(nodeId))
  const toggleNode = useTreeStore((state) => state.toggleNode)
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)
  const openDialog = useUIStore((state) => state.openDialog)
  const setPGQuery = useUIStore((state) => state.setPGQuery)
  const toast = useToast()
  const queryClient = useQueryClient()

  // Fetch schemas when expanded and service is parsed
  const { data: schemaData, isLoading: loadingSchemas } = useQuery({
    queryKey: ['pgschemas', service.name],
    queryFn: async () => {
      const response = await fetch(`/api/pg/services/${encodeURIComponent(service.name)}/schemas`)
      if (!response.ok) throw new Error('Failed to fetch schemas')
      return response.json() as Promise<{ schemas: { name: string; tables: { name: string; schema: string; columns: { name: string; type: string; nullable: boolean }[] }[] }[] }>
    },
    enabled: isExpanded && service.is_parsed,
    staleTime: 60000,
  })

  const node: TreeNode = {
    id: nodeId,
    name: service.name,
    type: 'pgservice',
    serviceName: service.name,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
    toggleNode(nodeId)
  }

  const handleParse = async (e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await api.parsePGService(service.name)
      toast({
        title: 'Schema Parsed',
        description: `Successfully parsed schema for ${service.name}`,
        status: 'success',
        duration: 3000,
      })
      queryClient.invalidateQueries({ queryKey: ['pgservices'] })
      queryClient.invalidateQueries({ queryKey: ['pgschemas', service.name] })
    } catch (err) {
      toast({
        title: 'Parse Failed',
        description: (err as Error).message,
        status: 'error',
        duration: 5000,
      })
    }
  }

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('confirm', {
      mode: 'delete',
      title: 'Delete PostgreSQL Service',
      message: `Are you sure you want to remove "${service.name}" from pg_service.conf?`,
      data: { pgServiceName: service.name },
    })
  }

  const handleQuery = (e: React.MouseEvent) => {
    e.stopPropagation()
    setPGQuery({
      serviceName: service.name,
    })
  }

  const handleUpload = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('pgupload', {
      mode: 'create',
      data: { serviceName: service.name },
    })
  }

  const handleRefresh = (e: React.MouseEvent) => {
    e.stopPropagation()
    queryClient.invalidateQueries({ queryKey: ['pgschemas', service.name] })
  }

  const subtitle = service.host ? `${service.host}:${service.port || '5432'}/${service.dbname || ''}` : ''

  return (
    <Box>
      <TreeNodeRow
        node={{ ...node, name: service.name }}
        isExpanded={isExpanded}
        isSelected={isSelected}
        isLoading={loadingSchemas}
        onClick={handleClick}
        onDelete={handleDelete}
        onQuery={service.is_parsed ? handleQuery : undefined}
        onUpload={handleUpload}
        onRefresh={service.is_parsed ? handleRefresh : undefined}
        level={2}
        count={schemaData?.schemas?.length}
      />
      {/* Show parse button or schemas */}
      {isExpanded && !service.is_parsed && (
        <Box pl={8} py={2}>
          <Button
            size="sm"
            colorScheme="blue"
            variant="outline"
            leftIcon={<FiDatabase size={14} />}
            onClick={handleParse}
          >
            Parse Schema
          </Button>
          <Text fontSize="xs" color="gray.500" mt={1}>
            {subtitle}
          </Text>
        </Box>
      )}
      {isExpanded && service.is_parsed && schemaData?.schemas && (
        <Box pl={4}>
          {schemaData.schemas.map((schema) => (
            <PGSchemaNode
              key={schema.name}
              serviceName={service.name}
              schema={schema}
            />
          ))}
        </Box>
      )}
    </Box>
  )
}
