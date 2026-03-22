/**
 * Hosting API - Geospatial Hosting Service APIs
 *
 * This module provides API client functions for the geospatial hosting service,
 * including authentication, products, orders, subscriptions, and instances.
 */

import { handleResponse } from './common'

const HOSTING_BASE = '/api/v1'

// ============================================================================
// Types
// ============================================================================

export interface HostingUser {
  id: string
  email: string
  first_name: string
  last_name: string
  avatar_url?: string
  is_active: boolean
  is_admin: boolean
  created_at: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  first_name?: string
  last_name?: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export interface Product {
  id: string
  name: string
  slug: string
  description: string
  image_url?: string
  icon?: string
  is_available: boolean
  features?: string[]
  created_at: string
}

export interface Package {
  id: string
  product_id: string
  name: string
  slug: string
  price_monthly: number
  price_yearly: number
  features: string[]
  cpu_limit?: string
  memory_limit?: string
  storage_limit?: string
  is_popular: boolean
  sort_order: number
  product?: Product
}

export interface Cluster {
  id: string
  code: string
  name: string
  region: string
  domain: string
  is_active: boolean
  instance_count?: number
  capacity?: number
}

export interface CreateOrderRequest {
  package_id: string
  cluster_id: string
  app_name: string
  billing_cycle: 'monthly' | 'yearly'
  coupon_code?: string
}

export interface SalesOrder {
  id: string
  user_id: string
  package_id: string
  cluster_id: string
  app_name: string
  billing_cycle: string
  status: string
  subtotal_amount: number
  discount_amount: number
  tax_amount: number
  total_amount: number
  currency: string
  payment_provider?: string
  payment_id?: string
  created_at: string
  updated_at: string
  package?: Package
  cluster?: Cluster
}

export interface OrderSummary {
  id: string
  product_name: string
  package_name: string
  cluster_name: string
  app_name: string
  status: string
  total_amount: number
  currency: string
  created_at: string
}

export interface CheckoutRequest {
  success_url: string
  cancel_url: string
}

export interface CheckoutResponse {
  checkout_url: string
  session_id?: string
  reference?: string
}

export interface Instance {
  id: string
  user_id: string
  sales_order_id: string
  product_id: string
  package_id: string
  cluster_id: string
  name: string
  display_name?: string
  status: InstanceStatus
  url?: string
  internal_url?: string
  vault_path?: string
  admin_username?: string
  health_status: HealthStatus
  health_message?: string
  last_health_check?: string
  cpu_usage?: number
  memory_usage?: number
  storage_usage?: number
  expires_at?: string
  deleted_at?: string
  created_at: string
  updated_at: string
  product?: Product
  package?: Package
  cluster?: Cluster
}

export type InstanceStatus =
  | 'pending'
  | 'deploying'
  | 'starting_up'
  | 'online'
  | 'offline'
  | 'error'
  | 'maintenance'
  | 'deleting'
  | 'deleted'

export type HealthStatus =
  | 'unknown'
  | 'healthy'
  | 'degraded'
  | 'unhealthy'

export interface InstanceSummary {
  id: string
  name: string
  display_name: string
  product_name: string
  package_name: string
  cluster_name: string
  status: InstanceStatus
  health_status: HealthStatus
  url?: string
  created_at: string
}

export interface InstanceCredentials {
  url: string
  admin_username: string
  admin_password: string
  database_host?: string
  database_port?: number
  database_name?: string
  database_user?: string
  database_pass?: string
  extra?: Record<string, string>
}

export interface InstanceHealthCheck {
  instance_id: string
  status: string
  response_time: number
  status_code?: number
  error?: string
  checked_at: string
}

export interface Activity {
  id: string
  instance_id: string
  user_id?: string
  activity_type: string
  status: string
  jenkins_build_number?: number
  jenkins_build_url?: string
  argocd_app_name?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  created_at: string
}

export interface Subscription {
  id: string
  user_id: string
  instance_id: string
  stripe_subscription_id?: string
  paystack_subscription_id?: string
  status: string
  current_period_start: string
  current_period_end: string
  cancel_at_period_end: boolean
  created_at: string
}

export interface PaymentConfig {
  providers: string[]
  stripe_publishable_key?: string
  paystack_public_key?: string
}

// ============================================================================
// Token Management
// ============================================================================

const TOKEN_KEY = 'hosting_access_token'
const REFRESH_TOKEN_KEY = 'hosting_refresh_token'

export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function getStoredRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function storeTokens(tokens: AuthTokens): void {
  localStorage.setItem(TOKEN_KEY, tokens.access_token)
  localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token)
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

function authHeaders(): HeadersInit {
  const token = getStoredToken()
  if (token) {
    return { Authorization: `Bearer ${token}` }
  }
  return {}
}

// ============================================================================
// Authentication API
// ============================================================================

export async function register(request: RegisterRequest): Promise<AuthTokens> {
  const response = await fetch(`${HOSTING_BASE}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  })
  const tokens = await handleResponse<AuthTokens>(response)
  storeTokens(tokens)
  return tokens
}

export async function login(request: LoginRequest): Promise<AuthTokens> {
  const response = await fetch(`${HOSTING_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  })
  const tokens = await handleResponse<AuthTokens>(response)
  storeTokens(tokens)
  return tokens
}

export async function logout(): Promise<void> {
  try {
    await fetch(`${HOSTING_BASE}/auth/logout`, {
      method: 'POST',
      headers: authHeaders(),
    })
  } finally {
    clearTokens()
  }
}

export async function refreshTokens(): Promise<AuthTokens> {
  const refreshToken = getStoredRefreshToken()
  if (!refreshToken) {
    throw new Error('No refresh token available')
  }

  const response = await fetch(`${HOSTING_BASE}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
  const tokens = await handleResponse<AuthTokens>(response)
  storeTokens(tokens)
  return tokens
}

export async function getProfile(): Promise<HostingUser> {
  const response = await fetch(`${HOSTING_BASE}/auth/profile`, {
    headers: authHeaders(),
  })
  return handleResponse<HostingUser>(response)
}

export async function updateProfile(data: Partial<HostingUser>): Promise<HostingUser> {
  const response = await fetch(`${HOSTING_BASE}/auth/profile`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(data),
  })
  return handleResponse<HostingUser>(response)
}

export async function changePassword(request: ChangePasswordRequest): Promise<void> {
  const response = await fetch(`${HOSTING_BASE}/auth/change-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(request),
  })
  return handleResponse<void>(response)
}

