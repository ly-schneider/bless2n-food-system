"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef } from "react"
import { useCart } from "@/contexts/cart-context"
import { getPaymentStatus } from "@/lib/api/payments"
import { addOrder } from "@/lib/orders-storage"

export default function PaymentNextClient() {
  const sp = useSearchParams()
  const router = useRouter()
  const { clearCart } = useCart()
  const clearedRef = useRef(false)

  useEffect(() => {
    const resolve = async () => {
      const orderIdParam = sp.get("order_id")

      if (orderIdParam) {
        await handleOrder(orderIdParam)
        return
      }

      try {
        const raw = sessionStorage.getItem("bfs.pending_order")
        if (raw) {
          const parsed = JSON.parse(raw) as { orderId?: string }
          if (parsed?.orderId) {
            await handleOrder(parsed.orderId)
            return
          }
        }
      } catch {}

      router.replace("/food/checkout")
    }

    const handleOrder = async (orderId: string) => {
      try {
        const status = await getPaymentStatus(orderId)
        if (status.status === "paid" || status.status === "pending") {
          if (!clearedRef.current) {
            clearedRef.current = true
            addOrder(orderId)
            clearCart()
          }
          try {
            sessionStorage.removeItem("bfs.pending_order")
          } catch {}
          router.replace(`/food/checkout/success?order_id=${encodeURIComponent(orderId)}`)
        } else if (status.status === "cancelled") {
          router.replace("/food/checkout/cancel")
        } else {
          router.replace("/food/checkout/error")
        }
      } catch {
        router.replace("/food/checkout/error")
      }
    }

    void resolve()
  }, [sp, clearCart, router])

  return (
    <div className="flex min-h-[70vh] items-center justify-center">
      <p className="text-muted-foreground">Zahlungsstatus wird geprüft…</p>
    </div>
  )
}
