"use client"

import { useRouter } from "next/navigation"
import { useEffect } from "react"
import { AuthNudgeBanner } from "@/components/auth/auth-nudge"
import { CheckoutClient } from "@/components/payment/checkout-client"
import { useAuth } from "@/contexts/auth-context"
import { useCart } from "@/contexts/cart-context"

export default function PaymentPage() {
  const { cart } = useCart()
  const router = useRouter()
  const { accessToken } = useAuth()

  useEffect(() => {
    if (cart.items.length === 0) router.replace("/food/checkout")
  }, [cart.items.length, router])

  return (
    <div className="flex flex-1 flex-col">
      <main className="container mx-auto flex-1 overflow-y-auto px-4 pt-4 pb-28">
        <h2 className="mb-4 text-2xl">Mit TWINT bezahlen</h2>
        <AuthNudgeBanner />
        <CheckoutClient />
        {!accessToken && (
          <p className="text-muted-foreground text-xs">
            Wir erheben nur die minimal nötigen Daten, um deine Bestellung vorzubereiten und zu übergeben. Zahlungsdaten
            werden von unseren Zahlungspartnern verarbeitet.
          </p>
        )}
      </main>
    </div>
  )
}
