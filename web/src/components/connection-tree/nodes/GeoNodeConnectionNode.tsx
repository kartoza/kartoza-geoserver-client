import { Box, Text } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useTreeStore } from '../../../stores/treeStore'
import { useUIStore } from '../../../stores/uiStore'
import type { TreeNode, GeoNodeConnection } from '../../../types'
import * as api from '../../../api'
import { TreeNodeRow } from '../TreeNodeRow'
import { GeoNodeResourceCategoryNode } from './GeoNodeResourceCategoryNode'

interface GeoNodeConnectionNodeProps {
  connection: GeoNodeConnection
}

export function GeoNodeConnectionNode({ connection }: GeoNodeConnectionNodeProps) {
  const nodeId = `geonodeconn-${connection.id}`
  const isExpanded = useTreeStore((state) => state.isExpanded(nodeId))
  const toggleNode = useTreeStore((state) => state.toggleNode)
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)
  const openDialog = useUIStore((state) => state.openDialog)

  // Prefetch resources when expanded
  const { data: datasetsData, isLoading: datasetsLoading } = useQuery({
    queryKey: ['geonodedatasets', connection.id],
    queryFn: () => api.getGeoNodeDatasets(connection.id, 1, 100),
    enabled: isExpanded,
    staleTime: 30000,
  })

  const { data: mapsData, isLoading: mapsLoading } = useQuery({
    queryKey: ['geonodemaps', connection.id],
    queryFn: () => api.getGeoNodeMaps(connection.id, 1, 100),
    enabled: isExpanded,
    staleTime: 30000,
  })

  const { data: documentsData, isLoading: documentsLoading } = useQuery({
    queryKey: ['geonodedocuments', connection.id],
    queryFn: () => api.getGeoNodeDocuments(connection.id, 1, 100),
    enabled: isExpanded,
    staleTime: 30000,
  })

  const { data: geostoriesData, isLoading: geostoriesLoading } = useQuery({
    queryKey: ['geonodegeostories', connection.id],
    queryFn: () => api.getGeoNodeGeoStories(connection.id, 1, 100),
    enabled: isExpanded,
    staleTime: 30000,
  })

  const { data: dashboardsData, isLoading: dashboardsLoading } = useQuery({
    queryKey: ['geonodedashboards', connection.id],
    queryFn: () => api.getGeoNodeDashboards(connection.id, 1, 100),
    enabled: isExpanded,
    staleTime: 30000,
  })

  const isLoading = datasetsLoading || mapsLoading || documentsLoading || geostoriesLoading || dashboardsLoading

  const node: TreeNode = {
    id: nodeId,
    name: connection.name,
    type: 'geonodeconnection',
    geonodeConnectionId: connection.id,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
    toggleNode(nodeId)
  }

  const handleUpload = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('geonodeupload', {
      mode: 'create',
      data: {
        connectionId: connection.id,
        connectionName: connection.name,
      },
    })
  }

  return (
    <Box>
      <TreeNodeRow
        node={node}
        isExpanded={isExpanded}
        isSelected={isSelected}
        isLoading={isLoading}
        onClick={handleClick}
        onUpload={handleUpload}
        level={2}
      />
      {isExpanded && (
        <Box pl={4}>
          {/* Datasets */}
          <GeoNodeResourceCategoryNode
            connectionId={connection.id}
            connectionUrl={connection.url}
            category="datasets"
            name="Datasets"
            resources={datasetsData?.datasets || []}
            total={datasetsData?.total || 0}
            isLoading={datasetsLoading}
          />

          {/* Maps */}
          <GeoNodeResourceCategoryNode
            connectionId={connection.id}
            connectionUrl={connection.url}
            category="maps"
            name="Maps"
            resources={mapsData?.maps || []}
            total={mapsData?.total || 0}
            isLoading={mapsLoading}
          />

          {/* Documents */}
          <GeoNodeResourceCategoryNode
            connectionId={connection.id}
            connectionUrl={connection.url}
            category="documents"
            name="Documents"
            resources={documentsData?.documents || []}
            total={documentsData?.total || 0}
            isLoading={documentsLoading}
          />

          {/* GeoStories */}
          <GeoNodeResourceCategoryNode
            connectionId={connection.id}
            connectionUrl={connection.url}
            category="geostories"
            name="GeoStories"
            resources={geostoriesData?.geostories || []}
            total={geostoriesData?.total || 0}
            isLoading={geostoriesLoading}
          />

          {/* Dashboards */}
          <GeoNodeResourceCategoryNode
            connectionId={connection.id}
            connectionUrl={connection.url}
            category="dashboards"
            name="Dashboards"
            resources={dashboardsData?.dashboards || []}
            total={dashboardsData?.total || 0}
            isLoading={dashboardsLoading}
          />

          {!isLoading &&
            (!datasetsData?.datasets?.length &&
             !mapsData?.maps?.length &&
             !documentsData?.documents?.length &&
             !geostoriesData?.geostories?.length &&
             !dashboardsData?.dashboards?.length) && (
            <Box px={2} py={2}>
              <Text color="gray.500" fontSize="sm">
                No resources found
              </Text>
            </Box>
          )}
        </Box>
      )}
    </Box>
  )
}
