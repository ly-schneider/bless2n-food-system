import {
  AssignProductsToStationRequest,
  AssignProductsToStationResponse,
  CreateStationRequest,
  CreateStationResponse,
  GetStationProductsResponse,
  GetStationResponse,
  ListStationsResponse,
  RemoveProductFromStationResponse,
  RequestStationRequest,
  RequestStationResponse,
  UpdateStationStatusRequest,
  UpdateStationStatusResponse,
} from "../../types/station"
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

  if (!response.ok) {
    const error = await response.text()
    throw new Error(error || `HTTP ${response.status}`)
  }

  const data = (await response.json()) as unknown
  return data as T
}

// Public Station API methods
export async function requestStation(request: RequestStationRequest): Promise<RequestStationResponse> {
  return apiRequest<RequestStationResponse>("/v1/stations/request", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

export async function listStations(params?: {
  activeOnly?: boolean
  limit?: number
  offset?: number
}): Promise<ListStationsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.activeOnly) searchParams.set("active_only", "true")
  if (params?.limit) searchParams.set("limit", params.limit.toString())
  if (params?.offset) searchParams.set("offset", params.offset.toString())

  const queryString = searchParams.toString()
  const endpoint = `/v1/stations${queryString ? `?${queryString}` : ""}`

  return apiRequest<ListStationsResponse>(endpoint)
}

export async function getStation(id: string): Promise<GetStationResponse> {
  return apiRequest<GetStationResponse>(`/v1/stations/${id}`)
}

export async function getStationProducts(id: string): Promise<GetStationProductsResponse> {
  return apiRequest<GetStationProductsResponse>(`/v1/stations/${id}/products`)
}

// Admin Station API methods
export async function createStation(request: CreateStationRequest): Promise<CreateStationResponse> {
  return apiRequest<CreateStationResponse>("/v1/admin/stations", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

export async function updateStationStatus(
  id: string,
  request: UpdateStationStatusRequest
): Promise<UpdateStationStatusResponse> {
  return apiRequest<UpdateStationStatusResponse>(`/v1/admin/stations/${id}/status`, {
    method: "PUT",
    body: JSON.stringify(request),
  })
}

export async function assignProductsToStation(
  id: string,
  request: AssignProductsToStationRequest
): Promise<AssignProductsToStationResponse> {
  return apiRequest<AssignProductsToStationResponse>(`/v1/admin/stations/${id}/products`, {
    method: "POST",
    body: JSON.stringify(request),
  })
}

export async function removeProductFromStation(
  stationId: string,
  productId: string
): Promise<RemoveProductFromStationResponse> {
  return apiRequest<RemoveProductFromStationResponse>(`/v1/admin/stations/${stationId}/products/${productId}`, {
    method: "DELETE",
  })
}

const StationAPI = {
  requestStation,
  listStations,
  getStation,
  getStationProducts,
  createStation,
  updateStationStatus,
  assignProductsToStation,
  removeProductFromStation,
}

export default StationAPI
