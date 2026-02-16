// Utility functions for QueryPanel

import type { Column, Condition, OrderBy } from './types'

/**
 * Build SQL query from visual builder components
 */
export function buildSQL(
  selectedSchema: string,
  selectedTable: string,
  selectedColumns: Column[],
  conditions: Condition[],
  orderBy: OrderBy[],
  limit: number,
  offset: number,
  distinct: boolean
): string {
  if (!selectedTable) return ''

  const cols = selectedColumns.length > 0
    ? selectedColumns.map(c => {
        if (c.aggregate) {
          return `${c.aggregate}(${c.name})${c.alias ? ` AS ${c.alias}` : ''}`
        }
        return c.alias ? `${c.name} AS ${c.alias}` : c.name
      }).join(', ')
    : '*'

  let sql = `SELECT ${distinct ? 'DISTINCT ' : ''}${cols}\nFROM ${selectedSchema}.${selectedTable}`

  if (conditions.length > 0) {
    const whereClause = conditions.map((c, i) => {
      let clause = ''
      if (i > 0) clause += ` ${c.logic} `
      if (['IS NULL', 'IS NOT NULL'].includes(c.operator)) {
        clause += `${c.column} ${c.operator}`
      } else if (c.operator === 'IN') {
        clause += `${c.column} IN (${c.value})`
      } else if (['LIKE', 'ILIKE'].includes(c.operator)) {
        clause += `${c.column} ${c.operator} '%${c.value}%'`
      } else {
        clause += `${c.column} ${c.operator} '${c.value}'`
      }
      return clause
    }).join('')
    sql += `\nWHERE ${whereClause}`
  }

  // Group by if aggregates are used
  const hasAggregates = selectedColumns.some(c => c.aggregate)
  const nonAggColumns = selectedColumns.filter(c => !c.aggregate)
  if (hasAggregates && nonAggColumns.length > 0) {
    sql += `\nGROUP BY ${nonAggColumns.map(c => c.name).join(', ')}`
  }

  if (orderBy.length > 0) {
    sql += `\nORDER BY ${orderBy.map(o => `${o.column} ${o.direction}`).join(', ')}`
  }

  sql += `\nLIMIT ${limit}`
  if (offset > 0) {
    sql += ` OFFSET ${offset}`
  }

  return sql
}

/**
 * Export query results to CSV format
 */
export function exportToCSV(columns: string[], rows: Record<string, unknown>[]): string {
  const headers = columns.join(',')
  const rowsData = rows.map(row =>
    columns.map(col => {
      const val = row[col]
      if (val === null) return ''
      if (typeof val === 'string' && val.includes(',')) return `"${val}"`
      return String(val)
    }).join(',')
  ).join('\n')
  return `${headers}\n${rowsData}`
}

/**
 * Export query results to JSON format
 */
export function exportToJSON(rows: Record<string, unknown>[]): string {
  return JSON.stringify(rows, null, 2)
}

/**
 * Modify SQL query to add or update OFFSET clause
 */
export function modifySQLWithOffset(sql: string, limit: number, currentOffset: number): string {
  const modifiedSQL = sql.replace(/LIMIT \d+(\s+OFFSET \d+)?/i, `LIMIT ${limit} OFFSET ${currentOffset}`)
  return modifiedSQL.includes('OFFSET') ? modifiedSQL : `${sql} OFFSET ${currentOffset}`
}