export async function requestPasswordReset(email: string): Promise<void> {
  const response = await fetch(`${HOSTING_BASE}/auth/reset-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email }),
  })
  return handleResponse<void>(response)
}

export async function confirmPasswordReset(token: string, newPassword: string): Promise<void> {
  const response = await fetch(`${HOSTING_BASE}/auth/reset-confirm`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, new_password: newPassword }),
  })
  return handleResponse<void>(response)
}

export function isAuthenticated(): boolean {
  return !!getStoredToken()
}

// ============================================================================
// Products API
// ============================================================================

export async function getProducts(): Promise<Product[]> {
  const response = await fetch(`${HOSTING_BASE}/products`)
  const data = await handleResponse<{ products: Product[] }>(response)
  return data.products
}

export async function getProduct(slug: string): Promise<Product> {
  const response = await fetch(`${HOSTING_BASE}/products/${encodeURIComponent(slug)}`)
  return handleResponse<Product>(response)
}

export async function getProductPackages(productSlug: string): Promise<Package[]> {
  const response = await fetch(`${HOSTING_BASE}/products/${encodeURIComponent(productSlug)}/packages`)
  const data = await handleResponse<{ packages: Package[] }>(response)
  return data.packages
}

export async function getPackage(packageId: string): Promise<Package> {
  const response = await fetch(`${HOSTING_BASE}/packages/${packageId}`)
  return handleResponse<Package>(response)
}

export async function getClusters(): Promise<Cluster[]> {
  const response = await fetch(`${HOSTING_BASE}/clusters`)
  const data = await handleResponse<{ clusters: Cluster[] }>(response)
  return data.clusters
}

// ============================================================================
// Orders API
// ============================================================================

export async function createOrder(request: CreateOrderRequest): Promise<SalesOrder> {
  const response = await fetch(`${HOSTING_BASE}/orders`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(request),
  })
  return handleResponse<SalesOrder>(response)
}

export async function getOrder(orderId: string): Promise<SalesOrder> {
  const response = await fetch(`${HOSTING_BASE}/orders/${orderId}`, {
    headers: authHeaders(),
  })
  return handleResponse<SalesOrder>(response)
}

export async function listOrders(): Promise<{ orders: OrderSummary[]; total: number }> {
  const response = await fetch(`${HOSTING_BASE}/orders`, {
    headers: authHeaders(),
  })
  return handleResponse<{ orders: OrderSummary[]; total: number }>(response)
}

export async function checkout(
  orderId: string,
  provider: 'stripe' | 'paystack',
  request: CheckoutRequest
): Promise<CheckoutResponse> {
  const response = await fetch(`${HOSTING_BASE}/orders/${orderId}/checkout/${provider}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(request),
  })
  return handleResponse<CheckoutResponse>(response)
}

