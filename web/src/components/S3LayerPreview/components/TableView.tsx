import {
  Box,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  HStack,
  IconButton,
  Text,
  Spinner,
} from '@chakra-ui/react'
import { FiChevronLeft, FiChevronRight } from 'react-icons/fi'
import type { S3AttributeTableResponse } from '../../../types'

interface TableViewProps {
  tableData: S3AttributeTableResponse | null
  tableLoading: boolean
  tableOffset: number
  tableLimit: number
  onOffsetChange: (offset: number) => void
}

export default function TableView({
  tableData,
  tableLoading,
  tableOffset,
  tableLimit,
  onOffsetChange,
}: TableViewProps) {
  if (tableLoading && !tableData) {
    return (
      <Box h="100%" display="flex" alignItems="center" justifyContent="center">
        <Spinner size="lg" />
      </Box>
    )
  }

  if (!tableData) {
    return (
      <Box h="100%" display="flex" alignItems="center" justifyContent="center">
        <Text color="gray.500">No data available</Text>
      </Box>
    )
  }

  const handlePrevPage = () => {
    const newOffset = Math.max(0, tableOffset - tableLimit)
    onOffsetChange(newOffset)
  }

  const handleNextPage = () => {
    if (tableData.hasMore) {
      onOffsetChange(tableOffset + tableLimit)
    }
  }

  return (
    <Box h="100%" overflow="auto">
      <Box overflowX="auto">
        <Table size="sm" variant="striped">
          <Thead position="sticky" top={0} bg="gray.700" zIndex={1}>
            <Tr>
              {tableData.fields.map((field) => (
                <Th key={field} color="gray.300">{field}</Th>
              ))}
            </Tr>
          </Thead>
          <Tbody>
            {tableData.rows.map((row, idx) => (
              <Tr key={idx}>
                {tableData.fields.map((field) => (
                  <Td key={field} maxW="200px" overflow="hidden" textOverflow="ellipsis">
                    {String(row[field] ?? '')}
                  </Td>
                ))}
              </Tr>
            ))}
          </Tbody>
        </Table>
      </Box>
      <HStack justify="space-between" p={2} bg="gray.700" position="sticky" bottom={0}>
        <IconButton
          aria-label="Previous page"
          icon={<FiChevronLeft />}
          size="sm"
          onClick={handlePrevPage}
          isDisabled={tableOffset === 0}
        />
        <Text fontSize="sm">
          Showing {tableOffset + 1} - {Math.min(tableOffset + tableLimit, tableData.total)} of {tableData.total}
        </Text>
        <IconButton
          aria-label="Next page"
          icon={<FiChevronRight />}
          size="sm"
          onClick={handleNextPage}
          isDisabled={!tableData.hasMore}
        />
      </HStack>
    </Box>
  )
}
