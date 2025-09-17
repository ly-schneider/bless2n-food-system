import { BaseEntity, ListResponse, MessageResponse } from './common'

export type OrderStatus = "pending" | "paid" | "cancelled" | "refunded"

export interface OrderItem {
  productId: string
  quantity: number
  price: number
  name?: string
}

export interface Order extends BaseEntity {
  customerId?: string
  contactEmail?: string
  total: number
  status: OrderStatus
}

export interface CreateOrderRequest {
  items: OrderItem[]
  contactEmail?: string
  total: number
}

export interface UpdateOrderRequest {
  status?: OrderStatus
  contactEmail?: string
}

export interface UpdateOrderStatusRequest {
  status: OrderStatus
}

// Response types
export type CreateOrderResponse = Order
export type UpdateOrderResponse = Order
export type GetOrderResponse = Order
export type ListOrdersResponse = ListResponse<Order>
export type DeleteOrderResponse = MessageResponse

export interface UpdateOrderStatusResponse extends MessageResponse {
  id: string
  status: OrderStatus
}

// Redemption types
export interface RedemptionItem {
  productId: string
  quantity: number
}

export interface RedeemOrderItemsRequest {
  orderId: string
  items: RedemptionItem[]
}

export interface RedeemOrderItemsResponse extends MessageResponse {
  redeemedItems: RedemptionItem[]
}

// Cart context types for compatibility
export interface OrderCartItem extends OrderItem {
  name: string
  description?: string
}

export interface CartContext {
  items: OrderCartItem[]
  total: number
  addItem: (item: OrderCartItem) => void
  removeItem: (productId: string) => void
  updateQuantity: (productId: string, quantity: number) => void
  clearCart: () => void
  getItemCount: () => number
}