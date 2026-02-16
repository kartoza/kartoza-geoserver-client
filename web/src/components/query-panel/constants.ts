// Constants for QueryPanel

export const OPERATORS = [
  { value: '=', label: 'equals' },
  { value: '!=', label: 'not equals' },
  { value: '<', label: 'less than' },
  { value: '<=', label: 'less or equal' },
  { value: '>', label: 'greater than' },
  { value: '>=', label: 'greater or equal' },
  { value: 'LIKE', label: 'contains' },
  { value: 'ILIKE', label: 'contains (case insensitive)' },
  { value: 'IS NULL', label: 'is null' },
  { value: 'IS NOT NULL', label: 'is not null' },
  { value: 'IN', label: 'in list' },
]

export const AGGREGATES = [
  { value: '', label: 'None' },
  { value: 'COUNT', label: 'Count' },
  { value: 'SUM', label: 'Sum' },
  { value: 'AVG', label: 'Average' },
  { value: 'MIN', label: 'Min' },
  { value: 'MAX', label: 'Max' },
  { value: 'ST_Extent', label: 'Extent (geo)' },
  { value: 'ST_Union', label: 'Union (geo)' },
]

export const AI_EXAMPLE_QUESTIONS = [
  'Show me all records from the last 30 days',
  'Find the top 10 largest by area',
  'Count records grouped by type',
  'Show records where name contains "park"',
  'Find all polygons that intersect with each other',
]
