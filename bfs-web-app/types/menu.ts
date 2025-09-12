export interface MenuCategory {
  id: string
  name: string
  description: string
  sortOrder: number
  isActive: boolean
  createdAt: Date
  updatedAt: Date
}

export interface MenuItemModifier {
  id: string
  name: string
  price: number
  isRequired: boolean
  maxSelections?: number
  options: ModifierOption[]
}

export interface ModifierOption {
  id: string
  name: string
  price: number
  isAvailable: boolean
}

export interface MenuItem {
  id: string
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
  createdAt: Date
  updatedAt: Date
}

export interface NutritionInfo {
  calories: number
  protein: number
  carbs: number
  fat: number
  fiber: number
  sodium: number
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
