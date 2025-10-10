import type { ListResponse, ProductDTO } from "@/types"
import { apiRequest } from "../api"

export interface ListProductsParams {
  categoryId?: string
  limit?: number
  offset?: number
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

  return apiRequest<ListResponse<ProductDTO>>(endpoint)
}
