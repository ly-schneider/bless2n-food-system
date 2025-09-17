import { BaseEntity } from './common'
import { MenuItem, ModifierOption } from './menu'

export interface CartItemModifier {
  modifierId: string
  name: string
  selectedOptions: ModifierOption[]
  totalPrice: number
}

export interface CartItem extends BaseEntity {
  menuItem: MenuItem
  quantity: number
  modifiers: CartItemModifier[]
  specialInstructions?: string
  unitPrice: number
  totalPrice: number
  addedAt: Date
}

export interface Cart extends BaseEntity {
  userId?: string
  guestId?: string
  items: CartItem[]
  subtotal: number
  tax: number
  discount: number
  total: number
  estimatedPreparationTime: number
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