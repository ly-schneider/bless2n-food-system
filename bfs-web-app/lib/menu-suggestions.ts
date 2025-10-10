import { Cart, CartItem } from "@/types/cart"
import { MenuDTO, MenuSlotDTO } from "@/types/menu"
import { ProductDTO } from "@/types/product"

export interface MenuSuggestion {
  menuProduct: ProductDTO
  configuration: Record<string, string> // slotId -> productId
  sourceItems: { slotId: string; cartItem: CartItem }[]
  savingsCents: number
}

function canMapCartItemToSlot(cartItem: CartItem, slot: MenuSlotDTO): boolean {
  if (!slot.menuSlotItems) return false
  return slot.menuSlotItems.some((p) => p.id === cartItem.product.id)
}

function buildConfiguration(
  slots: MenuSlotDTO[],
  chosen: { slotId: string; cartItem: CartItem }[]
): Record<string, string> {
  const cfg: Record<string, string> = {}
  for (const { slotId, cartItem } of chosen) {
    cfg[slotId] = cartItem.product.id
  }
  // Ensure all slots are present (fallback none)
  for (const s of slots) {
    if (!(s.id in cfg)) cfg[s.id] = ""
  }
  return cfg
}

export function findBestMenuSuggestion(cart: Cart, allProducts: ProductDTO[]): MenuSuggestion | null {
  const menus = allProducts.filter((p) => p.type === "menu" && p.menu && p.menu.slots && p.menu.slots.length > 0)
  if (menus.length === 0) return null

  // Consider only simple products in cart with available quantity
  const simpleItems = cart.items.filter((i) => i.product.type === "simple" && i.quantity > 0)
  if (simpleItems.length === 0) return null

  let best: MenuSuggestion | null = null

  for (const menu of menus) {
    const slots = (menu.menu as MenuDTO).slots as MenuSlotDTO[]

    // For each slot, try to find a cart item that fits it
    const chosen: { slotId: string; cartItem: CartItem }[] = []
    const usedItemIds = new Set<string>()

    for (const slot of slots) {
      const match = simpleItems.find((it) => !usedItemIds.has(it.id) && canMapCartItemToSlot(it, slot))
      if (!match) {
        // This menu cannot be built from current cart
        chosen.length = 0
        break
      }
      chosen.push({ slotId: slot.id, cartItem: match })
      usedItemIds.add(match.id)
    }

    if (chosen.length === slots.length) {
      const sumSimple = chosen.reduce((sum, ch) => sum + ch.cartItem.product.priceCents, 0)
      const savings = sumSimple - menu.priceCents
      if (savings > 0) {
        const suggestion: MenuSuggestion = {
          menuProduct: menu,
          configuration: buildConfiguration(slots, chosen),
          sourceItems: chosen,
          savingsCents: savings,
        }
        if (!best || suggestion.savingsCents > best.savingsCents) {
          best = suggestion
        }
      }
    }
  }

  return best
}
