import {
  AssignProductToStationsRequest,
  AssignProductToStationsResponse,
  CreateCategoryRequest,
  CreateCategoryResponse,
  CreateProductBundleRequest,
  CreateProductBundleResponse,
  CreateProductRequest,
  CreateProductResponse,
  DeleteCategoryResponse,
  DeleteProductResponse,
  GetCategoryResponse,
  GetProductResponse,
  ListCategoriesResponse,
  ListProductsResponse,
  SetCategoryActiveResponse,
  SetProductActiveResponse,
  UpdateCategoryRequest,
  UpdateCategoryResponse,
  UpdateProductBundleRequest,
  UpdateProductBundleResponse,
  UpdateProductRequest,
  UpdateProductResponse,
  UpdateProductStockRequest,
  UpdateProductStockResponse,
} from "@/types"
import { AuthService } from "../auth"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"

async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const accessToken = await AuthService.getAccessToken()

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(accessToken && { Authorization: `Bearer ${accessToken}` }),
      ...options.headers,
    },
  })

  console.log(response)

  if (!response.ok) {
    const error = await response.text()
    throw new Error(error || `HTTP ${response.status}`)
  }

  const data = (await response.json()) as unknown
  console.log(data)
  return data as T
}

// Category API methods
export async function createCategory(request: CreateCategoryRequest): Promise<CreateCategoryResponse> {
  return apiRequest<CreateCategoryResponse>("/v1/admin/categories", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

export async function getCategory(id: string): Promise<GetCategoryResponse> {
  return apiRequest<GetCategoryResponse>(`/v1/admin/categories/${id}`)
}

export async function updateCategory(id: string, request: UpdateCategoryRequest): Promise<UpdateCategoryResponse> {
  return apiRequest<UpdateCategoryResponse>(`/v1/admin/categories/${id}`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

export async function deleteCategory(id: string): Promise<DeleteCategoryResponse> {
  return apiRequest<DeleteCategoryResponse>(`/v1/admin/categories/${id}`, {
    method: "DELETE",
  })
}

export async function listCategories(params?: {
  activeOnly?: boolean
  limit?: number
  offset?: number
}): Promise<ListCategoriesResponse> {
  const searchParams = new URLSearchParams()
  if (params?.activeOnly) searchParams.set("active_only", "true")
  if (params?.limit) searchParams.set("limit", params.limit.toString())
  if (params?.offset) searchParams.set("offset", params.offset.toString())

  const queryString = searchParams.toString()
  const endpoint = `/v1/admin/categories${queryString ? `?${queryString}` : ""}`

  return apiRequest<ListCategoriesResponse>(endpoint)
}

export async function setCategoryActive(id: string, active: boolean): Promise<SetCategoryActiveResponse> {
  return apiRequest<SetCategoryActiveResponse>(`/v1/admin/categories/${id}/status?active=${active}`, {
    method: "PUT",
  })
}

// Product API methods
export async function createProduct(request: CreateProductRequest): Promise<CreateProductResponse> {
  return apiRequest<CreateProductResponse>("/v1/admin/products", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

export async function getProduct(id: string): Promise<GetProductResponse> {
  return apiRequest<GetProductResponse>(`/v1/admin/products/${id}`)
}

export async function updateProduct(id: string, request: UpdateProductRequest): Promise<UpdateProductResponse> {
  return apiRequest<UpdateProductResponse>(`/v1/admin/products/${id}`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

export async function deleteProduct(id: string): Promise<DeleteProductResponse> {
  return apiRequest<DeleteProductResponse>(`/v1/admin/products/${id}`, {
    method: "DELETE",
  })
}

export async function listProducts(params?: {
  categoryId?: string
  activeOnly?: boolean
  limit?: number
  offset?: number
}): Promise<ListProductsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.categoryId) searchParams.set("category_id", params.categoryId)
  if (params?.activeOnly) searchParams.set("active_only", "true")
  if (params?.limit) searchParams.set("limit", params.limit.toString())
  if (params?.offset) searchParams.set("offset", params.offset.toString())

  const queryString = searchParams.toString()
  const endpoint = `/v1/admin/products${queryString ? `?${queryString}` : ""}`

  return apiRequest<ListProductsResponse>(endpoint)
}

// Public list products (no admin auth required)
export async function listPublicProducts(params?: {
  categoryId?: string
  activeOnly?: boolean
  limit?: number
  offset?: number
}): Promise<ListProductsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.categoryId) searchParams.set("category_id", params.categoryId)
  if (params?.limit) searchParams.set("limit", params.limit.toString())
  if (params?.offset) searchParams.set("offset", params.offset.toString())

  const queryString = searchParams.toString()
  const endpoint = `/v1/products${queryString ? `?${queryString}` : ""}`

  return apiRequest<ListProductsResponse>(endpoint)
}

export async function setProductActive(id: string, active: boolean): Promise<SetProductActiveResponse> {
  return apiRequest<SetProductActiveResponse>(`/v1/admin/products/${id}/status?active=${active}`, {
    method: "PUT",
  })
}

export async function updateProductStock(
  id: string,
  request: UpdateProductStockRequest
): Promise<UpdateProductStockResponse> {
  return apiRequest<UpdateProductStockResponse>(`/v1/admin/products/${id}/stock`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

export async function assignProductToStations(
  id: string,
  request: AssignProductToStationsRequest
): Promise<AssignProductToStationsResponse> {
  return apiRequest<AssignProductToStationsResponse>(`/v1/admin/products/${id}/stations`, {
    method: "POST",
    body: JSON.stringify(request),
  })
}

// Product Bundle API methods
export async function createProductBundle(request: CreateProductBundleRequest): Promise<CreateProductBundleResponse> {
  return apiRequest<CreateProductBundleResponse>("/v1/admin/products/bundles", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

export async function updateProductBundle(
  id: string,
  request: UpdateProductBundleRequest
): Promise<UpdateProductBundleResponse> {
  return apiRequest<UpdateProductBundleResponse>(`/v1/admin/products/bundles/${id}`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

const ProductAPI = {
  createCategory,
  getCategory,
  updateCategory,
  deleteCategory,
  listCategories,
  setCategoryActive,
  createProduct,
  getProduct,
  updateProduct,
  deleteProduct,
  listProducts,
  listPublicProducts,
  setProductActive,
  updateProductStock,
  assignProductToStations,
  createProductBundle,
  updateProductBundle,
}

export default ProductAPI
