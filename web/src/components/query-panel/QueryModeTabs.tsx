// Query Mode Tabs Component

import {
  Tabs,
  TabList,
  Tab,
  Icon,
  useColorModeValue,
} from '@chakra-ui/react'
import { FiCpu, FiGrid, FiCode } from 'react-icons/fi'
import type { QueryMode } from './types'

interface QueryModeTabsProps {
  activeMode: QueryMode
  onModeChange: (mode: QueryMode) => void
}

export const QueryModeTabs: React.FC<QueryModeTabsProps> = ({
  activeMode,
  onModeChange,
}) => {
  const headerBg = useColorModeValue('gray.50', 'gray.700')

  const getTabIndex = (mode: QueryMode): number => {
    switch (mode) {
      case 'ai': return 0
      case 'visual': return 1
      case 'sql': return 2
    }
  }

  const getModeFromIndex = (index: number): QueryMode => {
    switch (index) {
      case 0: return 'ai'
      case 1: return 'visual'
      case 2: return 'sql'
      default: return 'visual'
    }
  }

  return (
    <Tabs
      index={getTabIndex(activeMode)}
      onChange={(i) => onModeChange(getModeFromIndex(i))}
      variant="soft-rounded"
      colorScheme="blue"
      px={4}
      py={3}
      bg={headerBg}
    >
      <TabList>
        <Tab gap={2}>
          <Icon as={FiCpu} />
          AI Natural Language
        </Tab>
        <Tab gap={2}>
          <Icon as={FiGrid} />
          Visual Builder
        </Tab>
        <Tab gap={2}>
          <Icon as={FiCode} />
          SQL Editor
        </Tab>
      </TabList>
    </Tabs>
  )
}
