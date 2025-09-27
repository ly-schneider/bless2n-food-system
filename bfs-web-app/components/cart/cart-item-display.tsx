"use client"

import { Minus, Pen, Plus, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { CartItem } from "@/types/cart"
import { formatChf } from "@/lib/utils"

export interface CartItemDisplayProps {
  item: CartItem
  onUpdateQuantity: (quantity: number) => void
  onRemove: () => void
  onEdit: () => void
}

export function CartItemDisplay({ item, onUpdateQuantity, onRemove, onEdit }: CartItemDisplayProps) {
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
      <div className="flex w-full flex-col gap-4">
        {isMenuProduct ? (
          <>
            <div className="flex flex-row justify-between">
              <div className="flex flex-col gap-1">
                <h3 className="font-family-secondary truncate text-lg font-medium">{item.product.name}</h3>
                {hasConfiguration && (
                  <div className="flex flex-row flex-wrap gap-1.5">
                    {Object.entries(item.configuration || {}).map(([slotId, productId]) => {
                      const slot = item.product.menu?.slots?.find((s) => s.id === slotId)
                      const slotItem = slot?.menuSlotItems?.find((si) => si.id === productId)

                      if (slot && slotItem) {
                        return (
                          <div
                            key={slotId}
                            className="text-muted-foreground border-border rounded-lg border p-1 text-xs"
                          >
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
                  className="text-destructive hover:text-destructive size-10 bg-inherit"
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
            </div>
            <div className="flex flex-row items-center justify-between">
              <h4 className="font-family-secondary truncate text-base">{formatChf(item.product.priceCents)}</h4>
              <div className="flex flex-row items-center gap-2">
                <Button onClick={onEdit} size="icon" variant="outline" className="size-10 shrink-0 bg-inherit">
                  <Pen className="size-4" />
                </Button>
                <Button
                  onClick={() => onUpdateQuantity(item.quantity - 1)}
                  size="icon"
                  variant="outline"
                  className="size-10 shrink-0 bg-inherit"
                >
                  <Minus className="size-4" />
                </Button>
                <span className="min-w-4 text-center font-medium">{item.quantity}</span>
                <Button
                  onClick={() => onUpdateQuantity(item.quantity + 1)}
                  size="icon"
                  variant="outline"
                  className="size-10 shrink-0 bg-inherit"
                >
                  <Plus className="size-4" />
                </Button>
              </div>
            </div>
          </>
        ) : (
          <div className="flex flex-row items-center justify-between">
            <div className="flex flex-col gap-0">
              <h3 className="font-family-secondary truncate text-lg font-medium">{item.product.name}</h3>
              <h4 className="font-family-secondary truncate text-base">{formatChf(item.product.priceCents)}</h4>
            </div>
            <div className="flex flex-row items-center gap-2">
              <Button
                onClick={() => onUpdateQuantity(item.quantity - 1)}
                size="icon"
                variant="outline"
                className="size-10 shrink-0 bg-inherit"
              >
                <Minus className="size-4" />
              </Button>
              <span className="min-w-4 text-center font-medium">{item.quantity}</span>
              <Button
                onClick={() => onUpdateQuantity(item.quantity + 1)}
                size="icon"
                variant="outline"
                className="size-10 shrink-0 bg-inherit"
              >
                <Plus className="size-4" />
              </Button>
              <Button
                onClick={onRemove}
                size="icon"
                variant="outline"
                className="text-destructive hover:text-destructive size-10 shrink-0 bg-inherit"
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
