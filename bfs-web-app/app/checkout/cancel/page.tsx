"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { createCheckoutSession } from "@/lib/api/payments"

export default function CheckoutCancelPage() {
  const router = useRouter()
  const { cart } = useCart()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleRetry = async () => {
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
    <div className="flex min-h-[70vh] flex-col items-center justify-center gap-10 px-4 pb-36">
      <div className="relative h-72 w-72">
        {/* Concentric rings (error red gradient) */}
        <div className="absolute inset-0 rounded-full bg-gradient-to-br from-red-200/60 to-red-400/40 blur-sm" />
        <div className="absolute inset-8 rounded-full border-8 border-red-300/60" />
        <div className="absolute inset-16 rounded-full border-8 border-red-400/50" />
        <div className="absolute inset-24 flex items-center justify-center rounded-full bg-red-500/80">
          <X className="h-14 w-14 text-white" />
        </div>
      </div>
      <h1 className="text-3xl font-semibold">Bezahlung Abgebrochen</h1>

      {/* Bottom fixed action buttons stacked */}
      <div className="fixed inset-x-0 bottom-0 p-4">
        <div className="flex flex-col gap-3">
          <Button className="rounded-pill h-12 w-full text-base font-medium" onClick={handleRetry} disabled={loading}>
            {loading ? "Weiterleiten…" : "Erneut versuchen"}
          </Button>
          <Button
            variant="outline"
            className="rounded-pill h-12 w-full bg-white text-base font-medium text-black"
            onClick={() => router.push("/checkout?from=cancel")}
          >
            Zurück zum Warenkorb
          </Button>
          {error && <p className="text-center text-sm text-red-600">{error}</p>}
        </div>
      </div>
    </div>
  )
}
