import {
  VStack,
  Card,
  CardBody,
  Flex,
  HStack,
  Box,
  Icon,
  Heading,
  Text,
  Spacer,
  Button,
  Badge,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  useColorModeValue,
  useDisclosure,
} from '@chakra-ui/react'
import { FiServer, FiSettings, FiPlus, FiUpload } from 'react-icons/fi'
import { useQuery } from '@tanstack/react-query'
import * as api from '../../api/client'
import { useConnectionStore } from '../../stores/connectionStore'
import { useUIStore } from '../../stores/uiStore'
import { SettingsDialog } from '../dialogs/SettingsDialog'

interface ConnectionPanelProps {
  connectionId: string
}

export default function ConnectionPanel({ connectionId }: ConnectionPanelProps) {
  const connections = useConnectionStore((state) => state.connections)
  const connection = connections.find((c) => c.id === connectionId)
  const openDialog = useUIStore((state) => state.openDialog)
  const cardBg = useColorModeValue('white', 'gray.800')
  const settingsDisclosure = useDisclosure()

  const { data: serverInfo } = useQuery({
    queryKey: ['serverInfo', connectionId],
    queryFn: () => api.getServerInfo(connectionId),
  })

  const { data: workspaces } = useQuery({
    queryKey: ['workspaces', connectionId],
    queryFn: () => api.getWorkspaces(connectionId),
  })

  if (!connection) return null

  return (
    <VStack spacing={6} align="stretch">
      {/* Connection Header */}
      <Card
        bg="linear-gradient(135deg, #0a3a50 0%, #175a77 50%, #2d7d9b 100%)"
        color="white"
      >
        <CardBody py={8} px={6}>
          <Flex align="center" wrap="wrap" gap={4}>
            <HStack spacing={4}>
              <Box bg="whiteAlpha.200" p={3} borderRadius="lg">
                <Icon as={FiServer} boxSize={8} />
              </Box>
              <VStack align="start" spacing={0}>
                <Heading size="lg" color="white">{connection.name}</Heading>
                <Text color="white" opacity={0.9}>{connection.url}</Text>
              </VStack>
            </HStack>
            <Spacer />
            <HStack>
              <Button
                variant="outline"
                color="white"
                borderColor="whiteAlpha.400"
                _hover={{ bg: 'whiteAlpha.200' }}
                leftIcon={<FiSettings />}
                onClick={settingsDisclosure.onOpen}
              >
                Service Metadata
              </Button>
              <Badge colorScheme="green" fontSize="md" px={4} py={2}>
                Connected
              </Badge>
            </HStack>
          </Flex>
        </CardBody>
      </Card>

      {/* Settings Dialog */}
      <SettingsDialog
        isOpen={settingsDisclosure.isOpen}
        onClose={settingsDisclosure.onClose}
        connectionId={connectionId}
        connectionName={connection.name}
      />

      {/* Stats */}
      <SimpleGrid columns={{ base: 1, md: 3 }} spacing={4}>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Workspaces</StatLabel>
              <StatNumber color="kartoza.700">{workspaces?.length ?? 0}</StatNumber>
              <StatHelpText>Total workspaces</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        {serverInfo && (
          <>
            <Card bg={cardBg}>
              <CardBody>
                <Stat>
                  <StatLabel>GeoServer</StatLabel>
                  <StatNumber color="kartoza.700" fontSize="xl">{serverInfo.GeoServerVersion}</StatNumber>
                  <StatHelpText>Version</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card bg={cardBg}>
              <CardBody>
                <Stat>
                  <StatLabel>GeoTools</StatLabel>
                  <StatNumber color="kartoza.700" fontSize="xl">{serverInfo.GeoToolsVersion}</StatNumber>
                  <StatHelpText>Version</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </>
        )}
      </SimpleGrid>

      {/* Actions */}
      <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
        <Button
          size="lg"
          variant="accent"
          leftIcon={<FiPlus />}
          onClick={() => openDialog('workspace', { mode: 'create', data: { connectionId } })}
          py={8}
        >
          Create New Workspace
        </Button>
        <Button
          size="lg"
          variant="outline"
          leftIcon={<FiUpload />}
          onClick={() => openDialog('upload', { mode: 'create' })}
          py={8}
        >
          Upload Data
        </Button>
      </SimpleGrid>
    </VStack>
  )
}
