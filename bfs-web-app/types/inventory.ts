export type InventoryReason = "opening_balance" | "sale" | "refund" | "manual_adjust" | "correction"

export interface InventoryLedger {
  id: string
  productId: string
  delta: number
  reason: InventoryReason
  createdAt: string // ISO date
}
