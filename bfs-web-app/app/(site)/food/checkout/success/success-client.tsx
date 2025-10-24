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
  const { cart, clearCart } = useCart()
  const clearedRef = useRef(false)

  const orderIdParam = sp.get("order_id")
  const pi = sp.get("payment_intent")
  const redirectStatus = sp.get("redirect_status")

  const [orderId, setOrderId] = useState<string | null>(orderIdParam)

  useEffect(() => {
    if (redirectStatus === "failed") {
      router.replace("/food/checkout/error")
      return
    }
    if (redirectStatus === "canceled") {
      router.replace("/food/checkout/cancel")
      return
    }

    const resolvePI = async () => {
      if (!pi) return
      try {
        const s = await getPaymentStatus(pi)
        if (s.status === "succeeded") {
          const oid = s.metadata?.order_id || null
          if (oid) {
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
    resolvePI()
  }, [pi, redirectStatus, clearCart, router, cart.items, cart.totalCents])

  useEffect(() => {
    if (redirectStatus === "failed") {
      router.replace("/food/checkout/error")
      return
    }
    if (redirectStatus === "canceled") {
      router.replace("/food/checkout/cancel")
      return
    }
    if (orderIdParam && !clearedRef.current) {
      addOrder(orderIdParam, cart.items, cart.totalCents)
      clearedRef.current = true
      clearCart()
    }
  }, [orderIdParam, redirectStatus, clearCart, cart.items, cart.totalCents, router])

  useEffect(() => {
    if (orderId) addOrder(orderId)
  }, [orderId])

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
