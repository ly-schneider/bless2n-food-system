import type { GratisInfo, OrderSyncStatus, PosPaymentMethod, QueuedOrder, QueuedOrderItem, ReceiptItem } from "@/types/order-queue"
import { generateLocalId } from "./idempotency"

const STORAGE_KEY = "bfs-pos-order-queue"
const MAX_RETRY_ATTEMPTS = 5
const RETRY_DELAYS = [2000, 4000, 8000, 16000, 32000]
const SYNC_INTERVAL = 30000

const PERMANENT_ERROR_CODES = ["product_not_free", "insufficient_remaining_redemptions"]

type SyncListener = (order: QueuedOrder) => void
type StateListener = (state: { orders: QueuedOrder[]; isOnline: boolean }) => void
type PaymentErrorListener = (order: QueuedOrder, errorCode: string, errorMessage: string) => void

export class OrderQueueManager {
  private orders: QueuedOrder[] = []
  private isOnline = true
  private syncTimeout: ReturnType<typeof setTimeout> | null = null
  private syncListeners: Set<SyncListener> = new Set()
  private stateListeners: Set<StateListener> = new Set()
  private paymentErrorListeners: Set<PaymentErrorListener> = new Set()
  private token: string
  private deviceId: string
  private isSyncing = false

  constructor(token: string, deviceId: string) {
    this.token = token
    this.deviceId = deviceId
    this.loadFromStorage()
    this.setupNetworkListeners()
    this.startPeriodicSync()
  }

  private loadFromStorage() {
    try {
      const stored = sessionStorage.getItem(STORAGE_KEY)
      if (stored) {
        const parsed = JSON.parse(stored) as QueuedOrder[]
        this.orders = parsed.filter((o) => o.status !== "paid")
      }
    } catch {
      this.orders = []
    }
  }

