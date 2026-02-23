import { useState, useEffect } from 'react'
import {
  Box,
  VStack,
  HStack,
  Text,
  Icon,
  Badge,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Spinner,
  useColorModeValue,
  useToast,
  IconButton,
  Tooltip,
} from '@chakra-ui/react'
import {
  FiCloud,
  FiFolder,
  FiFile,
  FiPlay,
  FiUsers,
  FiRefreshCw,
  FiTrash2,
  FiDownload,
  FiUpload,
} from 'react-icons/fi'
import type { TreeNode, QFieldCloudProject, QFieldCloudFile, QFieldCloudJob, QFieldCloudCollaborator, QFieldCloudDelta } from '../../types'
import * as api from '../../api/client'
import { useUIStore } from '../../stores/uiStore'

interface QFieldCloudPanelProps {
  node: TreeNode
}

export function QFieldCloudPanel({ node }: QFieldCloudPanelProps) {
  const bgColor = useColorModeValue('white', 'gray.800')
  const cardBg = useColorModeValue('gray.50', 'gray.700')
  const borderColor = useColorModeValue('gray.200', 'gray.600')
  const toast = useToast()
  const openDialog = useUIStore((state) => state.openDialog)

  const [projects, setProjects] = useState<QFieldCloudProject[]>([])
  const [files, setFiles] = useState<QFieldCloudFile[]>([])
  const [jobs, setJobs] = useState<QFieldCloudJob[]>([])
  const [collaborators, setCollaborators] = useState<QFieldCloudCollaborator[]>([])
  const [deltas, setDeltas] = useState<QFieldCloudDelta[]>([])
  const [isLoading, setIsLoading] = useState(false)

  const connectionId = node.qfieldcloudConnectionId || ''
  const projectId = node.qfieldcloudProjectId || ''

  useEffect(() => {
    loadData()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [node.id, connectionId, projectId])

  const loadData = async () => {
    if (!connectionId) return
    setIsLoading(true)
    try {
      switch (node.type) {
        case 'qfieldcloudprojects':
          setProjects(await api.listQFieldCloudProjects(connectionId))
          break
        case 'qfieldcloudfiles':
          if (projectId) setFiles(await api.listQFieldCloudFiles(connectionId, projectId))
          break
        case 'qfieldcloudjobs':
          if (projectId) setJobs(await api.listQFieldCloudJobs(connectionId, projectId))
          break
        case 'qfieldcloudcollaborators':
          if (projectId) setCollaborators(await api.listQFieldCloudCollaborators(connectionId, projectId))
          break
        case 'qfieldclouddeltas':
          if (projectId) setDeltas(await api.listQFieldCloudDeltas(connectionId, projectId))
          break
      }
    } catch (err) {
      toast({
        title: 'Failed to load data',
        description: err instanceof Error ? err.message : 'Unknown error',
        status: 'error',
        duration: 5000,
        isClosable: true,
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleDeleteProject = async (proj: QFieldCloudProject) => {
    openDialog('confirm', {
      mode: 'delete',
      title: 'Delete Project',
      message: `Delete project "${proj.name}"? This cannot be undone.`,
      onConfirm: async () => {
        try {
          await api.deleteQFieldCloudProject(connectionId, proj.id)
          toast({ title: 'Project deleted', status: 'success', duration: 3000 })
          loadData()
        } catch (err) {
          toast({
            title: 'Failed to delete project',
            description: err instanceof Error ? err.message : 'Unknown error',
            status: 'error',
            duration: 5000,
          })
        }
      },
    })
  }

  const handleDeleteFile = async (filename: string) => {
    openDialog('confirm', {
      mode: 'delete',
      title: 'Delete File',
      message: `Delete file "${filename}"?`,
      onConfirm: async () => {
        try {
          await api.deleteQFieldCloudFile(connectionId, projectId, filename)
          toast({ title: 'File deleted', status: 'success', duration: 3000 })
          loadData()
        } catch (err) {
          toast({
            title: 'Failed to delete file',
            description: err instanceof Error ? err.message : 'Unknown error',
            status: 'error',
            duration: 5000,
          })
        }
      },
    })
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !projectId) return
    try {
      await api.uploadQFieldCloudFile(connectionId, projectId, file)
      toast({ title: 'File uploaded', status: 'success', duration: 3000 })
      loadData()
    } catch (err) {
      toast({
        title: 'Upload failed',
        description: err instanceof Error ? err.message : 'Unknown error',
        status: 'error',
        duration: 5000,
      })
    }
    // Reset file input
    e.target.value = ''
  }

  const handleTriggerPackage = async () => {
    if (!projectId) return
    try {
      await api.createQFieldCloudJob(connectionId, projectId, { type: 'package' })
      toast({ title: 'Package job triggered', status: 'success', duration: 3000 })
      if (node.type === 'qfieldcloudjobs') loadData()
    } catch (err) {
      toast({
        title: 'Failed to trigger job',
        description: err instanceof Error ? err.message : 'Unknown error',
        status: 'error',
        duration: 5000,
      })
    }
  }

  const getJobStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'finished': case 'success': return 'green'
      case 'failed': case 'error': return 'red'
      case 'started': case 'queued': return 'blue'
      default: return 'gray'
    }
  }

  const getDeltaStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'applied': return 'green'
      case 'conflict': return 'orange'
      case 'error': return 'red'
      case 'pending': return 'blue'
      default: return 'gray'
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  const renderHeader = (icon: React.ElementType, title: string) => (
    <HStack justify="space-between" mb={4}>
      <HStack>
        <Icon as={icon} color="blue.500" boxSize={5} />
        <Text fontWeight="bold" fontSize="lg">{title}</Text>
        {isLoading && <Spinner size="sm" />}
      </HStack>
      <Tooltip label="Refresh">
        <IconButton
          aria-label="Refresh"
          icon={<FiRefreshCw />}
          size="sm"
          variant="ghost"
          onClick={loadData}
          isLoading={isLoading}
        />
      </Tooltip>
    </HStack>
  )

  // ─── Connection node ────────────────────────────────────────────────────────
  if (node.type === 'qfieldcloudconnection') {
    return (
      <Box p={4} bg={bgColor} borderRadius="md">
        <HStack mb={4}>
          <Icon as={FiCloud} color="blue.500" boxSize={6} />
          <Text fontWeight="bold" fontSize="xl">{node.name}</Text>
        </HStack>
        <Text color="gray.500">
          Select "Projects" in the tree to browse and manage QFieldCloud projects.
        </Text>
      </Box>
    )
  }

  // ─── Projects list ───────────────────────────────────────────────────────────
  if (node.type === 'qfieldcloudprojects') {
    return (
      <Box p={4} bg={bgColor} borderRadius="md">
        {renderHeader(FiFolder, 'Projects')}
        {projects.length === 0 && !isLoading && (
          <Text color="gray.500">No projects found.</Text>
        )}
        <VStack spacing={3} align="stretch">
          {projects.map((proj) => (
            <Box key={proj.id} p={3} bg={cardBg} borderRadius="md" borderWidth={1} borderColor={borderColor}>
              <HStack justify="space-between">
                <VStack align="start" spacing={0}>
                  <HStack>
                    <Text fontWeight="semibold">{proj.name}</Text>
                    {proj.is_public && <Badge colorScheme="green" size="sm">Public</Badge>}
                    {proj.needs_repackaging && <Badge colorScheme="orange" size="sm">Needs repackaging</Badge>}
                  </HStack>
                  <Text fontSize="sm" color="gray.500">{proj.description || 'No description'}</Text>
                  <Text fontSize="xs" color="gray.400">
                    Owner: {proj.owner} &bull; {formatBytes(proj.file_storage_bytes)}
                  </Text>
                </VStack>
                <Tooltip label="Delete project">
                  <IconButton
                    aria-label="Delete project"
                    icon={<FiTrash2 />}
                    size="sm"
                    colorScheme="red"
                    variant="ghost"
                    onClick={() => handleDeleteProject(proj)}
                  />
                </Tooltip>
              </HStack>
            </Box>
          ))}
        </VStack>
      </Box>
    )
  }

  // ─── Single project ──────────────────────────────────────────────────────────
  if (node.type === 'qfieldcloudproject') {
    return (
      <Box p={4} bg={bgColor} borderRadius="md">
        <HStack mb={4}>
          <Icon as={FiFolder} color="blue.500" boxSize={6} />
          <Text fontWeight="bold" fontSize="xl">{node.name}</Text>
        </HStack>
        <Text color="gray.500" mb={2}>
          Select a sub-category in the tree (Files, Jobs, Collaborators, Deltas) to manage this project.
        </Text>
        <Button
          leftIcon={<FiPlay />}
          colorScheme="blue"
          size="sm"
          onClick={handleTriggerPackage}
        >
          Trigger Repackage
        </Button>
      </Box>
    )
  }

  // ─── Files ───────────────────────────────────────────────────────────────────
  if (node.type === 'qfieldcloudfiles') {
    return (
      <Box p={4} bg={bgColor} borderRadius="md">
        {renderHeader(FiFile, 'Project Files')}
        <HStack mb={4}>
          <Button
            as="label"
            leftIcon={<FiUpload />}
            colorScheme="blue"
            size="sm"
            cursor="pointer"
          >
            Upload File
            <input type="file" style={{ display: 'none' }} onChange={handleFileUpload} />
          </Button>
        </HStack>
        {files.length === 0 && !isLoading ? (
          <Text color="gray.500">No files in this project.</Text>
        ) : (
          <Table size="sm" variant="simple">
            <Thead>
              <Tr>
                <Th>Name</Th>
                <Th>Size</Th>
                <Th>Last Modified</Th>
                <Th>Actions</Th>
              </Tr>
            </Thead>
            <Tbody>
              {files.map((f) => (
                <Tr key={f.name}>
                  <Td>
                    <HStack>
                      <Icon as={FiFile} color="gray.400" />
                      <Text fontSize="sm">{f.name}</Text>
                      {f.is_packaging_file && <Badge colorScheme="purple" size="sm">packaging</Badge>}
                    </HStack>
                  </Td>
                  <Td fontSize="sm">{formatBytes(f.size)}</Td>
                  <Td fontSize="sm">{new Date(f.last_modified).toLocaleDateString()}</Td>
                  <Td>
                    <HStack spacing={1}>
                      <Tooltip label="Download">
                        <IconButton
                          aria-label="Download"
                          icon={<FiDownload />}
                          size="xs"
                          variant="ghost"
                          as="a"
                          href={api.getQFieldCloudFileDownloadUrl(connectionId, projectId, f.name)}
                        />
                      </Tooltip>
                      <Tooltip label="Delete">
                        <IconButton
                          aria-label="Delete"
                          icon={<FiTrash2 />}
                          size="xs"
                          colorScheme="red"
                          variant="ghost"
                          onClick={() => handleDeleteFile(f.name)}
                        />
                      </Tooltip>
                    </HStack>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Box>
    )
  }

  // ─── Jobs ─────────────────────────────────────────────────────────────────────
  if (node.type === 'qfieldcloudjobs') {
    return (
      <Box p={4} bg={bgColor} borderRadius="md">
        {renderHeader(FiPlay, 'Jobs')}
        <HStack mb={4}>
          <Button leftIcon={<FiPlay />} colorScheme="blue" size="sm" onClick={handleTriggerPackage}>
            Trigger Repackage
          </Button>
        </HStack>
        {jobs.length === 0 && !isLoading ? (
          <Text color="gray.500">No jobs for this project.</Text>
        ) : (
          <Table size="sm" variant="simple">
            <Thead>
              <Tr>
                <Th>Type</Th>
                <Th>Status</Th>
                <Th>Created</Th>
                <Th>Finished</Th>
              </Tr>
            </Thead>
            <Tbody>
              {jobs.map((job) => (
                <Tr key={job.id}>
                  <Td fontSize="sm">{job.type}</Td>
                  <Td>
                    <Badge colorScheme={getJobStatusColor(job.status)}>{job.status}</Badge>
                  </Td>
                  <Td fontSize="sm">{new Date(job.created_at).toLocaleString()}</Td>
                  <Td fontSize="sm">{job.finished_at ? new Date(job.finished_at).toLocaleString() : '—'}</Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Box>
    )
  }

  // ─── Collaborators ────────────────────────────────────────────────────────────
  if (node.type === 'qfieldcloudcollaborators') {
    return (
      <Box p={4} bg={bgColor} borderRadius="md">
        {renderHeader(FiUsers, 'Collaborators')}
        {collaborators.length === 0 && !isLoading ? (
          <Text color="gray.500">No collaborators for this project.</Text>
        ) : (
          <Table size="sm" variant="simple">
            <Thead>
              <Tr>
                <Th>Username</Th>
                <Th>Role</Th>
              </Tr>
            </Thead>
            <Tbody>
              {collaborators.map((c) => (
                <Tr key={c.collaborator}>
                  <Td fontSize="sm">{c.collaborator}</Td>
                  <Td>
                    <Badge colorScheme={c.role === 'admin' ? 'red' : c.role === 'manager' ? 'orange' : 'blue'}>
                      {c.role}
                    </Badge>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Box>
    )
  }

  // ─── Deltas ───────────────────────────────────────────────────────────────────
  if (node.type === 'qfieldclouddeltas') {
    return (
      <Box p={4} bg={bgColor} borderRadius="md">
        {renderHeader(FiRefreshCw, 'Deltas (Offline Sync)')}
        {deltas.length === 0 && !isLoading ? (
          <Text color="gray.500">No deltas for this project.</Text>
        ) : (
          <Table size="sm" variant="simple">
            <Thead>
              <Tr>
                <Th>ID</Th>
                <Th>Status</Th>
                <Th>Client</Th>
                <Th>Updated</Th>
              </Tr>
            </Thead>
            <Tbody>
              {deltas.map((d) => (
                <Tr key={d.id}>
                  <Td fontSize="xs" fontFamily="mono">{d.id.substring(0, 8)}…</Td>
                  <Td>
                    <Badge colorScheme={getDeltaStatusColor(d.status)}>{d.status}</Badge>
                  </Td>
                  <Td fontSize="xs" fontFamily="mono">{d.client_id.substring(0, 8)}…</Td>
                  <Td fontSize="sm">{new Date(d.updated_at).toLocaleString()}</Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Box>
    )
  }

  // ─── Root container ───────────────────────────────────────────────────────────
  return (
    <Box p={4} bg={bgColor} borderRadius="md">
      <HStack mb={4}>
        <Icon as={FiCloud} color="blue.500" boxSize={6} />
        <Text fontWeight="bold" fontSize="xl">QFieldCloud</Text>
      </HStack>
      <Text color="gray.500">
        Add a QFieldCloud connection using the settings panel to manage your field data collection projects.
      </Text>
    </Box>
  )
}
