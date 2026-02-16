// SQL Editor Mode Component

import {
  Box,
  VStack,
  HStack,
  Text,
  IconButton,
} from '@chakra-ui/react'
import { motion } from 'framer-motion'
import { FiCopy } from 'react-icons/fi'
import { SQLEditor } from '../SQLEditor'
import { springs } from '../../utils/animations'
import type { SchemaInfo } from './types'

const MotionBox = motion(Box)

interface SQLModeProps {
  serviceName: string
  schemas: SchemaInfo[]
  sqlQuery: string
  onSqlChange: (sql: string) => void
  onCopySQL: () => void
}

export const SQLMode: React.FC<SQLModeProps> = ({
  serviceName,
  schemas,
  sqlQuery,
  onSqlChange,
  onCopySQL,
}) => {
  return (
    <MotionBox
      key="sql"
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 20 }}
      transition={springs.snappy}
      h="100%"
    >
      <VStack spacing={4} align="stretch" h="100%">
        <HStack justify="space-between">
          <Text fontWeight="600" fontSize="sm" color="gray.600">
            Write your SQL query
          </Text>
          <HStack>
            <IconButton
              aria-label="Copy"
              icon={<FiCopy />}
              size="sm"
              variant="ghost"
              onClick={onCopySQL}
            />
          </HStack>
        </HStack>
        <Box flex={1} minH="300px">
          <SQLEditor
            value={sqlQuery}
            onChange={onSqlChange}
            height="100%"
            serviceName={serviceName}
            schemas={schemas}
            placeholder="SELECT * FROM schema.table WHERE ..."
          />
        </Box>
      </VStack>
    </MotionBox>
  )
}
