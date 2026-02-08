import { useState, useEffect } from 'react'
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  Switch,
  Input,
  Textarea,
  VStack,
  HStack,
  Box,
  Text,
  Icon,
  Badge,
  SimpleGrid,
  Divider,
  Spinner,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Tag,
  TagLabel,
  TagCloseButton,
  Wrap,
  WrapItem,
  Code,
  Accordion,
  AccordionItem,
  AccordionButton,
  AccordionPanel,
  AccordionIcon,
  IconButton,
  useToast,
} from '@chakra-ui/react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { FiLayers, FiEye, FiSearch, FiInfo, FiGlobe, FiLink, FiPlus, FiTrash2 } from 'react-icons/fi'
import { useUIStore } from '../../stores/uiStore'
import { useTreeStore } from '../../stores/treeStore'
import * as api from '../../api/client'
import type { LayerMetadataUpdate, MetadataLink } from '../../types'

export default function LayerDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const dialogData = useUIStore((state) => state.dialogData)
  const closeDialog = useUIStore((state) => state.closeDialog)
  const selectedNode = useTreeStore((state) => state.selectedNode)
  const queryClient = useQueryClient()
  const toast = useToast()

  const [formData, setFormData] = useState<LayerMetadataUpdate>({
    enabled: true,
    advertised: true,
    queryable: true,
  })
  const [keywordInput, setKeywordInput] = useState('')
  const [metadataLinks, setMetadataLinks] = useState<MetadataLink[]>([])
  const [newLink, setNewLink] = useState<MetadataLink>({
    type: 'text/html',
    metadataType: 'TC211',
    content: '',
  })

  const isOpen = activeDialog === 'layer'

  const connectionId = (dialogData?.data?.connectionId as string) || selectedNode?.connectionId || ''
  const workspace = (dialogData?.data?.workspace as string) || selectedNode?.workspace || ''
  const layerName = (dialogData?.data?.layerName as string) || selectedNode?.name || ''

  // Fetch comprehensive layer metadata
  const { data: metadata, isLoading } = useQuery({
    queryKey: ['layerMetadata', connectionId, workspace, layerName],
    queryFn: () => api.getLayerFullMetadata(connectionId, workspace, layerName),
    enabled: isOpen && !!connectionId && !!workspace && !!layerName,
  })

  useEffect(() => {
    if (metadata) {
      setFormData({
        title: metadata.title || '',
        abstract: metadata.abstract || '',
        keywords: metadata.keywords || [],
        srs: metadata.srs || '',
        enabled: metadata.enabled,
        advertised: metadata.advertised,
        queryable: metadata.queryable,
        attributionTitle: metadata.attributionTitle || '',
        attributionHref: metadata.attributionHref || '',
      })
      setMetadataLinks(metadata.metadataLinks || [])
    }
  }, [metadata])

  const updateMutation = useMutation({
    mutationFn: (data: LayerMetadataUpdate) =>
      api.updateLayerMetadata(connectionId, workspace, layerName, {
        ...data,
        metadataLinks: metadataLinks,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['layerMetadata', connectionId, workspace, layerName] })
      queryClient.invalidateQueries({ queryKey: ['layer', connectionId, workspace, layerName] })
      queryClient.invalidateQueries({ queryKey: ['layers', connectionId, workspace] })
      toast({
        title: 'Layer updated',
        description: 'Layer metadata has been updated successfully.',
        status: 'success',
        duration: 3000,
      })
      closeDialog()
    },
    onError: (error: Error) => {
      toast({
        title: 'Error updating layer',
        description: error.message,
        status: 'error',
        duration: 5000,
      })
    },
  })

  const handleChange = <K extends keyof LayerMetadataUpdate>(field: K, value: LayerMetadataUpdate[K]) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
  }

  const handleAddKeyword = () => {
    if (keywordInput.trim() && !formData.keywords?.includes(keywordInput.trim())) {
      setFormData((prev) => ({
        ...prev,
        keywords: [...(prev.keywords || []), keywordInput.trim()],
      }))
      setKeywordInput('')
    }
  }

  const handleRemoveKeyword = (keyword: string) => {
    setFormData((prev) => ({
      ...prev,
      keywords: prev.keywords?.filter((k) => k !== keyword) || [],
    }))
  }

  const handleAddMetadataLink = () => {
    if (newLink.content.trim()) {
      setMetadataLinks((prev) => [...prev, { ...newLink }])
      setNewLink({ type: 'text/html', metadataType: 'TC211', content: '' })
    }
  }

  const handleRemoveMetadataLink = (index: number) => {
    setMetadataLinks((prev) => prev.filter((_, i) => i !== index))
  }

  const handleSubmit = () => {
    updateMutation.mutate(formData)
  }

  if (!isOpen) return null

  return (
    <Modal isOpen={isOpen} onClose={closeDialog} size="2xl" isCentered scrollBehavior="inside">
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent borderRadius="xl" overflow="hidden" maxH="85vh">
        {/* Gradient Header */}
        <Box
          bg="linear-gradient(135deg, #1B6B9B 0%, #3B9DD9 100%)"
          px={6}
          py={4}
        >
          <HStack spacing={3}>
            <Box bg="whiteAlpha.200" p={2} borderRadius="lg">
              <Icon as={FiLayers} boxSize={5} color="white" />
            </Box>
            <Box>
              <Text color="white" fontWeight="600" fontSize="lg">
                Edit Layer Metadata
              </Text>
              <Text color="whiteAlpha.800" fontSize="sm">
                {layerName}
              </Text>
            </Box>
            {metadata && (
              <Badge
                ml="auto"
                colorScheme={metadata.storeType === 'coveragestore' ? 'orange' : 'blue'}
                px={2}
                py={1}
                borderRadius="md"
              >
                {metadata.storeType === 'coveragestore' ? 'Raster' : 'Vector'}
              </Badge>
            )}
          </HStack>
        </Box>
        <ModalCloseButton color="white" />

        <ModalBody py={4} overflowY="auto">
          {isLoading ? (
            <VStack py={10}>
              <Spinner size="xl" color="kartoza.500" />
              <Text color="gray.500">Loading layer metadata...</Text>
            </VStack>
          ) : (
            <Tabs variant="enclosed" colorScheme="kartoza">
              <TabList>
                <Tab><HStack spacing={2}><Icon as={FiInfo} /><Text>Basic Info</Text></HStack></Tab>
                <Tab><HStack spacing={2}><Icon as={FiGlobe} /><Text>Description</Text></HStack></Tab>
                <Tab><HStack spacing={2}><Icon as={FiLink} /><Text>Attribution</Text></HStack></Tab>
              </TabList>

              <TabPanels>
                {/* Basic Info Tab */}
                <TabPanel px={0} py={4}>
                  <VStack spacing={4} align="stretch">
                    {/* Layer Settings */}
                    <FormControl
                      display="flex"
                      alignItems="center"
                      justifyContent="space-between"
                      p={3}
                      bg="gray.50"
                      borderRadius="lg"
                    >
                      <HStack>
                        <Icon as={FiLayers} color="green.500" />
                        <Box>
                          <FormLabel mb={0} fontWeight="500">Enabled</FormLabel>
                          <Text fontSize="xs" color="gray.500">
                            Layer is available for use in GeoServer
                          </Text>
                        </Box>
                      </HStack>
                      <Switch
                        colorScheme="green"
                        size="lg"
                        isChecked={formData.enabled}
                        onChange={(e) => handleChange('enabled', e.target.checked)}
                      />
                    </FormControl>

                    <FormControl
                      display="flex"
                      alignItems="center"
                      justifyContent="space-between"
                      p={3}
                      bg="gray.50"
                      borderRadius="lg"
                    >
                      <HStack>
                        <Icon as={FiEye} color="blue.500" />
                        <Box>
                          <FormLabel mb={0} fontWeight="500">Advertised</FormLabel>
                          <Text fontSize="xs" color="gray.500">
                            Layer appears in GetCapabilities responses
                          </Text>
                        </Box>
                      </HStack>
                      <Switch
                        colorScheme="blue"
                        size="lg"
                        isChecked={formData.advertised}
                        onChange={(e) => handleChange('advertised', e.target.checked)}
                      />
                    </FormControl>

                    <FormControl
                      display="flex"
                      alignItems="center"
                      justifyContent="space-between"
                      p={3}
                      bg="gray.50"
                      borderRadius="lg"
                    >
                      <HStack>
                        <Icon as={FiSearch} color="purple.500" />
                        <Box>
                          <FormLabel mb={0} fontWeight="500">Queryable</FormLabel>
                          <Text fontSize="xs" color="gray.500">
                            Layer supports GetFeatureInfo requests
                          </Text>
                        </Box>
                      </HStack>
                      <Switch
                        colorScheme="purple"
                        size="lg"
                        isChecked={formData.queryable}
                        onChange={(e) => handleChange('queryable', e.target.checked)}
                      />
                    </FormControl>

                    <Divider />

                    {/* Data Information (read-only) */}
                    <Accordion allowToggle>
                      <AccordionItem border="none">
                        <AccordionButton bg="gray.50" borderRadius="lg" _hover={{ bg: 'gray.100' }}>
                          <Box flex="1" textAlign="left">
                            <Text fontWeight="500" color="gray.600">Data Information</Text>
                          </Box>
                          <AccordionIcon />
                        </AccordionButton>
                        <AccordionPanel pb={4}>
                          <SimpleGrid columns={2} spacing={4}>
                            <Box>
                              <Text fontSize="xs" color="gray.500">Workspace</Text>
                              <Text fontWeight="medium">{workspace}</Text>
                            </Box>
                            <Box>
                              <Text fontSize="xs" color="gray.500">Store</Text>
                              <Text fontWeight="medium">{metadata?.store || 'Unknown'}</Text>
                            </Box>
                            <Box>
                              <Text fontSize="xs" color="gray.500">Native Name</Text>
                              <Code fontSize="sm">{metadata?.nativeName || metadata?.name}</Code>
                            </Box>
                            <Box>
                              <Text fontSize="xs" color="gray.500">Default Style</Text>
                              <Text fontWeight="medium">{metadata?.defaultStyle || 'None'}</Text>
                            </Box>
                            <Box>
                              <Text fontSize="xs" color="gray.500">Native CRS</Text>
                              <Code fontSize="sm">{metadata?.nativeCRS || 'Unknown'}</Code>
                            </Box>
                            <Box>
                              <Text fontSize="xs" color="gray.500">Declared SRS</Text>
                              <Code fontSize="sm">{metadata?.srs || 'Unknown'}</Code>
                            </Box>
                          </SimpleGrid>

                          {metadata?.nativeBoundingBox && (
                            <Box mt={4}>
                              <Text fontSize="xs" color="gray.500" mb={2}>Native Bounding Box</Text>
                              <Code display="block" p={2} borderRadius="md" fontSize="xs">
                                ({metadata.nativeBoundingBox.minx.toFixed(4)}, {metadata.nativeBoundingBox.miny.toFixed(4)}) -
                                ({metadata.nativeBoundingBox.maxx.toFixed(4)}, {metadata.nativeBoundingBox.maxy.toFixed(4)})
                              </Code>
                            </Box>
                          )}

                          {metadata?.latLonBoundingBox && (
                            <Box mt={4}>
                              <Text fontSize="xs" color="gray.500" mb={2}>Lat/Lon Bounding Box</Text>
                              <Code display="block" p={2} borderRadius="md" fontSize="xs">
                                ({metadata.latLonBoundingBox.minx.toFixed(4)}, {metadata.latLonBoundingBox.miny.toFixed(4)}) -
                                ({metadata.latLonBoundingBox.maxx.toFixed(4)}, {metadata.latLonBoundingBox.maxy.toFixed(4)})
                              </Code>
                            </Box>
                          )}
                        </AccordionPanel>
                      </AccordionItem>
                    </Accordion>
                  </VStack>
                </TabPanel>

                {/* Description Tab */}
                <TabPanel px={0} py={4}>
                  <VStack spacing={4} align="stretch">
                    <FormControl>
                      <FormLabel fontWeight="500">Title</FormLabel>
                      <Input
                        value={formData.title || ''}
                        onChange={(e) => handleChange('title', e.target.value)}
                        placeholder="Human-readable layer title"
                        bg="gray.50"
                        borderRadius="lg"
                      />
                      <Text fontSize="xs" color="gray.500" mt={1}>
                        A short descriptive name for the layer
                      </Text>
                    </FormControl>

                    <FormControl>
                      <FormLabel fontWeight="500">Abstract / Description</FormLabel>
                      <Textarea
                        value={formData.abstract || ''}
                        onChange={(e) => handleChange('abstract', e.target.value)}
                        placeholder="Detailed description of the layer content, purpose, and data source..."
                        bg="gray.50"
                        borderRadius="lg"
                        rows={4}
                      />
                      <Text fontSize="xs" color="gray.500" mt={1}>
                        A longer description of the layer for metadata and documentation
                      </Text>
                    </FormControl>

                    <FormControl>
                      <FormLabel fontWeight="500">Keywords</FormLabel>
                      <HStack>
                        <Input
                          value={keywordInput}
                          onChange={(e) => setKeywordInput(e.target.value)}
                          placeholder="Add a keyword..."
                          bg="gray.50"
                          borderRadius="lg"
                          onKeyPress={(e) => e.key === 'Enter' && handleAddKeyword()}
                        />
                        <IconButton
                          aria-label="Add keyword"
                          icon={<FiPlus />}
                          colorScheme="kartoza"
                          onClick={handleAddKeyword}
                          borderRadius="lg"
                        />
                      </HStack>
                      <Wrap mt={2} spacing={2}>
                        {formData.keywords?.map((keyword) => (
                          <WrapItem key={keyword}>
                            <Tag size="md" colorScheme="blue" borderRadius="full">
                              <TagLabel>{keyword}</TagLabel>
                              <TagCloseButton onClick={() => handleRemoveKeyword(keyword)} />
                            </Tag>
                          </WrapItem>
                        ))}
                      </Wrap>
                      <Text fontSize="xs" color="gray.500" mt={1}>
                        Keywords help users find this layer in searches
                      </Text>
                    </FormControl>
                  </VStack>
                </TabPanel>

                {/* Attribution Tab */}
                <TabPanel px={0} py={4}>
                  <VStack spacing={4} align="stretch">
                    <Box p={4} bg="blue.50" borderRadius="lg" borderLeft="4px solid" borderLeftColor="blue.400">
                      <Text fontSize="sm" color="blue.700">
                        <strong>Attribution</strong> provides credit to the data source and is displayed in WMS GetCapabilities and map clients.
                      </Text>
                    </Box>

                    <FormControl>
                      <FormLabel fontWeight="500">Attribution Title</FormLabel>
                      <Input
                        value={formData.attributionTitle || ''}
                        onChange={(e) => handleChange('attributionTitle', e.target.value)}
                        placeholder="e.g., Data provided by..."
                        bg="gray.50"
                        borderRadius="lg"
                      />
                    </FormControl>

                    <FormControl>
                      <FormLabel fontWeight="500">Attribution Link (URL)</FormLabel>
                      <Input
                        value={formData.attributionHref || ''}
                        onChange={(e) => handleChange('attributionHref', e.target.value)}
                        placeholder="https://example.com/data-source"
                        bg="gray.50"
                        borderRadius="lg"
                      />
                    </FormControl>

                    <Divider />

                    <Box>
                      <Text fontWeight="500" mb={3}>Metadata Links</Text>
                      <Text fontSize="xs" color="gray.500" mb={3}>
                        Links to external metadata documents (ISO 19115, FGDC, etc.)
                      </Text>

                      {metadataLinks.length > 0 && (
                        <VStack spacing={2} mb={4} align="stretch">
                          {metadataLinks.map((link, index) => (
                            <HStack
                              key={index}
                              p={3}
                              bg="gray.50"
                              borderRadius="lg"
                              justify="space-between"
                            >
                              <VStack align="start" spacing={0}>
                                <HStack>
                                  <Badge colorScheme="purple">{link.metadataType}</Badge>
                                  <Badge colorScheme="gray">{link.type}</Badge>
                                </HStack>
                                <Text fontSize="sm" color="gray.600" isTruncated maxW="300px">
                                  {link.content}
                                </Text>
                              </VStack>
                              <IconButton
                                aria-label="Remove link"
                                icon={<FiTrash2 />}
                                size="sm"
                                colorScheme="red"
                                variant="ghost"
                                onClick={() => handleRemoveMetadataLink(index)}
                              />
                            </HStack>
                          ))}
                        </VStack>
                      )}

                      <Box p={4} bg="gray.50" borderRadius="lg">
                        <Text fontWeight="500" fontSize="sm" mb={3}>Add Metadata Link</Text>
                        <VStack spacing={3}>
                          <HStack w="100%">
                            <FormControl flex={1}>
                              <FormLabel fontSize="xs">Type</FormLabel>
                              <Input
                                size="sm"
                                value={newLink.type}
                                onChange={(e) => setNewLink({ ...newLink, type: e.target.value })}
                                placeholder="text/html"
                              />
                            </FormControl>
                            <FormControl flex={1}>
                              <FormLabel fontSize="xs">Metadata Type</FormLabel>
                              <Input
                                size="sm"
                                value={newLink.metadataType}
                                onChange={(e) => setNewLink({ ...newLink, metadataType: e.target.value })}
                                placeholder="TC211"
                              />
                            </FormControl>
                          </HStack>
                          <FormControl>
                            <FormLabel fontSize="xs">URL</FormLabel>
                            <HStack>
                              <Input
                                size="sm"
                                value={newLink.content}
                                onChange={(e) => setNewLink({ ...newLink, content: e.target.value })}
                                placeholder="https://example.com/metadata.xml"
                              />
                              <IconButton
                                aria-label="Add link"
                                icon={<FiPlus />}
                                size="sm"
                                colorScheme="kartoza"
                                onClick={handleAddMetadataLink}
                              />
                            </HStack>
                          </FormControl>
                        </VStack>
                      </Box>
                    </Box>
                  </VStack>
                </TabPanel>
              </TabPanels>
            </Tabs>
          )}
        </ModalBody>

        <ModalFooter
          gap={3}
          borderTop="1px solid"
          borderTopColor="gray.100"
          bg="gray.50"
        >
          <Button variant="ghost" onClick={closeDialog} borderRadius="lg">
            Cancel
          </Button>
          <Button
            colorScheme="kartoza"
            onClick={handleSubmit}
            isLoading={updateMutation.isPending}
            borderRadius="lg"
            px={6}
          >
            Save Changes
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
