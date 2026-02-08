import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  VStack,
  HStack,
  Box,
  Text,
  Icon,
  Badge,
  SimpleGrid,
  Divider,
  Spinner,
  Code,
  Accordion,
  AccordionItem,
  AccordionButton,
  AccordionPanel,
  AccordionIcon,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { FiDatabase, FiImage, FiFolder, FiInfo } from 'react-icons/fi'
import { useUIStore } from '../../stores/uiStore'
import { useTreeStore } from '../../stores/treeStore'
import * as api from '../../api/client'
import type { CoverageStore } from '../../types'

export default function StoreDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const dialogData = useUIStore((state) => state.dialogData)
  const closeDialog = useUIStore((state) => state.closeDialog)
  const selectedNode = useTreeStore((state) => state.selectedNode)

  const isOpen = activeDialog === 'datastore' || activeDialog === 'coveragestore'
  const isDataStore = activeDialog === 'datastore'
  const isEditMode = dialogData?.mode === 'edit'

  const connectionId = (dialogData?.data?.connectionId as string) || selectedNode?.connectionId || ''
  const workspace = (dialogData?.data?.workspace as string) || selectedNode?.workspace || ''
  const storeName = (dialogData?.data?.storeName as string) || selectedNode?.name || ''

  // Fetch store details
  const { data: store, isLoading } = useQuery({
    queryKey: [isDataStore ? 'datastores' : 'coveragestores', connectionId, workspace, storeName],
    queryFn: () =>
      isDataStore
        ? api.getDataStore(connectionId, workspace, storeName)
        : api.getCoverageStore(connectionId, workspace, storeName),
    enabled: isOpen && isEditMode && !!connectionId && !!workspace && !!storeName,
  })

  if (!isOpen) return null

  return (
    <Modal isOpen={isOpen} onClose={closeDialog} size="xl" isCentered>
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent borderRadius="xl" overflow="hidden" maxH="85vh">
        {/* Gradient Header */}
        <Box
          bg={isDataStore
            ? "linear-gradient(135deg, #1B6B9B 0%, #3B9DD9 100%)"
            : "linear-gradient(135deg, #D4922A 0%, #E8A331 100%)"
          }
          px={6}
          py={4}
        >
          <HStack spacing={3}>
            <Box bg="whiteAlpha.200" p={2} borderRadius="lg">
              <Icon as={isDataStore ? FiDatabase : FiImage} boxSize={5} color="white" />
            </Box>
            <Box>
              <Text color="white" fontWeight="600" fontSize="lg">
                {isDataStore ? 'Data Store Details' : 'Coverage Store Details'}
              </Text>
              <Text color="whiteAlpha.800" fontSize="sm">
                {storeName}
              </Text>
            </Box>
          </HStack>
        </Box>
        <ModalCloseButton color="white" />

        <ModalBody py={6} overflowY="auto">
          {isLoading ? (
            <VStack py={10}>
              <Spinner size="xl" color="kartoza.500" />
              <Text color="gray.500">Loading store details...</Text>
            </VStack>
          ) : store ? (
            <VStack spacing={6} align="stretch">
              {/* Basic Info */}
              <Box p={4} bg="gray.50" borderRadius="lg">
                <HStack mb={3}>
                  <Icon as={FiInfo} color="kartoza.500" />
                  <Text fontWeight="bold" color="gray.600">
                    Store Information
                  </Text>
                </HStack>
                <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                  <Box>
                    <Text fontSize="xs" color="gray.500">Name</Text>
                    <Text fontWeight="medium">{store.name}</Text>
                  </Box>
                  <Box>
                    <Text fontSize="xs" color="gray.500">Type</Text>
                    <HStack>
                      <Text fontWeight="medium">{store.type || 'Unknown'}</Text>
                      <Badge colorScheme={isDataStore ? 'blue' : 'orange'}>
                        {isDataStore ? 'Vector' : 'Raster'}
                      </Badge>
                    </HStack>
                  </Box>
                  <Box>
                    <Text fontSize="xs" color="gray.500">Workspace</Text>
                    <HStack>
                      <Icon as={FiFolder} color="gray.400" boxSize={4} />
                      <Text fontWeight="medium">{workspace}</Text>
                    </HStack>
                  </Box>
                  <Box>
                    <Text fontSize="xs" color="gray.500">Status</Text>
                    <Badge colorScheme={store.enabled ? 'green' : 'gray'} fontSize="sm">
                      {store.enabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </Box>
                </SimpleGrid>
              </Box>

              {/* Description (only for coverage stores) */}
              {!isDataStore && (store as CoverageStore).description && (
                <>
                  <Divider />
                  <Box>
                    <Text fontWeight="bold" color="gray.600" mb={2}>
                      Description
                    </Text>
                    <Text color="gray.600">{(store as CoverageStore).description}</Text>
                  </Box>
                </>
              )}

              {/* Store Type Info */}
              <Accordion allowToggle>
                <AccordionItem border="none">
                  <AccordionButton
                    bg="gray.50"
                    borderRadius="lg"
                    _hover={{ bg: 'gray.100' }}
                  >
                    <Box flex="1" textAlign="left">
                      <Text fontWeight="500" color="gray.600">
                        Technical Details
                      </Text>
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                  <AccordionPanel pb={4}>
                    <VStack align="stretch" spacing={2}>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color="gray.500">Store Type:</Text>
                        <Code fontSize="sm">{store.type || 'Unknown'}</Code>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color="gray.500">Workspace:</Text>
                        <Code fontSize="sm">{workspace}</Code>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color="gray.500">Enabled:</Text>
                        <Code fontSize="sm">{store.enabled ? 'true' : 'false'}</Code>
                      </HStack>
                    </VStack>
                  </AccordionPanel>
                </AccordionItem>
              </Accordion>

              {/* Info Note */}
              <Box
                p={4}
                bg="blue.50"
                borderRadius="lg"
                borderLeft="4px solid"
                borderLeftColor="blue.400"
              >
                <Text fontSize="sm" color="blue.700">
                  <strong>Note:</strong> Store configuration (connection parameters, file paths, etc.) 
                  can be modified through the GeoServer admin interface. This view shows the current 
                  store metadata.
                </Text>
              </Box>
            </VStack>
          ) : (
            <VStack py={10}>
              <Text color="gray.500">Store not found</Text>
            </VStack>
          )}
        </ModalBody>

        <ModalFooter
          gap={3}
          borderTop="1px solid"
          borderTopColor="gray.100"
          bg="gray.50"
        >
          <Button
            colorScheme="kartoza"
            onClick={closeDialog}
            borderRadius="lg"
            px={6}
          >
            Close
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
