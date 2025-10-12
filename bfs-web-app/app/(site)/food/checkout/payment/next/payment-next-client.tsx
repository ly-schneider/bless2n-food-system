"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef } from "react"
import { useCart } from "@/contexts/cart-context"
import { getPaymentStatus } from "@/lib/api/payments"
import { addOrder } from "@/lib/orders-storage"

export default function PaymentNextClient() {
  const sp = useSearchParams()
  const router = useRouter()
  const { cart, clearCart } = useCart()
  const clearedRef = useRef(false)

  useEffect(() => {
    const redirectStatus = sp.get("redirect_status")
    // If Stripe already indicates the outcome, route immediately
    if (redirectStatus === "failed") {
      router.replace("/food/checkout/error")
      return
    }
    if (redirectStatus === "canceled") {
      router.replace("/food/checkout/cancel")
      return
    }

    const resolve = async () => {
      // PI might be provided by Stripe or our own param `pi`
      const pi = sp.get("payment_intent") || sp.get("pi")
      if (!pi) {
        // Fall back to session if present
        try {
          const raw = sessionStorage.getItem("bfs.pi.current")
          if (raw) {
            const parsed = JSON.parse(raw) as { pi?: string }
            if (parsed?.pi) {
              await handlePI(parsed.pi)
              return
            }
          }
        } catch {}
        // No PI: go back to checkout
        router.replace("/food/checkout")
        return
      }
      await handlePI(pi)
    }

    const handlePI = async (pi: string) => {
      try {
        const s = await getPaymentStatus(pi)
        if (s.status === "succeeded") {
          const oid = s.metadata?.order_id || null
          if (oid) {
            // Persist a snapshot of the cart to the order before clearing
            addOrder(oid, cart.items, cart.totalCents)
          }
          if (!clearedRef.current) {
            clearedRef.current = true
            clearCart()
          }
          try {
            sessionStorage.removeItem("bfs.pi.current")
          } catch {}
          // Prefer explicit order id passthrough for success view
          router.replace(oid ? `/food/checkout/success?order_id=${encodeURIComponent(oid)}` : "/food/checkout/success")
        } else if (s.status === "canceled") {
          router.replace("/food/checkout/cancel")
        } else {
          // requires_payment_method or any other non-success -> error
          router.replace("/food/checkout/error")
        }
      } catch {
        router.replace("/food/checkout/error")
      }
    }

    void resolve()
    // We intentionally depend only on search params and cart snapshot values
  }, [sp])

  return (
    <div className="flex min-h-[70vh] items-center justify-center">
      <p className="text-muted-foreground">Zahlungsstatus wird geprüft…</p>
    </div>
  )
}
