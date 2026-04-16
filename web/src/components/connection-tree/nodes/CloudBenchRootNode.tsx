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

const PROVIDER_NODE_IDS: Record<string, string> = {
  geoserver: 'geoserver',
  postgres: 'postgresql',
  s3: 's3storage-root',
  iceberg: 'iceberg-root',
  qgis: 'qgisprojects-root',
  geonode: 'geonode-root',
  qfieldcloud: 'qfieldcloud-root',
  mergin: 'merginmaps-root',
}

export function CloudBenchRootNode({ connections }: CloudBenchRootNodeProps) {
  const nodeId = 'cloudbench-root'
  const isExpanded = useTreeStore((state) => state.isExpanded(nodeId))
  const toggleNode = useTreeStore((state) => state.toggleNode)
  const expandNode = useTreeStore((state) => state.expandNode)

  // Provider enablement
  const fetchProviders = useProvidersStore((state) => state.fetchProviders)
  const isProviderEnabled = useProvidersStore((state) => state.isProviderEnabled)
  const providers = useProvidersStore((state) => state.providers)

  // Fetch providers on mount
  useEffect(() => {
    fetchProviders()
  }, [fetchProviders])

  // Auto-expand all provider nodes once providers are loaded
  useEffect(() => {
    providers
      .filter((p) => p.enabled)
      .forEach((p) => {
        const nodeId = PROVIDER_NODE_IDS[p.id]
        if (nodeId) expandNode(nodeId)
      })
  }, [providers]) // eslint-disable-line react-hooks/exhaustive-deps

  // Auto-expand root on mount
  useEffect(() => {
    if (!isExpanded) {
      toggleNode(nodeId)
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <Box>
      <>
        {/* GeoServer Section */}
        {(isProviderEnabled('geoserver')) && (
          <GeoServerRootNode connections={connections}/>
        )}
        {/* PostgreSQL Section */}
        {(isProviderEnabled('postgres')) && (
          <PostgreSQLRootNode/>
        )}
        {/* S3 Storage Section */}
        {(isProviderEnabled('s3')) && (
          <S3StorageRootNode/>
        )}
        {/* Apache Iceberg Section */}
        {(isProviderEnabled('iceberg')) && (
          <IcebergRootNode/>
        )}
        {/* QGIS Projects Section */}
        {(isProviderEnabled('qgis')) && (
          <QGISProjectsRootNode/>
        )}
        {/* GeoNode Section */}
        {(isProviderEnabled('geonode')) && (
          <GeoNodeRootNode/>
        )}
        {/* QFieldCloud Section */}
        {(isProviderEnabled('qfieldcloud')) && (
          <QFieldCloudRootNode/>
        )}
        {/* Mergin Maps Section */}
        {(isProviderEnabled('mergin')) && (
          <MerginMapsRootNode/>
        )}
      </>
    </Box>
  )
}
