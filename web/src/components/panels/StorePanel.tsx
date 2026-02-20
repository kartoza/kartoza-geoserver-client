import {
  VStack,
  Card,
  CardBody,
  Flex,
  HStack,
  Box,
  Icon,
  Heading,
  Text,
  Spacer,
  Button,
  Badge,
  SimpleGrid,
  Divider,
  useColorModeValue,
} from '@chakra-ui/react'
import { FiDatabase, FiImage, FiMap, FiEdit3 } from 'react-icons/fi'
import { useQuery } from '@tanstack/react-query'
import * as api from '../../api/client'
import { useUIStore } from '../../stores/uiStore'

interface StorePanelProps {
  connectionId: string
  workspace: string
  storeName: string
  storeType: 'datastore' | 'coveragestore'
}

export default function StorePanel({
  connectionId,
  workspace,
  storeName,
  storeType,
}: StorePanelProps) {
  const cardBg = useColorModeValue('white', 'gray.800')
  const setPreview = useUIStore((state) => state.setPreview)
  const openDialog = useUIStore((state) => state.openDialog)

  const isDataStore = storeType === 'datastore'

  const { data: store } = useQuery({
    queryKey: [storeType + 's', connectionId, workspace, storeName],
    queryFn: () =>
      isDataStore
        ? api.getDataStore(connectionId, workspace, storeName)
        : api.getCoverageStore(connectionId, workspace, storeName),
  })

  const handlePreview = async () => {
    try {
      const { url } = await api.startPreview({
        connId: connectionId,
        workspace,
        layerName: storeName,
        storeName,
        storeType,
        layerType: isDataStore ? 'vector' : 'raster',
      })
      setPreview({
        url,
        layerName: storeName,
        workspace,
        connectionId,
        storeName,
        storeType,
        layerType: isDataStore ? 'vector' : 'raster',
      })
    } catch (err) {
      useUIStore.getState().setError((err as Error).message)
    }
  }

  return (
    <VStack spacing={6} align="stretch">
      <Card
        bg="linear-gradient(135deg, #0a3a50 0%, #175a77 50%, #2d7d9b 100%)"
        color="white"
      >
        <CardBody py={8} px={6}>
          <Flex align="center" wrap="wrap" gap={4}>
            <HStack spacing={4}>
              <Box bg="whiteAlpha.200" p={3} borderRadius="lg">
                <Icon as={isDataStore ? FiDatabase : FiImage} boxSize={8} />
              </Box>
              <VStack align="start" spacing={1}>
                <Heading size="lg" color="white">{storeName}</Heading>
                <HStack>
                  <Badge colorScheme={isDataStore ? 'blue' : 'orange'}>
                    {isDataStore ? 'Data Store' : 'Coverage Store'}
                  </Badge>
                  {store?.enabled && <Badge colorScheme="green">Enabled</Badge>}
                </HStack>
              </VStack>
            </HStack>
            <Spacer />
            <HStack wrap="wrap" gap={2}>
              <Button
                variant="solid"
                bg="whiteAlpha.200"
                color="white"
                _hover={{ bg: 'whiteAlpha.300' }}
                leftIcon={<FiMap />}
                onClick={handlePreview}
              >
                Preview on Map
              </Button>
              <Button
                variant="outline"
                color="white"
                borderColor="whiteAlpha.400"
                _hover={{ bg: 'whiteAlpha.200' }}
                leftIcon={<FiEdit3 />}
                onClick={() => openDialog(storeType, {
                  mode: 'edit',
                  data: { connectionId, workspace, storeName }
                })}
              >
                Edit Store
              </Button>
            </HStack>
          </Flex>
        </CardBody>
      </Card>

      <Card bg={cardBg}>
        <CardBody>
          <VStack align="start" spacing={3}>
            <Heading size="sm" color="gray.600">Store Details</Heading>
            <Divider />
            <SimpleGrid columns={2} spacing={4} w="100%">
              <Box>
                <Text fontSize="xs" color="gray.500">Workspace</Text>
                <Text fontWeight="medium">{workspace}</Text>
              </Box>
              <Box>
                <Text fontSize="xs" color="gray.500">Type</Text>
                <Text fontWeight="medium">{store?.type || 'Unknown'}</Text>
              </Box>
            </SimpleGrid>
          </VStack>
        </CardBody>
      </Card>
    </VStack>
  )
}
