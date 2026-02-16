import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  VStack,
  HStack,
  Text,
  Switch,
  FormControl,
  FormLabel,
  Box,
  Icon,
  Divider,
} from '@chakra-ui/react'
import { FiSettings, FiEye } from 'react-icons/fi'
import { SiPostgresql } from 'react-icons/si'
import { useUIStore } from '../../stores/uiStore'

export default function AppSettingsDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const closeDialog = useUIStore((state) => state.closeDialog)
  const settings = useUIStore((state) => state.settings)
  const setShowHiddenPGServices = useUIStore((state) => state.setShowHiddenPGServices)

  const isOpen = activeDialog === 'settings'

  return (
    <Modal isOpen={isOpen} onClose={closeDialog} size="md" isCentered>
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent borderRadius="xl" overflow="hidden">
        {/* Header */}
        <Box
          bg="linear-gradient(90deg, #dea037 0%, #417d9b 100%)"
          px={6}
          py={4}
        >
          <HStack spacing={3}>
            <Box bg="whiteAlpha.200" p={2} borderRadius="lg">
              <Icon as={FiSettings} boxSize={5} color="white" />
            </Box>
            <Box>
              <Text color="white" fontWeight="600" fontSize="lg">
                Settings
              </Text>
              <Text color="whiteAlpha.800" fontSize="sm">
                Configure application preferences
              </Text>
            </Box>
          </HStack>
        </Box>
        <ModalCloseButton color="white" />

        <ModalBody py={6}>
          <VStack spacing={6} align="stretch">
            {/* PostgreSQL Section */}
            <Box>
              <HStack spacing={2} mb={4}>
                <Icon as={SiPostgresql} color="blue.600" />
                <Text fontWeight="600" color="gray.700">
                  PostgreSQL Services
                </Text>
              </HStack>

              <FormControl display="flex" alignItems="center" justifyContent="space-between">
                <HStack spacing={3}>
                  <Icon as={FiEye} color="gray.500" />
                  <Box>
                    <FormLabel htmlFor="show-hidden-pg" mb={0} cursor="pointer">
                      Show hidden services
                    </FormLabel>
                    <Text fontSize="xs" color="gray.500">
                      Display hidden PostgreSQL services in tree and dashboard
                    </Text>
                  </Box>
                </HStack>
                <Switch
                  id="show-hidden-pg"
                  colorScheme="blue"
                  isChecked={settings.showHiddenPGServices}
                  onChange={(e) => setShowHiddenPGServices(e.target.checked)}
                />
              </FormControl>
            </Box>

            <Divider />

            {/* Info Section */}
            <Box>
              <Text fontSize="sm" color="gray.500">
                Hidden services are commented out in your pg_service.conf file.
                They won't be used by applications but can be restored later.
              </Text>
            </Box>
          </VStack>
        </ModalBody>

        <ModalFooter
          borderTop="1px solid"
          borderTopColor="gray.100"
          bg="gray.50"
        >
          <Button onClick={closeDialog} colorScheme="blue" borderRadius="lg">
            Done
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
