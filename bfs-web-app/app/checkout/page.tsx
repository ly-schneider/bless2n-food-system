"use client"

import { ArrowLeft, ShoppingCart } from "lucide-react"
import { useRouter } from "next/navigation"
import { useState } from "react"
import { CartItemDisplay } from "@/components/cart/cart-item-display"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { formatChf } from "@/lib/utils"
import { CartItem } from "@/types/cart"
import { createCheckoutSession } from "@/lib/api/payments"

export default function CheckoutPage() {
  const { cart, updateQuantity, removeFromCart } = useCart()
  const [editingItem, setEditingItem] = useState<CartItem | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()

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

      <main className="container mx-auto px-4 pb-28 pt-4">
        <h2 className="text-2xl mb-4">Warenkorb</h2>

        <div className="flex-1 overflow-y-auto">
          {cart.items.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <ShoppingCart className="text-muted-foreground mb-4 size-12" />
              <p className="text-muted-foreground text-lg">Ihr Warenkorb ist leer</p>
              <p className="text-muted-foreground mt-1 text-sm">Fügen Sie Produkte hinzu, um zu beginnen</p>
            </div>
          ) : (
            <div className="divide-border -my-5 flex max-h-[60vh] flex-col divide-y overflow-y-auto [&>*]:py-5">
              {cart.items.map((item) => (
                <CartItemDisplay
                  key={item.id}
                  item={item}
                  onUpdateQuantity={(quantity) => updateQuantity(item.id, quantity)}
                  onRemove={() => removeFromCart(item.id)}
                  onEdit={() => setEditingItem(item)}
                />
              ))}
            </div>
          )}
        </div>
      </main>

      {cart.items.length > 0 && (
        <div className="border-border fixed right-0 bottom-0 left-0 z-50 border-t p-4 shadow-lg">
          <div className="container mx-auto flex flex-col gap-3">
            <div className="flex items-center justify-between py-4">
              <div className="flex flex-col">
                <p className="text-lg font-semibold">Total</p>
                <span className="text-muted-foreground text-sm">{cart.items.length} Produkte</span>
              </div>
              <p className="text-lg font-semibold">{formatChf(cart.totalCents)}</p>
            </div>

            <div className="flex items-center justify-between gap-3">
              <Button
                onClick={() => router.back()}
                size="icon"
                variant="outline"
                className="rounded-full bg-white text-black size-12 shrink-0"
              >
                <ArrowLeft className="size-5" />
              </Button>

              <Button
                className="rounded-pill h-12 text-base font-medium flex-1 md:flex-none md:min-w-64"
                onClick={handlePay}
                disabled={loading}
              >
                {loading ? "Weiterleiten…" : "Mit TWINT bezahlen"}
              </Button>
            </div>
            {error && <p className="text-red-600 text-sm">{error}</p>}
          </div>
        </div>
      )}

      {editingItem && (
        <ProductConfigurationModal
          product={editingItem.product}
          isOpen={true}
          onClose={() => setEditingItem(null)}
          initialConfiguration={editingItem.configuration}
          editingItemId={editingItem.id}
        />
      )}
    </div>
  )
}
