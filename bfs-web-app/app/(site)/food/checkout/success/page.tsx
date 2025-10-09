"use client"

import { Check } from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { getPaymentStatus } from "@/lib/api/payments"
import { addOrder } from "@/lib/orders-storage"

export default function CheckoutSuccessPage() {
  const sp = useSearchParams()
  const router = useRouter()
  const { cart, clearCart } = useCart()
  const clearedRef = useRef(false)

  // Legacy (Stripe Checkout) param
  const orderIdParam = sp.get("order_id")
  // Payment Element return params
  const pi = sp.get("payment_intent")
  const redirectStatus = sp.get("redirect_status")

  const [orderId, setOrderId] = useState<string | null>(orderIdParam)

  useEffect(() => {
    // If Stripe appended redirect_status, honor it directly
    if (redirectStatus === "failed") {
      router.replace("/food/checkout/error")
      return
    }
    if (redirectStatus === "canceled") {
      router.replace("/food/checkout/cancel")
      return
    }

    // If returning from Payment Element, confirm final status and extract order id from metadata
    const resolvePI = async () => {
      if (!pi) return
      try {
        const s = await getPaymentStatus(pi)
        if (s.status === "succeeded") {
          const oid = s.metadata?.order_id || null
          if (oid) {
            // Persist a snapshot of the cart to the order before clearing
            addOrder(oid, cart.items, cart.totalCents)
            setOrderId(oid)
          }
          if (!clearedRef.current) {
            clearedRef.current = true
            clearCart()
          }
          try {
            sessionStorage.removeItem("bfs.pi.current")
          } catch {}
        } else {
          // Determine redirect based on status if available
          if (redirectStatus === "canceled") {
            router.replace("/food/checkout/cancel")
          } else {
            router.replace("/food/checkout/error")
          }
        }
      } catch {
        router.replace("/food/checkout/error")
      }
    }
    void resolvePI()
     
  }, [pi, redirectStatus, clearCart, router, cart.items, cart.totalCents])

  // Legacy Checkout path: store order snapshot and clear cart if provided
  useEffect(() => {
    // If redirect indicated failure/cancel, do not treat legacy param as success
    if (redirectStatus === "failed") {
      router.replace("/food/checkout/error")
      return
    }
    if (redirectStatus === "canceled") {
      router.replace("/food/checkout/cancel")
      return
    }
    if (orderIdParam && !clearedRef.current) {
      // Persist items for the legacy flow
      addOrder(orderIdParam, cart.items, cart.totalCents)
      clearedRef.current = true
      clearCart()
    }
  }, [orderIdParam, redirectStatus, clearCart, cart.items, cart.totalCents, router])

  useEffect(() => {
    if (orderId) addOrder(orderId)
  }, [orderId])

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center gap-10 px-4 pb-28">
      <div className="relative h-72 w-72">
        <div className="absolute inset-0 rounded-full bg-gradient-to-br from-green-200/60 to-green-400/40 blur-sm" />
        <div className="absolute inset-8 rounded-full border-8 border-green-300/60" />
        <div className="absolute inset-16 rounded-full border-8 border-green-400/50" />
        <div className="absolute inset-24 rounded-full bg-green-500/80 flex items-center justify-center">
          <Check className="h-14 w-14 text-white" />
        </div>
      </div>
      <h1 className="text-3xl font-semibold">Bezahlung erfolgreich</h1>
      <div className="max-w-xl mx-auto fixed inset-x-0 bottom-0 p-4">
        <Button
          className="rounded-pill w-full h-12 text-base font-medium"
          onClick={() => router.push(orderId ? `/food/orders/${orderId}?from=success` : "/")}
        >
          Weiter
        </Button>
      </div>
    </div>
  )
}
