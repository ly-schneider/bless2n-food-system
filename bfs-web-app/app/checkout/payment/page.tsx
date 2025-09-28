"use client"

import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import Header from "@/components/layout/header"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { createCheckoutSession } from "@/lib/api/payments"
import { formatChf } from "@/lib/utils"

export default function PaymentPage() {
  const { cart } = useCart()
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (cart.items.length === 0) {
      router.replace("/checkout")
    }
  }, [cart.items.length, router])

  const handlePay = async () => {
    setLoading(true)
    setError(null)
    try {
      const items = cart.items.map((i) => ({
        productId: i.product.id,
        quantity: i.quantity,
        configuration: i.configuration,
      }))
      const res = await createCheckoutSession({ items })
      window.location.href = res.url
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : "Fehler beim Starten der Zahlung"
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen">
      <Header />
      <main className="container mx-auto px-4 py-8">
        <h2 className="text-2xl mb-2">Mit TWINT bezahlen</h2>
        <p className="text-muted-foreground mb-6">Sie werden zu Stripe Checkout weitergeleitet.</p>

        <div className="border rounded-md p-4 mb-6">
          <div className="flex items-center justify-between">
            <span>Gesamtsumme</span>
            <strong>{formatChf(cart.totalCents)}</strong>
          </div>
        </div>

        {error && <p className="text-red-600 mb-4">{error}</p>}

        <div className="flex gap-3">
          <Button variant="outline" onClick={() => router.back()} disabled={loading}>
            Zurück
          </Button>
          <Button onClick={handlePay} disabled={loading || cart.items.length === 0}>
            {loading ? "Weiterleiten…" : "TWINT Checkout öffnen"}
          </Button>
        </div>
      </main>
    </div>
  )
}
