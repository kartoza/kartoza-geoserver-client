import { useState, useCallback, useRef, useEffect } from 'react'
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  Box,
  Text,
  VStack,
  HStack,
  Progress,
  List,
  ListItem,
  ListIcon,
  Icon,
  Badge,
  useToast,
  useColorModeValue,
  Checkbox,
  Divider,
  Spinner,
} from '@chakra-ui/react'
import { FiFile, FiCheck, FiX, FiUploadCloud, FiLayers, FiDatabase, FiPause, FiPlay } from 'react-icons/fi'
import { useQueryClient } from '@tanstack/react-query'
import { useUIStore } from '../../stores/uiStore'
import { useTreeStore } from '../../stores/treeStore'
import { useConnectionStore } from '../../stores/connectionStore'
import * as api from '../../api'
import { useChunkedUpload } from '../../hooks/useChunkedUpload'

interface FileUpload {
  file: File
  progress: number
  chunkProgress: number
  status: 'pending' | 'uploading' | 'paused' | 'success' | 'error' | 'cancelled'
  speedBps: number
  etaSeconds: number
  chunksUploaded: number
  chunksTotal: number
  geoServerSent: number
  geoServerTotal: number
  error?: string
  storeName?: string
  storeType?: string
}

interface AvailableLayer {
  name: string
  selected: boolean
}

function formatSpeed(bps: number): string {
  if (bps <= 0) return ''
  if (bps < 1024) return `${bps.toFixed(0)} B/s`
  if (bps < 1024 * 1024) return `${(bps / 1024).toFixed(1)} KB/s`
  return `${(bps / (1024 * 1024)).toFixed(1)} MB/s`
}

function formatEta(seconds: number): string {
  if (!isFinite(seconds) || seconds <= 0) return ''
  if (seconds < 60) return `${Math.ceil(seconds)}s`
  const m = Math.floor(seconds / 60)
  const s = Math.ceil(seconds % 60)
  return `${m}m ${s}s`
}

