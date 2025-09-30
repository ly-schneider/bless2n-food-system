"use client"

import { useMemo } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { useCart } from "@/contexts/cart-context"
import { formatChf } from "@/lib/utils"
import type { MenuSuggestion } from "@/lib/menu-suggestions"
import type { CartItem } from "@/types/cart"
import { CartItemDisplay } from "./cart-item-display"

interface InlineMenuGroupProps {
  suggestion: MenuSuggestion
  items: CartItem[]
  onEditItem: (item: CartItem) => void
}

export function InlineMenuGroup({ suggestion, items, onEditItem }: InlineMenuGroupProps) {
  const { updateQuantity, removeFromCart, addToCart } = useCart()

  const sumSimpleCents = useMemo(() => items.reduce((sum, it) => sum + it.product.priceCents, 0), [items])

  const applyConversion = () => {
    for (const ci of items) {
      updateQuantity(ci.id, ci.quantity - 1)
    }
    addToCart(suggestion.menuProduct, suggestion.configuration)
  }

  return (
    <Card className="border-none !p-0 shadow-none rounded-[11px]">
      <div className="p-4">
        <div className="mb-1 flex items-center justify-between gap-2">
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold">Menü‑Vorschlag: {suggestion.menuProduct.name}</p>
            <p className="text-xs text-muted-foreground">Spare {formatChf(suggestion.savingsCents)} — Deine Auswahl bleibt erhalten.</p>
          </div>
          <span className="ml-2 shrink-0 rounded-full bg-green-100 px-2 py-0.5 text-[10px] font-medium text-green-800">Bestes Angebot</span>
        </div>

        <div className="mb-2">
          <Collapsible>
            <CollapsibleTrigger className="text-xs text-muted-foreground underline underline-offset-2">
              Preisvergleich
            </CollapsibleTrigger>
            <CollapsibleContent>
              <div className="mt-2 space-y-1 text-xs">
                <div className="flex items-center justify-between">
                  <span>Aktuell</span>
                  <span className="font-medium">{formatChf(sumSimpleCents)}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span>{suggestion.menuProduct.name}</span>
                  <span className="font-medium">{formatChf(suggestion.menuProduct.priceCents)}</span>
                </div>
                <div className="mt-1 flex items-center justify-between border-t pt-1">
                  <span>Gespart</span>
                  <span className="font-semibold text-green-700">{formatChf(suggestion.savingsCents)}</span>
                </div>
              </div>
            </CollapsibleContent>
          </Collapsible>
        </div>

        <div className="divide-border divide-y">
          {items.map((item) => (
            <div key={item.id} className="py-3">
              <CartItemDisplay
                item={item}
                onUpdateQuantity={(q) => updateQuantity(item.id, q)}
                onRemove={() => removeFromCart(item.id)}
                onEdit={() => onEditItem(item)}
              />
            </div>
          ))}
        </div>
        <div className="mt-3 flex items-center justify-between text-sm">
          <span className="font-medium">Menü‑Preis</span>
          <span className="font-medium">{formatChf(suggestion.menuProduct.priceCents)}</span>
        </div>
        <div className="mt-3">
          <Button className="w-full" onClick={applyConversion}>
            Jetzt wechseln und sparen
          </Button>
        </div>
      </div>
    </Card>
  )
}
