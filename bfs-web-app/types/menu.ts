import { BaseEntity } from './common'

export interface MenuCategory extends BaseEntity {
  name: string
  description: string
  sortOrder: number
  isActive: boolean
}

export interface ModifierOption {
  id: string
  name: string
  price: number
  isAvailable: boolean
}

export interface MenuItemModifier {
  id: string
  name: string
  price: number
  isRequired: boolean
  maxSelections?: number
  options: ModifierOption[]
}

export interface NutritionInfo {
  calories: number
  protein: number
  carbs: number
  fat: number
  fiber: number
  sodium: number
}

export interface MenuItem extends BaseEntity {
  categoryId: string
  name: string
  description: string
  price: number
  image?: string
  ingredients: string[]
  allergens: string[]
  nutritionInfo?: NutritionInfo
  modifiers: MenuItemModifier[]
  isAvailable: boolean
  sortOrder: number
  preparationTime: number
}

export interface MenuWithCategories {
  categories: MenuCategory[]
  items: MenuItem[]
}

export interface CreateMenuItemRequest {
  categoryId: string
  name: string
  description: string
  price: number
  image?: string
  ingredients: string[]
  allergens: string[]
  nutritionInfo?: NutritionInfo
  modifiers: Omit<MenuItemModifier, "id">[]
  preparationTime: number
}

export interface UpdateMenuItemRequest extends Partial<CreateMenuItemRequest> {
  id: string
  isAvailable?: boolean
}