export default function UploadDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const closeDialog = useUIStore((state) => state.closeDialog)
  const selectedNode = useTreeStore((state) => state.selectedNode)
  const activeConnectionId = useConnectionStore((state) => state.activeConnectionId)
  const queryClient = useQueryClient()
  const toast = useToast()

  const [files, setFiles] = useState<FileUpload[]>([])
  const [isUploading, setIsUploading] = useState(false)
  const [uploadComplete, setUploadComplete] = useState(false)
  const [availableLayers, setAvailableLayers] = useState<AvailableLayer[]>([])
  const [loadingLayers, setLoadingLayers] = useState(false)
  const [publishingLayers, setPublishingLayers] = useState(false)
  const [currentStore, setCurrentStore] = useState<{ name: string; type: string } | null>(null)
  const [currentFileIndex, setCurrentFileIndex] = useState<number>(-1)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const dropzoneBg = useColorModeValue('gray.50', 'gray.700')
  const dropzoneBorder = useColorModeValue('gray.300', 'gray.600')

  const isOpen = activeDialog === 'upload'
  const connectionId = selectedNode?.connectionId || activeConnectionId
  const workspace = selectedNode?.workspace

  const chunkedUpload = useChunkedUpload()

  // Reset state when dialog opens
  useEffect(() => {
    if (isOpen) {
      setFiles([])
      setUploadComplete(false)
      setAvailableLayers([])
      setCurrentStore(null)
      setCurrentFileIndex(-1)
      chunkedUpload.reset()
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen])

  useEffect(() => {
    if (currentFileIndex < 0) return
    const { state } = chunkedUpload
    setFiles((prev) =>
      prev.map((f, idx) => {
        if (idx !== currentFileIndex) return f

        const overallPct =
          state.chunksTotal > 0
            ? Math.round((state.chunksUploaded / state.chunksTotal) * 100)
            : f.progress

        let status = f.status
        if (state.status === 'uploading') status = 'uploading'
        else if (state.status === 'paused') status = 'paused'
        else if (state.status === 'completed') status = 'success'
        else if (state.status === 'error') status = 'error'
        else if (state.status === 'cancelled') status = 'cancelled'

        return {
          ...f,
          status,
          progress: overallPct,
          chunkProgress: state.currentChunkProgress,
          speedBps: state.speedBps,
          etaSeconds: state.etaSeconds,
          chunksUploaded: state.chunksUploaded,
          chunksTotal: state.chunksTotal,
          geoServerSent: state.geoServerSent,
          geoServerTotal: state.geoServerTotal,
          error: state.error,
        }
      })
    )
  }, [chunkedUpload.state, currentFileIndex]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    const droppedFiles = Array.from(e.dataTransfer.files)
    addFiles(droppedFiles)
  }, [])

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files)
      addFiles(selectedFiles)
    }
  }

  const addFiles = (newFiles: File[]) => {
    const supportedExtensions = ['.shp', '.zip', '.gpkg', '.tif', '.tiff', '.sld', '.css']
    const validFiles = newFiles.filter((file) =>
      supportedExtensions.some((ext) => file.name.toLowerCase().endsWith(ext))
    )

    if (validFiles.length < newFiles.length) {
      toast({
        title: 'Some files skipped',
        description: 'Only shapefiles, GeoPackages, GeoTIFFs, and styles are supported',
        status: 'warning',
        duration: 3000,
      })
    }

    setFiles((prev) => [
      ...prev,
      ...validFiles.map((file) => ({
        file,
        progress: 0,
        chunkProgress: 0,
        status: 'pending' as const,
        speedBps: 0,
        etaSeconds: 0,
        chunksUploaded: 0,
        chunksTotal: 0,
        geoServerSent: 0,
        geoServerTotal: 0,
      })),
    ])
    setUploadComplete(false)
    setAvailableLayers([])
  }

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index))
  }

  const handleUpload = async () => {
    if (!connectionId || !workspace) {
      toast({
        title: 'No workspace selected',
        description: 'Please select a workspace in the tree first',
        status: 'warning',
        duration: 3000,
      })
      return
    }

    if (files.length === 0) {
      toast({
        title: 'No files selected',
        description: 'Please select files to upload',
        status: 'warning',
        duration: 3000,
      })
      return
    }

    setIsUploading(true)
    let lastGpkgStore: { name: string; type: string } | null = null

    for (let i = 0; i < files.length; i++) {
      if (files[i].status !== 'pending') continue

      setCurrentFileIndex(i)
      chunkedUpload.reset()

      setFiles((prev) =>
        prev.map((f, idx) =>
          idx === i ? { ...f, status: 'uploading' as const } : f
        )
      )

      try {
        const result = await chunkedUpload.start(connectionId, workspace, files[i].file)

        setFiles((prev) =>
          prev.map((f, idx) =>
            idx === i
              ? {
                  ...f,
                  status: 'success' as const,
                  progress: 100,
                  chunkProgress: 100,
                  storeName: result.storeName,
                  storeType: result.storeType,
                }
              : f
          )
        )

        if (files[i].file.name.toLowerCase().endsWith('.gpkg') && result.storeName) {
          lastGpkgStore = { name: result.storeName, type: result.storeType || 'datastore' }
        }
      } catch (err) {
        const errMsg = (err as Error).message
        if (errMsg === 'Upload cancelled') {
          setFiles((prev) =>
            prev.map((f, idx) =>
              idx === i ? { ...f, status: 'cancelled' as const } : f
            )
          )
        } else {
          setFiles((prev) =>
            prev.map((f, idx) =>
              idx === i ? { ...f, status: 'error' as const, error: errMsg } : f
            )
          )
        }
      }
    }

    setCurrentFileIndex(-1)
    setIsUploading(false)
    setUploadComplete(true)

    queryClient.invalidateQueries({ queryKey: ['datastores', connectionId, workspace] })
    queryClient.invalidateQueries({ queryKey: ['coveragestores', connectionId, workspace] })
    queryClient.invalidateQueries({ queryKey: ['styles', connectionId, workspace] })
    queryClient.invalidateQueries({ queryKey: ['layers', connectionId, workspace] })

    const successCount = files.filter((f) => f.status === 'success').length
    if (successCount > 0) {
      toast({
        title: 'Upload complete',
        description: `${successCount} file(s) uploaded successfully`,
        status: 'success',
        duration: 3000,
      })
    }

    if (lastGpkgStore && lastGpkgStore.type === 'datastore') {
      setCurrentStore(lastGpkgStore)
      await loadAvailableLayers(lastGpkgStore.name)
    }
  }

  const loadAvailableLayers = async (storeName: string) => {
    if (!connectionId || !workspace) return

    setLoadingLayers(true)
    try {
      const layers = await api.getAvailableFeatureTypes(connectionId, workspace, storeName)
      if (layers.length > 0) {
        setAvailableLayers(layers.map((name) => ({ name, selected: true })))
      }
    } catch (err) {
      console.error('Failed to load available layers:', err)
    } finally {
      setLoadingLayers(false)
    }
  }

  const toggleLayerSelection = (index: number) => {
    setAvailableLayers((prev) =>
      prev.map((layer, i) =>
        i === index ? { ...layer, selected: !layer.selected } : layer
      )
    )
  }

  const selectAllLayers = () => {
    setAvailableLayers((prev) =>
      prev.map((layer) => ({ ...layer, selected: true }))
    )
  }

  const deselectAllLayers = () => {
    setAvailableLayers((prev) =>
      prev.map((layer) => ({ ...layer, selected: false }))
    )
  }

  const handlePublishLayers = async () => {
    if (!connectionId || !workspace || !currentStore) return

    const selectedLayers = availableLayers.filter((l) => l.selected).map((l) => l.name)
    if (selectedLayers.length === 0) {
      toast({
        title: 'No layers selected',
        description: 'Please select at least one layer to publish',
        status: 'warning',
        duration: 3000,
      })
      return
    }

    setPublishingLayers(true)
    try {
      const result = await api.publishFeatureTypes(connectionId, workspace, currentStore.name, selectedLayers)

      if (result.published.length > 0) {
        toast({
          title: 'Layers published',
          description: `${result.published.length} layer(s) published successfully`,
          status: 'success',
          duration: 3000,
        })

        setAvailableLayers((prev) =>
          prev.filter((l) => !result.published.includes(l.name))
        )

        queryClient.invalidateQueries({ queryKey: ['layers', connectionId, workspace] })
      }

      if (result.errors.length > 0) {
        toast({
          title: 'Some layers failed',
          description: result.errors.join(', '),
          status: 'error',
          duration: 5000,
        })
      }
    } catch (err) {
      toast({
        title: 'Publish failed',
        description: (err as Error).message,
        status: 'error',
        duration: 5000,
      })
    } finally {
      setPublishingLayers(false)
    }
  }

  const handleClose = () => {
    setFiles([])
    setUploadComplete(false)
    setAvailableLayers([])
    setCurrentStore(null)
    setCurrentFileIndex(-1)
    chunkedUpload.reset()
    closeDialog()
  }

  const getFileIcon = (status: FileUpload['status']) => {
    switch (status) {
      case 'success':
        return FiCheck
      case 'error':
        return FiX
      case 'cancelled':
        return FiX
      default:
        return FiFile
    }
  }

  const getFileColor = (status: FileUpload['status']) => {
    switch (status) {
      case 'success':
        return 'green.500'
      case 'error':
        return 'red.500'
      case 'cancelled':
        return 'orange.500'
      default:
        return 'gray.500'
    }
  }

  const hasPendingUploads = files.some((f) => f.status === 'pending')
  const selectedLayerCount = availableLayers.filter((l) => l.selected).length

  const uploadingFiles = files.filter((f) => f.status === 'uploading' || f.status === 'paused')
  const totalFiles = files.length
  const completedFiles = files.filter((f) => f.status === 'success' || f.status === 'error' || f.status === 'cancelled').length

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="lg" isCentered>
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent borderRadius="xl" overflow="hidden" maxH="85vh">
        {/* Gradient Header */}
        <Box
          bg="linear-gradient(135deg, #0a3a50 0%, #175a77 50%, #2d7d9b 100%)"
          px={6}
          py={4}
        >
          <HStack spacing={3}>
            <Box bg="whiteAlpha.200" p={2} borderRadius="lg">
              <Icon as={FiUploadCloud} boxSize={5} color="white" />
            </Box>
            <Box flex="1">
              <Text color="white" fontWeight="600" fontSize="lg">
                Upload Files
              </Text>
              <HStack spacing={2}>
                <Text color="whiteAlpha.800" fontSize="sm">
                  Target:
                </Text>
                {workspace ? (
                  <Badge bg="whiteAlpha.200" color="white" fontSize="xs">
                    {workspace}
                  </Badge>
                ) : (
                  <Text color="whiteAlpha.600" fontSize="sm" fontStyle="italic">
                    No workspace selected
                  </Text>
                )}
              </HStack>
            </Box>
          </HStack>
        </Box>
        <ModalCloseButton color="white" />

        <ModalBody py={6} overflowY="auto">
          <VStack spacing={4}>
            {/* Dropzone - hide after upload complete */}
            {!uploadComplete && (
              <Box
                w="100%"
                p={8}
                bg={dropzoneBg}
                border="2px dashed"
                borderColor={dropzoneBorder}
                borderRadius="xl"
                textAlign="center"
                cursor="pointer"
                onClick={() => fileInputRef.current?.click()}
                onDrop={handleDrop}
                onDragOver={(e) => e.preventDefault()}
                _hover={{
                  borderColor: 'kartoza.500',
                  bg: 'kartoza.50',
                }}
                transition="all 0.2s"
              >
                <VStack spacing={3}>
                  <Box
                    bg="kartoza.50"
                    p={4}
                    borderRadius="full"
                  >
                    <Icon as={FiUploadCloud} boxSize={10} color="kartoza.500" />
                  </Box>
                  <VStack spacing={1}>
                    <Text fontWeight="600" color="gray.700">
                      Drop files here or click to browse
                    </Text>
                    <Text fontSize="sm" color="gray.500">
                      Shapefile (.zip), GeoPackage (.gpkg), GeoTIFF (.tif), SLD, CSS
                    </Text>
                  </VStack>
                </VStack>
                <input
                  ref={fileInputRef}
                  type="file"
                  multiple
                  accept=".zip,.shp,.gpkg,.tif,.tiff,.sld,.css"
                  style={{ display: 'none' }}
                  onChange={handleFileSelect}
                />
              </Box>
            )}

            {/* Overall progress when uploading multiple files */}
            {isUploading && totalFiles > 1 && (
              <Box w="100%" p={3} bg={dropzoneBg} borderRadius="lg" border="1px solid" borderColor="gray.200">
                <HStack justify="space-between" mb={2}>
                  <Text fontSize="sm" fontWeight="500" color="gray.700">
                    File {completedFiles + (uploadingFiles.length > 0 ? 1 : 0)} of {totalFiles}
                  </Text>
                  <Text fontSize="xs" color="gray.500">
                    {completedFiles} completed
                  </Text>
                </HStack>
                <Progress
                  value={totalFiles > 0 ? (completedFiles / totalFiles) * 100 : 0}
                  size="sm"
                  colorScheme="kartoza"
                  borderRadius="full"
                />
              </Box>
            )}

            {/* File list */}
            {files.length > 0 && (
              <List spacing={2} w="100%">
                {files.map((upload, index) => (
                  <ListItem
                    key={index}
                    p={3}
                    bg={dropzoneBg}
                    borderRadius="lg"
                    border="1px solid"
                    borderColor={
                      upload.status === 'success'
                        ? 'green.200'
                        : upload.status === 'error'
                        ? 'red.200'
                        : upload.status === 'cancelled'
                        ? 'orange.200'
                        : 'gray.200'
                    }
                  >
                    <VStack align="stretch" spacing={2}>
                      <Box display="flex" alignItems="center" gap={2}>
                        <ListIcon
                          as={getFileIcon(upload.status)}
                          color={getFileColor(upload.status)}
                          boxSize={5}
                        />
                        <Text flex="1" fontSize="sm" noOfLines={1} fontWeight="500">
                          {upload.file.name}
                        </Text>
                        {upload.status === 'pending' && (
                          <Button
                            size="xs"
                            variant="ghost"
                            colorScheme="red"
                            onClick={() => removeFile(index)}
                            borderRadius="md"
                          >
                            Remove
                          </Button>
                        )}
                        {upload.status === 'success' && (
                          <Badge colorScheme="green" borderRadius="md">Complete</Badge>
                        )}
                        {upload.status === 'cancelled' && (
                          <Badge colorScheme="orange" borderRadius="md">Cancelled</Badge>
                        )}
                        {(upload.status === 'uploading' || upload.status === 'paused') &&
                          !(upload.chunksUploaded >= upload.chunksTotal && upload.chunksTotal > 0) && (
                          <HStack spacing={1}>
                            {upload.status === 'uploading' ? (
                              <Button
                                size="xs"
                                variant="outline"
                                colorScheme="yellow"
                                onClick={chunkedUpload.pause}
                                leftIcon={<FiPause />}
                                borderRadius="md"
                              >
                                Pause
                              </Button>
                            ) : (
                              <Button
                                size="xs"
                                variant="outline"
                                colorScheme="green"
                                onClick={chunkedUpload.resume}
                                leftIcon={<FiPlay />}
                                borderRadius="md"
                              >
                                Resume
                              </Button>
                            )}
                            <Button
                              size="xs"
                              variant="ghost"
                              colorScheme="red"
                              onClick={chunkedUpload.cancel}
                              borderRadius="md"
                            >
                              Cancel
                            </Button>
                          </HStack>
                        )}
                      </Box>

                      {/* GeoServer upload progress (finalizing phase) */}
                      {(upload.status === 'uploading' || upload.status === 'paused') &&
                        upload.chunksUploaded >= upload.chunksTotal &&
                        upload.chunksTotal > 0 && (
                        <Box>
                          <HStack justify="space-between" mb={1}>
                            <Text fontSize="xs" color="blue.500" fontWeight="500">
                              Sending to GeoServer…
                            </Text>
                            {upload.geoServerTotal > 0 && (
                              <Text fontSize="xs" color="gray.500">
                                {Math.round((upload.geoServerSent / upload.geoServerTotal) * 100)}%
                              </Text>
                            )}
                          </HStack>
                          <Progress
                            value={
                              upload.geoServerTotal > 0
                                ? (upload.geoServerSent / upload.geoServerTotal) * 100
                                : undefined
                            }
                            isIndeterminate={upload.geoServerTotal === 0}
                            size="sm"
                            colorScheme="blue"
                            borderRadius="full"
                          />
                        </Box>
                      )}

                      {/* Overall progress bar */}
                      {(upload.status === 'uploading' || upload.status === 'paused') && upload.chunksTotal > 0 && (
                        <>
                          <Box>
                            <HStack justify="space-between" mb={1}>
                              <Text fontSize="xs" color="gray.500">
                                {upload.chunksUploaded >= upload.chunksTotal
                                  ? 'All chunks received'
                                  : `Overall: ${upload.chunksUploaded}/${upload.chunksTotal} chunks`}
                              </Text>
                              <HStack spacing={3}>
                                {upload.speedBps > 0 && (
                                  <Text fontSize="xs" color="gray.500">
                                    {formatSpeed(upload.speedBps)}
                                  </Text>
                                )}
                                {upload.etaSeconds > 0 && (
                                  <Text fontSize="xs" color="gray.500">
                                    ETA: {formatEta(upload.etaSeconds)}
                                  </Text>
                                )}
                              </HStack>
                            </HStack>
                            <Progress
                              value={upload.progress}
                              size="sm"
                              colorScheme={upload.status === 'paused' ? 'yellow' : 'kartoza'}
                              borderRadius="full"
                            />
                          </Box>

                          {/* Current chunk progress bar */}
                          {upload.chunksTotal > 1 && (
                            <Box>
                              <Text fontSize="xs" color="gray.400" mb={1}>
                                Current chunk: {upload.chunkProgress}%
                              </Text>
                              <Progress
                                value={upload.chunkProgress}
                                size="xs"
                                colorScheme="blue"
                                borderRadius="full"
                                opacity={0.7}
                              />
                            </Box>
                          )}
                        </>
                      )}

                      {(upload.status === 'uploading' || upload.status === 'paused') && upload.chunksTotal <= 1 && (
                        <Progress
                          value={upload.chunkProgress}
                          size="sm"
                          colorScheme={upload.status === 'paused' ? 'yellow' : 'kartoza'}
                          borderRadius="full"
                        />
                      )}

                      {upload.status === 'error' && (
                        <Text fontSize="xs" color="red.500">
                          {upload.error}
                        </Text>
                      )}
                    </VStack>
                  </ListItem>
                ))}
              </List>
            )}

            {/* Available layers section */}
            {uploadComplete && availableLayers.length > 0 && (
              <>
                <Divider />
                <Box w="100%">
                  <HStack justify="space-between" mb={3}>
                    <HStack>
                      <Icon as={FiLayers} color="kartoza.500" />
                      <Text fontWeight="600">Available Layers in {currentStore?.name}</Text>
                    </HStack>
                    <HStack spacing={2}>
                      <Button size="xs" variant="ghost" onClick={selectAllLayers}>
                        Select All
                      </Button>
                      <Button size="xs" variant="ghost" onClick={deselectAllLayers}>
                        Deselect All
                      </Button>
                    </HStack>
                  </HStack>
                  <Text fontSize="sm" color="gray.500" mb={3}>
                    The GeoPackage contains multiple layers. Select which layers to publish:
                  </Text>
                  <VStack align="stretch" spacing={2} maxH="200px" overflowY="auto">
                    {availableLayers.map((layer, index) => (
                      <Box
                        key={layer.name}
                        p={2}
                        bg={dropzoneBg}
                        borderRadius="md"
                        border="1px solid"
                        borderColor={layer.selected ? 'kartoza.200' : 'gray.200'}
                      >
                        <Checkbox
                          isChecked={layer.selected}
                          onChange={() => toggleLayerSelection(index)}
                          colorScheme="kartoza"
                        >
                          <HStack spacing={2}>
                            <Icon as={FiDatabase} color="gray.500" boxSize={4} />
                            <Text fontSize="sm">{layer.name}</Text>
                          </HStack>
                        </Checkbox>
                      </Box>
                    ))}
                  </VStack>
                </Box>
              </>
            )}

            {/* Loading layers indicator */}
            {loadingLayers && (
              <HStack spacing={2} color="gray.500">
                <Spinner size="sm" />
                <Text fontSize="sm">Checking for available layers...</Text>
              </HStack>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter
          gap={3}
          borderTop="1px solid"
          borderTopColor="gray.100"
          bg="gray.50"
        >
          <Button variant="ghost" onClick={handleClose} borderRadius="lg">
            {uploadComplete ? 'Close' : 'Cancel'}
          </Button>

          {/* Show publish button if there are available layers */}
          {uploadComplete && availableLayers.length > 0 && (
            <Button
              colorScheme="green"
              onClick={handlePublishLayers}
              isLoading={publishingLayers}
              loadingText="Publishing..."
              isDisabled={selectedLayerCount === 0}
              leftIcon={<FiLayers />}
              borderRadius="lg"
              px={6}
            >
              Publish {selectedLayerCount > 0 && `(${selectedLayerCount})`}
            </Button>
          )}

          {/* Show upload button if there are pending uploads */}
          {(!uploadComplete || hasPendingUploads) && (
            <Button
              colorScheme="kartoza"
              onClick={handleUpload}
              isLoading={isUploading}
              loadingText="Uploading..."
              isDisabled={files.length === 0 || !workspace || !hasPendingUploads}
              leftIcon={<FiUploadCloud />}
              borderRadius="lg"
              px={6}
            >
              Upload {files.filter((f) => f.status === 'pending').length > 0 &&
                `(${files.filter((f) => f.status === 'pending').length})`}
            </Button>
          )}
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
