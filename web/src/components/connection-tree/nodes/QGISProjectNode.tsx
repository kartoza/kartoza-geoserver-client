import { Box } from '@chakra-ui/react'
import { useTreeStore } from '../../../stores/treeStore'
import { useUIStore } from '../../../stores/uiStore'
import type { TreeNode, QGISProject } from '../../../types'
import { TreeNodeRow } from '../TreeNodeRow'

interface QGISProjectNodeProps {
  project: QGISProject
}

export function QGISProjectNode({ project }: QGISProjectNodeProps) {
  const nodeId = `qgisproject-${project.id}`
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)
  const openDialog = useUIStore((state) => state.openDialog)
  const setQGISPreview = useUIStore((state) => state.setQGISPreview)

  const node: TreeNode = {
    id: nodeId,
    name: project.name,
    type: 'qgisproject',
    qgisProjectId: project.id,
    qgisProjectPath: project.path,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
  }

  const handlePreview = (e: React.MouseEvent) => {
    e.stopPropagation()
    // Show preview in the right pane (like GeoServer layers)
    setQGISPreview({
      projectId: project.id,
      projectName: project.name,
    })
  }

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('qgisproject', {
      mode: 'edit',
      data: { projectId: project.id },
    })
  }

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    openDialog('confirm', {
      mode: 'delete',
      title: 'Delete QGIS Project',
      message: `Are you sure you want to remove "${project.name}" from the list? The file will not be deleted from disk.`,
      data: { qgisProjectId: project.id },
    })
  }

  return (
    <Box>
      <TreeNodeRow
        node={node}
        isExpanded={false}
        isSelected={isSelected}
        isLoading={false}
        onClick={handleClick}
        onPreview={handlePreview}
        onEdit={handleEdit}
        onDelete={handleDelete}
        level={2}
        isLeaf
      />
    </Box>
  )
}
