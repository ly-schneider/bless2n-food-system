import { MenuItem, ModifierOption } from "./menu"

export interface CartItemModifier {
  modifierId: string
  name: string
  selectedOptions: ModifierOption[]
  totalPrice: number
}

export interface CartItem {
  id: string
  menuItem: MenuItem
  quantity: number
  modifiers: CartItemModifier[]
  specialInstructions?: string
  unitPrice: number
  totalPrice: number
  addedAt: Date
}

export interface Cart {
  id: string
  userId?: string
  guestId?: string
  items: CartItem[]
  subtotal: number
  tax: number
  discount: number
  total: number
  estimatedPreparationTime: number
  createdAt: Date
  updatedAt: Date
}

export interface AddToCartRequest {
  menuItemId: string
  quantity: number
  modifiers: {
    modifierId: string
    selectedOptionIds: string[]
  }[]
  specialInstructions?: string
}

export interface UpdateCartItemRequest {
  cartItemId: string
  quantity: number
  modifiers?: {
    modifierId: string
    selectedOptionIds: string[]
  }[]
  specialInstructions?: string
}

export interface CartSummary {
  itemCount: number
  subtotal: number
  tax: number
  total: number
}
