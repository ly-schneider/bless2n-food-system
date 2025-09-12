// Station types matching backend API

export interface Station {
  id: string
  name: string
  location: string
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

// Station API requests/responses
export interface RequestStationRequest {
  businessName: string
  contactEmail: string
  contactName: string
  location: string
  description?: string
}

export interface RequestStationResponse {
  message: string
}

export interface CreateStationRequest {
  name: string
  location: string
  description?: string
}

export interface CreateStationResponse {
  id: string
  name: string
  location: string
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface GetStationResponse {
  id: string
  name: string
  location: string
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface ListStationsResponse {
  stations: Station[]
  total: number
  limit: number
  offset: number
}

export interface UpdateStationStatusRequest {
  isActive: boolean
}

export interface UpdateStationStatusResponse {
  id: string
  isActive: boolean
  message: string
}

export interface AssignProductsToStationRequest {
  productIds: string[]
}

export interface AssignProductsToStationResponse {
  message: string
}

export interface RemoveProductFromStationResponse {
  message: string
}

export interface GetStationProductsResponse {
  products: StationProduct[]
  total: number
}

export interface StationProduct {
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
