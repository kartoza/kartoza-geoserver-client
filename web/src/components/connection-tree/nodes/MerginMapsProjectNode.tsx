import { Box } from '@chakra-ui/react'
import { useTreeStore, generateNodeId } from '../../../stores/treeStore'
import type { TreeNode, MerginMapsProject } from '../../../types'
import { TreeNodeRow } from '../TreeNodeRow'

interface MerginMapsProjectNodeProps {
  connectionId: string
  project: MerginMapsProject
}

export function MerginMapsProjectNode({ connectionId, project }: MerginMapsProjectNodeProps) {
  const projectPath = `${project.namespace}/${project.name}`
  const nodeId = generateNodeId('merginmapsproject', `${connectionId}-${projectPath}`)
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)

  const node: TreeNode = {
    id: nodeId,
    name: projectPath,
    type: 'merginmapsproject',
    merginMapsConnectionId: connectionId,
    merginMapsNamespace: project.namespace,
    merginMapsProjectName: project.name,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
  }

  return (
    <Box>
      <TreeNodeRow
        node={node}
        isExpanded={false}
        isSelected={isSelected}
        isLoading={false}
        onClick={handleClick}
        level={3}
      />
    </Box>
  )
}
