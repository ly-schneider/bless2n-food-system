import type { ListResponse, ProductDTO, ProductSummaryDTO } from "@/types"
import { apiRequest } from "../api"

export interface ListProductsParams {
  categoryId?: string
  limit?: number
  offset?: number
}

interface RawSlotOption {
  productId?: string
  id?: string
  name?: string
  priceCents?: number
  image?: string | null
  jeton?: { id: string; name: string; color: string } | null
}

interface RawMenuSlot {
  id: string
  name: string
  sequence?: number
  options?: RawSlotOption[]
}

interface RawProduct extends ProductDTO {
  stock?: number
  menuSlots?: RawMenuSlot[]
}

function normalizeProduct(raw: RawProduct): ProductDTO {
  const product = raw as ProductDTO

  const isMenu = raw.type === "menu"
  const stock = raw.stock
  if (!isMenu && typeof stock === "number") {
    product.availableQuantity = stock
    product.isAvailable = stock > 0
    product.isLowStock = stock > 0 && stock <= 10
  }

  const menuSlots = raw.menuSlots
  if (menuSlots && menuSlots.length > 0) {
    product.menu = {
      slots: menuSlots.map((slot) => ({
        id: slot.id,
        name: slot.name,
        sequence: slot.sequence ?? 0,
        options: Array.isArray(slot.options)
          ? slot.options.map((opt) => ({
              id: opt.productId ?? opt.id ?? "",
              name: opt.name ?? "",
              priceCents: opt.priceCents ?? 0,
              type: "simple" as const,
              image: opt.image ?? null,
              isActive: true,
              isAvailable: true,
              category: null as unknown as ProductSummaryDTO["category"],
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

  const raw = await apiRequest<ListResponse<RawProduct>>(endpoint)
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
