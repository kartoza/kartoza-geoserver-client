import {
  Box,
  Flex,
  VStack,
  HStack,
  Text,
  Icon,
  Badge,
  Progress,
  Spinner,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Tooltip,
  IconButton,
} from '@chakra-ui/react'
import { keyframes, css } from '@emotion/react'
import { useQuery } from '@tanstack/react-query'
import {
  FiServer,
  FiLayers,
  FiDatabase,
  FiCheckCircle,
  FiAlertTriangle,
  FiXCircle,
  FiRefreshCw,
  FiClock,
  FiHardDrive,
  FiActivity,
} from 'react-icons/fi'
import * as api from '../api/client'
import type { ServerStatus } from '../types'

// Keyframe animations
const pulseKeyframes = keyframes`
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
`

const spinKeyframes = keyframes`
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
`

// Format bytes to human readable
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

interface ServerCardProps {
  server: ServerStatus
  isAlert?: boolean
}

function ServerCard({ server, isAlert = false }: ServerCardProps) {
  const borderColor = server.online ? 'green.400' : 'red.400'
  const bgColor = isAlert ? 'red.50' : server.online ? 'white' : 'gray.50'
  const statusIcon = server.online ? FiCheckCircle : FiXCircle
  const statusColor = server.online ? 'green.500' : 'red.500'

  return (
    <Box
      bg={bgColor}
      borderRadius="xl"
      border="2px solid"
      borderColor={borderColor}
      p={4}
      position="relative"
      transition="all 0.3s ease"
      _hover={{
        transform: 'translateY(-2px)',
        boxShadow: 'lg',
      }}
      css={!server.online ? css`animation: ${pulseKeyframes} 2s ease-in-out infinite;` : undefined}
    >
      {/* Status indicator */}
      <Box
        position="absolute"
        top={3}
        right={3}
      >
        <Tooltip label={server.online ? 'Online' : server.error || 'Offline'}>
          <span>
            <Icon
              as={statusIcon}
              color={statusColor}
              boxSize={5}
            />
          </span>
        </Tooltip>
      </Box>

      {/* Server info */}
      <VStack align="start" spacing={3}>
        <HStack>
          <Icon as={FiServer} color="kartoza.500" boxSize={6} />
          <VStack align="start" spacing={0}>
            <Text fontWeight="bold" fontSize="lg" noOfLines={1}>
              {server.connectionName}
            </Text>
            <Text fontSize="xs" color="gray.500" noOfLines={1}>
              {server.url}
            </Text>
          </VStack>
        </HStack>

        {server.online ? (
          <>
            {/* Version badge */}
            {server.geoserverVersion && (
              <Badge colorScheme="blue" fontSize="xs">
                GeoServer {server.geoserverVersion}
              </Badge>
            )}

            {/* Stats grid */}
            <SimpleGrid columns={2} spacing={2} w="100%">
              <Stat size="sm">
                <StatLabel fontSize="xs" color="gray.500">
                  <HStack spacing={1}>
                    <Icon as={FiLayers} boxSize={3} />
                    <Text>Layers</Text>
                  </HStack>
                </StatLabel>
                <StatNumber fontSize="xl" color="purple.600">
                  {server.layerCount}
                </StatNumber>
              </Stat>

              <Stat size="sm">
                <StatLabel fontSize="xs" color="gray.500">
                  <HStack spacing={1}>
                    <Icon as={FiDatabase} boxSize={3} />
                    <Text>Stores</Text>
                  </HStack>
                </StatLabel>
                <StatNumber fontSize="xl" color="blue.600">
                  {server.dataStoreCount + server.coverageCount}
                </StatNumber>
              </Stat>
            </SimpleGrid>

            {/* Memory usage */}
            {server.memoryTotal > 0 && (
              <Box w="100%">
                <HStack justify="space-between" mb={1}>
                  <HStack spacing={1}>
                    <Icon as={FiHardDrive} boxSize={3} color="gray.500" />
                    <Text fontSize="xs" color="gray.500">Memory</Text>
                  </HStack>
                  <Text fontSize="xs" color="gray.600">
                    {formatBytes(server.memoryUsed)} / {formatBytes(server.memoryTotal)}
                  </Text>
                </HStack>
                <Progress
                  value={server.memoryUsedPct}
                  size="sm"
                  colorScheme={server.memoryUsedPct > 80 ? 'red' : server.memoryUsedPct > 60 ? 'yellow' : 'green'}
                  borderRadius="full"
                />
              </Box>
            )}

            {/* Response time */}
            <HStack spacing={2} fontSize="xs" color="gray.500">
              <Icon as={FiClock} boxSize={3} />
              <Text>Response: {server.responseTimeMs}ms</Text>
            </HStack>
          </>
        ) : (
          <Box w="100%">
            <Alert status="error" variant="subtle" borderRadius="md" py={2}>
              <AlertIcon boxSize={4} />
              <Text fontSize="sm" noOfLines={2}>
                {server.error || 'Server is offline'}
              </Text>
            </Alert>
          </Box>
        )}
      </VStack>
    </Box>
  )
}

