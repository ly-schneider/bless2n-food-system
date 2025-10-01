import { apiRequest } from "../api"

export type OrderStatus = "pending" | "paid" | "cancelled" | "refunded"

export interface OrderSummaryDTO {
  id: string
  status: OrderStatus
  createdAt: string
}

export interface ListResponse<T> {
  items: T[]
  count: number
}

export async function listMyOrders(accessToken?: string) {
  return apiRequest<ListResponse<OrderSummaryDTO>>(
    "/v1/orders",
    {
      method: "GET",
      headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    }
  )
}

