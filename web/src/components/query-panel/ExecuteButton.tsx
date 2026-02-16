// Execute Query Button Component

import {
  Flex,
  Button,
  Tooltip,
  Spinner,
} from '@chakra-ui/react'
import { FiPlay, FiUpload } from 'react-icons/fi'
import type { QueryResult } from './types'

interface ExecuteButtonProps {
  executing: boolean
  result: QueryResult | null
  generatedSQL: string
  sqlQuery: string
  activeMode: 'ai' | 'visual' | 'sql'
  borderColor: string
  bgColor: string
  onExecute: () => void
  onPublishToGeoServer?: (sql: string, name: string) => void
}

export const ExecuteButton: React.FC<ExecuteButtonProps> = ({
  executing,
  result,
  generatedSQL,
  sqlQuery,
  activeMode,
  borderColor,
  bgColor,
  onExecute,
  onPublishToGeoServer,
}) => {
  return (
    <Flex
      p={4}
      gap={2}
      borderTop="1px solid"
      borderColor={borderColor}
      bg={bgColor}
      flexShrink={0}
    >
      <Button
        colorScheme="blue"
        leftIcon={executing ? <Spinner size="sm" /> : <FiPlay />}
        onClick={onExecute}
        isDisabled={executing}
        flex={1}
        borderRadius="xl"
        size="lg"
      >
        {executing ? 'Executing...' : 'Execute Query'}
      </Button>
      {onPublishToGeoServer && result && (
        <Tooltip label="Publish as SQL View Layer to GeoServer">
          <Button
            colorScheme="green"
            leftIcon={<FiUpload />}
            onClick={() => {
              const sql = activeMode === 'visual' ? generatedSQL : sqlQuery
              onPublishToGeoServer(sql, 'sql_view_' + Date.now())
            }}
            borderRadius="xl"
            size="lg"
          >
            Publish
          </Button>
        </Tooltip>
      )}
    </Flex>
  )
}