export async function cancelOrder(orderId: string): Promise<void> {
  const response = await fetch(`${HOSTING_BASE}/orders/${orderId}/cancel`, {
    method: 'POST',
    headers: authHeaders(),
  })
  return handleResponse<void>(response)
}

export async function getPaymentConfig(): Promise<PaymentConfig> {
  const response = await fetch(`${HOSTING_BASE}/payment/config`)
  return handleResponse<PaymentConfig>(response)
}

// ============================================================================
// Instances API
// ============================================================================

export async function listInstances(): Promise<{ instances: InstanceSummary[]; total: number }> {
  const response = await fetch(`${HOSTING_BASE}/instances`, {
    headers: authHeaders(),
  })
  return handleResponse<{ instances: InstanceSummary[]; total: number }>(response)
}

export async function getInstance(instanceId: string): Promise<Instance> {
  const response = await fetch(`${HOSTING_BASE}/instances/${instanceId}`, {
    headers: authHeaders(),
  })
  return handleResponse<Instance>(response)
}

export async function getInstanceCredentials(instanceId: string): Promise<InstanceCredentials> {
  const response = await fetch(`${HOSTING_BASE}/instances/${instanceId}/credentials`, {
    headers: authHeaders(),
  })
  return handleResponse<InstanceCredentials>(response)
}

export async function getInstanceHealth(instanceId: string): Promise<InstanceHealthCheck> {
  const response = await fetch(`${HOSTING_BASE}/instances/${instanceId}/health`, {
    headers: authHeaders(),
  })
  return handleResponse<InstanceHealthCheck>(response)
}

export async function restartInstance(instanceId: string): Promise<void> {
  const response = await fetch(`${HOSTING_BASE}/instances/${instanceId}/restart`, {
    method: 'POST',
    headers: authHeaders(),
  })
  return handleResponse<void>(response)
}

export async function deleteInstance(instanceId: string): Promise<void> {
  const response = await fetch(`${HOSTING_BASE}/instances/${instanceId}`, {
    method: 'DELETE',
    headers: authHeaders(),
  })
  return handleResponse<void>(response)
}

export async function listInstanceActivities(instanceId: string): Promise<{ activities: Activity[]; total: number }> {
  const response = await fetch(`${HOSTING_BASE}/instances/${instanceId}/activities`, {
    headers: authHeaders(),
  })
  return handleResponse<{ activities: Activity[]; total: number }>(response)
}

