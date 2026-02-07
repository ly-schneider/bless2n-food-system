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
  return apiRequest<ListResponse<OrderSummaryDTO>>("/v1/orders?scope=mine", {
    method: "GET",
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
  })
}

export interface OrderLineDTO {
  id: string
  orderId: string
  lineType: "simple" | "bundle" | "component"
  productId: string
  title: string
  quantity: number
  unitPriceCents: number
  parentLineId?: string | null
  menuSlotId?: string | null
  menuSlotName?: string | null
  productImage?: string | null
  childLines?: OrderLineDTO[] | null
}

export interface PublicOrderDetailsDTO {
  id: string
  status: OrderStatus
  totalCents: number
  createdAt: string
  lines?: OrderLineDTO[]
}

export async function getOrderPublicById(orderId: string) {
  return apiRequest<PublicOrderDetailsDTO>(`/v1/orders/${encodeURIComponent(orderId)}`, {
    method: "GET",
  })
}
