export interface Category {
  id: string
  name: string
  isActive: boolean
  // Required zero-based sort position; lower comes first
  position: number
  createdAt: string // ISO date
  updatedAt: string // ISO date
}

export interface CategoryDTO {
  id: string
  name: string
  isActive: boolean
  // Required zero-based sort position; lower comes first
  position: number
}
