import { useTreeStore, generateNodeId } from '../../../stores/treeStore'
import type { TreeNode } from '../../../types'
import { TreeNodeRow } from '../TreeNodeRow'
import type { PGColumnRowProps } from '../types'

export function PGColumnRow({ serviceName, schemaName, tableName, column }: PGColumnRowProps) {
  const nodeId = generateNodeId('pgcolumn', serviceName, schemaName, `${tableName}:${column.name}`)
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)

  const node: TreeNode = {
    id: nodeId,
    name: `${column.name} (${column.type}${column.nullable ? ', nullable' : ''})`,
    type: 'pgcolumn',
    serviceName,
    schemaName,
    tableName,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
  }

  return (
    <TreeNodeRow
      node={node}
      isExpanded={false}
      isSelected={isSelected}
      isLoading={false}
      onClick={handleClick}
      level={5}
      isLeaf
    />
  )
}
