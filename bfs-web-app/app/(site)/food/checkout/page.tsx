"use client"

import { ArrowLeft, ShoppingCart } from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef, useState } from "react"
import { AuthNudgeBanner } from "@/components/auth/auth-nudge"
import { CartItemDisplay } from "@/components/cart/cart-item-display"
import { InlineMenuGroup } from "@/components/cart/inline-menu-group"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { useBestMenuSuggestion } from "@/components/cart/use-best-menu-suggestion"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { formatChf } from "@/lib/utils"
import { CartItem } from "@/types/cart"

export default function CheckoutPage() {
  const { cart, updateQuantity, removeFromCart } = useCart()
  const [editingItem, setEditingItem] = useState<CartItem | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()
  const sp = useSearchParams()
  const from = sp.get("from")
  const footerRef = useRef<HTMLDivElement>(null)
  const [footerHeight, setFooterHeight] = useState(0)

  // Measure footer height to ensure last cart item isn’t obscured
  useEffect(() => {
    const el = footerRef.current
    if (!el) return

    const update = () => setFooterHeight(el.offsetHeight || 0)
    update()

    let ro: ResizeObserver | null = null
    if (typeof ResizeObserver !== "undefined") {
      ro = new ResizeObserver(() => update())
      ro.observe(el)
    }

    const onResize = () => update()
    window.addEventListener("resize", onResize)

    return () => {
      window.removeEventListener("resize", onResize)
      if (ro) ro.disconnect()
    }
  }, [cart.items.length, loading, error])

  const handlePay = async () => {
    setLoading(true)
    setError(null)
    try {
      router.push("/food/checkout/payment")
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : "Fehler beim Starten der Zahlung"
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  const { suggestion, contiguous, startIndex, endIndex } = useBestMenuSuggestion()

  return (
    <div className="flex flex-1 flex-col">

      <main
        className="container mx-auto flex-1 overflow-y-auto px-4 pt-4"
        style={{ paddingBottom: footerHeight ? footerHeight + 16 : 16 }}
      >
        <h2 className="text-2xl mb-4">Warenkorb</h2>

        {/* Soft sign-in encouragement */}
        <AuthNudgeBanner />

        <div>
          {cart.items.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <ShoppingCart className="text-muted-foreground mb-4 size-10" />
              <p className="text-muted-foreground text-lg font-semibold">Warenkorb ist leer</p>
            </div>
          ) : (
            <div className="mt-4 -mb-5 flex flex-col divide-y divide-border [&>*]:py-5 [&>*:first-child]:pt-0">
              {(() => {
                const rows: React.ReactNode[] = []
                if (suggestion && contiguous && startIndex >= 0 && endIndex >= startIndex) {
                  for (let i = 0; i < startIndex; i++) {
                    const item = cart.items[i]!
                    rows.push(
                      <CartItemDisplay
                        key={item.id}
                        item={item}
                        onUpdateQuantity={(q) => updateQuantity(item.id, q)}
                        onRemove={() => removeFromCart(item.id)}
                        onEdit={() => setEditingItem(item)}
                      />
                    )
                  }
                  const grouped = cart.items.slice(startIndex, endIndex + 1)
                  rows.push(
                    <InlineMenuGroup
                      key={`group-${grouped.map((g) => g.id).join('-')}`}
                      suggestion={suggestion}
                      items={grouped}
                      onEditItem={(it) => setEditingItem(it)}
                    />
                  )
                  for (let i = endIndex + 1; i < cart.items.length; i++) {
                    const item = cart.items[i]!
                    rows.push(
                      <CartItemDisplay
                        key={item.id}
                        item={item}
                        onUpdateQuantity={(q) => updateQuantity(item.id, q)}
                        onRemove={() => removeFromCart(item.id)}
                        onEdit={() => setEditingItem(item)}
                      />
                    )
                  }
                } else {
                  for (const item of cart.items) {
                    rows.push(
                      <CartItemDisplay
                        key={item.id}
                        item={item}
                        onUpdateQuantity={(q) => updateQuantity(item.id, q)}
                        onRemove={() => removeFromCart(item.id)}
                        onEdit={() => setEditingItem(item)}
                      />
                    )
                  }
                }
                return rows
              })()}
            </div>
          )}
        </div>
      </main>

      <div ref={footerRef} className="fixed right-0 bottom-0 left-0 z-50 p-4 bg-background">
        <div className="max-w-xl mx-auto flex flex-col gap-3">
          {cart.items.length > 0 && (
            <div className="flex items-center justify-between py-2">
              <div className="flex flex-col">
                <p className="text-lg font-semibold">Total</p>
                <span className="text-muted-foreground text-sm">{cart.items.length} Produkte</span>
              </div>
              <p className="text-lg font-semibold">{formatChf(cart.totalCents)}</p>
            </div>
          )}

          {cart.items.length > 0 ? (
            <div className="flex items-center justify-between gap-3">
              <Button
                onClick={() => {
                  if (from === "cancel") {
                    router.push("/")
                  } else {
                    router.back()
                  }
                }}
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
          ) : (
            <div>
              <Button
                variant="outline"
                className="rounded-pill h-12 w-full bg-white text-black text-base font-medium"
                onClick={() => {
                  if (from === "cancel") {
                    router.push("/")
                  } else {
                    router.back()
                  }
                }}
              >
                <ArrowLeft className="mr-2 size-5" />
                Zurück
              </Button>
            </div>
          )}

          {cart.items.length > 0 && error && <p className="text-red-600 text-sm">{error}</p>}
        </div>
      </div>

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
