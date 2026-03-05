"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef } from "react"
import { useCart } from "@/contexts/cart-context"
import { getPaymentStatus } from "@/lib/api/payments"
import { addOrder } from "@/lib/orders-storage"

const MAX_POLLS = 8
const POLL_INTERVAL_MS = 1500

export default function PaymentNextClient() {
  const sp = useSearchParams()
  const router = useRouter()
  const { clearCart } = useCart()
  const clearedRef = useRef(false)
  const pollingRef = useRef(false)

  useEffect(() => {
    const result = sp.get("result")
    const orderIdParam = sp.get("order_id")

    if (result === "cancel") {
      router.replace("/food/checkout/cancel")
      return
    }
    if (result === "failed") {
      router.replace("/food/checkout/error")
      return
    }

    const resolve = async () => {
      if (orderIdParam) {
        await pollOrder(orderIdParam)
        return
      }

      try {
        const raw = sessionStorage.getItem("bfs.pending_order")
        if (raw) {
          const parsed = JSON.parse(raw) as { orderId?: string }
          if (parsed?.orderId) {
            await pollOrder(parsed.orderId)
            return
          }
        }
      } catch {}

      router.replace("/food/checkout")
    }

    const pollOrder = async (orderId: string) => {
      if (pollingRef.current) return
      pollingRef.current = true

      for (let attempt = 0; attempt < MAX_POLLS; attempt++) {
        try {
          const res = await getPaymentStatus(orderId)

          if (res.status === "paid") {
            if (!clearedRef.current) {
              clearedRef.current = true
              addOrder(orderId)
              clearCart()
            }
            try {
              sessionStorage.removeItem("bfs.pending_order")
            } catch {}
            router.replace(`/food/checkout/success?order_id=${encodeURIComponent(orderId)}`)
            return
          }

          if (res.status === "cancelled") {
            router.replace("/food/checkout/cancel")
            return
          }

          if (res.status !== "pending") {
            router.replace("/food/checkout/error")
            return
          }
        } catch {
          router.replace("/food/checkout/error")
          return
        }

        if (attempt < MAX_POLLS - 1) {
          await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS))
        }
      }

      router.replace("/food/checkout/error")
    }

    void resolve()
  }, [sp, clearCart, router])

  return (
    <div className="flex min-h-[70vh] items-center justify-center">
      <p className="text-muted-foreground">Zahlungsstatus wird geprüft…</p>
    </div>
  )
}
