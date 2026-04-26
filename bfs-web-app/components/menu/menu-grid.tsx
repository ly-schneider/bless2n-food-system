"use client"

import Image from "next/image"
import { useState } from "react"
import { CartButtons } from "@/components/cart/cart-buttons"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { useCart } from "@/contexts/cart-context"
import { formatChf } from "@/lib/utils"
import { ListResponse, ProductDTO } from "@/types"

export function MenuGrid({ products }: { products: ListResponse<ProductDTO> }) {
  const getCatPos = (p?: { position?: number | null } | null) => {
    const v = p?.position
    return typeof v === "number" && isFinite(v) ? v : 1_000_000
  }

  const activeItems = products.items.filter((it) => it.isActive !== false)

  const sortedProducts = [...activeItems].sort((a, b) => {
    const pa = getCatPos(a.category)
    const pb = getCatPos(b.category)
    if (pa !== pb) return pa - pb
    return a.name.localeCompare(b.name)
  })

  return (
    <div className="xs:grid-cols-2 grid grid-cols-1 gap-3 md:gap-5 lg:grid-cols-3 xl:grid-cols-4">
      {sortedProducts.map((product) => (
        <MenuProductCard key={product.id} product={product} />
      ))}
    </div>
  )
}

function MenuProductCard({ product }: { product: ProductDTO }) {
  const [isConfigModalOpen, setIsConfigModalOpen] = useState(false)
  const { addToCart, getItemQuantity, getTotalProductQuantity } = useCart()
  const isMenu = product.type === "menu"
  const isAvailable = isMenu || product.isAvailable !== false
  const isLowStock = !isMenu && product.isLowStock === true
  const availableQty = isMenu ? null : (product.availableQuantity ?? null)
  const isActive = product.isActive !== false
  const disabled = !isAvailable || !isActive

  // Determine current quantity for max check, mirroring CartButtons logic
  const quantity = product.type === "menu" ? getTotalProductQuantity(product.id) : getItemQuantity(product.id)
  const maxQty = typeof availableQty === "number" ? availableQty : undefined
  const reachedMax = typeof maxQty === "number" && quantity >= maxQty

  const handleConfigureProduct = () => {
    setIsConfigModalOpen(true)
  }

  const handleCardActivate = () => {
    if (disabled || reachedMax) return
    if (product.type === "menu") {
      // Open configuration for menu products when no specific config is chosen
      handleConfigureProduct()
    } else {
      addToCart(product)
    }
  }

  const onCardKeyDown: React.KeyboardEventHandler<HTMLDivElement> = (e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault()
      handleCardActivate()
    }
  }

  return (
    <>
      <Card
        role="button"
        tabIndex={disabled || reachedMax ? -1 : 0}
        aria-disabled={disabled || reachedMax}
        onClick={handleCardActivate}
        onKeyDown={onCardKeyDown}
        className={
          "gap-0 overflow-hidden rounded-[11px] p-0 transition-shadow hover:shadow-lg " +
          (disabled || reachedMax ? "" : "cursor-pointer")
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
                unoptimized={product.image.includes("localhost") || product.image.includes("127.0.0.1")}
              />
            ) : (
              <div className="absolute inset-0 flex items-center justify-center text-zinc-500">Kein Bild</div>
            )}
            {!isAvailable && (
              <div className="absolute inset-0 z-10 grid place-items-center rounded-[11px] bg-black/55">
                <span className="rounded-full bg-red-400 px-3 py-1 text-sm font-medium text-white">Ausverkauft</span>
              </div>
            )}
            {isAvailable && !isActive && (
              <div className="absolute inset-0 z-10 grid place-items-center rounded-[11px] bg-black/55">
                <span className="rounded-full bg-zinc-700 px-3 py-1 text-sm font-medium text-white">Nicht aktiv</span>
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
          <div className="flex items-start justify-between gap-2">
            <div className="flex min-w-0 flex-col">
              <h3 className="font-family-secondary text-lg">{product.name}</h3>
              {product.description && (
                <p className="text-muted-foreground mt-0.5 line-clamp-2 text-xs whitespace-pre-line">
                  {product.description}
                </p>
              )}
              <p className="font-family-secondary mt-1 text-base">{formatChf(product.priceCents)}</p>
            </div>
            <div className="flex items-center" onClick={(e) => e.stopPropagation()}>
              <CartButtons product={product} onConfigureProduct={handleConfigureProduct} disabled={disabled} />
            </div>
          </div>
        </CardContent>
      </Card>

      <ProductConfigurationModal
        product={product}
        isOpen={isConfigModalOpen}
        onClose={() => setIsConfigModalOpen(false)}
      />
    </>
  )
}

export default MenuGrid
