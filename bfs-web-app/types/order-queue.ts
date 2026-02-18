import type { Cents } from "./common"

export type OrderSyncStatus = "pending" | "syncing" | "synced" | "failed" | "paid"

export type PosPaymentMethod =
  | "cash"
  | "card"
  | "twint"
  | "gratis_guest"
  | "gratis_vip"
  | "gratis_staff"
  | "gratis_100club"

export interface GratisInfo {
  type: "guest" | "vip" | "staff" | "100club"
  elvantoPersonId?: string
  elvantoPersonName?: string
  freeQuantity?: number
}

export interface Club100DiscountItem {
  cartItemId: string
  productId: string
  productName: string
  quantity: number
  discountedQuantity: number
  unitPriceCents: number
}

export interface Club100Discount {
  person: {
    id: string
    firstName: string
    lastName: string
    remaining: number
    max: number
  }
  discountedItems: Club100DiscountItem[]
  totalDiscountCents: number
  freeProductIds: string[]
}

export interface QueuedOrderItem {
  productId: string
  quantity: number
  menuSelections?: Array<{ slotId: string; productId: string }>
}

export interface ReceiptItem {
  title: string
  quantity: number
  unitPriceCents: Cents
  configuration?: Array<{ slot: string; choice: string }>
}

export interface QueuedOrder {
  localId: string
  idempotencyKey: string
  items: QueuedOrderItem[]
  totalCents: Cents
  paymentMethod: PosPaymentMethod
  gratisInfo?: GratisInfo
  status: OrderSyncStatus
  serverId?: string
  attemptCount: number
  lastAttemptAt?: number
  lastError?: string
  receiptData?: {
    items: ReceiptItem[]
    pickupQr: string | null
    orderTimestamp: number
  }
  createdAt: number
}

export interface OrderQueueState {
  orders: QueuedOrder[]
  isOnline: boolean
}
