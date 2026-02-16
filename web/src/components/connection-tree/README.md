# ConnectionTree Component Architecture

This directory contains the refactored ConnectionTree component, split into logical, maintainable modules.

## Directory Structure

```
connection-tree/
├── README.md                   # This file
├── ConnectionTree.tsx          # Main component entry point
├── index.ts                    # Public API exports
├── types.ts                    # TypeScript interfaces and types
├── utils.ts                    # Utility functions (icons, colors)
├── TreeNodeRow.tsx            # Reusable tree row component
├── DatasetRow.tsx             # Reusable dataset row component
└── nodes/                     # Node type implementations
    ├── index.ts               # Node exports
    ├── CloudBenchRootNode.tsx        # Root application node
    ├── GeoServerRootNode.tsx         # GeoServer section root
    ├── PostgreSQLRootNode.tsx        # PostgreSQL section root
    ├── ConnectionNode.tsx            # Individual GeoServer connection
    ├── WorkspaceNode.tsx             # GeoServer workspace
    ├── CategoryNode.tsx              # Category containers (Layers, Stores, etc.)
    ├── ItemNode.tsx                  # Individual items (layers, stores, etc.)
    ├── PGServiceNode.tsx             # PostgreSQL service
    ├── PGSchemaNode.tsx              # PostgreSQL schema
    ├── PGTableNode.tsx               # PostgreSQL table/view
    ├── PGColumnRow.tsx               # PostgreSQL column (leaf)
    ├── DataStoreContentsNode.tsx     # DataStore contents with publish UI
    └── CoverageStoreContentsNode.tsx # CoverageStore contents
```

## Module Responsibilities

### Core Files

- **ConnectionTree.tsx**: Main component that initializes the tree and renders the root node
- **index.ts**: Public API - exports components, types, and utilities for external use
- **types.ts**: All TypeScript interfaces and type definitions
- **utils.ts**: Shared utility functions (icon mapping, color mapping)

### Shared Components

- **TreeNodeRow.tsx**: Reusable component for rendering tree rows with consistent styling and actions
- **DatasetRow.tsx**: Specialized row component for feature types and coverages

### Node Components (nodes/)

Each node component is responsible for:
1. Managing its own state (expanded/collapsed, selected)
2. Fetching its own data when needed
3. Rendering its row using TreeNodeRow
4. Rendering its children when expanded
5. Handling user actions (edit, delete, preview, etc.)

## Backwards Compatibility

The original `/components/ConnectionTree.tsx` file now re-exports from this directory, maintaining full backwards compatibility:

```typescript
// Old import still works
import ConnectionTree from '@/components/ConnectionTree'

// New import also works
import ConnectionTree from '@/components/connection-tree'
```

## Usage

### Basic Usage

```typescript
import ConnectionTree from '@/components/connection-tree'

function Sidebar() {
  return <ConnectionTree />
}
```

### Advanced Usage

```typescript
import {
  ConnectionTree,
  TreeNodeRow,
  getNodeIconComponent,
  getNodeColor
} from '@/components/connection-tree'

// Use individual components or utilities
```

## Node Hierarchy

```
CloudBench Root
├── GeoServer
│   └── Connection
│       └── Workspace
│           ├── Data Stores
│           │   └── DataStore
│           │       ├── Published FeatureTypes
│           │       └── Available FeatureTypes (unpublished)
│           ├── Coverage Stores
│           │   └── CoverageStore
│           │       └── Coverages
│           ├── Layers
│           │   └── Layer
│           ├── Styles
│           │   └── Style
│           └── Layer Groups
│               └── LayerGroup
└── PostgreSQL
    └── PGService
        └── PGSchema
            └── PGTable/PGView
                └── PGColumn
```

## Adding New Node Types

To add a new node type:

1. Add the type to `types.ts` if needed
2. Create a new component in `nodes/` directory
3. Add icon mapping to `utils.ts` `getNodeIconComponent()`
4. Add color mapping to `utils.ts` `getNodeColor()`
5. Export from `nodes/index.ts`
6. Integrate into parent node's render

## Testing

After making changes:

1. Run the build: `npm run build`
2. Start the development server: `npm run dev`
3. Test the tree functionality in the browser
4. Verify all node types expand/collapse correctly
5. Test all actions (edit, delete, preview, etc.)
