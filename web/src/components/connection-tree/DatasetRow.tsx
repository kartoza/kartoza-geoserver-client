import {
  Box,
  Flex,
  Text,
  IconButton,
  Icon,
  Tooltip,
  Badge,
  useColorModeValue,
  Checkbox,
} from '@chakra-ui/react'
import {
  FiEye,
  FiPlus,
} from 'react-icons/fi'
import { getNodeIconComponent, getNodeColor } from './utils'
import type { DatasetRowProps } from './types'

export function DatasetRow({
  name,
  isPublished,
  isCoverage = false,
  bg,
  isSelected,
  onToggleSelect,
  onPublish,
  onPreview,
}: DatasetRowProps) {
  const hoverBg = useColorModeValue('gray.100', 'gray.600')
  const iconType = isCoverage ? 'coverage' : 'featuretype'
  const NodeIcon = getNodeIconComponent(iconType)
  const nodeColor = getNodeColor(iconType)

  return (
    <Flex
      align="center"
      py={1.5}
      px={2}
      pl={6}
      bg={bg}
      _hover={{ bg: hoverBg }}
      borderRadius="md"
      mx={1}
      my={0.5}
      role="group"
    >
      {!isPublished && onToggleSelect && (
        <Checkbox
          size="sm"
          isChecked={isSelected}
          onChange={onToggleSelect}
          mr={2}
          colorScheme="kartoza"
        />
      )}
      <Box
        p={1}
        borderRadius="md"
        mr={2}
      >
        <Icon
          as={NodeIcon}
          boxSize={3.5}
          color={nodeColor}
        />
      </Box>
      <Text
        flex="1"
        fontSize="sm"
        noOfLines={1}
      >
        {name}
      </Text>
      {isPublished && (
        <Badge colorScheme="green" fontSize="2xs" mr={2}>
          Published
        </Badge>
      )}
      <Flex
        gap={1}
        opacity={0}
        _groupHover={{ opacity: 1 }}
        transition="opacity 0.15s"
      >
        {onPreview && (
          <Tooltip label="Preview" fontSize="xs">
            <IconButton
              aria-label="Preview"
              icon={<FiEye size={12} />}
              size="xs"
              variant="ghost"
              colorScheme="kartoza"
              onClick={(e) => {
                e.stopPropagation()
                onPreview()
              }}
            />
          </Tooltip>
        )}
        {!isPublished && onPublish && (
          <Tooltip label="Publish as Layer" fontSize="xs">
            <IconButton
              aria-label="Publish"
              icon={<FiPlus size={12} />}
              size="xs"
              variant="ghost"
              colorScheme="green"
              onClick={(e) => {
                e.stopPropagation()
                onPublish()
              }}
            />
          </Tooltip>
        )}
      </Flex>
    </Flex>
  )
}