export default function Dashboard() {
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ['dashboard'],
    queryFn: api.getDashboard,
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  if (isLoading) {
    return (
      <Flex h="100%" align="center" justify="center">
        <VStack spacing={4}>
          <Spinner size="xl" color="kartoza.500" thickness="4px" />
          <Text color="gray.500">Loading server status...</Text>
        </VStack>
      </Flex>
    )
  }

  if (error) {
    return (
      <Flex h="100%" align="center" justify="center" p={8}>
        <Alert status="error" borderRadius="lg">
          <AlertIcon />
          <AlertTitle>Failed to load dashboard</AlertTitle>
          <AlertDescription>{(error as Error).message}</AlertDescription>
        </Alert>
      </Flex>
    )
  }

  if (!data || data.servers.length === 0) {
    return (
      <Flex h="100%" align="center" justify="center" p={8}>
        <VStack spacing={4} textAlign="center">
          <Icon as={FiServer} boxSize={16} color="gray.300" />
          <Text color="gray.500" fontSize="lg">No servers configured</Text>
          <Text color="gray.400" fontSize="sm">
            Add a GeoServer connection to get started
          </Text>
        </VStack>
      </Flex>
    )
  }

  // Separate alert servers from healthy ones
  const healthyServers = data.servers.filter(s => s.online)
  const alertServers = data.servers.filter(s => !s.online)

  return (
    <Box h="100%" overflowY="auto" p={6}>
      <VStack spacing={6} align="stretch">
        {/* Header with refresh button */}
        <Flex justify="space-between" align="center">
          <HStack spacing={3}>
            <Icon as={FiActivity} boxSize={6} color="kartoza.500" />
            <Text fontSize="xl" fontWeight="bold">Server Dashboard</Text>
          </HStack>
          <HStack spacing={4}>
            {/* Summary stats */}
            <HStack spacing={6} display={{ base: 'none', md: 'flex' }}>
              <HStack spacing={2}>
                <Icon as={FiCheckCircle} color="green.500" />
                <Text fontWeight="bold" color="green.600">{data.onlineCount}</Text>
                <Text color="gray.500">Online</Text>
              </HStack>
              {data.offlineCount > 0 && (
                <HStack spacing={2}>
                  <Icon as={FiXCircle} color="red.500" />
                  <Text fontWeight="bold" color="red.600">{data.offlineCount}</Text>
                  <Text color="gray.500">Offline</Text>
                </HStack>
              )}
              <HStack spacing={2}>
                <Icon as={FiLayers} color="purple.500" />
                <Text fontWeight="bold" color="purple.600">{data.totalLayers}</Text>
                <Text color="gray.500">Layers</Text>
              </HStack>
            </HStack>
            <Tooltip label="Refresh status">
              <IconButton
                aria-label="Refresh"
                icon={<FiRefreshCw />}
                variant="ghost"
                colorScheme="kartoza"
                onClick={() => refetch()}
                css={isFetching ? css`animation: ${spinKeyframes} 1s linear infinite;` : undefined}
              />
            </Tooltip>
          </HStack>
        </Flex>

        {/* Alert section for offline/error servers */}
        {alertServers.length > 0 && (
          <Box>
            <HStack spacing={2} mb={3}>
              <Icon as={FiAlertTriangle} color="red.500" />
              <Text fontWeight="bold" color="red.600">
                Servers Requiring Attention ({alertServers.length})
              </Text>
            </HStack>
            <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={4}>
              {alertServers.map(server => (
                <ServerCard key={server.connectionId} server={server} isAlert />
              ))}
            </SimpleGrid>
          </Box>
        )}

        {/* Healthy servers section */}
        {healthyServers.length > 0 && (
          <Box>
            {alertServers.length > 0 && (
              <HStack spacing={2} mb={3}>
                <Icon as={FiCheckCircle} color="green.500" />
                <Text fontWeight="bold" color="green.600">
                  Online Servers ({healthyServers.length})
                </Text>
              </HStack>
            )}
            <SimpleGrid columns={{ base: 1, md: 2, lg: 3, xl: 4 }} spacing={4}>
              {healthyServers.map(server => (
                <ServerCard key={server.connectionId} server={server} />
              ))}
            </SimpleGrid>
          </Box>
        )}
      </VStack>
    </Box>
  )
}
