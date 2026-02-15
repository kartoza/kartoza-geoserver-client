import React from 'react'
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  Text,
  Box,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
} from '@chakra-ui/react'
import { useUIStore } from '../../stores/uiStore'

// Error boundary class component
class ErrorBoundary extends React.Component<
  { children: React.ReactNode; onError?: (error: Error) => void },
  { hasError: boolean; error: Error | null }
> {
  constructor(props: { children: React.ReactNode; onError?: (error: Error) => void }) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('DataViewerDialog Error:', error, errorInfo)
    this.props.onError?.(error)
  }

  render() {
    if (this.state.hasError) {
      return (
        <Alert status="error">
          <AlertIcon />
          <Box>
            <AlertTitle>Dialog Error</AlertTitle>
            <AlertDescription>
              {this.state.error?.message || 'An unknown error occurred'}
            </AlertDescription>
          </Box>
        </Alert>
      )
    }

    return this.props.children
  }
}

function DataViewerContent() {
  console.log('[DataViewerContent] Rendering')

  const dialogData = useUIStore((state) => state.dialogData)

  console.log('[DataViewerContent] dialogData:', dialogData)

  // Safely extract data
  const data = dialogData?.data
  const serviceName = String(data?.serviceName ?? 'unknown')
  const schemaName = String(data?.schemaName ?? 'unknown')
  const tableName = String(data?.tableName ?? 'unknown')

  console.log('[DataViewerContent] Extracted:', { serviceName, schemaName, tableName })

  return (
    <Box>
      <Text>Service: {serviceName}</Text>
      <Text>Schema: {schemaName}</Text>
      <Text>Table: {tableName}</Text>
      <Text mt={4} color="green.500">If you see this, the dialog is working!</Text>
    </Box>
  )
}

export default function DataViewerDialog() {
  console.log('[DataViewerDialog] Component rendering')

  const activeDialog = useUIStore((state) => state.activeDialog)
  const closeDialog = useUIStore((state) => state.closeDialog)

  const isOpen = activeDialog === 'dataviewer'

  console.log('[DataViewerDialog] isOpen:', isOpen, 'activeDialog:', activeDialog)

  if (!isOpen) {
    return null
  }

  return (
    <Modal isOpen={isOpen} onClose={closeDialog} size="xl">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Data Viewer</ModalHeader>
        <ModalCloseButton />
        <ModalBody pb={6}>
          <ErrorBoundary>
            <DataViewerContent />
          </ErrorBoundary>
        </ModalBody>
      </ModalContent>
    </Modal>
  )
}
