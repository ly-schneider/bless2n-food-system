import {
  CreateOrderRequest,
  CreateOrderResponse,
  DeleteOrderResponse,
  GetOrderResponse,
  ListOrdersResponse,
  UpdateOrderRequest,
  UpdateOrderResponse,
  UpdateOrderStatusRequest,
  UpdateOrderStatusResponse,
} from "@/types"
import { AuthService } from "../auth"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"

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

// Order API methods
export async function createOrder(request: CreateOrderRequest): Promise<CreateOrderResponse> {
  return apiRequest<CreateOrderResponse>("/v1/orders", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

export async function getOrder(id: string): Promise<GetOrderResponse> {
  return apiRequest<GetOrderResponse>(`/v1/orders/${id}`)
}

export async function updateOrder(id: string, request: UpdateOrderRequest): Promise<UpdateOrderResponse> {
  return apiRequest<UpdateOrderResponse>(`/v1/orders/${id}`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

export async function deleteOrder(id: string): Promise<DeleteOrderResponse> {
  return apiRequest<DeleteOrderResponse>(`/v1/orders/${id}`, {
    method: "DELETE",
  })
}

export async function listOrders(params?: {
  customerId?: string
  status?: string
  limit?: number
  offset?: number
}): Promise<ListOrdersResponse> {
  const searchParams = new URLSearchParams()
  if (params?.customerId) searchParams.set("customer_id", params.customerId)
  if (params?.status) searchParams.set("status", params.status)
  if (params?.limit) searchParams.set("limit", params.limit.toString())
  if (params?.offset) searchParams.set("offset", params.offset.toString())

  const queryString = searchParams.toString()
  const endpoint = `/v1/orders${queryString ? `?${queryString}` : ""}`

  return apiRequest<ListOrdersResponse>(endpoint)
}

export async function getMyOrders(params?: { limit?: number; offset?: number }): Promise<ListOrdersResponse> {
  const searchParams = new URLSearchParams()
  if (params?.limit) searchParams.set("limit", params.limit.toString())
  if (params?.offset) searchParams.set("offset", params.offset.toString())

  const queryString = searchParams.toString()
  const endpoint = `/v1/orders/my${queryString ? `?${queryString}` : ""}`

  return apiRequest<ListOrdersResponse>(endpoint)
}

// Admin Order API methods
export async function updateOrderStatus(
  id: string,
  request: UpdateOrderStatusRequest
): Promise<UpdateOrderStatusResponse> {
  return apiRequest<UpdateOrderStatusResponse>(`/v1/admin/orders/${id}/status`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

const OrderAPI = {
  createOrder,
  getOrder,
  updateOrder,
  deleteOrder,
  listOrders,
  getMyOrders,
  updateOrderStatus,
}

export default OrderAPI
