import type { ListResponse, ProductDTO } from "@/types"
import { apiRequest } from "../api"

export interface ListProductsParams {
  categoryId?: string
  limit?: number
  offset?: number
}

/**
 * The backend returns menu slot data as `product.menuSlots[]` with option items
 * containing `productId`. The frontend expects `product.menu.slots[]` with items
 * containing `id`. This function normalizes the response shape.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function normalizeProduct(raw: any): ProductDTO {
  const product = raw as ProductDTO

  // Compute availability from stock (backend returns stock as inventory count)
  // Menu products don't have stock - they're always available
  const isMenu = raw.type === "menu"
  const stock = raw.stock as number | undefined
  if (!isMenu && typeof stock === "number") {
    product.availableQuantity = stock
    product.isAvailable = stock > 0
    product.isLowStock = stock > 0 && stock <= 10
  }

  // Map menuSlots (backend shape) → menu.slots (frontend shape)
  const menuSlots = raw.menuSlots
  if (menuSlots && Array.isArray(menuSlots) && menuSlots.length > 0) {
    product.menu = {
      slots: menuSlots.map((slot: any) => ({
        // eslint-disable-line @typescript-eslint/no-explicit-any
        id: slot.id,
        name: slot.name,
        sequence: slot.sequence ?? 0,
        options: Array.isArray(slot.options)
          ? slot.options.map((opt: any) => ({
              // eslint-disable-line @typescript-eslint/no-explicit-any
              id: opt.productId ?? opt.id,
              name: opt.name ?? "",
              priceCents: opt.priceCents ?? 0,
              type: "simple" as const,
              image: opt.image ?? null,
              isActive: true,
              isAvailable: true,
              category: null,
              jeton: opt.jeton ? { id: opt.jeton.id, name: opt.jeton.name, color: opt.jeton.color } : undefined,
            }))
          : null,
      })),
    }
  }

  return product
}

export async function listProducts(params: ListProductsParams = {}): Promise<ListResponse<ProductDTO>> {
  const searchParams = new URLSearchParams()

  if (params.categoryId) {
    searchParams.append("category_id", params.categoryId)
  }

  if (params.limit !== undefined) {
    searchParams.append("limit", params.limit.toString())
  }

  if (params.offset !== undefined) {
    searchParams.append("offset", params.offset.toString())
  }

  const queryString = searchParams.toString()
  const endpoint = `/v1/products${queryString ? `?${queryString}` : ""}`

  const raw = await apiRequest<ListResponse<ProductDTO>>(endpoint)
  const items = (raw.items || []).map(normalizeProduct)

  // Build stock map from simple products to apply to menu slot options
  const stockMap = new Map<string, { stock: number; isAvailable: boolean; isLowStock: boolean }>()
  for (const p of items) {
    if (p.type !== "menu" && typeof p.availableQuantity === "number") {
      stockMap.set(p.id, {
        stock: p.availableQuantity,
        isAvailable: p.isAvailable !== false,
        isLowStock: p.isLowStock === true,
      })
    }
  }

  // Apply stock to menu slot options
  for (const p of items) {
    if (p.type === "menu" && p.menu?.slots) {
      for (const slot of p.menu.slots) {
        if (slot.options) {
          for (const opt of slot.options) {
            const stockInfo = stockMap.get(opt.id)
            if (stockInfo) {
              opt.availableQuantity = stockInfo.stock
              opt.isAvailable = stockInfo.isAvailable
              opt.isLowStock = stockInfo.isLowStock
            }
          }
        }
      }
    }
  }

  return { ...raw, items }
}
