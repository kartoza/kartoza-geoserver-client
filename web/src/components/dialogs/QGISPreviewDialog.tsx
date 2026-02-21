import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalFooter,
  ModalCloseButton,
  Button,
  Box,
} from '@chakra-ui/react'
import { useUIStore } from '../../stores/uiStore'
import QGISMapLibrePreview from '../QGISMapLibrePreview'

export default function QGISPreviewDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const dialogData = useUIStore((state) => state.dialogData)
  const closeDialog = useUIStore((state) => state.closeDialog)

  const isOpen = activeDialog === 'qgispreview'
  const projectId = dialogData?.data?.projectId as string | undefined
  const projectName = dialogData?.data?.projectName as string | undefined

  const handleClose = () => {
    closeDialog()
  }

  if (!projectId) return null

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="6xl" isCentered>
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent borderRadius="xl" overflow="hidden" maxH="90vh">
        <ModalCloseButton zIndex={10} />

        <Box h="75vh">
          <QGISMapLibrePreview
            projectId={projectId}
            projectName={projectName}
          />
        </Box>

        <ModalFooter
          gap={3}
          borderTop="1px solid"
          borderTopColor="gray.100"
          bg="gray.50"
        >
          <Button variant="ghost" onClick={handleClose} borderRadius="lg">
            Close
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
