"use client"

import { Check } from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { getPaymentStatus } from "@/lib/api/payments"
import { addOrder } from "@/lib/orders-storage"

export default function CheckoutSuccessClient() {
  const sp = useSearchParams()
  const router = useRouter()
  const { clearCart } = useCart()
  const clearedRef = useRef(false)

  const orderIdParam = sp.get("order_id")

  const [orderId, setOrderId] = useState<string | null>(orderIdParam)
  const [loading, setLoading] = useState(!orderIdParam)

  useEffect(() => {
    const resolvePendingOrder = async () => {
      if (orderIdParam) {
        setOrderId(orderIdParam)
        setLoading(false)

        try {
          const status = await getPaymentStatus(orderIdParam)
          if (status.status !== "paid" && status.status !== "pending") {
            router.replace("/food/checkout/cancel")
            return
          }
        } catch {}

        if (!clearedRef.current) {
          clearedRef.current = true
          addOrder(orderIdParam)
          clearCart()
        }
        try {
          sessionStorage.removeItem("bfs.pending_order")
        } catch {}
        return
      }

      try {
        const raw = sessionStorage.getItem("bfs.pending_order")
        if (raw) {
          const parsed = JSON.parse(raw) as { orderId?: string }
          if (parsed?.orderId) {
            setOrderId(parsed.orderId)

            if (!clearedRef.current) {
              clearedRef.current = true
              addOrder(parsed.orderId)
              clearCart()
            }
            sessionStorage.removeItem("bfs.pending_order")
            setLoading(false)
            return
          }
        }
      } catch {}

      setLoading(false)
    }

    resolvePendingOrder()
  }, [orderIdParam, clearCart, router])

  if (loading) {
    return (
      <div className="flex min-h-[70vh] items-center justify-center">
        <p className="text-muted-foreground">Zahlungsstatus wird geprüft…</p>
      </div>
    )
  }

  return (
    <div className="flex min-h-[70vh] flex-col items-center justify-center gap-10 px-4 pb-28">
      <div className="relative h-64 w-64 sm:h-72 sm:w-72">
        <div className="absolute inset-0 rounded-full bg-gradient-to-br from-green-200/60 to-green-400/40 blur-sm" />
        <div className="absolute inset-8 rounded-full border-8 border-green-300/60" />
        <div className="absolute inset-16 rounded-full border-8 border-green-400/50" />
        <div className="absolute inset-24 flex items-center justify-center rounded-full bg-green-500/80">
          <Check className="h-10 w-10 text-white sm:h-14 sm:w-14" />
        </div>
      </div>
      <h1 className="text-center text-2xl font-semibold sm:text-3xl">Bezahlung erfolgreich</h1>
      <div className="fixed inset-x-0 bottom-0 mx-auto max-w-xl p-4">
        <Button
          className="rounded-pill h-12 w-full text-base font-medium"
          onClick={() => router.push(orderId ? `/food/orders/${orderId}?from=success` : "/")}
        >
          Weiter
        </Button>
      </div>
    </div>
  )
}
