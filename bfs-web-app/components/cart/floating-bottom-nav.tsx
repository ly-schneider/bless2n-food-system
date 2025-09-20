"use client"

import { useState } from "react"
import { ShoppingCart, Trash2, Pen, Minus, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Drawer, DrawerClose, DrawerContent, DrawerHeader, DrawerTitle, DrawerFooter } from "@/components/ui/drawer"
import { useCart } from "@/contexts/cart-context"
import { CartItem } from "@/types/cart"
import { ProductConfigurationModal } from "./product-configuration-modal"

export function FloatingBottomNav() {
  const { cart, updateQuantity, removeFromCart } = useCart()
  const [isCartOpen, setIsCartOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<CartItem | null>(null)

  const totalItems = cart.items.reduce((sum, item) => sum + item.quantity, 0)

  return (
    totalItems > 0 && (
      <>
        <div className="border-border fixed right-0 bottom-0 left-0 z-50 border-t bg-white p-4 shadow-lg">
          <div className="container mx-auto flex items-center gap-3">
            <Button
              className="rounded-pill h-12 flex-1 text-base font-medium lg:max-w-xs"
              disabled={cart.items.length === 0}
            >
              Jetzt bezahlen • CHF {(cart.totalCents / 100).toFixed(2)}
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
                <div className="flex flex-col divide-y divide-border -my-8 [&>*]:py-8">
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
                <div className="flex flex-col gap-3">
                  <div className="flex items-center justify-between text-lg font-semibold">
                    <span>Gesamt:</span>
                    <span>CHF {(cart.totalCents / 100).toFixed(2)}</span>
                  </div>
                  <div className="flex gap-2">
                    <DrawerClose asChild>
                      <Button variant="outline" className="flex-1">
                        Weiter einkaufen
                      </Button>
                    </DrawerClose>
                    <Button className="flex-1">Zur Kasse • CHF {(cart.totalCents / 100).toFixed(2)}</Button>
                  </div>
                </div>
              </DrawerFooter>
            )}
          </DrawerContent>
        </Drawer>

        {editingItem && (
          <ProductConfigurationModal
            product={editingItem.product}
            isOpen={true}
            onClose={() => setEditingItem(null)}
            initialConfiguration={editingItem.configuration}
          />
        )}
      </>
    )
  )
}

interface CartItemDisplayProps {
  item: CartItem
  onUpdateQuantity: (quantity: number) => void
  onRemove: () => void
  onEdit: () => void
}

function CartItemDisplay({ item, onUpdateQuantity, onRemove, onEdit }: CartItemDisplayProps) {
  const isMenuProduct = item.product.type === "menu"
  const hasConfiguration = item.configuration && Object.keys(item.configuration).length > 0

  return (
    <div className="flex items-center gap-3">
      {item.product.image && (
        <div className="h-20 w-20 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
          <img
            src={item.product.image}
            alt={"Produktbild von " + item.product.name}
            className="h-full w-full rounded-[11px] object-cover"
          />
        </div>
      )}
      <div className="flex flex-col gap-4 w-full">
        <div className="flex flex-row justify-between">
          <div className="flex flex-col gap-1">
            <h3 className="font-family-secondary truncate text-lg font-medium">{item.product.name}</h3>
            {isMenuProduct && hasConfiguration && (
              <div className="flex flex-row flex-wrap gap-1.5">
                {Object.entries(item.configuration || {}).map(([slotId, productId]) => {
                  const slot = item.product.menu?.slots?.find((s) => s.id === slotId)
                  const slotItem = slot?.menuSlotItems?.find((si) => si.id === productId)

                  if (slot && slotItem) {
                    return (
                      <div key={slotId} className="text-muted-foreground border-border rounded-lg border p-1 text-xs">
                        <p className="font-medium">
                          {slot.name}: {slotItem.name}
                        </p>
                      </div>
                    )
                  }
                  return null
                })}
              </div>
            )}
          </div>
          <div>
            <Button
              onClick={onRemove}
              size="icon"
              variant="outline"
              className="text-destructive hover:text-destructive size-10"
            >
              <Trash2 className="size-4" />
            </Button>
          </div>
        </div>
        <div className="flex flex-row items-center justify-between">
          <h4 className="font-family-secondary truncate text-base">
            CHF {(item.product.priceCents / 100).toFixed(2)}.-
          </h4>
          <div className="flex flex-row items-center gap-2">
            {isMenuProduct && (
              <Button onClick={onEdit} size="icon" variant="outline" className="size-10 shrink-0">
                <Pen className="size-4" />
              </Button>
            )}
            <Button
              onClick={() => onUpdateQuantity(item.quantity - 1)}
              size="icon"
              variant="outline"
              className="size-10 shrink-0"
            >
              <Minus className="size-4" />
            </Button>

            <span className="min-w-4 text-center font-medium">{item.quantity}</span>

            <Button
              onClick={() => onUpdateQuantity(item.quantity + 1)}
              size="icon"
              variant="outline"
              className="size-10 shrink-0"
            >
              <Plus className="size-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
