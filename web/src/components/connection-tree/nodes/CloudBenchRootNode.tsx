import { useEffect } from 'react'
import { Box } from '@chakra-ui/react'
import { useTreeStore } from '../../../stores/treeStore'
import { useProvidersStore } from '../../../stores/providersStore'
import { GeoServerRootNode } from './GeoServerRootNode'
import { PostgreSQLRootNode } from './PostgreSQLRootNode'
import { S3StorageRootNode } from './S3StorageRootNode'
import { QGISProjectsRootNode } from './QGISProjectsRootNode'
import { GeoNodeRootNode } from './GeoNodeRootNode'
import { IcebergRootNode } from './IcebergRootNode'
import { QFieldCloudRootNode } from './QFieldCloudRootNode'
import { MerginMapsRootNode } from './MerginMapsRootNode'

interface CloudBenchRootNodeProps {
  connections: { id: string; name: string; url: string }[]
}

export function CloudBenchRootNode({ connections }: CloudBenchRootNodeProps) {
  const nodeId = 'cloudbench-root'
  const isExpanded = useTreeStore((state) => state.isExpanded(nodeId))
  const toggleNode = useTreeStore((state) => state.toggleNode)

  // Provider enablement
  const fetchProviders = useProvidersStore((state) => state.fetchProviders)
  const isProviderEnabled = useProvidersStore((state) => state.isProviderEnabled)
  const providers = useProvidersStore((state) => state.providers)

  // Fetch providers on mount
  useEffect(() => {
    fetchProviders()
  }, [fetchProviders])

  // Auto-expand root on mount
  useEffect(() => {
    if (!isExpanded) {
      toggleNode(nodeId)
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // If providers haven't loaded yet, show all (graceful degradation)
  const showAll = providers.length === 0

  return (
    <Box>
      <>
        {/* GeoServer Section */}
        {(showAll || isProviderEnabled('geoserver')) && (
          <GeoServerRootNode connections={connections} />
        )}
        {/* PostgreSQL Section */}
        {(showAll || isProviderEnabled('postgres')) && (
          <PostgreSQLRootNode />
        )}
        {/* S3 Storage Section */}
        {(showAll || isProviderEnabled('s3')) && (
          <S3StorageRootNode />
        )}
        {/* Apache Iceberg Section */}
        {(showAll || isProviderEnabled('iceberg')) && (
          <IcebergRootNode />
        )}
        {/* QGIS Projects Section */}
        {(showAll || isProviderEnabled('qgis')) && (
          <QGISProjectsRootNode />
        )}
        {/* GeoNode Section */}
        {(showAll || isProviderEnabled('geonode')) && (
          <GeoNodeRootNode />
        )}
        {/* QFieldCloud Section */}
        {(showAll || isProviderEnabled('qfieldcloud')) && (
          <QFieldCloudRootNode />
        )}
        {/* Mergin Maps Section */}
        {(showAll || isProviderEnabled('mergin')) && (
          <MerginMapsRootNode />
        )}
      </>
    </Box>
  )
}
