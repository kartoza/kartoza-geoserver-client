import { useState, useEffect } from 'react'
import {
  Box,
  VStack,
  HStack,
  Heading,
  Text,
  Card,
  CardBody,
  CardHeader,
  Button,
  Badge,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Alert,
  AlertIcon,
  Spinner,
  useColorModeValue,
  Divider,
  Icon,
  IconButton,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Code,
  useClipboard,
  Tooltip,
  Progress,
} from '@chakra-ui/react'
import {
  FaServer,
  FaDatabase,
  FaGlobe,
  FaExternalLinkAlt,
  FaKey,
  FaRedo,
  FaTrash,
  FaCopy,
  FaCheck,
  FaCircle,
} from 'react-icons/fa'
import * as hostingApi from '../../api/hosting'
import type { Instance, InstanceCredentials, Activity } from '../../api/hosting'

interface InstancePanelProps {
  instanceId: string
}

export default function InstancePanel({ instanceId }: InstancePanelProps) {
  const [instance, setInstance] = useState<Instance | null>(null)
  const [credentials, setCredentials] = useState<InstanceCredentials | null>(null)
  const [activities, setActivities] = useState<Activity[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const { isOpen: isCredsOpen, onOpen: onCredsOpen, onClose: onCredsClose } = useDisclosure()
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure()

  const cardBg = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')

  useEffect(() => {
    loadInstance()
    loadActivities()
  }, [instanceId])

  const loadInstance = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await hostingApi.getInstance(instanceId)
      setInstance(data)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setIsLoading(false)
    }
  }

  const loadActivities = async () => {
    try {
      const data = await hostingApi.listInstanceActivities(instanceId)
      setActivities(data.activities)
    } catch (err) {
      // Ignore activity loading errors
    }
  }

  const loadCredentials = async () => {
    try {
      const creds = await hostingApi.getInstanceCredentials(instanceId)
      setCredentials(creds)
      onCredsOpen()
    } catch (err) {
      setError((err as Error).message)
    }
  }

  const handleRestart = async () => {
    setActionLoading('restart')
    try {
      await hostingApi.restartInstance(instanceId)
      await loadInstance()
      await loadActivities()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setActionLoading(null)
    }
  }

  const handleDelete = async () => {
    setActionLoading('delete')
    try {
      await hostingApi.deleteInstance(instanceId)
      onDeleteClose()
      await loadInstance()
      await loadActivities()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setActionLoading(null)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return 'green'
      case 'starting_up':
      case 'deploying':
        return 'yellow'
      case 'offline':
      case 'maintenance':
        return 'gray'
      case 'error':
        return 'red'
      case 'deleting':
      case 'deleted':
        return 'gray'
      default:
        return 'gray'
    }
  }

  const getHealthColor = (health: string) => {
    switch (health) {
      case 'healthy':
        return 'green'
      case 'degraded':
        return 'yellow'
      case 'unhealthy':
        return 'red'
      default:
        return 'gray'
    }
  }

  const getProductIcon = (productId: string) => {
    switch (productId) {
      case 'geoserver':
        return FaServer
      case 'geonode':
        return FaGlobe
      case 'postgis':
        return FaDatabase
      default:
        return FaServer
    }
  }

  if (isLoading) {
    return (
      <Box p={6} textAlign="center">
        <Spinner size="lg" />
      </Box>
    )
  }

  if (error) {
    return (
      <Box p={6}>
        <Alert status="error">
          <AlertIcon />
          {error}
        </Alert>
      </Box>
    )
  }

  if (!instance) {
    return (
      <Box p={6}>
        <Alert status="warning">
          <AlertIcon />
          Instance not found
        </Alert>
      </Box>
    )
  }

  return (
    <Box p={6} h="full" overflowY="auto">
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <HStack justify="space-between" flexWrap="wrap">
          <HStack spacing={4}>
            <Icon
              as={getProductIcon(instance.product_id)}
              boxSize={10}
              color="blue.500"
            />
            <VStack align="start" spacing={0}>
              <Heading size="md">{instance.display_name || instance.name}</Heading>
              <HStack>
                <Badge colorScheme={getStatusColor(instance.status)}>
                  <HStack spacing={1}>
                    <Icon as={FaCircle} boxSize={2} />
                    <Text>{instance.status}</Text>
                  </HStack>
                </Badge>
                {instance.health_status !== 'unknown' && (
                  <Badge colorScheme={getHealthColor(instance.health_status)}>
                    {instance.health_status}
                  </Badge>
                )}
              </HStack>
            </VStack>
          </HStack>

          <HStack spacing={2}>
            {instance.url && instance.status === 'online' && (
              <Button
                as="a"
                href={instance.url}
                target="_blank"
                rel="noopener noreferrer"
                leftIcon={<FaExternalLinkAlt />}
                colorScheme="blue"
                size="sm"
              >
                Open
              </Button>
            )}
            {instance.status === 'online' && (
              <Button
                leftIcon={<FaKey />}
                onClick={loadCredentials}
                size="sm"
              >
                Credentials
              </Button>
            )}
            {['online', 'offline', 'error'].includes(instance.status) && (
              <Button
                leftIcon={<FaRedo />}
                onClick={handleRestart}
                isLoading={actionLoading === 'restart'}
                size="sm"
              >
                Restart
              </Button>
            )}
            {!['deleting', 'deleted'].includes(instance.status) && (
              <IconButton
                aria-label="Delete instance"
                icon={<FaTrash />}
                colorScheme="red"
                variant="ghost"
                onClick={onDeleteOpen}
                size="sm"
              />
            )}
          </HStack>
        </HStack>

        <Tabs>
          <TabList>
            <Tab>Overview</Tab>
            <Tab>Activity</Tab>
            <Tab>Resources</Tab>
          </TabList>

          <TabPanels>
            {/* Overview Tab */}
            <TabPanel px={0}>
              <VStack spacing={4} align="stretch">
                {/* Instance Details */}
                <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
                  <CardHeader>
                    <Heading size="sm">Instance Details</Heading>
                  </CardHeader>
                  <CardBody pt={0}>
                    <Table variant="simple" size="sm">
                      <Tbody>
                        <Tr>
                          <Td fontWeight="bold" w="200px">Product</Td>
                          <Td>{instance.product?.name || instance.product_id}</Td>
                        </Tr>
                        <Tr>
                          <Td fontWeight="bold">Package</Td>
                          <Td>{instance.package?.name || instance.package_id}</Td>
                        </Tr>
                        <Tr>
                          <Td fontWeight="bold">Cluster</Td>
                          <Td>{instance.cluster?.name || instance.cluster_id}</Td>
                        </Tr>
                        <Tr>
                          <Td fontWeight="bold">URL</Td>
                          <Td>
                            {instance.url ? (
                              <a href={instance.url} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--chakra-colors-blue-500)' }}>
                                {instance.url}
                              </a>
                            ) : (
                              'Not available yet'
                            )}
                          </Td>
                        </Tr>
                        <Tr>
                          <Td fontWeight="bold">Created</Td>
                          <Td>{hostingApi.formatDateTime(instance.created_at)}</Td>
                        </Tr>
                        {instance.last_health_check && (
                          <Tr>
                            <Td fontWeight="bold">Last Health Check</Td>
                            <Td>{hostingApi.formatDateTime(instance.last_health_check)}</Td>
                          </Tr>
                        )}
                      </Tbody>
                    </Table>
                  </CardBody>
                </Card>

                {/* Package Features */}
                {instance.package?.features && instance.package.features.length > 0 && (
                  <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
                    <CardHeader>
                      <Heading size="sm">Package Features</Heading>
                    </CardHeader>
                    <CardBody pt={0}>
                      <VStack align="start" spacing={2}>
                        {instance.package.features.map((feature, idx) => (
                          <HStack key={idx}>
                            <Icon as={FaCheck} color="green.500" />
                            <Text>{feature}</Text>
                          </HStack>
                        ))}
                        {instance.package.cpu_limit && (
                          <HStack>
                            <Icon as={FaCheck} color="green.500" />
                            <Text>{instance.package.cpu_limit} CPU</Text>
                          </HStack>
                        )}
                        {instance.package.memory_limit && (
                          <HStack>
                            <Icon as={FaCheck} color="green.500" />
                            <Text>{instance.package.memory_limit} Memory</Text>
                          </HStack>
                        )}
                        {instance.package.storage_limit && (
                          <HStack>
                            <Icon as={FaCheck} color="green.500" />
                            <Text>{instance.package.storage_limit} Storage</Text>
                          </HStack>
                        )}
                      </VStack>
                    </CardBody>
                  </Card>
                )}
              </VStack>
            </TabPanel>

            {/* Activity Tab */}
            <TabPanel px={0}>
              <VStack spacing={4} align="stretch">
                {activities.length === 0 ? (
                  <Text color="gray.500">No recent activity</Text>
                ) : (
                  <Table variant="simple" size="sm">
                    <Thead>
                      <Tr>
                        <Th>Type</Th>
                        <Th>Status</Th>
                        <Th>Started</Th>
                        <Th>Duration</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {activities.map((activity) => (
                        <Tr key={activity.id}>
                          <Td>
                            <Badge>{activity.activity_type}</Badge>
                          </Td>
                          <Td>
                            <Badge
                              colorScheme={
                                activity.status === 'success' ? 'green' :
                                activity.status === 'error' ? 'red' :
                                activity.status === 'running' ? 'yellow' :
                                'gray'
                              }
                            >
                              {activity.status}
                            </Badge>
                          </Td>
                          <Td>
                            {activity.started_at
                              ? hostingApi.formatDateTime(activity.started_at)
                              : hostingApi.formatDateTime(activity.created_at)}
                          </Td>
                          <Td>
                            {activity.started_at && activity.completed_at && (
                              (() => {
                                const duration = new Date(activity.completed_at).getTime() - new Date(activity.started_at).getTime()
                                const seconds = Math.floor(duration / 1000)
                                const minutes = Math.floor(seconds / 60)
                                if (minutes > 0) {
                                  return `${minutes}m ${seconds % 60}s`
                                }
                                return `${seconds}s`
                              })()
                            )}
                          </Td>
                        </Tr>
                      ))}
                    </Tbody>
                  </Table>
                )}
              </VStack>
            </TabPanel>

            {/* Resources Tab */}
            <TabPanel px={0}>
              <VStack spacing={4} align="stretch">
                {instance.cpu_usage !== undefined && (
                  <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
                    <CardBody>
                      <VStack align="stretch" spacing={2}>
                        <HStack justify="space-between">
                          <Text fontWeight="bold">CPU Usage</Text>
                          <Text>{Math.round(instance.cpu_usage * 100)}%</Text>
                        </HStack>
                        <Progress
                          value={instance.cpu_usage * 100}
                          colorScheme={instance.cpu_usage > 0.8 ? 'red' : instance.cpu_usage > 0.6 ? 'yellow' : 'green'}
                          size="sm"
                          borderRadius="full"
                        />
                      </VStack>
                    </CardBody>
                  </Card>
                )}

                {instance.memory_usage !== undefined && (
                  <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
                    <CardBody>
                      <VStack align="stretch" spacing={2}>
                        <HStack justify="space-between">
                          <Text fontWeight="bold">Memory Usage</Text>
                          <Text>{Math.round(instance.memory_usage * 100)}%</Text>
                        </HStack>
                        <Progress
                          value={instance.memory_usage * 100}
                          colorScheme={instance.memory_usage > 0.8 ? 'red' : instance.memory_usage > 0.6 ? 'yellow' : 'green'}
                          size="sm"
                          borderRadius="full"
                        />
                      </VStack>
                    </CardBody>
                  </Card>
                )}

                {instance.storage_usage !== undefined && (
                  <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
                    <CardBody>
                      <VStack align="stretch" spacing={2}>
                        <HStack justify="space-between">
                          <Text fontWeight="bold">Storage Usage</Text>
                          <Text>{Math.round(instance.storage_usage * 100)}%</Text>
                        </HStack>
                        <Progress
                          value={instance.storage_usage * 100}
                          colorScheme={instance.storage_usage > 0.8 ? 'red' : instance.storage_usage > 0.6 ? 'yellow' : 'green'}
                          size="sm"
                          borderRadius="full"
                        />
                      </VStack>
                    </CardBody>
                  </Card>
                )}

                {instance.cpu_usage === undefined && instance.memory_usage === undefined && (
                  <Text color="gray.500">Resource metrics not available yet</Text>
                )}
              </VStack>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>

      {/* Credentials Modal */}
      <CredentialsModal
        isOpen={isCredsOpen}
        onClose={onCredsClose}
        credentials={credentials}
      />

      {/* Delete Confirmation Modal */}
      <Modal isOpen={isDeleteOpen} onClose={onDeleteClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Delete Instance</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Alert status="warning" mb={4}>
              <AlertIcon />
              This action cannot be undone.
            </Alert>
            <Text>
              Are you sure you want to delete <strong>{instance.name}</strong>?
              All data will be permanently removed.
            </Text>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onDeleteClose}>
              Cancel
            </Button>
            <Button
              colorScheme="red"
              onClick={handleDelete}
              isLoading={actionLoading === 'delete'}
            >
              Delete
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  )
}

