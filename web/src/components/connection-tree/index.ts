// Re-export everything for backwards compatibility
export { default } from './ConnectionTree'
export { default as ConnectionTree } from './ConnectionTree'

// Export types for advanced usage
export type {
  ConnectionNodeProps,
  WorkspaceNodeProps,
  CategoryNodeProps,
  ItemNodeProps,
  PGServiceNodeProps,
  PGSchemaNodeProps,
  PGTableNodeProps,
  PGColumnRowProps,
  DataStoreContentsNodeProps,
  CoverageStoreContentsNodeProps,
  DatasetRowProps,
  TreeNodeRowProps,
} from './types'

// Export utilities for advanced usage
export { getNodeIconComponent, getNodeColor } from './utils'

// Export components for advanced usage
export { TreeNodeRow } from './TreeNodeRow'
export { DatasetRow } from './DatasetRow'
export * from './nodes'
