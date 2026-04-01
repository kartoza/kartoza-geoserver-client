import { create } from 'zustand'
import type { TreeNode, NodeType } from '../types'
import { setNodeUrlParam, clearNodeUrlParam, getParentNodeIds } from '../utils/nodeUrl'

interface TreeState {
  expandedNodes: Set<string>
  selectedNode: TreeNode | null

  // Actions
  toggleNode: (nodeId: string) => void
  expandNode: (nodeId: string) => void
  collapseNode: (nodeId: string) => void
  selectNode: (node: TreeNode | null) => void
  restoreNode: (node: TreeNode) => void
  isExpanded: (nodeId: string) => boolean
  clearSelection: () => void
  reset: () => void
}

export const useTreeStore = create<TreeState>((set, get) => ({
  expandedNodes: new Set<string>(),
  selectedNode: null,

  toggleNode: (nodeId: string) => {
    set(state => {
      const newExpanded = new Set(state.expandedNodes)
      if (newExpanded.has(nodeId)) {
        newExpanded.delete(nodeId)
      } else {
        newExpanded.add(nodeId)
      }
      return { expandedNodes: newExpanded }
    })
  },

  expandNode: (nodeId: string) => {
    set(state => {
      const newExpanded = new Set(state.expandedNodes)
      newExpanded.add(nodeId)
      return { expandedNodes: newExpanded }
    })
  },

  collapseNode: (nodeId: string) => {
    set(state => {
      const newExpanded = new Set(state.expandedNodes)
      newExpanded.delete(nodeId)
      return { expandedNodes: newExpanded }
    })
  },

  selectNode: (node: TreeNode | null) => {
    set({ selectedNode: node })
    if (node) {
      setNodeUrlParam(node.id)
    } else {
      clearNodeUrlParam()
    }
  },

  restoreNode: (node: TreeNode) => {
    const parentIds = getParentNodeIds(node.id)
    const selfExpanding = ['datastore', 'coveragestore', 'layer', 'layergroups', 'styles'].includes(node.type)
    set(state => {
      const newExpanded = new Set(state.expandedNodes)
      parentIds.forEach(id => newExpanded.add(id))
      if (selfExpanding) newExpanded.add(node.id)
      return { expandedNodes: newExpanded, selectedNode: node }
    })
  },

  isExpanded: (nodeId: string) => {
    return get().expandedNodes.has(nodeId)
  },

  clearSelection: () => {
    set({ selectedNode: null })
    clearNodeUrlParam()
  },

  reset: () => {
    set({ expandedNodes: new Set(), selectedNode: null })
  },
}))

// Helper function to generate a unique node ID
export function generateNodeId(
  type: NodeType | string,
  connectionId?: string,
  workspace?: string,
  name?: string
): string {
  const parts = [type]
  if (connectionId) parts.push(connectionId)
  if (workspace) parts.push(workspace)
  if (name) parts.push(name)
  return parts.join(':')
}

// Helper function to get icon for node type
export function getNodeIcon(type: NodeType | string): string {
  switch (type) {
    case 'root':
      return '🌍'
    case 'connection':
      return '🖥️'
    case 'workspace':
      return '📁'
    case 'datastores':
      return '💾'
    case 'coveragestores':
      return '🖼️'
    case 'datastore':
      return '🗃️'
    case 'coveragestore':
      return '📷'
    case 'layers':
      return '📑'
    case 'layer':
      return '📄'
    case 'styles':
      return '🎨'
    case 'style':
      return '🖌️'
    case 'layergroups':
      return '📚'
    case 'layergroup':
      return '📘'
    default:
      return '❓'
  }
}