// Credentials Modal Component
function CredentialsModal({
  isOpen,
  onClose,
  credentials,
}: {
  isOpen: boolean
  onClose: () => void
  credentials: InstanceCredentials | null
}) {
  if (!credentials) return null

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Instance Credentials</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <VStack spacing={4} align="stretch">
            <Alert status="warning">
              <AlertIcon />
              Keep these credentials secure. They provide full access to your instance.
            </Alert>

            <CopyableField label="URL" value={credentials.url} />
            <CopyableField label="Username" value={credentials.admin_username} />
            <CopyableField label="Password" value={credentials.admin_password} isPassword />

            {credentials.database_host && (
              <>
                <Divider />
                <Heading size="xs">Database Credentials</Heading>
                <CopyableField label="Host" value={credentials.database_host} />
                {credentials.database_port && (
                  <CopyableField label="Port" value={String(credentials.database_port)} />
                )}
                {credentials.database_name && (
                  <CopyableField label="Database" value={credentials.database_name} />
                )}
                {credentials.database_user && (
                  <CopyableField label="User" value={credentials.database_user} />
                )}
                {credentials.database_pass && (
                  <CopyableField label="Password" value={credentials.database_pass} isPassword />
                )}
              </>
            )}
          </VStack>
        </ModalBody>
        <ModalFooter>
          <Button colorScheme="blue" onClick={onClose}>
            Close
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}

// Copyable Field Component
function CopyableField({
  label,
  value,
  isPassword = false,
}: {
  label: string
  value: string
  isPassword?: boolean
}) {
  const { hasCopied, onCopy } = useClipboard(value)
  const [showPassword, setShowPassword] = useState(false)

  return (
    <HStack justify="space-between">
      <Text fontWeight="bold" minW="100px">{label}</Text>
      <HStack flex={1}>
        <Code p={2} flex={1} fontSize="sm">
          {isPassword && !showPassword ? '••••••••' : value}
        </Code>
        {isPassword && (
          <IconButton
            aria-label={showPassword ? 'Hide' : 'Show'}
            icon={showPassword ? <FaCheck /> : <FaKey />}
            size="sm"
            variant="ghost"
            onClick={() => setShowPassword(!showPassword)}
          />
        )}
        <Tooltip label={hasCopied ? 'Copied!' : 'Copy'}>
          <IconButton
            aria-label="Copy"
            icon={hasCopied ? <FaCheck /> : <FaCopy />}
            size="sm"
            variant="ghost"
            onClick={onCopy}
          />
        </Tooltip>
      </HStack>
    </HStack>
  )
}
