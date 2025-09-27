"use client"

import { useState } from "react"
import { ShoppingCart, Trash2, Pen, Minus, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Drawer, DrawerClose, DrawerContent, DrawerHeader, DrawerTitle, DrawerFooter } from "@/components/ui/drawer"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from "@/components/ui/sheet"
import { useCart } from "@/contexts/cart-context"
import { CartItem } from "@/types/cart"
import { ProductConfigurationModal } from "./product-configuration-modal"
import { formatChf } from "@/lib/utils"
import { useIsMobile } from "@/hooks/use-mobile"
import { useRouter } from "next/navigation"
import { CartItemDisplay } from "@/components/cart/cart-item-display"

export function FloatingBottomNav() {
  const { cart, updateQuantity, removeFromCart } = useCart()
  const [isCartOpen, setIsCartOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<CartItem | null>(null)
  const isMobile = useIsMobile()
  const router = useRouter()

  const totalItems = cart.items.reduce((sum, item) => sum + item.quantity, 0)

  return (
    totalItems > 0 && (
      <>
        <div className="border-border fixed right-0 bottom-0 left-0 z-50 border-t bg-white p-4 shadow-lg">
          <div className="container mx-auto flex items-center justify-center gap-3">
            <Button
              className="rounded-pill h-12 text-base font-medium flex-1 lg:flex-none lg:max-w-xs px-6 md:px-8 lg:px-10 xl:px-12"
              disabled={cart.items.length === 0}
              onClick={() => router.push("/checkout")}
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
                  <div className="divide-border -my-5 flex flex-col divide-y [&>*]:py-5">
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

              {cart.items.length > 0 && (
                <DrawerFooter>
                  <div className="border-border mt-12 flex flex-col gap-3 border-t pt-4">
                    <div className="flex items-center justify-between pb-2">
                      <div className="flex flex-col">
                        <p className="text-lg font-semibold">Total</p>
                        <span className="text-muted-foreground text-sm">{cart.items.length} Produkte</span>
                      </div>
                      <p className="text-lg font-semibold">{formatChf(cart.totalCents)}</p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        className="rounded-pill h-12 flex-1 text-base font-medium lg:max-w-xs"
                        onClick={() => {
                          setIsCartOpen(false)
                          router.push("/checkout")
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
            <SheetContent side="right" className="bg-primary-foreground rounded-l-3xl w-full sm:max-w-sm md:max-w-md lg:max-w-lg xl:max-w-xl">
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
                  <div className="divide-border -my-5 flex flex-col divide-y [&>*]:py-5">
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

              {cart.items.length > 0 && (
                <SheetFooter>
                  <div className="border-border mt-12 flex w-full flex-col gap-3 border-t pt-4">
                    <div className="flex items-center justify-between pb-2">
                      <div className="flex flex-col">
                        <p className="text-lg font-semibold">Total</p>
                        <span className="text-muted-foreground text-sm">{cart.items.length} Produkte</span>
                      </div>
                      <p className="text-lg font-semibold">{formatChf(cart.totalCents)}</p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        className="rounded-pill h-12 w-full text-base font-medium"
                        onClick={() => {
                          setIsCartOpen(false)
                          router.push("/checkout")
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
