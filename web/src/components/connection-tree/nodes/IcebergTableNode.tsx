import { Box, HStack, Text, Badge } from '@chakra-ui/react'
import { useTreeStore, generateNodeId } from '../../../stores/treeStore'
import { useUIStore } from '../../../stores/uiStore'
import type { TreeNode, IcebergTable } from '../../../types'
import { TreeNodeRow } from '../TreeNodeRow'

interface IcebergTableNodeProps {
  connectionId: string
  connectionName: string
  namespace: string
  table: IcebergTable
}

export function IcebergTableNode({ connectionId, connectionName, namespace, table }: IcebergTableNodeProps) {
  const nodeId = generateNodeId('icebergtable', `${connectionId}:${namespace}:${table.name}`)
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)
  const openDialog = useUIStore((state) => state.openDialog)
  const setIcebergPreview = useUIStore((state) => state.setIcebergPreview)

  const node: TreeNode = {
    id: nodeId,
    name: table.name,
    type: 'icebergtable',
    icebergConnectionId: connectionId,
    icebergNamespace: namespace,
    icebergTableName: table.name,
    icebergHasGeometry: table.hasGeometry,
    icebergGeometryColumns: table.geometryColumns,
    icebergRowCount: table.rowCount,
    icebergSnapshotCount: table.snapshotCount,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
  }

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('confirm', {
      mode: 'delete',
      title: 'Delete Iceberg Table',
      message: `Are you sure you want to delete table "${table.name}"?`,
      data: {
        icebergConnectionId: connectionId,
        icebergNamespace: namespace,
        icebergTableName: table.name,
      },
    })
  }

  const handlePreview = (e: React.MouseEvent) => {
    e.stopPropagation()
    setIcebergPreview({
      connectionId,
      connectionName,
      namespace,
      tableName: table.name,
    })
  }

  const handleShowData = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('icebergtabledata', {
      mode: 'view',
      data: {
        connectionId,
        namespace,
        tableName: table.name,
      },
    })
  }

  const handleQuery = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('icebergquery', {
      mode: 'create',
      data: {
        connectionId,
        namespace,
        tableName: table.name,
      },
    })
  }

  // Show geometry badge if table has geometry columns
  const badges = []
  if (table.hasGeometry) {
    badges.push(
      <Badge key="geo" colorScheme="green" fontSize="2xs" px={1}>
        Geo
      </Badge>
    )
  }
  if (table.snapshotCount && table.snapshotCount > 1) {
    badges.push(
      <Badge key="snapshots" colorScheme="blue" fontSize="2xs" px={1}>
        {table.snapshotCount} snapshots
      </Badge>
    )
  }

  return (
    <Box>
      <TreeNodeRow
        node={node}
        isExpanded={false}
        isSelected={isSelected}
        isLoading={false}
        onClick={handleClick}
        onDelete={handleDelete}
        onPreview={table.hasGeometry ? handlePreview : undefined}
        onShowData={handleShowData}
        onQuery={handleQuery}
        level={4}
        isLeaf={true}
      />
      {badges.length > 0 && (
        <HStack ml="52px" mt={-1} mb={1} gap={1}>
          {badges}
          {table.rowCount !== undefined && (
            <Text fontSize="2xs" color="gray.500">
              {table.rowCount.toLocaleString()} rows
            </Text>
          )}
        </HStack>
      )}
    </Box>
  )
}
