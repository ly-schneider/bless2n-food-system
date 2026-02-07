"use client"

import { Minus, Pen, Plus, Trash2 } from "lucide-react"
import Image from "next/image"
import { Button } from "@/components/ui/button"
import { formatChf } from "@/lib/utils"
import { CartItem } from "@/types/cart"

export interface CartItemDiscountInfo {
  discountedQuantity: number
  unitPriceCents: number
}

export interface CartItemDisplayProps {
  item: CartItem
  onUpdateQuantity: (quantity: number) => void
  onRemove: () => void
  onEdit: () => void
  isPOS?: boolean
  discountInfo?: CartItemDiscountInfo
  maxQuantity?: number | null
}

export function CartItemDisplay({
  item,
  onUpdateQuantity,
  onRemove,
  onEdit,
  isPOS = false,
  discountInfo,
  maxQuantity,
}: CartItemDisplayProps) {
  const isMenuProduct = item.product.type === "menu"
  const hasConfiguration = item.configuration && Object.keys(item.configuration).length > 0
  const isFullyDiscounted = discountInfo && discountInfo.discountedQuantity >= item.quantity
  const isPartiallyDiscounted = discountInfo && discountInfo.discountedQuantity > 0 && discountInfo.discountedQuantity < item.quantity
  const atMaxQuantity = maxQuantity != null && item.quantity >= maxQuantity

  return (
    <div className="flex items-center gap-3">
      {item.product.image && (
        <div
          className={`relative shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6] ${
            isPOS ? "h-16 w-16" : "h-20 w-20"
          }`}
        >
          <Image
            src={item.product.image}
            alt={"Produktbild von " + item.product.name}
            fill
            sizes={isPOS ? "128px" : "160px"}
            quality={90}
            className="h-full w-full rounded-[11px] object-cover"
            unoptimized={item.product.image.includes("localhost") || item.product.image.includes("127.0.0.1")}
          />
        </div>
      )}
      <div className="flex w-full flex-col gap-4">
        {isMenuProduct ? (
          <>
            <div className="flex flex-row justify-between">
              <div className="flex flex-col gap-1">
                <h3 className={`font-family-secondary truncate font-medium ${isPOS ? "text-sm" : "text-lg"}`}>
                  {item.product.name}
                </h3>
                {hasConfiguration && (
                  <div className="flex flex-row flex-wrap gap-1.5">
                    {Object.entries(item.configuration || {}).map(([slotId, productId]) => {
                      const slot = item.product.menu?.slots?.find((s) => s.id === slotId)
                      const slotItem = slot?.options?.find((si) => si.id === productId)

                      if (slot && slotItem) {
                        return (
                          <div
                            key={slotId}
                            className={`text-muted-foreground border-border rounded-lg border p-1 ${
                              isPOS ? "text-[10px]" : "text-xs"
                            }`}
                          >
                            <p className="font-medium">
                              {slot.name}:{" "}
                              {slotItem.name.includes(" Burger") ? slotItem.name.replace(" Burger", "") : slotItem.name}
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
                  className={`text-destructive hover:text-destructive shrink-0 bg-inherit ${
                    isPOS ? "size-9" : "size-10"
                  }`}
                >
                  <Trash2 className={"size-4"} />
                </Button>
              </div>
            </div>
            <div className="flex flex-row items-center justify-between">
              <div className="flex items-center gap-2">
                {isFullyDiscounted ? (
                  <>
                    <h4 className={`font-family-secondary truncate text-muted-foreground line-through ${isPOS ? "text-xs" : "text-base"}`}>
                      {formatChf(item.product.priceCents * item.quantity)}
                    </h4>
                    <span className={`font-medium text-green-600 ${isPOS ? "text-xs" : "text-base"}`}>Gratis</span>
                  </>
                ) : isPartiallyDiscounted ? (
                  <>
                    <h4 className={`font-family-secondary truncate text-muted-foreground line-through ${isPOS ? "text-xs" : "text-base"}`}>
                      {formatChf(item.product.priceCents * item.quantity)}
                    </h4>
                    <span className={`font-medium text-green-600 ${isPOS ? "text-xs" : "text-base"}`}>
                      {formatChf(item.product.priceCents * (item.quantity - discountInfo!.discountedQuantity))}
                    </span>
                    <span className={`text-muted-foreground ${isPOS ? "text-[10px]" : "text-xs"}`}>
                      ({discountInfo!.discountedQuantity}× gratis)
                    </span>
                  </>
                ) : (
                  <h4 className={`font-family-secondary truncate ${isPOS ? "text-xs" : "text-base"}`}>
                    {formatChf(item.product.priceCents)}
                  </h4>
                )}
              </div>
              <div className="flex flex-row items-center gap-2">
                <Button
                  onClick={onEdit}
                  size="icon"
                  variant="outline"
                  className={`shrink-0 bg-inherit ${isPOS ? "size-9" : "size-10"}`}
                >
                  <Pen className={"size-4"} />
                </Button>
                <Button
                  onClick={() => onUpdateQuantity(item.quantity - 1)}
                  size="icon"
                  variant="outline"
                  className={`shrink-0 bg-inherit ${isPOS ? "size-9" : "size-10"}`}
                >
                  <Minus className={"size-4"} />
                </Button>
                <span className={`min-w-4 text-center font-medium ${isPOS ? "text-xs" : ""}`}>{item.quantity}</span>
                <Button
                  onClick={() => onUpdateQuantity(item.quantity + 1)}
                  size="icon"
                  variant="outline"
                  className={`shrink-0 bg-inherit ${isPOS ? "size-9" : "size-10"}`}
                  disabled={atMaxQuantity}
                >
                  <Plus className={"size-4"} />
                </Button>
              </div>
            </div>
          </>
        ) : (
          <div className="flex flex-row items-center justify-between">
            <div className="flex flex-col gap-0">
              <h3 className={`font-family-secondary truncate font-medium ${isPOS ? "text-sm" : "text-lg"}`}>
                {item.product.name}
              </h3>
              <div className="flex items-center gap-2">
                {isFullyDiscounted ? (
                  <>
                    <h4 className={`font-family-secondary truncate text-muted-foreground line-through ${isPOS ? "text-xs" : "text-sm"}`}>
                      {formatChf(item.product.priceCents * item.quantity)}
                    </h4>
                    <span className={`font-medium text-green-600 ${isPOS ? "text-xs" : "text-sm"}`}>Gratis</span>
                  </>
                ) : isPartiallyDiscounted ? (
                  <>
                    <h4 className={`font-family-secondary truncate text-muted-foreground line-through ${isPOS ? "text-xs" : "text-sm"}`}>
                      {formatChf(item.product.priceCents * item.quantity)}
                    </h4>
                    <span className={`font-medium text-green-600 ${isPOS ? "text-xs" : "text-sm"}`}>
                      {formatChf(item.product.priceCents * (item.quantity - discountInfo!.discountedQuantity))}
                    </span>
                    <span className={`text-muted-foreground ${isPOS ? "text-[10px]" : "text-xs"}`}>
                      ({discountInfo!.discountedQuantity}× gratis)
                    </span>
                  </>
                ) : (
                  <h4 className={`font-family-secondary truncate ${isPOS ? "text-xs" : "text-sm"}`}>
                    {formatChf(item.product.priceCents)}
                  </h4>
                )}
              </div>
            </div>
            <div className="flex flex-row items-center gap-2">
              <Button
                onClick={() => onUpdateQuantity(item.quantity - 1)}
                size="icon"
                variant="outline"
                className={`shrink-0 bg-inherit ${isPOS ? "size-9" : "size-10"}`}
              >
                <Minus className={"size-4"} />
              </Button>
              <span className={`min-w-4 text-center font-medium ${isPOS ? "text-xs" : ""}`}>{item.quantity}</span>
              <Button
                onClick={() => onUpdateQuantity(item.quantity + 1)}
                size="icon"
                variant="outline"
                className={`shrink-0 bg-inherit ${isPOS ? "size-9" : "size-10"}`}
                disabled={atMaxQuantity}
              >
                <Plus className={"size-4"} />
              </Button>
              <Button
                onClick={onRemove}
                size="icon"
                variant="outline"
                className={`text-destructive hover:text-destructive shrink-0 bg-inherit ${
                  isPOS ? "size-9" : "size-10"
                }`}
              >
                <Trash2 className={"size-4"} />
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
