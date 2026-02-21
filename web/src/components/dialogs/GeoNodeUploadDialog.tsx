import { useState, useEffect, useRef, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
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
  Input,
  VStack,
  HStack,
  Text,
  Icon,
  Box,
  useToast,
  Progress,
  Textarea,
  Badge,
  Alert,
  AlertIcon,
  useColorModeValue,
} from '@chakra-ui/react'
import { FiUpload, FiFile, FiCheckCircle, FiAlertCircle } from 'react-icons/fi'
import { TbWorld } from 'react-icons/tb'
import { useQueryClient } from '@tanstack/react-query'
import { useUIStore } from '../../stores/uiStore'
import * as api from '../../api/client'

// Helper to format file size
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// Check if file type is supported for GeoNode upload
function getSupportedFileInfo(filename: string): { supported: boolean; type: string; description: string } {
  const ext = filename.split('.').pop()?.toLowerCase() || ''
  switch (ext) {
    case 'gpkg':
      return { supported: true, type: 'GeoPackage', description: 'Multi-layer support - each layer becomes a dataset' }
    case 'shp':
    case 'zip':
      return { supported: true, type: 'Shapefile', description: 'Include .shx, .dbf, .prj files in ZIP' }
    case 'tif':
    case 'tiff':
    case 'geotiff':
      return { supported: true, type: 'GeoTIFF', description: 'Raster dataset' }
    case 'geojson':
    case 'json':
      return { supported: true, type: 'GeoJSON', description: 'Vector dataset' }
    case 'kml':
      return { supported: true, type: 'KML', description: 'Keyhole Markup Language' }
    case 'csv':
      return { supported: true, type: 'CSV', description: 'Requires coordinate columns' }
    default:
      return { supported: false, type: 'Unknown', description: 'File type may not be supported' }
  }
}

