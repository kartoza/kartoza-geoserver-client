import { useState, useEffect } from 'react'
import { Box, Text, HStack, Icon } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { FaCircle } from 'react-icons/fa'
import { useTreeStore } from '../../../stores/treeStore'
import type { TreeNode } from '../../../types'
import * as hostingApi from '../../../api/hosting'
import { TreeNodeRow } from '../TreeNodeRow'

export function HostingRootNode() {
  const nodeId = 'hosting-root'
  const isExpanded = useTreeStore((state) => state.isExpanded(nodeId))
  const toggleNode = useTreeStore((state) => state.toggleNode)
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)

  const [isAuthenticated, setIsAuthenticated] = useState(hostingApi.isAuthenticated())

  // Check auth status on mount and periodically
  useEffect(() => {
    const checkAuth = () => {
      setIsAuthenticated(hostingApi.isAuthenticated())
    }
    checkAuth()
    const interval = setInterval(checkAuth, 5000)
    return () => clearInterval(interval)
  }, [])

  // Fetch user profile if authenticated
  const { data: user } = useQuery({
    queryKey: ['hosting-profile'],
    queryFn: () => hostingApi.getProfile(),
    enabled: isAuthenticated,
    staleTime: 60000,
    retry: false,
  })

  // Fetch instances if authenticated
  const { data: instancesData, isLoading: instancesLoading } = useQuery({
    queryKey: ['hosting-instances'],
    queryFn: () => hostingApi.listInstances(),
    enabled: isAuthenticated,
    staleTime: 30000,
  })

  const node: TreeNode = {
    id: nodeId,
    name: 'Hosting',
    type: 'hosting',
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
    toggleNode(nodeId)
  }

  const handleShop = (e: React.MouseEvent) => {
    e.stopPropagation()
    const shopNode: TreeNode = {
      id: 'hosting-shop',
      name: 'Shop',
      type: 'hostingshop',
    }
    selectNode(shopNode)
  }

  return (
    <Box>
      <TreeNodeRow
        node={node}
        isExpanded={isExpanded}
        isSelected={isSelected}
        isLoading={instancesLoading}
        onClick={handleClick}
        onAdd={handleShop}
        level={1}
        count={instancesData?.instances?.length}
      />
      {isExpanded && (
        <Box pl={4}>
          {isAuthenticated ? (
            <>
              {/* User Account Node */}
              <HostingAccountNode user={user} />

              {/* Instances */}
              {instancesData?.instances && instancesData.instances.length > 0 ? (
                instancesData.instances.map((instance) => (
                  <HostingInstanceNode
                    key={instance.id}
                    instance={instance}
                  />
                ))
              ) : (
                <Box px={2} py={3}>
                  <Text color="gray.500" fontSize="sm">
                    No instances yet. Click + to browse the shop.
                  </Text>
                </Box>
              )}
            </>
          ) : (
            <Box px={2} py={3}>
              <Text color="gray.500" fontSize="sm">
                Click + to browse products or sign in to view your instances.
              </Text>
            </Box>
          )}
        </Box>
      )}
    </Box>
  )
}

// Account Node Component
function HostingAccountNode({ user }: { user?: hostingApi.HostingUser }) {
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)

  const nodeId = 'hosting-account'
  const node: TreeNode = {
    id: nodeId,
    name: user?.email || 'Account',
    type: 'hostingaccount',
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    // Clicking on account opens the shop (profile management)
    const shopNode: TreeNode = {
      id: 'hosting-shop',
      name: 'Shop',
      type: 'hostingshop',
    }
    selectNode(shopNode)
  }

  return (
    <TreeNodeRow
      node={node}
      isExpanded={false}
      isSelected={isSelected}
      isLoading={false}
      onClick={handleClick}
      level={2}
      isLeaf
    />
  )
}

// Instance Node Component
function HostingInstanceNode({ instance }: { instance: hostingApi.InstanceSummary }) {
  const selectNode = useTreeStore((state) => state.selectNode)
  const selectedNode = useTreeStore((state) => state.selectedNode)

  const nodeId = `hosting-instance-${instance.id}`
  const node: TreeNode = {
    id: nodeId,
    name: instance.display_name || instance.name,
    type: 'hostinginstance',
    hostingInstanceId: instance.id,
    hostingInstanceStatus: instance.status,
    hostingInstanceUrl: instance.url,
    hostingProductName: instance.product_name,
    hostingPackageName: instance.package_name,
  }

  const isSelected = selectedNode?.id === nodeId

  const handleClick = () => {
    selectNode(node)
  }

  const getStatusColor = () => {
    switch (instance.status) {
      case 'online':
        return 'green.500'
      case 'starting_up':
      case 'deploying':
        return 'yellow.500'
      case 'offline':
      case 'maintenance':
        return 'gray.500'
      case 'error':
        return 'red.500'
      default:
        return 'gray.500'
    }
  }

  return (
    <Box position="relative">
      <HStack spacing={0} align="center">
        <TreeNodeRow
          node={node}
          isExpanded={false}
          isSelected={isSelected}
          isLoading={false}
          onClick={handleClick}
          level={2}
          isLeaf
        />
      </HStack>
      {/* Status indicator */}
      <Box
        position="absolute"
        left="44px"
        top="50%"
        transform="translateY(-50%)"
        pointerEvents="none"
      >
        <Icon as={FaCircle} color={getStatusColor()} boxSize={2} />
      </Box>
    </Box>
  )
}
