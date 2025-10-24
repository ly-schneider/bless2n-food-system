"use client"

import { Minus, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { CartItemConfiguration, ProductDTO } from "@/types"

interface CartButtonsProps {
  product: ProductDTO
  configuration?: CartItemConfiguration
  onConfigureProduct?: () => void
  disabled?: boolean
}

export function CartButtons({ product, configuration, onConfigureProduct, disabled }: CartButtonsProps) {
  const { addToCart, updateQuantity, getItemQuantity, getTotalProductQuantity, cart } = useCart()

  // For menu products without specific configuration, show total quantity across all configurations
  const quantity = configuration
    ? getItemQuantity(product.id, configuration)
    : product.type === "menu"
      ? getTotalProductQuantity(product.id)
      : getItemQuantity(product.id, configuration)

  const maxQty = typeof product.availableQuantity === "number" ? product.availableQuantity : undefined
  const reachedMax = typeof maxQty === "number" && quantity >= maxQty

  const handleAdd = () => {
    if (disabled || reachedMax) return
    if (product.type === "menu" && !configuration && onConfigureProduct) {
      onConfigureProduct()
    } else {
      addToCart(product, configuration)
    }
  }

  const handleRemove = () => {
    if (quantity > 0) {
      if (configuration) {
        // For specific configuration, update that item
        const itemId = `${product.id}-${JSON.stringify(configuration)}`
        const currentItem = cart.items.find((item) => item.id === itemId)
        if (currentItem) {
          updateQuantity(currentItem.id, currentItem.quantity - 1)
        }
      } else if (product.type === "menu") {
        // For menu products without configuration, remove one from the last added configuration
        const productItems = cart.items.filter((item) => item.product.id === product.id)
        if (productItems.length > 0) {
          const lastItem = productItems[productItems.length - 1]
          if (lastItem) {
            updateQuantity(lastItem.id, lastItem.quantity - 1)
          }
        }
      } else {
        // For simple products
        const itemId = `${product.id}-`
        const currentItem = cart.items.find((item) => item.id === itemId)
        if (currentItem) {
          updateQuantity(currentItem.id, currentItem.quantity - 1)
        }
      }
    }
  }

  const handleIncrease = () => {
    if (disabled || reachedMax) return
    if (configuration) {
      // For specific configuration, update that item
      const itemId = `${product.id}-${JSON.stringify(configuration)}`
      const currentItem = cart.items.find((item) => item.id === itemId)
      if (currentItem) {
        updateQuantity(currentItem.id, currentItem.quantity + 1)
      }
    } else if (product.type === "menu") {
      // For menu products without configuration, open configuration modal instead
      if (onConfigureProduct) {
        onConfigureProduct()
      }
    } else {
      // For simple products
      const itemId = `${product.id}-`
      const currentItem = cart.items.find((item) => item.id === itemId)
      if (currentItem) {
        updateQuantity(currentItem.id, currentItem.quantity + 1)
      }
    }
  }

  if (quantity === 0) {
    return (
      <Button
        onClick={handleAdd}
        size="icon"
        variant="ghost"
        disabled={disabled || reachedMax}
        className="bg-foreground text-background hover:bg-foreground hover:text-background size-8 cursor-pointer rounded-full disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Plus className="size-4" />
      </Button>
    )
  }

  return (
    <div className="flex items-center gap-0">
      <Button
        onClick={handleRemove}
        size="icon"
        variant="ghost"
        className="bg-foreground text-background hover:bg-foreground hover:text-background size-8 cursor-pointer rounded-full"
      >
        <Minus className="size-4" />
      </Button>

      <span className="min-w-6 text-center font-medium sm:min-w-8">{quantity}</span>

      <Button
        onClick={handleIncrease}
        size="icon"
        variant="ghost"
        disabled={disabled || reachedMax}
        className="bg-foreground text-background hover:bg-foreground hover:text-background size-8 cursor-pointer rounded-full disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Plus className="size-4" />
      </Button>
    </div>
  )
}