// ============================================================================
// Subscriptions API
// ============================================================================

export async function listSubscriptions(): Promise<{ subscriptions: Subscription[]; total: number }> {
  const response = await fetch(`${HOSTING_BASE}/subscriptions`, {
    headers: authHeaders(),
  })
  return handleResponse<{ subscriptions: Subscription[]; total: number }>(response)
}

export async function getSubscription(subscriptionId: string): Promise<Subscription> {
  const response = await fetch(`${HOSTING_BASE}/subscriptions/${subscriptionId}`, {
    headers: authHeaders(),
  })
  return handleResponse<Subscription>(response)
}

export async function cancelSubscription(subscriptionId: string): Promise<void> {
  const response = await fetch(`${HOSTING_BASE}/subscriptions/${subscriptionId}/cancel`, {
    method: 'POST',
    headers: authHeaders(),
  })
  return handleResponse<void>(response)
}

// ============================================================================
// Utility functions
// ============================================================================

export function getStatusBadgeColor(status: InstanceStatus): string {
  switch (status) {
    case 'online':
      return 'green'
    case 'starting_up':
    case 'deploying':
      return 'yellow'
    case 'offline':
    case 'maintenance':
      return 'gray'
    case 'error':
      return 'red'
    case 'deleting':
    case 'deleted':
      return 'gray'
    default:
      return 'gray'
  }
}

export function getHealthBadgeColor(health: HealthStatus): string {
  switch (health) {
    case 'healthy':
      return 'green'
    case 'degraded':
      return 'yellow'
    case 'unhealthy':
      return 'red'
    default:
      return 'gray'
  }
}

export function formatPrice(amount: number, currency: string = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
  }).format(amount)
}

export function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function calculateYearlySavings(monthly: number, yearly: number): number {
  const yearlyIfMonthly = monthly * 12
  return Math.round(((yearlyIfMonthly - yearly) / yearlyIfMonthly) * 100)
}

// ============================================================================
// Admin Types
// ============================================================================

export interface AdminDashboardStats {
  total_users: number
  active_users: number
  total_instances: number
  online_instances: number
  total_orders: number
  monthly_revenue: number
  total_revenue: number
  pending_orders: number
  instances_by_status: Record<string, number>
  revenue_by_product: Record<string, number>
}

export interface AdminUserListItem {
  id: string
  email: string
  first_name: string
  last_name: string
  is_active: boolean
  is_admin: boolean
  instance_count: number
  total_spent: number
  created_at: string
  last_login_at?: string
}

export interface AdminUserListResult {
  users: AdminUserListItem[]
  total_count: number
  page: number
  page_size: number
  total_pages: number
}

export interface AdminInstanceListItem {
  id: string
  name: string
  user_id: string
  user_email: string
  product_name: string
  package_name: string
  status: string
  health_status: string
  url: string
  created_at: string
  last_health_check?: string
}

export interface AdminInstanceListResult {
  instances: AdminInstanceListItem[]
  total_count: number
  page: number
  page_size: number
  total_pages: number
}

export interface AdminOrderListItem {
  id: string
  user_id: string
  user_email: string
  product_name: string
  package_name: string
  status: string
  total_amount: number
  currency: string
  payment_method: string
  created_at: string
  paid_at?: string
}

export interface AdminOrderListResult {
  orders: AdminOrderListItem[]
  total_count: number
  page: number
  page_size: number
  total_pages: number
}

export interface RevenueDataPoint {
  date: string
  revenue: number
  orders: number
}

export interface RevenueChartData {
  data: RevenueDataPoint[]
  period: string
  group_by: string
}

export interface SystemHealth {
  status: 'healthy' | 'degraded' | 'unhealthy'
  components: Record<string, {
    status: 'healthy' | 'unhealthy'
    message?: string
    last_checked: string
  }>
  timestamp: string
}

export interface UpdateUserAdminRequest {
  is_active?: boolean
  is_admin?: boolean
}

