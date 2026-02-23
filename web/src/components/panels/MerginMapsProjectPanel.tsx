import {
  Box,
  VStack,
  HStack,
  Text,
  Icon,
  Badge,
  Divider,
  Skeleton,
  useColorModeValue,
} from '@chakra-ui/react'
import { FiFolder, FiCalendar, FiDatabase } from 'react-icons/fi'
import { useQuery } from '@tanstack/react-query'
import type { TreeNode } from '../../types'
import * as api from '../../api/client'

interface MerginMapsProjectPanelProps {
  node: TreeNode
}

export function MerginMapsProjectPanel({ node }: MerginMapsProjectPanelProps) {
  const bgColor = useColorModeValue('white', 'gray.800')
  const cardBg = useColorModeValue('gray.50', 'gray.700')
  const headingColor = useColorModeValue('gray.800', 'white')
  const labelColor = useColorModeValue('gray.600', 'gray.300')
  const sectionLabelColor = useColorModeValue('gray.600', 'gray.300')

  const namespace = node.merginMapsNamespace
  const projectName = node.merginMapsProjectName
  const connectionId = node.merginMapsConnectionId

  // Fetch projects to find the selected project's metadata
  const { data: projectsData, isLoading } = useQuery({
    queryKey: ['merginmapsprojects', connectionId],
    queryFn: () => api.getMerginMapsProjects(connectionId!),
    enabled: !!connectionId,
    staleTime: 30000,
  })

  const project = projectsData?.projects.find(
    (p) => p.namespace === namespace && p.name === projectName
  )

  const formatBytes = (bytes?: number) => {
    if (!bytes) return '—'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  return (
    <Box p={6} bg={bgColor} h="100%" overflowY="auto">
      <VStack align="stretch" spacing={6}>
        {/* Header */}
        <HStack spacing={3}>
          <Box bg="green.100" p={3} borderRadius="xl">
            <Icon as={FiFolder} boxSize={6} color="green.600" />
          </Box>
          <Box>
            <Text fontSize="xl" fontWeight="700" color={headingColor}>
              {projectName || '—'}
            </Text>
            <Text fontSize="sm" color={labelColor}>
              {namespace}
            </Text>
          </Box>
        </HStack>

        <Divider />

        {/* Project Details */}
        <Box bg={cardBg} borderRadius="xl" p={4}>
          <Text fontWeight="600" fontSize="sm" color={sectionLabelColor} mb={3} textTransform="uppercase" letterSpacing="wider">
            Project Details
          </Text>
          {isLoading ? (
            <VStack align="stretch" spacing={2}>
              <Skeleton h="20px" borderRadius="md" />
              <Skeleton h="20px" borderRadius="md" />
              <Skeleton h="20px" borderRadius="md" />
            </VStack>
          ) : (
            <VStack align="stretch" spacing={3}>
              <HStack>
                <Icon as={FiFolder} color="gray.400" boxSize={4} />
                <Text fontSize="sm" color={labelColor}>Full path:</Text>
                <Text fontSize="sm" fontWeight="500">{namespace}/{projectName}</Text>
              </HStack>

              {project?.version && (
                <HStack>
                  <Icon as={FiDatabase} color="gray.400" boxSize={4} />
                  <Text fontSize="sm" color={labelColor}>Version:</Text>
                  <Badge colorScheme="green" variant="subtle">{project.version}</Badge>
                </HStack>
              )}

              {project?.disk_usage !== undefined && (
                <HStack>
                  <Icon as={FiDatabase} color="gray.400" boxSize={4} />
                  <Text fontSize="sm" color={labelColor}>Size:</Text>
                  <Text fontSize="sm" fontWeight="500">{formatBytes(project.disk_usage)}</Text>
                </HStack>
              )}

              {project?.created && (
                <HStack>
                  <Icon as={FiCalendar} color="gray.400" boxSize={4} />
                  <Text fontSize="sm" color={labelColor}>Created:</Text>
                  <Text fontSize="sm" fontWeight="500">
                    {new Date(project.created).toLocaleDateString()}
                  </Text>
                </HStack>
              )}

              {project?.public !== undefined && (
                <HStack>
                  <Text fontSize="sm" color={labelColor}>Visibility:</Text>
                  <Badge colorScheme={project.public ? 'blue' : 'orange'} variant="subtle">
                    {project.public ? 'Public' : 'Private'}
                  </Badge>
                </HStack>
              )}

              {project?.tags && project.tags.length > 0 && (
                <HStack flexWrap="wrap">
                  <Text fontSize="sm" color={labelColor}>Tags:</Text>
                  {project.tags.map((tag) => (
                    <Badge key={tag} colorScheme="gray" variant="subtle" fontSize="xs">
                      {tag}
                    </Badge>
                  ))}
                </HStack>
              )}
            </VStack>
          )}
        </Box>
      </VStack>
    </Box>
  )
}
