// Type definitions for QueryPanel

export interface Column {
  name: string
  alias?: string
  aggregate?: string
}

export interface Condition {
  column: string
  operator: string
  value: string
  logic: 'AND' | 'OR'
}

export interface OrderBy {
  column: string
  direction: 'ASC' | 'DESC'
}

export interface SchemaInfo {
  name: string
  tables: TableInfo[]
}

export interface TableInfo {
  name: string
  columns: ColumnInfo[]
  has_geometry?: boolean
  geometry_column?: string
}

export interface ColumnInfo {
  name: string
  type: string
  nullable: boolean
}

export interface QueryResult {
  columns: ColumnInfo[]
  rows: Record<string, unknown>[]
  row_count: number
  total_count?: number
  duration_ms: number
  sql: string
  has_more?: boolean
}

export interface QueryPanelProps {
  serviceName: string
  initialSchema?: string
  initialTable?: string
  initialSQL?: string
  onClose?: () => void
  onPublishToGeoServer?: (sql: string, name: string) => void
}

export interface AIResponse {
  sql?: string
  explanation?: string
  confidence?: number
  warnings?: string[]
}

export type QueryMode = 'ai' | 'visual' | 'sql'
