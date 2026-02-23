import {
  Card,
  CardBody,
  VStack,
  HStack,
  Icon,
  Text,
  Badge,
  Button,
} from '@chakra-ui/react'
import { FiLayers, FiMap } from 'react-icons/fi'
import { useQuery } from '@tanstack/react-query'
import * as api from '../../api'
import { useUIStore } from '../../stores/uiStore'

interface LayerCardProps {
  name: string
  connectionId: string
  workspace: string
  enabled: boolean
}

export default function LayerCard({
  name,
  connectionId,
  workspace,
  enabled,
}: LayerCardProps) {
  const setPreview = useUIStore((state) => state.setPreview)

  const { data: layer } = useQuery({
    queryKey: ['layer', connectionId, workspace, name],
    queryFn: () => api.getLayer(connectionId, workspace, name),
  })

  const handlePreview = async () => {
    try {
      const { url } = await api.startPreview({
        connId: connectionId,
        workspace,
        layerName: name,
        storeName: layer?.store,
        storeType: layer?.storeType,
        layerType: layer?.storeType === 'coveragestore' ? 'raster' : 'vector',
      })
      setPreview({
        url,
        layerName: name,
        workspace,
        connectionId,
        storeName: layer?.store,
        storeType: layer?.storeType,
        layerType: layer?.storeType === 'coveragestore' ? 'raster' : 'vector',
      })
    } catch (err) {
      useUIStore.getState().setError((err as Error).message)
    }
  }

  return (
    <Card variant="outline" size="sm" _hover={{ shadow: 'md' }} transition="all 0.2s">
      <CardBody py={4} px={4}>
        <VStack align="stretch" spacing={3}>
          <HStack justify="space-between">
            <HStack>
              <Icon as={FiLayers} color="kartoza.500" boxSize={5} />
              <Text fontWeight="medium" noOfLines={1}>{name}</Text>
            </HStack>
            <Badge colorScheme={enabled ? 'green' : 'gray'} size="sm">
              {enabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </HStack>
          {layer?.defaultStyle && (
            <Text fontSize="xs" color="gray.500">Style: {layer.defaultStyle}</Text>
          )}
          <Button
            size="sm"
            colorScheme="kartoza"
            leftIcon={<FiMap />}
            onClick={handlePreview}
          >
            Preview on Map
          </Button>
        </VStack>
      </CardBody>
    </Card>
  )
}
