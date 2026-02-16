import { Box } from '@chakra-ui/react'
import ConnectionTree from './ConnectionTree'

export default function Sidebar() {
  return (
    <Box h="100%">
      <Box p={2}>
        <ConnectionTree />
      </Box>
    </Box>
  )
}
