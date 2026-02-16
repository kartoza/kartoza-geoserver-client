import { useState } from 'react'
import { Box, Flex, Text, Button, useToast, useColorModeValue } from '@chakra-ui/react'
import { useQueryClient } from '@tanstack/react-query'
import { FiUpload } from 'react-icons/fi'
import { useUIStore } from '../../../stores/uiStore'
import * as api from '../../../api/client'
import { DatasetRow } from '../DatasetRow'
import type { DataStoreContentsNodeProps } from '../types'

export function DataStoreContentsNode({
  connectionId,
  workspace,
  storeName,
  featureTypes,
  availableFeatureTypes,
}: DataStoreContentsNodeProps) {
  const toast = useToast()
  const queryClient = useQueryClient()
  const [selectedForPublish, setSelectedForPublish] = useState<Set<string>>(new Set())
  const [isPublishing, setIsPublishing] = useState(false)
  const setPreview = useUIStore((state) => state.setPreview)

  const toggleSelection = (name: string) => {
    const newSelection = new Set(selectedForPublish)
    if (newSelection.has(name)) {
      newSelection.delete(name)
    } else {
      newSelection.add(name)
    }
    setSelectedForPublish(newSelection)
  }

  const selectAll = () => {
    setSelectedForPublish(new Set(availableFeatureTypes))
  }

  const handlePublishSelected = async () => {
    if (selectedForPublish.size === 0) return

    setIsPublishing(true)
    try {
      const result = await api.publishFeatureTypes(
        connectionId,
        workspace,
        storeName,
        Array.from(selectedForPublish)
      )

      if (result.published.length > 0) {
        toast({
          title: 'Layers Published',
          description: `Successfully published ${result.published.length} layer(s)`,
          status: 'success',
          duration: 3000,
        })
        // Refresh queries
        queryClient.invalidateQueries({ queryKey: ['featuretypes', connectionId, workspace, storeName] })
        queryClient.invalidateQueries({ queryKey: ['available-featuretypes', connectionId, workspace, storeName] })
        queryClient.invalidateQueries({ queryKey: ['layers', connectionId, workspace] })
        setSelectedForPublish(new Set())
      }

      if (result.errors.length > 0) {
        toast({
          title: 'Some layers failed to publish',
          description: result.errors.join(', '),
          status: 'warning',
          duration: 5000,
        })
      }
    } catch (error) {
      toast({
        title: 'Failed to publish layers',
        description: error instanceof Error ? error.message : 'Unknown error',
        status: 'error',
        duration: 5000,
      })
    } finally {
      setIsPublishing(false)
    }
  }

  const handlePublishSingle = async (featureTypeName: string) => {
    try {
      await api.publishFeatureType(connectionId, workspace, storeName, featureTypeName)
      toast({
        title: 'Layer Published',
        description: `Successfully published ${featureTypeName}`,
        status: 'success',
        duration: 3000,
      })
      // Refresh queries
      queryClient.invalidateQueries({ queryKey: ['featuretypes', connectionId, workspace, storeName] })
      queryClient.invalidateQueries({ queryKey: ['available-featuretypes', connectionId, workspace, storeName] })
      queryClient.invalidateQueries({ queryKey: ['layers', connectionId, workspace] })
    } catch (error) {
      toast({
        title: 'Failed to publish layer',
        description: error instanceof Error ? error.message : 'Unknown error',
        status: 'error',
        duration: 5000,
      })
    }
  }

  const handlePreviewPublished = (featureTypeName: string) => {
    api.startPreview({
      connId: connectionId,
      workspace,
      layerName: featureTypeName,
      storeName,
      storeType: 'datastore',
      layerType: 'vector',
    }).then(({ url }) => {
      setPreview({
        url,
        layerName: featureTypeName,
        workspace,
        connectionId,
        storeName,
        storeType: 'datastore',
        layerType: 'vector',
      })
    }).catch((err) => {
      toast({
        title: 'Preview failed',
        description: err.message,
        status: 'error',
        duration: 3000,
      })
    })
  }

  const bgAvailable = useColorModeValue('yellow.50', 'yellow.900')
  const bgPublished = useColorModeValue('green.50', 'green.900')

  return (
    <Box>
      {/* Published feature types */}
      {featureTypes.length > 0 && (
        <Box mb={2}>
          <Text fontSize="xs" fontWeight="600" color="gray.500" px={2} py={1}>
            Published ({featureTypes.length})
          </Text>
          {featureTypes.map((ft) => (
            <DatasetRow
              key={ft.name}
              name={ft.name}
              isPublished
              bg={bgPublished}
              onPreview={() => handlePreviewPublished(ft.name)}
            />
          ))}
        </Box>
      )}

      {/* Unpublished feature types */}
      {availableFeatureTypes.length > 0 && (
        <Box>
          <Flex align="center" justify="space-between" px={2} py={1}>
            <Text fontSize="xs" fontWeight="600" color="gray.500">
              Available to Publish ({availableFeatureTypes.length})
            </Text>
            <Flex gap={1}>
              <Button
                size="xs"
                variant="ghost"
                onClick={selectAll}
                isDisabled={selectedForPublish.size === availableFeatureTypes.length}
              >
                Select All
              </Button>
              {selectedForPublish.size > 0 && (
                <Button
                  size="xs"
                  colorScheme="kartoza"
                  leftIcon={<FiUpload size={12} />}
                  onClick={handlePublishSelected}
                  isLoading={isPublishing}
                >
                  Publish ({selectedForPublish.size})
                </Button>
              )}
            </Flex>
          </Flex>
          {availableFeatureTypes.map((ftName) => (
            <DatasetRow
              key={ftName}
              name={ftName}
              isPublished={false}
              bg={bgAvailable}
              isSelected={selectedForPublish.has(ftName)}
              onToggleSelect={() => toggleSelection(ftName)}
              onPublish={() => handlePublishSingle(ftName)}
            />
          ))}
        </Box>
      )}

      {featureTypes.length === 0 && availableFeatureTypes.length === 0 && (
        <Text fontSize="xs" color="gray.500" px={2} py={2} fontStyle="italic">
          No datasets in this store
        </Text>
      )}
    </Box>
  )
}
