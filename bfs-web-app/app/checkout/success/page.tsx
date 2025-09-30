"use client"

import { useEffect, useRef } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { addOrder } from "@/lib/orders-storage"

export default function CheckoutSuccessPage() {
  const sp = useSearchParams()
  const orderId = sp.get("order_id")
  const { clearCart } = useCart()
  const clearedRef = useRef(false)
  const router = useRouter()

  useEffect(() => {
    if (!clearedRef.current) {
      clearedRef.current = true
      clearCart()
    }
  }, [clearCart])

  // Persist order id for anonymous users
  useEffect(() => {
    if (orderId) {
      addOrder(orderId)
    }
  }, [orderId])

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center gap-10 px-4 pb-28">
      <div className="relative h-72 w-72">
        {/* Concentric rings */}
        <div className="absolute inset-0 rounded-full bg-gradient-to-br from-green-200/60 to-green-400/40 blur-sm" />
        <div className="absolute inset-8 rounded-full border-8 border-green-300/60" />
        <div className="absolute inset-16 rounded-full border-8 border-green-400/50" />
        <div className="absolute inset-24 rounded-full bg-green-500/80 flex items-center justify-center">
          <Check className="h-14 w-14 text-white" />
        </div>
      </div>
      <h1 className="text-3xl font-semibold">Bezahlung erfolgreich</h1>
      {/* Bottom fixed full-width action bar */}
      <div className="fixed inset-x-0 bottom-0 p-4">
        <Button
          className="rounded-pill w-full h-12 text-base font-medium"
          onClick={() => router.push(orderId ? `/checkout/qr?order_id=${orderId}&from=success` : "/")}
        >
          Weiter
        </Button>
      </div>
    </div>
  )
}