export default function GeoNodeUploadDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const dialogData = useUIStore((state) => state.dialogData)
  const closeDialog = useUIStore((state) => state.closeDialog)
  const queryClient = useQueryClient()
  const toast = useToast()
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Dialog data
  const connectionId = dialogData?.data?.connectionId as string | undefined
  const connectionName = dialogData?.data?.connectionName as string | undefined

  // Form state
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [title, setTitle] = useState('')
  const [abstract, setAbstract] = useState('')
  const [fileInfo, setFileInfo] = useState<{ supported: boolean; type: string; description: string } | null>(null)

  // Upload state
  const [isUploading, setIsUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [uploadResult, setUploadResult] = useState<{ success: boolean; message: string } | null>(null)

  const isOpen = activeDialog === 'geonodeupload'
  const dropzoneBg = useColorModeValue('gray.50', 'gray.700')
  const dropzoneBorderColor = useColorModeValue('gray.300', 'gray.600')

  // Reset form when dialog opens
  useEffect(() => {
    if (isOpen) {
      setSelectedFile(null)
      setTitle('')
      setAbstract('')
      setFileInfo(null)
      setUploadProgress(0)
      setUploadResult(null)
    }
  }, [isOpen])

  const handleFileSelect = useCallback((file: File) => {
    setSelectedFile(file)
    setUploadResult(null)
    // Set title from filename (without extension) if not already set
    if (!title) {
      const nameWithoutExt = file.name.replace(/\.[^/.]+$/, '')
      setTitle(nameWithoutExt)
    }
    // Check file type support
    setFileInfo(getSupportedFileInfo(file.name))
  }, [title])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    const file = e.dataTransfer.files[0]
    if (file) {
      handleFileSelect(file)
    }
  }, [handleFileSelect])

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
  }, [])

  const handleUpload = async () => {
    if (!selectedFile || !connectionId) {
      toast({
        title: 'Missing required fields',
        description: 'Please select a file',
        status: 'warning',
        duration: 3000,
      })
      return
    }

    setIsUploading(true)
    setUploadProgress(0)

    // Simulate progress since we don't have real progress tracking
    const progressInterval = setInterval(() => {
      setUploadProgress((prev) => Math.min(prev + 10, 90))
    }, 500)

    try {
      const result = await api.uploadGeoNodeDataset(
        connectionId,
        selectedFile,
        title || undefined,
        abstract || undefined
      )

      clearInterval(progressInterval)
      setUploadProgress(100)

      if (result.success) {
        setUploadResult({
          success: true,
          message: result.message || 'Upload successful! Processing may take a few moments.',
        })

        // Refresh datasets list
        queryClient.invalidateQueries({ queryKey: ['geonodedatasets', connectionId] })
        queryClient.invalidateQueries({ queryKey: ['geonoderesources', connectionId] })

        toast({
          title: 'Upload successful',
          description: 'Your dataset is being processed by GeoNode',
          status: 'success',
          duration: 3000,
        })
      } else {
        setUploadResult({
          success: false,
          message: result.error || 'Upload failed',
        })
      }
    } catch (err) {
      clearInterval(progressInterval)
      setUploadResult({
        success: false,
        message: (err as Error).message,
      })
      toast({
        title: 'Upload failed',
        description: (err as Error).message,
        status: 'error',
        duration: 5000,
      })
    } finally {
      setIsUploading(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={closeDialog} size="lg" isCentered>
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent borderRadius="xl" overflow="hidden">
        {/* Gradient Header */}
        <Box
          bg="linear-gradient(135deg, #0d7377 0%, #14919b 50%, #2dc2c9 100%)"
          p={4}
        >
          <HStack spacing={3}>
            <Box bg="whiteAlpha.200" p={2} borderRadius="lg">
              <Icon as={TbWorld} boxSize={5} color="white" />
            </Box>
            <Box>
              <Text color="white" fontWeight="600" fontSize="lg">
                Upload to GeoNode
              </Text>
              <Text color="whiteAlpha.800" fontSize="sm">
                {connectionName || 'Upload geospatial data as datasets'}
              </Text>
            </Box>
          </HStack>
        </Box>
        <ModalCloseButton color="white" />

        <ModalBody py={6}>
          <VStack spacing={4}>
            {/* File Drop Zone */}
            <FormControl>
              <FormLabel fontWeight="500" color="gray.700">File</FormLabel>
              <Box
                border="2px dashed"
                borderColor={selectedFile ? 'teal.400' : dropzoneBorderColor}
                borderRadius="lg"
                p={6}
                bg={selectedFile ? 'teal.50' : dropzoneBg}
                textAlign="center"
                cursor="pointer"
                transition="all 0.2s"
                _hover={{ borderColor: 'teal.400', bg: 'teal.50' }}
                onClick={() => fileInputRef.current?.click()}
                onDrop={handleDrop}
                onDragOver={handleDragOver}
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  hidden
                  accept=".gpkg,.shp,.zip,.tif,.tiff,.geotiff,.geojson,.json,.kml,.csv"
                  onChange={(e) => {
                    const file = e.target.files?.[0]
                    if (file) handleFileSelect(file)
                  }}
                />
                {selectedFile ? (
                  <VStack spacing={2}>
                    <Icon as={FiFile} boxSize={8} color="teal.500" />
                    <Text fontWeight="500" color="gray.700">{selectedFile.name}</Text>
                    <Text fontSize="sm" color="gray.500">{formatFileSize(selectedFile.size)}</Text>
                    {fileInfo && (
                      <Badge colorScheme={fileInfo.supported ? 'teal' : 'orange'}>
                        {fileInfo.type}
                      </Badge>
                    )}
                    {fileInfo && (
                      <Text fontSize="xs" color="gray.500">{fileInfo.description}</Text>
                    )}
                  </VStack>
                ) : (
                  <VStack spacing={2}>
                    <Icon as={FiUpload} boxSize={8} color="gray.400" />
                    <Text color="gray.600">
                      Drop a file here or click to browse
                    </Text>
                    <Text fontSize="sm" color="gray.500">
                      Supports GeoPackage, Shapefile, GeoTIFF, GeoJSON, KML
                    </Text>
                  </VStack>
                )}
              </Box>
            </FormControl>

            {/* Title */}
            <FormControl>
              <FormLabel fontWeight="500" color="gray.700">Title (optional)</FormLabel>
              <Input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Dataset title"
                size="lg"
                borderRadius="lg"
              />
            </FormControl>

            {/* Abstract */}
            <FormControl>
              <FormLabel fontWeight="500" color="gray.700">Abstract (optional)</FormLabel>
              <Textarea
                value={abstract}
                onChange={(e) => setAbstract(e.target.value)}
                placeholder="Brief description of the dataset"
                size="lg"
                borderRadius="lg"
                rows={3}
              />
            </FormControl>

            {/* GeoPackage multi-layer info */}
            {fileInfo?.type === 'GeoPackage' && (
              <Alert status="info" borderRadius="lg" variant="subtle">
                <AlertIcon />
                <Box>
                  <Text fontSize="sm" fontWeight="500">Multi-layer GeoPackage</Text>
                  <Text fontSize="xs" color="gray.600">
                    Each layer in the GeoPackage will be published as a separate dataset in GeoNode.
                  </Text>
                </Box>
              </Alert>
            )}

            {/* Upload Progress */}
            {isUploading && (
              <Box w="100%">
                <HStack justify="space-between" mb={2}>
                  <Text fontSize="sm" color="gray.600">Uploading...</Text>
                  <Text fontSize="sm" color="gray.600">{uploadProgress}%</Text>
                </HStack>
                <Progress
                  value={uploadProgress}
                  size="sm"
                  colorScheme="teal"
                  borderRadius="full"
                  hasStripe
                  isAnimated
                />
              </Box>
            )}

            {/* Upload Result */}
            <AnimatePresence>
              {uploadResult && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  style={{ width: '100%' }}
                >
                  <Alert
                    status={uploadResult.success ? 'success' : 'error'}
                    borderRadius="lg"
                    variant="subtle"
                  >
                    <AlertIcon as={uploadResult.success ? FiCheckCircle : FiAlertCircle} />
                    <Text fontSize="sm">{uploadResult.message}</Text>
                  </Alert>
                </motion.div>
              )}
            </AnimatePresence>
          </VStack>
        </ModalBody>

        <ModalFooter
          gap={3}
          borderTop="1px solid"
          borderTopColor="gray.100"
          bg="gray.50"
        >
          <Button variant="ghost" onClick={closeDialog} borderRadius="lg">
            {uploadResult?.success ? 'Close' : 'Cancel'}
          </Button>
          <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
            <Button
              colorScheme="teal"
              onClick={handleUpload}
              isLoading={isUploading}
              isDisabled={!selectedFile}
              borderRadius="lg"
              px={6}
              leftIcon={<FiUpload />}
            >
              Upload
            </Button>
          </motion.div>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
