import type { Cents } from "./common"
import type { ProductDTO } from "./index"

export interface CartItemConfiguration {
  [slotId: string]: string // slotId -> selected productId
}

export interface CartItem {
  id: string
  product: ProductDTO
  quantity: number
  configuration?: CartItemConfiguration
  totalPriceCents: Cents
}

export interface Cart {
  items: CartItem[]
  totalCents: Cents
}

export interface CartContextType {
  cart: Cart
  addToCart: (product: ProductDTO, configuration?: CartItemConfiguration) => void
  removeFromCart: (itemId: string) => void
  updateQuantity: (itemId: string, quantity: number) => void
  updateItemConfiguration: (oldItemId: string, product: ProductDTO, configuration?: CartItemConfiguration) => void
  clearCart: () => void
  getItemQuantity: (productId: string, configuration?: CartItemConfiguration) => number
  getTotalProductQuantity: (productId: string) => number
}
