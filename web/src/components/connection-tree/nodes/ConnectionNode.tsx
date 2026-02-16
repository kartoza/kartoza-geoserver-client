import { Box } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useTreeStore, generateNodeId } from '../../../stores/treeStore'
import { useUIStore } from '../../../stores/uiStore'
import type { TreeNode } from '../../../types'
import * as api from '../../../api/client'
import { TreeNodeRow } from '../TreeNodeRow'
import { WorkspaceNode } from './WorkspaceNode'
import type { ConnectionNodeProps } from '../types'

export function ConnectionNode({ connectionId, name, url }: ConnectionNodeProps) {
  const nodeId = generateNodeId('connection', connectionId)
  const isExpanded = useTreeStore((state) => state.isExpanded(nodeId))
  const toggleNode = useTreeStore((state) => state.toggleNode)
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)
  const openDialog = useUIStore((state) => state.openDialog)

  const { data: workspaces, isLoading } = useQuery({
    queryKey: ['workspaces', connectionId],
    queryFn: () => api.getWorkspaces(connectionId),
    staleTime: 30000,
  })

  const node: TreeNode = {
    id: nodeId,
    name,
    type: 'connection',
    connectionId,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
    toggleNode(nodeId)
  }

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('connection', { mode: 'edit', data: { connectionId } })
  }

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('confirm', {
      mode: 'delete',
      title: 'Delete Connection',
      message: `Are you sure you want to delete connection "${name}"?`,
      data: { connectionId },
    })
  }

  const handleOpenAdmin = (e: React.MouseEvent) => {
    e.stopPropagation()
    // GeoServer admin URL is typically the base URL + /web
    const adminUrl = url.replace(/\/rest\/?$/, '/web')
    window.open(adminUrl, '_blank', 'noopener,noreferrer')
  }

  return (
    <Box>
      <TreeNodeRow
        node={node}
        isExpanded={isExpanded}
        isSelected={isSelected}
        isLoading={isLoading}
        onClick={handleClick}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onOpenAdmin={handleOpenAdmin}
        level={2}
        count={workspaces?.length}
      />
      {isExpanded && workspaces && (
        <Box pl={4}>
          {workspaces.map((ws) => (
            <WorkspaceNode
              key={ws.name}
              connectionId={connectionId}
              workspace={ws.name}
            />
          ))}
        </Box>
      )}
    </Box>
  )
}
