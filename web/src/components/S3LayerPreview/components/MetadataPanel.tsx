import {
  SimpleGrid,
  Text,
  Badge,
  useColorModeValue,
} from '@chakra-ui/react'
import type { S3PreviewMetadata } from '../../../types'
import { formatSize, getFormatBadgeColor, getPreviewTypeBadgeColor } from '../utils/formatters'

interface MetadataPanelProps {
  metadata: S3PreviewMetadata
}

export default function MetadataPanel({ metadata }: MetadataPanelProps) {
  const metaBg = useColorModeValue('gray.50', 'gray.700')

  return (
    <SimpleGrid columns={2} spacing={2} p={3} bg={metaBg} borderRadius="md" fontSize="sm">
      <Text fontWeight="medium">Format:</Text>
      <Badge colorScheme={getFormatBadgeColor(metadata.format)}>{metadata.format.toUpperCase()}</Badge>

      <Text fontWeight="medium">Type:</Text>
      <Badge colorScheme={getPreviewTypeBadgeColor(metadata.previewType)}>{metadata.previewType}</Badge>

      <Text fontWeight="medium">Size:</Text>
      <Text>{formatSize(metadata.size)}</Text>

      {metadata.crs && (
        <>
          <Text fontWeight="medium">CRS:</Text>
          <Text>{metadata.crs}</Text>
        </>
      )}

      {metadata.bounds && (
        <>
          <Text fontWeight="medium">Bounds:</Text>
          <Text fontSize="xs">
            {metadata.bounds.minX.toFixed(4)}, {metadata.bounds.minY.toFixed(4)} to{' '}
            {metadata.bounds.maxX.toFixed(4)}, {metadata.bounds.maxY.toFixed(4)}
          </Text>
        </>
      )}

      {metadata.bandCount && (
        <>
          <Text fontWeight="medium">Bands:</Text>
          <Text>{metadata.bandCount}</Text>
        </>
      )}

      {metadata.featureCount && (
        <>
          <Text fontWeight="medium">Features:</Text>
          <Text>{metadata.featureCount.toLocaleString()}</Text>
        </>
      )}

      {metadata.fieldNames && metadata.fieldNames.length > 0 && (
        <>
          <Text fontWeight="medium">Fields:</Text>
          <Text fontSize="xs" noOfLines={2}>{metadata.fieldNames.join(', ')}</Text>
        </>
      )}
    </SimpleGrid>
  )
}
