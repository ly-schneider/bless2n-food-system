// Product and Category types matching backend API

export interface Category {
  id: string
  name: string
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface Product {
  id: string
  name: string
  description?: string
  price: number
  categoryId: string
  isActive: boolean
  stock: number
  createdAt: string
  updatedAt: string
}

export interface ProductBundle {
  id: string
  name: string
  description?: string
  price: number
  productIds: string[]
  isActive: boolean
  createdAt: string
  updatedAt: string
}

// Category API requests/responses
export interface CreateCategoryRequest {
  name: string
  description?: string
}

export interface CreateCategoryResponse {
  id: string
  name: string
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface UpdateCategoryRequest {
  name?: string
  description?: string
}

export interface UpdateCategoryResponse {
  id: string
  name: string
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface GetCategoryResponse {
  id: string
  name: string
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface ListCategoriesResponse {
  categories: Category[]
  total: number
  limit: number
  offset: number
}

export interface DeleteCategoryResponse {
  message: string
}

export interface SetCategoryActiveResponse {
  id: string
  isActive: boolean
  message: string
}

// Product API requests/responses
export interface CreateProductRequest {
  name: string
  description?: string
  price: number
  categoryId: string
  stock?: number
}

export interface CreateProductResponse {
  id: string
  name: string
  description?: string
  price: number
  categoryId: string
  isActive: boolean
  stock: number
  createdAt: string
  updatedAt: string
}

export interface UpdateProductRequest {
  name?: string
  description?: string
  price?: number
  categoryId?: string
  stock?: number
}

export interface UpdateProductResponse {
  id: string
  name: string
  description?: string
  price: number
  categoryId: string
  isActive: boolean
  stock: number
  createdAt: string
  updatedAt: string
}

export interface GetProductResponse {
  id: string
  name: string
  description?: string
  price: number
  categoryId: string
  isActive: boolean
  stock: number
  createdAt: string
  updatedAt: string
}

export interface ListProductsResponse {
  products: Product[]
  total: number
  limit: number
  offset: number
}

export interface DeleteProductResponse {
  message: string
}

export interface SetProductActiveResponse {
  id: string
  isActive: boolean
  message: string
}

export interface UpdateProductStockRequest {
  stock: number
}

export interface UpdateProductStockResponse {
  id: string
  stock: number
  message: string
}

export interface AssignProductToStationsRequest {
  stationIds: string[]
}

export interface AssignProductToStationsResponse {
  message: string
}

// Product Bundle API requests/responses
export interface CreateProductBundleRequest {
  name: string
  description?: string
  price: number
  productIds: string[]
}

export interface CreateProductBundleResponse {
  id: string
  name: string
  description?: string
  price: number
  productIds: string[]
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface UpdateProductBundleRequest {
  name?: string
  description?: string
  price?: number
  productIds?: string[]
}

export interface UpdateProductBundleResponse {
  id: string
  name: string
  description?: string
  price: number
  productIds: string[]
  isActive: boolean
  createdAt: string
  updatedAt: string
}
