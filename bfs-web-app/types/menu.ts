import type { ProductSummaryDTO } from "./product"

/** Full menu object returned by GET /v1/menus */
export interface Menu {
  id: string
  categoryId: string
  name: string
  image: string | null
  priceCents: number
  isActive: boolean
  createdAt: string
  updatedAt: string
  slots: MenuSlot[]
}

export interface MenuSlot {
  id: string
  menuProductId: string
  name: string
  sequence: number
  options: MenuSlotOption[]
}

export interface MenuSlotOption {
  menuSlotId: string
  optionProductId: string
  optionProduct?: {
    id: string
    name: string
    priceCents: number
    image: string | null
  }
}

export interface MenuSlotDTO {
  id: string
  name: string
  sequence: number
  options: MenuSlotItemDTO[] | null
}

export interface MenuSlotItem {
  id: string
  menuSlotId: string
  productId: string
}

export type MenuSlotItemDTO = ProductSummaryDTO

export interface MenuDTO {
  slots: MenuSlotDTO[] | null
}
