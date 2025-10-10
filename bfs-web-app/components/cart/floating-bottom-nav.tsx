"use client"

import { ShoppingCart } from "lucide-react"
import { useRouter } from "next/navigation"
import { useEffect, useRef, useState } from "react"
import { CartItemDisplay } from "@/components/cart/cart-item-display"
import { InlineMenuGroup } from "@/components/cart/inline-menu-group"
import { useBestMenuSuggestion } from "@/components/cart/use-best-menu-suggestion"
import { Button } from "@/components/ui/button"
import { Drawer, DrawerContent, DrawerFooter, DrawerHeader, DrawerTitle } from "@/components/ui/drawer"
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { useCart } from "@/contexts/cart-context"
import { useIsMobile } from "@/hooks/use-mobile"
import { formatChf } from "@/lib/utils"
import { CartItem } from "@/types/cart"
import { ProductConfigurationModal } from "./product-configuration-modal"

export function FloatingBottomNav() {
  const { cart, updateQuantity, removeFromCart } = useCart()
  const [isCartOpen, setIsCartOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<CartItem | null>(null)
  const isMobile = useIsMobile()
  const router = useRouter()
  const { suggestion, contiguous, startIndex, endIndex } = useBestMenuSuggestion()

  const totalItems = cart.items.reduce((sum, item) => sum + item.quantity, 0)

  // Measure the fixed action bar to add an in-flow spacer and avoid overlap
  const barRef = useRef<HTMLDivElement>(null)
  const [, setBarHeight] = useState(0)
  const [, setAppFooterHeight] = useState(0)

  useEffect(() => {
    const el = barRef.current
    if (!el) return
    const update = () => setBarHeight(el.offsetHeight || 0)
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
  }, [totalItems])

  // Also track global app footer height to avoid double spacing
  useEffect(() => {
    const footerEl = document.getElementById("app-footer")
    if (!footerEl) return
    const update = () => setAppFooterHeight(footerEl.offsetHeight || 0)
    update()
    let ro: ResizeObserver | null = null
    if (typeof ResizeObserver !== "undefined") {
      ro = new ResizeObserver(() => update())
      ro.observe(footerEl)
    }
    const onResize = () => update()
    window.addEventListener("resize", onResize)
    return () => {
      window.removeEventListener("resize", onResize)
      if (ro) ro.disconnect()
    }
  }, [])

  return (
    totalItems > 0 && (
      <>
        <div ref={barRef} className="fixed right-0 bottom-0 left-0 z-50 p-4">
          <div className="mx-auto flex max-w-xl items-center justify-center gap-3">
            <Button
              className="rounded-pill h-12 flex-1 px-6 text-base font-medium md:px-8 lg:max-w-xs lg:flex-none lg:px-10 xl:px-12"
              disabled={cart.items.length === 0}
              onClick={() => router.push("/food/checkout")}
            >
              Jetzt bezahlen
            </Button>
            <Button
              onClick={() => setIsCartOpen(true)}
              size="icon"
              variant="outline"
              className="relative size-12 shrink-0 rounded-full"
            >
              <ShoppingCart className="size-5" />
            </Button>
          </div>
        </div>

        {isMobile ? (
          <Drawer open={isCartOpen} onOpenChange={setIsCartOpen}>
            <DrawerContent className="rounded-t-3xl">
              <DrawerHeader>
                <DrawerTitle className="text-[1.35rem]">Warenkorb</DrawerTitle>
              </DrawerHeader>

              <div className="flex-1 overflow-y-auto px-5">
                {cart.items.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-12 text-center">
                    <ShoppingCart className="text-muted-foreground mb-4 size-12" />
                    <p className="text-muted-foreground text-lg">Ihr Warenkorb ist leer</p>
                    <p className="text-muted-foreground mt-1 text-sm">Fügen Sie Produkte hinzu, um zu beginnen</p>
                  </div>
                ) : (
                  <div className="divide-border mt-3 -mb-5 flex flex-col divide-y [&>*]:py-5 [&>*:first-child]:pt-0">
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
                            key={`group-${grouped.map((g) => g.id).join("-")}`}
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

              {cart.items.length > 0 && (
                <DrawerFooter>
                  <div className="border-border flex flex-col gap-3 border-t pt-4">
                    <div className="flex items-center justify-between pb-2">
                      <div className="flex flex-col">
                        <p className="text-lg font-semibold">Total</p>
                        <span className="text-muted-foreground text-sm">{cart.items.length} Produkte</span>
                      </div>
                      <p className="text-lg font-semibold">{formatChf(cart.totalCents)}</p>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Button
                        className="rounded-pill h-12 flex-1 text-base font-medium lg:max-w-xs"
                        onClick={() => {
                          setIsCartOpen(false)
                          router.push("/food/checkout")
                        }}
                      >
                        Jetzt bezahlen
                      </Button>
                    </div>
                  </div>
                </DrawerFooter>
              )}
            </DrawerContent>
          </Drawer>
        ) : (
          <Sheet open={isCartOpen} onOpenChange={setIsCartOpen}>
            <SheetContent
              side="right"
              className="bg-primary-foreground w-full rounded-l-3xl sm:max-w-sm md:max-w-md lg:max-w-lg xl:max-w-xl"
            >
              <SheetHeader>
                <SheetTitle className="text-[1.35rem]">Warenkorb</SheetTitle>
              </SheetHeader>

              <div className="flex-1 overflow-y-auto px-5">
                {cart.items.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-12 text-center">
                    <ShoppingCart className="text-muted-foreground mb-4 size-12" />
                    <p className="text-muted-foreground text-lg">Ihr Warenkorb ist leer</p>
                    <p className="text-muted-foreground mt-1 text-sm">Fügen Sie Produkte hinzu, um zu beginnen</p>
                  </div>
                ) : (
                  <div className="divide-border mt-3 -mb-5 flex flex-col divide-y [&>*]:py-5 [&>*:first-child]:pt-0">
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
                            key={`group-${grouped.map((g) => g.id).join("-")}`}
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

              {cart.items.length > 0 && (
                <SheetFooter>
                  <div className="border-border flex w-full flex-col gap-3 border-t pt-4">
                    <div className="flex items-center justify-between pb-2">
                      <div className="flex flex-col">
                        <p className="text-lg font-semibold">Total</p>
                        <span className="text-muted-foreground text-sm">{cart.items.length} Produkte</span>
                      </div>
                      <p className="text-lg font-semibold">{formatChf(cart.totalCents)}</p>
                    </div>
                    <div className="flex w-full flex-col gap-2">
                      <Button
                        className="rounded-pill h-12 w-full text-base font-medium"
                        onClick={() => {
                          setIsCartOpen(false)
                          router.push("/food/checkout")
                        }}
                      >
                        Jetzt bezahlen
                      </Button>
                    </div>
                  </div>
                </SheetFooter>
              )}
            </SheetContent>
          </Sheet>
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
      </>
    )
  )
}

// CartItemDisplay moved to components/cart/cart-item-display