// ============================================================================
// Admin API
// ============================================================================

export async function getAdminDashboard(): Promise<AdminDashboardStats> {
  const response = await fetch(`${HOSTING_BASE}/admin/dashboard`, {
    headers: authHeaders(),
  })
  return handleResponse<AdminDashboardStats>(response)
}

export interface AdminUserListOptions {
  page?: number
  page_size?: number
  search?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
  is_active?: boolean
  is_admin?: boolean
}

export async function getAdminUsers(options: AdminUserListOptions = {}): Promise<AdminUserListResult> {
  const params = new URLSearchParams()
  if (options.page) params.set('page', String(options.page))
  if (options.page_size) params.set('page_size', String(options.page_size))
  if (options.search) params.set('search', options.search)
  if (options.sort_by) params.set('sort_by', options.sort_by)
  if (options.sort_dir) params.set('sort_dir', options.sort_dir)
  if (options.is_active !== undefined) params.set('is_active', String(options.is_active))
  if (options.is_admin !== undefined) params.set('is_admin', String(options.is_admin))

  const response = await fetch(`${HOSTING_BASE}/admin/users?${params}`, {
    headers: authHeaders(),
  })
  return handleResponse<AdminUserListResult>(response)
}

export async function getAdminUser(userId: string): Promise<HostingUser> {
  const response = await fetch(`${HOSTING_BASE}/admin/users/${userId}`, {
    headers: authHeaders(),
  })
  return handleResponse<HostingUser>(response)
}

export async function updateAdminUser(userId: string, data: UpdateUserAdminRequest): Promise<HostingUser> {
  const response = await fetch(`${HOSTING_BASE}/admin/users/${userId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(data),
  })
  return handleResponse<HostingUser>(response)
}

export interface AdminInstanceListOptions {
  page?: number
  page_size?: number
  search?: string
  status?: string
  user_id?: string
  product?: string
}

export async function getAdminInstances(options: AdminInstanceListOptions = {}): Promise<AdminInstanceListResult> {
  const params = new URLSearchParams()
  if (options.page) params.set('page', String(options.page))
  if (options.page_size) params.set('page_size', String(options.page_size))
  if (options.search) params.set('search', options.search)
  if (options.status) params.set('status', options.status)
  if (options.user_id) params.set('user_id', options.user_id)
  if (options.product) params.set('product', options.product)

  const response = await fetch(`${HOSTING_BASE}/admin/instances?${params}`, {
    headers: authHeaders(),
  })
  return handleResponse<AdminInstanceListResult>(response)
}

export interface AdminOrderListOptions {
  page?: number
  page_size?: number
  status?: string
  user_id?: string
}

export async function getAdminOrders(options: AdminOrderListOptions = {}): Promise<AdminOrderListResult> {
  const params = new URLSearchParams()
  if (options.page) params.set('page', String(options.page))
  if (options.page_size) params.set('page_size', String(options.page_size))
  if (options.status) params.set('status', options.status)
  if (options.user_id) params.set('user_id', options.user_id)

  const response = await fetch(`${HOSTING_BASE}/admin/orders?${params}`, {
    headers: authHeaders(),
  })
  return handleResponse<AdminOrderListResult>(response)
}

export async function getAdminRevenue(
  period: '7d' | '30d' | '90d' | '1y' = '30d',
  groupBy: 'day' | 'week' | 'month' = 'day'
): Promise<RevenueChartData> {
  const params = new URLSearchParams({ period, group_by: groupBy })
  const response = await fetch(`${HOSTING_BASE}/admin/analytics/revenue?${params}`, {
    headers: authHeaders(),
  })
  return handleResponse<RevenueChartData>(response)
}

export async function getAdminHealth(): Promise<SystemHealth> {
  const response = await fetch(`${HOSTING_BASE}/admin/health`, {
    headers: authHeaders(),
  })
  return handleResponse<SystemHealth>(response)
}
