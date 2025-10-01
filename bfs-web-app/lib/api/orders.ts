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

export interface PublicOrderDetailsDTO {
  id: string
  status: OrderStatus
  totalCents: number
  createdAt: string
  items: Array<{
    id: string
    orderId: string
    productId: string
    title: string
    quantity: number
    pricePerUnitCents: number
    parentItemId?: string | null
    menuSlotId?: string | null
    menuSlotName?: string | null
    productImage?: string | null
  }>
}

export async function getOrderPublicById(orderId: string) {
  return apiRequest<PublicOrderDetailsDTO>(`/v1/orders/${encodeURIComponent(orderId)}`, {
    method: "GET",
  })
}
