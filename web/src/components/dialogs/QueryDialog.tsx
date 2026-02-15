import {
  Modal,
  ModalOverlay,
  ModalContent,
} from '@chakra-ui/react'
import { useUIStore } from '../../stores/uiStore'
import { QueryPanel } from '../QueryPanel'

export default function QueryDialog() {
  const activeDialog = useUIStore((state) => state.activeDialog)
  const dialogData = useUIStore((state) => state.dialogData)
  const closeDialog = useUIStore((state) => state.closeDialog)

  const isOpen = activeDialog === 'query'
  const data = dialogData?.data as {
    serviceName?: string
    schemaName?: string
    tableName?: string
    initialSQL?: string
  } | undefined

  if (!data?.serviceName) {
    return null
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={closeDialog}
      size="full"
      motionPreset="slideInBottom"
    >
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent
        m={4}
        borderRadius="2xl"
        overflow="hidden"
        h="calc(100vh - 32px)"
        maxH="calc(100vh - 32px)"
      >
        <QueryPanel
          serviceName={data.serviceName}
          initialSchema={data.schemaName}
          initialTable={data.tableName}
          initialSQL={data.initialSQL}
          onClose={closeDialog}
        />
      </ModalContent>
    </Modal>
  )
}
