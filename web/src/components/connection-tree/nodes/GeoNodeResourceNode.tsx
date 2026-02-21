import { Box } from '@chakra-ui/react'
import { useTreeStore } from '../../../stores/treeStore'
import type { TreeNode, NodeType, GeoNodeResource, GeoNodeDataset } from '../../../types'
import { TreeNodeRow } from '../TreeNodeRow'

interface GeoNodeResourceNodeProps {
  connectionId: string
  connectionUrl: string
  resource: GeoNodeResource
  nodeType: NodeType
  category: string
}

export function GeoNodeResourceNode({
  connectionId,
  connectionUrl,
  resource,
  nodeType,
  category,
}: GeoNodeResourceNodeProps) {
  const nodeId = `geonode-${category}-resource-${resource.pk}`
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)

  // Check if this is a dataset (has alternate field)
  const isDataset = category === 'datasets'
  const datasetResource = resource as GeoNodeDataset

  const node: TreeNode = {
    id: nodeId,
    name: resource.title,
    type: nodeType,
    geonodeConnectionId: connectionId,
    geonodeResourcePk: resource.pk,
    geonodeResourceUuid: resource.uuid,
    geonodeResourceType: resource.resource_type,
    geonodeThumbnailUrl: resource.thumbnail_url,
    geonodeDetailUrl: resource.detail_url ? `${connectionUrl}${resource.detail_url}` : undefined,
    geonodeAlternate: isDataset ? datasetResource.alternate : undefined,
    geonodeUrl: connectionUrl,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
  }

  return (
    <Box>
      <TreeNodeRow
        node={node}
        isExpanded={false}
        isSelected={isSelected}
        isLoading={false}
        onClick={handleClick}
        level={4}
      />
    </Box>
  )
}
