# QueryPanel Component Structure

The QueryPanel has been refactored from a monolithic 1516-line file into a modular, maintainable structure.

## Directory Structure

```
query-panel/
├── index.tsx                 # Main component orchestrator (345 lines)
├── types.ts                  # TypeScript interfaces and types (65 lines)
├── constants.ts              # SQL operators, aggregates, examples (34 lines)
├── utils.ts                  # Pure utility functions (97 lines)
├── hooks.ts                  # Custom React hooks (255 lines)
├── AIMode.tsx               # AI natural language query mode (215 lines)
├── VisualBuilder.tsx        # Visual query builder mode (491 lines)
├── SQLMode.tsx              # SQL editor mode (70 lines)
├── ResultsPanel.tsx         # Query results display (244 lines)
├── QueryPanelHeader.tsx     # Header component (63 lines)
├── QueryModeTabs.tsx        # Mode selection tabs (67 lines)
├── ExecuteButton.tsx        # Query execution button (73 lines)
└── README.md                # This file
```

## File Descriptions

### Core Files

**index.tsx** - Main component that:
- Orchestrates all sub-components
- Manages top-level state using custom hooks
- Coordinates data flow between components
- Handles query execution and result management

**types.ts** - Type definitions:
- `Column`, `Condition`, `OrderBy` - Visual builder types
- `SchemaInfo`, `TableInfo`, `ColumnInfo` - Database schema types
- `QueryResult` - Query execution results
- `QueryPanelProps` - Main component props
- `AIResponse` - AI query generation response
- `QueryMode` - Type for query modes ('ai' | 'visual' | 'sql')

**constants.ts** - Static configuration:
- `OPERATORS` - SQL comparison operators for WHERE clauses
- `AGGREGATES` - SQL aggregate functions
- `AI_EXAMPLE_QUESTIONS` - Example prompts for AI mode

**utils.ts** - Pure utility functions:
- `buildSQL()` - Generates SQL from visual builder state
- `exportToCSV()` - Exports results to CSV format
- `exportToJSON()` - Exports results to JSON format
- `modifySQLWithOffset()` - Modifies SQL for pagination

### Custom Hooks

**hooks.ts** - Reusable stateful logic:
- `useSchemas()` - Loads and manages database schemas
- `useAIProvider()` - Checks AI provider availability
- `useQueryExecution()` - Handles query execution and results
- `useAIQuery()` - Manages AI query generation
- `useResizableSplitter()` - Implements resizable panel splitter
- `useInfiniteScroll()` - Implements infinite scroll for results

### UI Components

**AIMode.tsx** - AI natural language query interface:
- Question input with example prompts
- AI provider status checking
- Generated SQL display with confidence scores
- Warning and explanation display

**VisualBuilder.tsx** - Visual query builder:
- Schema and table selection
- Column selection with aggregates and aliases
- WHERE conditions builder
- ORDER BY builder
- DISTINCT and LIMIT options
- Live SQL preview

**SQLMode.tsx** - Raw SQL editor:
- Full-featured SQL editor with syntax highlighting
- Schema-aware autocomplete
- Copy SQL functionality

**ResultsPanel.tsx** - Results display:
- Tabular data display with row numbers
- Column type information
- Export to CSV/JSON
- Infinite scroll for large result sets
- Loading states and error handling

**QueryPanelHeader.tsx** - Header bar:
- Service name display
- Schema count badge
- Close button

**QueryModeTabs.tsx** - Mode selection:
- Three tabs: AI, Visual Builder, SQL Editor
- Smooth mode switching

**ExecuteButton.tsx** - Action buttons:
- Execute query button with loading state
- Publish to GeoServer button (optional)

## Usage

The component maintains backwards compatibility. Import it the same way:

```typescript
import { QueryPanel } from '@/components/QueryPanel'
// or
import { QueryPanel } from '@/components/query-panel'
```

Both imports work identically. The original `QueryPanel.tsx` now re-exports from the refactored structure.

## Benefits of This Structure

1. **Maintainability** - Each file has a single responsibility
2. **Testability** - Components and hooks can be tested in isolation
3. **Reusability** - Hooks and utilities can be used elsewhere
4. **Readability** - Smaller files are easier to understand
5. **Collaboration** - Multiple developers can work on different parts
6. **Performance** - Easier to optimize individual components
7. **Type Safety** - Centralized type definitions prevent inconsistencies

## State Management

State is managed using a combination of:
- React hooks for component state
- Custom hooks for complex logic
- Props for parent-child communication
- No external state management library needed

## Key Features

- Three query modes: AI, Visual, SQL
- Real-time SQL generation from visual builder
- AI-powered natural language to SQL conversion
- Infinite scroll for large result sets
- Resizable split panels
- Export results to CSV/JSON
- Schema and table browsing
- WHERE conditions, ORDER BY, aggregates
- Query execution with timing information

## Performance Considerations

- Memoization via `useMemo` for derived state
- `useCallback` for stable function references
- AnimatePresence for smooth transitions
- Virtualization via infinite scroll
- Efficient re-render prevention

## Future Enhancements

Potential improvements:
- Query history component
- Saved queries/favorites
- Query execution plan visualization
- Result set filtering and sorting
- Column resize and reorder
- Cell editing for editable views
- Multi-query tabs
