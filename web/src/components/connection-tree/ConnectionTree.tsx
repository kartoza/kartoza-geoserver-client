import { useEffect } from 'react'
import { Box } from '@chakra-ui/react'
import { useConnectionStore } from '../../stores/connectionStore'
import { CloudBenchRootNode } from './nodes'

export default function ConnectionTree() {
  const connections = useConnectionStore((state) => state.connections)
  const fetchConnections = useConnectionStore((state) => state.fetchConnections)

  useEffect(() => {
    fetchConnections()
  }, [fetchConnections])

  return (
    <Box>
      {/* CloudBench Root Node */}
      <CloudBenchRootNode connections={connections} />
    </Box>
  )
}
