import { AuthService } from "../auth"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"

// Admin API types based on backend analysis
export interface ListCustomersResponse {
  customers: AdminCustomer[]
  total: number
  limit: number
  offset: number
}

export interface AdminCustomer {
  id: string
  email: string
  name: string
  isActive: boolean
  isEmailVerified: boolean
  createdAt: string
  updatedAt: string
}

export interface BanCustomerRequest {
  banned: boolean
  reason?: string
}

export interface BanCustomerResponse {
  id: string
  isActive: boolean
  message: string
}

export interface DeleteCustomerResponse {
  message: string
}

export interface InviteAdminRequest {
  email: string
  name: string
}

export interface InviteAdminResponse {
  message: string
  adminId: string
}

async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const accessToken = await AuthService.getAccessToken()

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(accessToken && { Authorization: `Bearer ${accessToken}` }),
      ...options.headers,
    },
  })

  if (!response.ok) {
    const error = await response.text()
    throw new Error(error || `HTTP ${response.status}`)
  }

  const data = (await response.json()) as unknown
  return data as T
}

// Customer Management
export async function listCustomers(params?: { limit?: number; offset?: number }): Promise<ListCustomersResponse> {
  const searchParams = new URLSearchParams()
  if (params?.limit) searchParams.set("limit", params.limit.toString())
  if (params?.offset) searchParams.set("offset", params.offset.toString())

  const queryString = searchParams.toString()
  const endpoint = `/v1/admin/customers${queryString ? `?${queryString}` : ""}`

  return apiRequest<ListCustomersResponse>(endpoint)
}

export async function banCustomer(id: string, request: BanCustomerRequest): Promise<BanCustomerResponse> {
  return apiRequest<BanCustomerResponse>(`/v1/admin/customers/${id}/ban`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

export async function deleteCustomer(id: string): Promise<DeleteCustomerResponse> {
  return apiRequest<DeleteCustomerResponse>(`/v1/admin/customers/${id}`, {
    method: "DELETE",
  })
}

// Admin Management
export async function inviteAdmin(request: InviteAdminRequest): Promise<InviteAdminResponse> {
  return apiRequest<InviteAdminResponse>("/v1/admin/invites", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

const AdminAPI = {
  listCustomers,
  banCustomer,
  deleteCustomer,
  inviteAdmin,
}

export default AdminAPI
