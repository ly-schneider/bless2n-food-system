import { WebSocketMessage } from './common'

export interface OrderUpdateMessage {
  orderId: string
  status: string
  estimatedTime?: number
}

export interface CartSyncMessage {
  cartId: string
  items: unknown[]
  total: number
}

export type OrderWSMessage = WebSocketMessage<OrderUpdateMessage>
export type CartWSMessage = WebSocketMessage<CartSyncMessage>