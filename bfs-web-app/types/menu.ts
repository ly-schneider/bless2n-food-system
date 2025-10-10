import type { ProductSummaryDTO } from "./product"

export interface MenuSlot {
  id: string
  productId: string
  name: string
  sequence: number
}

export interface MenuSlotDTO {
  id: string
  name: string
  sequence: number
  menuSlotItems: MenuSlotItemDTO[] | null
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