  private saveToStorage() {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(this.orders))
    } catch {}
  }

  private setupNetworkListeners() {
    if (typeof window === "undefined") return

    const updateOnline = () => {
      this.isOnline = navigator.onLine
      this.notifyStateListeners()
      if (this.isOnline) this.triggerSync()
    }

    window.addEventListener("online", updateOnline)
    window.addEventListener("offline", updateOnline)
    this.isOnline = navigator.onLine
  }

  private startPeriodicSync() {
    if (typeof window === "undefined") return

    const scheduleNext = () => {
      this.syncTimeout = setTimeout(() => {
        this.triggerSync()
        scheduleNext()
      }, SYNC_INTERVAL)
    }
    scheduleNext()
  }

  private notifyStateListeners() {
    const state = { orders: [...this.orders], isOnline: this.isOnline }
    Array.from(this.stateListeners).forEach((listener) => {
      try {
        listener(state)
      } catch {}
    })
  }

  private notifySyncListeners(order: QueuedOrder) {
    Array.from(this.syncListeners).forEach((listener) => {
      try {
        listener(order)
      } catch {}
    })
  }

  private notifyPaymentErrorListeners(order: QueuedOrder, errorCode: string, errorMessage: string) {
    Array.from(this.paymentErrorListeners).forEach((listener) => {
      try {
        listener(order, errorCode, errorMessage)
      } catch {}
    })
  }

  enqueue(
    items: QueuedOrderItem[],
    totalCents: number,
    paymentMethod: PosPaymentMethod,
    receiptData?: { items: ReceiptItem[]; pickupQr: string | null },
    gratisInfo?: GratisInfo
  ): QueuedOrder {
    const now = Date.now()
    const localId = generateLocalId()

    const order: QueuedOrder = {
      localId,
      idempotencyKey: localId,
      items,
      totalCents,
      paymentMethod,
      gratisInfo,
      status: "pending",
      attemptCount: 0,
      createdAt: now,
      receiptData: receiptData
        ? {
            ...receiptData,
            orderTimestamp: now,
          }
        : undefined,
    }

    this.orders.push(order)
    this.saveToStorage()
    this.notifyStateListeners()
    this.triggerSync()

    return order
  }

  triggerSync() {
    if (this.isSyncing || !this.isOnline) return

    const pending = this.orders.filter(
      (o) => o.status === "pending" || (o.status === "failed" && o.attemptCount < MAX_RETRY_ATTEMPTS)
    )
    if (pending.length === 0) return

    this.isSyncing = true
    this.syncNextOrder(pending, 0)
  }

  private async syncNextOrder(pending: QueuedOrder[], index: number) {
    if (index >= pending.length) {
      this.isSyncing = false
      return
    }

    const order = pending[index]
    if (!order) {
      this.isSyncing = false
      return
    }

    const shouldRetry =
      order.status === "failed" &&
      order.lastAttemptAt &&
      Date.now() - order.lastAttemptAt < RETRY_DELAYS[Math.min(order.attemptCount - 1, RETRY_DELAYS.length - 1)]!

    if (shouldRetry) {
      this.syncNextOrder(pending, index + 1)
      return
    }

    await this.syncOrder(order)
    this.syncNextOrder(pending, index + 1)
  }

  private async syncOrder(order: QueuedOrder) {
    const idx = this.orders.findIndex((o) => o.localId === order.localId)
    if (idx === -1) return

    this.orders[idx] = { ...order, status: "syncing" as OrderSyncStatus }
    this.saveToStorage()
    this.notifyStateListeners()

    try {
      const resOrder = await fetch(`/api/v1/orders`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.token}`,
          "Idempotency-Key": order.idempotencyKey,
        },
        body: JSON.stringify({ items: order.items }),
      })

      const orderJson = (await resOrder.json()) as { id?: string; detail?: string }
      if (!resOrder.ok || !orderJson.id) {
        throw new Error(orderJson.detail || "order_failed")
      }

      const serverId = orderJson.id
      this.orders[idx] = { ...this.orders[idx]!, serverId, status: "synced" }
      this.saveToStorage()
      this.notifyStateListeners()

      const paymentKey = `${order.idempotencyKey}:pay`
      const paymentBody: Record<string, unknown> = { method: order.paymentMethod, channel: "pos" }
      if (order.gratisInfo && order.gratisInfo.type === "100club") {
        paymentBody.club100 = {
          elvantoPersonId: order.gratisInfo.elvantoPersonId,
          elvantoPersonName: order.gratisInfo.elvantoPersonName,
          freeQuantity: order.gratisInfo.freeQuantity || 1,
        }
      }
      const resPay = await fetch(`/api/v1/orders/${serverId}/payment`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.token}`,
          "Idempotency-Key": paymentKey,
        },
        body: JSON.stringify(paymentBody),
      })

      const payJson = (await resPay.json()) as { code?: string; message?: string; detail?: string }
      if (!resPay.ok) {
        const errorCode = payJson.code || "payment_failed"
        const errorMessage = payJson.message || payJson.detail || "payment_failed"

        if (PERMANENT_ERROR_CODES.includes(errorCode)) {
          this.orders[idx] = {
            ...this.orders[idx]!,
            status: "failed",
            attemptCount: MAX_RETRY_ATTEMPTS,
            lastAttemptAt: Date.now(),
            lastError: errorCode,
          }
          this.saveToStorage()
          this.notifyStateListeners()
          this.notifyPaymentErrorListeners(this.orders[idx]!, errorCode, errorMessage)
          return
        }

        throw new Error(errorMessage)
      }

      this.orders[idx] = { ...this.orders[idx]!, status: "paid" }
      this.saveToStorage()
      this.notifyStateListeners()
      this.notifySyncListeners(this.orders[idx]!)
    } catch (e) {
      const currentOrder = this.orders[idx]!
      this.orders[idx] = {
        ...currentOrder,
        status: "failed",
        attemptCount: currentOrder.attemptCount + 1,
        lastAttemptAt: Date.now(),
        lastError: e instanceof Error ? e.message : "sync_failed",
      }
      this.saveToStorage()
      this.notifyStateListeners()

      if (this.orders[idx]!.attemptCount < MAX_RETRY_ATTEMPTS) {
        const delay = RETRY_DELAYS[Math.min(this.orders[idx]!.attemptCount - 1, RETRY_DELAYS.length - 1)]!
        setTimeout(() => this.triggerSync(), delay)
      }
    }
  }

  getOrder(localId: string): QueuedOrder | undefined {
    return this.orders.find((o) => o.localId === localId)
  }

  getPendingOrders(): QueuedOrder[] {
    return this.orders.filter((o) => o.status === "pending" || o.status === "syncing")
  }

  getFailedOrders(): QueuedOrder[] {
    return this.orders.filter((o) => o.status === "failed" && o.attemptCount >= MAX_RETRY_ATTEMPTS)
  }

  retryOrder(localId: string) {
    const idx = this.orders.findIndex((o) => o.localId === localId)
    if (idx === -1) return

    this.orders[idx] = { ...this.orders[idx]!, status: "pending", attemptCount: 0, lastError: undefined }
    this.saveToStorage()
    this.notifyStateListeners()
    this.triggerSync()
  }

  deleteOrder(localId: string) {
    this.orders = this.orders.filter((o) => o.localId !== localId)
    this.saveToStorage()
    this.notifyStateListeners()
  }

  clearPaidOrders() {
    this.orders = this.orders.filter((o) => o.status !== "paid")
    this.saveToStorage()
    this.notifyStateListeners()
  }

  onSync(listener: SyncListener): () => void {
    this.syncListeners.add(listener)
    return () => this.syncListeners.delete(listener)
  }

  onPaymentError(listener: PaymentErrorListener): () => void {
    this.paymentErrorListeners.add(listener)
    return () => this.paymentErrorListeners.delete(listener)
  }

  onStateChange(listener: StateListener): () => void {
    this.stateListeners.add(listener)
    listener({ orders: [...this.orders], isOnline: this.isOnline })
    return () => this.stateListeners.delete(listener)
  }

  changePaymentMethod(localId: string, newMethod: PosPaymentMethod, gratisInfo?: GratisInfo) {
    const idx = this.orders.findIndex((o) => o.localId === localId)
    if (idx === -1) return

    this.orders[idx] = {
      ...this.orders[idx]!,
      paymentMethod: newMethod,
      gratisInfo: gratisInfo,
      status: "pending",
      attemptCount: 0,
      lastError: undefined,
      idempotencyKey: `${this.orders[idx]!.idempotencyKey}:retry${Date.now()}`,
    }
    this.saveToStorage()
    this.notifyStateListeners()
    this.triggerSync()
  }

  getState() {
    return { orders: [...this.orders], isOnline: this.isOnline }
  }

  destroy() {
    if (this.syncTimeout) {
      clearTimeout(this.syncTimeout)
    }
    this.syncListeners.clear()
    this.stateListeners.clear()
    this.paymentErrorListeners.clear()
  }
}
