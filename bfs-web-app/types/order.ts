// Order types matching backend API

export type OrderStatus = "pending" | "paid" | "cancelled" | "refunded"

export interface Order {
  id: string
  customerId?: string
  contactEmail?: string
  total: number
  status: OrderStatus
  createdAt: string
  updatedAt: string
}

// Order API requests/responses
export interface CreateOrderRequest {
  items: OrderItem[]
  contactEmail?: string
  total: number
}

export interface CreateOrderResponse {
  id: string
  customerId?: string
  contactEmail?: string
  total: number
  status: OrderStatus
  createdAt: string
  updatedAt: string
}

export interface UpdateOrderRequest {
  status?: OrderStatus
  contactEmail?: string
}

export interface UpdateOrderResponse {
  id: string
  customerId?: string
  contactEmail?: string
  total: number
  status: OrderStatus
  createdAt: string
  updatedAt: string
}

export interface GetOrderResponse {
  id: string
  customerId?: string
  contactEmail?: string
  total: number
  status: OrderStatus
  createdAt: string
  updatedAt: string
}

export interface ListOrdersResponse {
  orders: Order[]
  total: number
  limit: number
  offset: number
}

export interface DeleteOrderResponse {
  message: string
}

export interface UpdateOrderStatusRequest {
  status: OrderStatus
}

export interface UpdateOrderStatusResponse {
  id: string
  status: OrderStatus
  message: string
}

// Order item for cart functionality
export interface OrderItem {
  productId: string
  quantity: number
  price: number
  name?: string
}

// Cart context types
export interface CartItem extends OrderItem {
  name: string
  description?: string
}

export interface CartContext {
  items: CartItem[]
  total: number
  addItem: (item: CartItem) => void
  removeItem: (productId: string) => void
  updateQuantity: (productId: string, quantity: number) => void
  clearCart: () => void
  getItemCount: () => number
}

// Redemption types (from backend analysis)
export interface RedemptionItem {
  productId: string
  quantity: number
}

export interface RedeemOrderItemsRequest {
  orderId: string
  items: RedemptionItem[]
}

export interface RedeemOrderItemsResponse {
  message: string
  redeemedItems: RedemptionItem[]
}
