"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { OrderQueueManager } from "@/lib/pos/order-queue"
import type { GratisInfo, PosPaymentMethod, QueuedOrder, QueuedOrderItem, ReceiptItem } from "@/types/order-queue"

interface UseOrderQueueOptions {
  token: string
  deviceId: string
  onOrderSynced?: (order: QueuedOrder) => void
  onPaymentError?: (order: QueuedOrder, errorCode: string, errorMessage: string) => void
}

export function useOrderQueue({ token, deviceId, onOrderSynced, onPaymentError }: UseOrderQueueOptions) {
  const managerRef = useRef<OrderQueueManager | null>(null)
  const [orders, setOrders] = useState<QueuedOrder[]>([])
  const [isOnline, setIsOnline] = useState(true)

  useEffect(() => {
    if (!token || !deviceId) return

    const manager = new OrderQueueManager(token, deviceId)
    managerRef.current = manager

    const unsubState = manager.onStateChange((state) => {
      setOrders(state.orders)
      setIsOnline(state.isOnline)
    })

    return () => {
      unsubState()
      manager.destroy()
      managerRef.current = null
    }
  }, [token, deviceId])

  useEffect(() => {
    if (!managerRef.current || !onOrderSynced) return

    return managerRef.current.onSync(onOrderSynced)
  }, [onOrderSynced])

  useEffect(() => {
    if (!managerRef.current || !onPaymentError) return

    return managerRef.current.onPaymentError(onPaymentError)
  }, [onPaymentError])

  const submitOrder = useCallback(
    (
      items: QueuedOrderItem[],
      totalCents: number,
      paymentMethod: PosPaymentMethod,
      receiptData?: { items: ReceiptItem[]; pickupQr: string | null },
      gratisInfo?: GratisInfo
    ): QueuedOrder | null => {
      if (!managerRef.current) return null
      return managerRef.current.enqueue(items, totalCents, paymentMethod, receiptData, gratisInfo)
    },
    []
  )

  const retryOrder = useCallback((localId: string) => {
    managerRef.current?.retryOrder(localId)
  }, [])

  const deleteOrder = useCallback((localId: string) => {
    managerRef.current?.deleteOrder(localId)
  }, [])

  const getOrder = useCallback((localId: string): QueuedOrder | undefined => {
    return managerRef.current?.getOrder(localId)
  }, [])

  const changePaymentMethod = useCallback((localId: string, newMethod: PosPaymentMethod, gratisInfo?: GratisInfo) => {
    managerRef.current?.changePaymentMethod(localId, newMethod, gratisInfo)
  }, [])

  const pendingCount = orders.filter((o) => o.status === "pending" || o.status === "syncing").length
  const failedOrders = orders.filter((o) => o.status === "failed" && o.attemptCount >= 5)

  return {
    orders,
    isOnline,
    pendingCount,
    failedOrders,
    submitOrder,
    retryOrder,
    deleteOrder,
    getOrder,
    changePaymentMethod,
  }
}
