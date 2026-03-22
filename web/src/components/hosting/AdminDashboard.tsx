/**
 * Admin Dashboard - Administrative overview and management interface
 *
 * This component provides an admin dashboard for managing the hosting platform,
 * including user management, instance oversight, and analytics.
 */

import { useState, useEffect, useCallback } from 'react'
import {
  AdminDashboardStats,
  AdminUserListItem,
  AdminInstanceListItem,
  AdminOrderListItem,
  RevenueDataPoint,
  SystemHealth,
  getAdminDashboard,
  getAdminUsers,
  getAdminInstances,
  getAdminOrders,
  getAdminRevenue,
  getAdminHealth,
  updateAdminUser,
  formatPrice,
  formatDate,
  formatDateTime,
  getStatusBadgeColor,
  getHealthBadgeColor,
} from '../../api/hosting'

type AdminTab = 'overview' | 'users' | 'instances' | 'orders' | 'health'

interface Props {
  onClose?: () => void
}

export function AdminDashboard({ onClose }: Props) {
  const [activeTab, setActiveTab] = useState<AdminTab>('overview')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Dashboard data
  const [stats, setStats] = useState<AdminDashboardStats | null>(null)
  const [users, setUsers] = useState<AdminUserListItem[]>([])
  const [instances, setInstances] = useState<AdminInstanceListItem[]>([])
  const [orders, setOrders] = useState<AdminOrderListItem[]>([])
  const [revenueData, setRevenueData] = useState<RevenueDataPoint[]>([])
  const [health, setHealth] = useState<SystemHealth | null>(null)

  // Pagination
  const [userPage, setUserPage] = useState(1)
  const [userTotal, setUserTotal] = useState(0)
  const [instancePage, setInstancePage] = useState(1)
  const [instanceTotal, setInstanceTotal] = useState(0)
  const [orderPage, setOrderPage] = useState(1)
  const [orderTotal, setOrderTotal] = useState(0)

  // Search and filters
  const [userSearch, setUserSearch] = useState('')
  const [instanceSearch, setInstanceSearch] = useState('')
  const [instanceStatus, setInstanceStatus] = useState('')

  const pageSize = 10

  const loadDashboard = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const [dashStats, healthData, revenue] = await Promise.all([
        getAdminDashboard(),
        getAdminHealth(),
        getAdminRevenue('30d', 'day'),
      ])
      setStats(dashStats)
      setHealth(healthData)
      setRevenueData(revenue.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard')
    } finally {
      setLoading(false)
    }
  }, [])

  const loadUsers = useCallback(async () => {
    try {
      const result = await getAdminUsers({
        page: userPage,
        page_size: pageSize,
        search: userSearch || undefined,
      })
      setUsers(result.users)
      setUserTotal(result.total_count)
    } catch (err) {
      console.error('Failed to load users:', err)
    }
  }, [userPage, userSearch])

  const loadInstances = useCallback(async () => {
    try {
      const result = await getAdminInstances({
        page: instancePage,
        page_size: pageSize,
        search: instanceSearch || undefined,
        status: instanceStatus || undefined,
      })
      setInstances(result.instances)
      setInstanceTotal(result.total_count)
    } catch (err) {
      console.error('Failed to load instances:', err)
    }
  }, [instancePage, instanceSearch, instanceStatus])

  const loadOrders = useCallback(async () => {
    try {
      const result = await getAdminOrders({
        page: orderPage,
        page_size: pageSize,
      })
      setOrders(result.orders)
      setOrderTotal(result.total_count)
    } catch (err) {
      console.error('Failed to load orders:', err)
    }
  }, [orderPage])

  useEffect(() => {
    loadDashboard()
  }, [loadDashboard])

  useEffect(() => {
    if (activeTab === 'users') {
      loadUsers()
    }
  }, [activeTab, loadUsers])

  useEffect(() => {
    if (activeTab === 'instances') {
      loadInstances()
    }
  }, [activeTab, loadInstances])

  useEffect(() => {
    if (activeTab === 'orders') {
      loadOrders()
    }
  }, [activeTab, loadOrders])

  const handleToggleUserActive = async (user: AdminUserListItem) => {
    try {
      await updateAdminUser(user.id, { is_active: !user.is_active })
      loadUsers()
    } catch (err) {
      console.error('Failed to update user:', err)
    }
  }

  const handleToggleUserAdmin = async (user: AdminUserListItem) => {
    try {
      await updateAdminUser(user.id, { is_admin: !user.is_admin })
      loadUsers()
    } catch (err) {
      console.error('Failed to update user:', err)
    }
  }

  if (loading && !stats) {
    return (
      <div className="admin-dashboard">
        <div className="admin-loading">Loading admin dashboard...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="admin-dashboard">
        <div className="admin-error">
          <span className="error-icon">!</span>
          <span>{error}</span>
          <button onClick={loadDashboard}>Retry</button>
        </div>
      </div>
    )
  }

  const userTotalPages = Math.ceil(userTotal / pageSize)
  const instanceTotalPages = Math.ceil(instanceTotal / pageSize)
  const orderTotalPages = Math.ceil(orderTotal / pageSize)

  return (
    <div className="admin-dashboard">
      <div className="admin-header">
        <h2>Admin Dashboard</h2>
        {onClose && (
          <button className="close-btn" onClick={onClose}>
            ×
          </button>
        )}
      </div>

      <div className="admin-tabs">
        <button
          className={`tab ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => setActiveTab('overview')}
        >
          Overview
        </button>
        <button
          className={`tab ${activeTab === 'users' ? 'active' : ''}`}
          onClick={() => setActiveTab('users')}
        >
          Users
        </button>
        <button
          className={`tab ${activeTab === 'instances' ? 'active' : ''}`}
          onClick={() => setActiveTab('instances')}
        >
          Instances
        </button>
        <button
          className={`tab ${activeTab === 'orders' ? 'active' : ''}`}
          onClick={() => setActiveTab('orders')}
        >
          Orders
        </button>
        <button
          className={`tab ${activeTab === 'health' ? 'active' : ''}`}
          onClick={() => setActiveTab('health')}
        >
          Health
        </button>
      </div>

      <div className="admin-content">
        {activeTab === 'overview' && stats && (
          <div className="overview-tab">
            <div className="stats-grid">
              <div className="stat-card">
                <div className="stat-label">Total Users</div>
                <div className="stat-value">{stats.total_users}</div>
                <div className="stat-detail">{stats.active_users} active</div>
              </div>
              <div className="stat-card">
                <div className="stat-label">Total Instances</div>
                <div className="stat-value">{stats.total_instances}</div>
                <div className="stat-detail">{stats.online_instances} online</div>
              </div>
              <div className="stat-card">
                <div className="stat-label">Monthly Revenue</div>
                <div className="stat-value">{formatPrice(stats.monthly_revenue)}</div>
                <div className="stat-detail">Total: {formatPrice(stats.total_revenue)}</div>
              </div>
              <div className="stat-card">
                <div className="stat-label">Pending Orders</div>
                <div className="stat-value">{stats.pending_orders}</div>
                <div className="stat-detail">{stats.total_orders} total</div>
              </div>
            </div>

            <div className="overview-sections">
              <div className="overview-section">
                <h3>Instances by Status</h3>
                <div className="status-list">
                  {Object.entries(stats.instances_by_status).map(([status, count]) => (
                    <div key={status} className="status-item">
                      <span className={`status-badge status-${getStatusBadgeColor(status as any)}`}>
                        {status}
                      </span>
                      <span className="status-count">{count}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="overview-section">
                <h3>Revenue by Product</h3>
                <div className="revenue-list">
                  {Object.entries(stats.revenue_by_product).map(([product, revenue]) => (
                    <div key={product} className="revenue-item">
                      <span className="product-name">{product}</span>
                      <span className="product-revenue">{formatPrice(revenue)}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {revenueData.length > 0 && (
              <div className="revenue-chart">
                <h3>Revenue Trend (30 days)</h3>
                <div className="chart-container">
                  <div className="simple-chart">
                    {revenueData.map((dp, i) => {
                      const maxRevenue = Math.max(...revenueData.map(d => d.revenue), 1)
                      const height = (dp.revenue / maxRevenue) * 100
                      return (
                        <div
                          key={i}
                          className="chart-bar"
                          style={{ height: `${height}%` }}
                          title={`${dp.date}: ${formatPrice(dp.revenue)}`}
                        />
                      )
                    })}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'users' && (
          <div className="users-tab">
            <div className="tab-header">
              <input
                type="text"
                placeholder="Search users..."
                value={userSearch}
                onChange={(e) => {
                  setUserSearch(e.target.value)
                  setUserPage(1)
                }}
                className="search-input"
              />
            </div>

            <table className="admin-table">
              <thead>
                <tr>
                  <th>Email</th>
                  <th>Name</th>
                  <th>Instances</th>
                  <th>Total Spent</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id}>
                    <td>
                      {user.email}
                      {user.is_admin && <span className="admin-badge">Admin</span>}
                    </td>
                    <td>{user.first_name} {user.last_name}</td>
                    <td>{user.instance_count}</td>
                    <td>{formatPrice(user.total_spent)}</td>
                    <td>
                      <span className={`status-badge status-${user.is_active ? 'green' : 'gray'}`}>
                        {user.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td>
                      <button
                        className="action-btn"
                        onClick={() => handleToggleUserActive(user)}
                        title={user.is_active ? 'Deactivate' : 'Activate'}
                      >
                        {user.is_active ? 'Deactivate' : 'Activate'}
                      </button>
                      <button
                        className="action-btn"
                        onClick={() => handleToggleUserAdmin(user)}
                        title={user.is_admin ? 'Remove Admin' : 'Make Admin'}
                      >
                        {user.is_admin ? 'Remove Admin' : 'Make Admin'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div className="pagination">
              <button
                disabled={userPage <= 1}
                onClick={() => setUserPage(p => p - 1)}
              >
                Previous
              </button>
              <span>Page {userPage} of {userTotalPages}</span>
              <button
                disabled={userPage >= userTotalPages}
                onClick={() => setUserPage(p => p + 1)}
              >
                Next
              </button>
            </div>
          </div>
        )}

        {activeTab === 'instances' && (
          <div className="instances-tab">
            <div className="tab-header">
              <input
                type="text"
                placeholder="Search instances..."
                value={instanceSearch}
                onChange={(e) => {
                  setInstanceSearch(e.target.value)
                  setInstancePage(1)
                }}
                className="search-input"
              />
              <select
                value={instanceStatus}
                onChange={(e) => {
                  setInstanceStatus(e.target.value)
                  setInstancePage(1)
                }}
                className="filter-select"
              >
                <option value="">All Statuses</option>
                <option value="online">Online</option>
                <option value="offline">Offline</option>
                <option value="deploying">Deploying</option>
                <option value="error">Error</option>
              </select>
            </div>

            <table className="admin-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>User</th>
                  <th>Product</th>
                  <th>Status</th>
                  <th>Health</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {instances.map((inst) => (
                  <tr key={inst.id}>
                    <td>
                      {inst.name}
                      {inst.url && (
                        <a href={inst.url} target="_blank" rel="noopener noreferrer" className="instance-link">
                          ↗
                        </a>
                      )}
                    </td>
                    <td>{inst.user_email}</td>
                    <td>{inst.product_name} / {inst.package_name}</td>
                    <td>
                      <span className={`status-badge status-${getStatusBadgeColor(inst.status as any)}`}>
                        {inst.status}
                      </span>
                    </td>
                    <td>
                      <span className={`status-badge status-${getHealthBadgeColor(inst.health_status as any)}`}>
                        {inst.health_status}
                      </span>
                    </td>
                    <td>{formatDate(inst.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div className="pagination">
              <button
                disabled={instancePage <= 1}
                onClick={() => setInstancePage(p => p - 1)}
              >
                Previous
              </button>
              <span>Page {instancePage} of {instanceTotalPages}</span>
              <button
                disabled={instancePage >= instanceTotalPages}
                onClick={() => setInstancePage(p => p + 1)}
              >
                Next
              </button>
            </div>
          </div>
        )}

        {activeTab === 'orders' && (
          <div className="orders-tab">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Order ID</th>
                  <th>User</th>
                  <th>Product</th>
                  <th>Amount</th>
                  <th>Status</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((order) => (
                  <tr key={order.id}>
                    <td>{order.id.slice(0, 8)}...</td>
                    <td>{order.user_email}</td>
                    <td>{order.product_name} / {order.package_name}</td>
                    <td>{formatPrice(order.total_amount / 100, order.currency)}</td>
                    <td>
                      <span className={`status-badge status-${order.status === 'paid' ? 'green' : order.status === 'pending' ? 'yellow' : 'gray'}`}>
                        {order.status}
                      </span>
                    </td>
                    <td>{formatDateTime(order.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div className="pagination">
              <button
                disabled={orderPage <= 1}
                onClick={() => setOrderPage(p => p - 1)}
              >
                Previous
              </button>
              <span>Page {orderPage} of {orderTotalPages}</span>
              <button
                disabled={orderPage >= orderTotalPages}
                onClick={() => setOrderPage(p => p + 1)}
              >
                Next
              </button>
            </div>
          </div>
        )}

        {activeTab === 'health' && health && (
          <div className="health-tab">
            <div className="system-health">
              <div className={`health-status health-${health.status}`}>
                <h3>System Status: {health.status.toUpperCase()}</h3>
                <div className="health-timestamp">Last updated: {formatDateTime(health.timestamp)}</div>
              </div>

              <div className="components-list">
                {Object.entries(health.components).map(([name, component]) => (
                  <div key={name} className={`component-card component-${component.status}`}>
                    <div className="component-header">
                      <span className="component-name">{name}</span>
                      <span className={`status-badge status-${component.status === 'healthy' ? 'green' : 'red'}`}>
                        {component.status}
                      </span>
                    </div>
                    {component.message && (
                      <div className="component-message">{component.message}</div>
                    )}
                    <div className="component-checked">
                      Checked: {formatDateTime(component.last_checked)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>

      <style>{`
        .admin-dashboard {
          padding: 20px;
          max-width: 1200px;
          margin: 0 auto;
        }

        .admin-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 20px;
        }

        .admin-header h2 {
          margin: 0;
        }

        .close-btn {
          background: none;
          border: none;
          font-size: 24px;
          cursor: pointer;
          color: var(--text-secondary);
        }

        .admin-tabs {
          display: flex;
          gap: 10px;
          border-bottom: 1px solid var(--border-color);
          margin-bottom: 20px;
        }

        .admin-tabs .tab {
          padding: 10px 20px;
          background: none;
          border: none;
          cursor: pointer;
          color: var(--text-secondary);
          border-bottom: 2px solid transparent;
        }

        .admin-tabs .tab.active {
          color: var(--primary-color);
          border-bottom-color: var(--primary-color);
        }

        .admin-tabs .tab:hover {
          color: var(--primary-color);
        }

        .admin-loading, .admin-error {
          text-align: center;
          padding: 40px;
        }

        .admin-error {
          color: var(--error-color);
        }

        .stats-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 20px;
          margin-bottom: 30px;
        }

        .stat-card {
          background: var(--card-bg);
          padding: 20px;
          border-radius: 8px;
          border: 1px solid var(--border-color);
        }

        .stat-label {
          color: var(--text-secondary);
          font-size: 14px;
        }

        .stat-value {
          font-size: 32px;
          font-weight: bold;
          margin: 8px 0;
        }

        .stat-detail {
          color: var(--text-secondary);
          font-size: 14px;
        }

        .overview-sections {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 20px;
          margin-bottom: 30px;
        }

        .overview-section {
          background: var(--card-bg);
          padding: 20px;
          border-radius: 8px;
          border: 1px solid var(--border-color);
        }

        .overview-section h3 {
          margin: 0 0 15px 0;
        }

        .status-list, .revenue-list {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }

        .status-item, .revenue-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .status-badge {
          padding: 2px 8px;
          border-radius: 4px;
          font-size: 12px;
          text-transform: capitalize;
        }

        .status-green { background: #dcfce7; color: #166534; }
        .status-yellow { background: #fef9c3; color: #854d0e; }
        .status-red { background: #fee2e2; color: #991b1b; }
        .status-gray { background: #f3f4f6; color: #374151; }

        .admin-badge {
          margin-left: 8px;
          padding: 2px 6px;
          background: #dbeafe;
          color: #1e40af;
          border-radius: 4px;
          font-size: 11px;
        }

        .revenue-chart {
          background: var(--card-bg);
          padding: 20px;
          border-radius: 8px;
          border: 1px solid var(--border-color);
        }

        .revenue-chart h3 {
          margin: 0 0 15px 0;
        }

        .chart-container {
          height: 150px;
          padding: 10px 0;
        }

        .simple-chart {
          display: flex;
          align-items: flex-end;
          gap: 2px;
          height: 100%;
        }

        .chart-bar {
          flex: 1;
          background: var(--primary-color);
          min-height: 4px;
          border-radius: 2px 2px 0 0;
          cursor: pointer;
        }

        .chart-bar:hover {
          background: var(--primary-hover);
        }

        .tab-header {
          display: flex;
          gap: 10px;
          margin-bottom: 15px;
        }

        .search-input {
          flex: 1;
          padding: 8px 12px;
          border: 1px solid var(--border-color);
          border-radius: 4px;
          font-size: 14px;
        }

        .filter-select {
          padding: 8px 12px;
          border: 1px solid var(--border-color);
          border-radius: 4px;
          font-size: 14px;
        }

        .admin-table {
          width: 100%;
          border-collapse: collapse;
        }

        .admin-table th, .admin-table td {
          padding: 12px;
          text-align: left;
          border-bottom: 1px solid var(--border-color);
        }

        .admin-table th {
          font-weight: 600;
          color: var(--text-secondary);
        }

        .action-btn {
          padding: 4px 8px;
          margin-right: 4px;
          background: var(--button-bg);
          border: 1px solid var(--border-color);
          border-radius: 4px;
          cursor: pointer;
          font-size: 12px;
        }

        .action-btn:hover {
          background: var(--button-hover-bg);
        }

        .instance-link {
          margin-left: 8px;
          color: var(--primary-color);
          text-decoration: none;
        }

        .pagination {
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 15px;
          margin-top: 20px;
        }

        .pagination button {
          padding: 8px 16px;
          background: var(--button-bg);
          border: 1px solid var(--border-color);
          border-radius: 4px;
          cursor: pointer;
        }

        .pagination button:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .system-health {
          max-width: 600px;
          margin: 0 auto;
        }

        .health-status {
          text-align: center;
          padding: 20px;
          border-radius: 8px;
          margin-bottom: 20px;
        }

        .health-healthy { background: #dcfce7; }
        .health-degraded { background: #fef9c3; }
        .health-unhealthy { background: #fee2e2; }

        .health-timestamp {
          color: var(--text-secondary);
          font-size: 14px;
          margin-top: 10px;
        }

        .components-list {
          display: flex;
          flex-direction: column;
          gap: 15px;
        }

        .component-card {
          padding: 15px;
          border-radius: 8px;
          border: 1px solid var(--border-color);
        }

        .component-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 10px;
        }

        .component-name {
          font-weight: 600;
          text-transform: capitalize;
        }

        .component-message {
          color: var(--text-secondary);
          font-size: 14px;
          margin-bottom: 10px;
        }

        .component-checked {
          color: var(--text-tertiary);
          font-size: 12px;
        }

        @media (max-width: 768px) {
          .overview-sections {
            grid-template-columns: 1fr;
          }

          .stats-grid {
            grid-template-columns: 1fr 1fr;
          }

          .admin-table {
            font-size: 14px;
          }
        }
      `}</style>
    </div>
  )
}

export default AdminDashboard
