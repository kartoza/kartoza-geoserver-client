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
import { FiEye } from 'react-icons/fi'
import * as api from '../../api'
import { useUIStore } from '../../stores/uiStore'

interface StoreCardProps {
  name: string
  type: string
  enabled: boolean
  icon: React.ElementType
  connectionId: string
  workspace: string
  storeType: 'datastore' | 'coveragestore'
}

export default function StoreCard({
  name,
  type,
  enabled,
  icon,
  connectionId,
  workspace,
  storeType,
}: StoreCardProps) {
  const setPreview = useUIStore((state) => state.setPreview)

  const handlePreview = async () => {
    try {
      const { url } = await api.startPreview({
        connId: connectionId,
        workspace,
        layerName: name,
        storeName: name,
        storeType,
        layerType: storeType === 'coveragestore' ? 'raster' : 'vector',
      })
      setPreview({
        url,
        layerName: name,
        workspace,
        connectionId,
        storeName: name,
        storeType,
        layerType: storeType === 'coveragestore' ? 'raster' : 'vector',
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
              <Icon as={icon} color="kartoza.500" boxSize={5} />
              <Text fontWeight="medium" noOfLines={1}>{name}</Text>
            </HStack>
            <Badge colorScheme={enabled ? 'green' : 'gray'} size="sm">
              {enabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </HStack>
          <Text fontSize="xs" color="gray.500">{type}</Text>
          <Button
            size="sm"
            variant="outline"
            leftIcon={<FiEye />}
            onClick={handlePreview}
          >
            Preview
          </Button>
        </VStack>
      </CardBody>
    </Card>
  )
}
