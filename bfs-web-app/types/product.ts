import { ActivatableNamedEntity, ListResponse, MessageResponse, StatusUpdateResponse } from './common'

export type Category = ActivatableNamedEntity

export interface Product extends ActivatableNamedEntity {
  categoryId: string
  type: string
  image: string
  price: number
}

export interface CreateCategoryRequest {
  name: string
  description?: string
}

export interface UpdateCategoryRequest {
  name?: string
  description?: string
}

export interface CreateProductRequest {
  name: string
  description?: string
  price: number
  categoryId: string
  stock?: number
}

export interface UpdateProductRequest {
  name?: string
  description?: string
  price?: number
  categoryId?: string
  stock?: number
}

export interface CreateProductBundleRequest {
  name: string
  description?: string
  price: number
  productIds: string[]
}

export interface UpdateProductBundleRequest {
  name?: string
  description?: string
  price?: number
  productIds?: string[]
}

export interface UpdateProductStockRequest {
  stock: number
}

export interface AssignProductToStationsRequest {
  stationIds: string[]
}

// Response types - using common patterns
export type CreateCategoryResponse = Category
export type UpdateCategoryResponse = Category
export type GetCategoryResponse = Category
export interface ListCategoriesResponse extends ListResponse<Category> {
  // Legacy compatibility field
  categories: Category[]
}
export type DeleteCategoryResponse = MessageResponse
export type SetCategoryActiveResponse = StatusUpdateResponse

export type CreateProductResponse = Product
export type UpdateProductResponse = Product
export type GetProductResponse = Product
export interface ListProductsResponse extends ListResponse<Product> {
  // Legacy compatibility field
  products: Product[]
}
export type DeleteProductResponse = MessageResponse
export type SetProductActiveResponse = StatusUpdateResponse

export interface UpdateProductStockResponse extends MessageResponse {
  id: string
  stock: number
}

export type AssignProductToStationsResponse = MessageResponse

export type CreateProductBundleResponse = ProductBundle
export type UpdateProductBundleResponse = ProductBundle