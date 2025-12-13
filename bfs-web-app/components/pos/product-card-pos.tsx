"use client"

import { Plus } from "lucide-react"
import Image from "next/image"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { useCart } from "@/contexts/cart-context"
import type { ProductDTO } from "@/types"

function formatPriceLabel(cents: number): string {
  const francs = Math.floor(cents / 100)
  const rappen = cents % 100
  if (rappen === 0) return `CHF ${francs}.-`
  const v = (cents / 100).toFixed(2)
  return `CHF ${v}`
}

export function ProductCardPOS({ product, onConfigure }: { product: ProductDTO; onConfigure: () => void }) {
  const { addToCart } = useCart()
  const isAvailable = product.isAvailable !== false
  const isLowStock = product.isLowStock === true
  const availableQty = product.availableQuantity ?? null
  const isActive = product.isActive !== false
  const disabled = !isAvailable || !isActive

  const handleAdd = () => {
    if (disabled) return
    if (product.type === "menu") onConfigure()
    else addToCart(product)
  }

  const onCardKeyDown: React.KeyboardEventHandler<HTMLDivElement> = (e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault()
      handleAdd()
    }
  }

  return (
    <Card
      role="button"
      tabIndex={disabled ? -1 : 0}
      aria-disabled={disabled}
      onClick={handleAdd}
      onKeyDown={onCardKeyDown}
      className={
        "gap-0 overflow-hidden rounded-[11px] p-0 transition-shadow hover:shadow-lg " +
        (disabled ? "" : "cursor-pointer")
      }
    >
      <CardHeader className="p-2">
        <div className="relative aspect-video rounded-[11px] rounded-t-lg bg-[#cec9c6]">
          {product.image ? (
            <Image
              src={product.image}
              alt={"Produktbild von " + product.name}
              fill
              sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
              quality={90}
              className="h-full w-full rounded-[11px] object-cover"
            />
          ) : (
            <div className="absolute inset-0 flex items-center justify-center text-zinc-500">Kein Bild</div>
          )}
          {!isAvailable && (
            <div className="absolute inset-0 z-10 grid place-items-center rounded-[11px] bg-black/55">
              <span className="rounded-full bg-red-400 px-3 py-1 text-sm font-medium text-white">Ausverkauft</span>
            </div>
          )}
          {isLowStock && isAvailable && isActive && (
            <div className="absolute top-1 left-2 z-10">
              <span className="rounded-full bg-amber-600 px-2 py-0.5 text-xs font-medium text-white">
                {availableQty !== null ? `Nur ${availableQty} übrig` : "Geringer Bestand"}
              </span>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent className="px-2 pt-0 pb-4">
        <div className="flex items-center justify-between">
          <div className="flex flex-col">
            <h3 className="font-family-secondary text-base">{product.name}</h3>
            <p className="font-family-secondary text-sm">{formatPriceLabel(product.priceCents)}</p>
          </div>
          <div className="flex items-center">
            <Button
              size="icon"
              onClick={(e) => {
                e.stopPropagation()
                handleAdd()
              }}
              aria-label={`Produkt ${product.name} hinzufügen`}
              className="bg-foreground hover:bg-foreground/90 rounded-[10px] text-white"
            >
              <Plus className="size-5" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
