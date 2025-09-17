import { ActivatableNamedEntity, ListResponse, MessageResponse, StatusUpdateResponse } from './common'
import { Product } from './product'

export interface Station extends ActivatableNamedEntity {
  location: string
}

export interface RequestStationRequest {
  businessName: string
  contactEmail: string
  contactName: string
  location: string
  description?: string
}

export interface CreateStationRequest {
  name: string
  location: string
  description?: string
}

export interface UpdateStationStatusRequest {
  isActive: boolean
}

export interface AssignProductsToStationRequest {
  productIds: string[]
}

// Response types
export type RequestStationResponse = MessageResponse
export type CreateStationResponse = Station
export type GetStationResponse = Station
export type ListStationsResponse = ListResponse<Station>
export type UpdateStationStatusResponse = StatusUpdateResponse
export type AssignProductsToStationResponse = MessageResponse
export type RemoveProductFromStationResponse = MessageResponse

export type GetStationProductsResponse = ListResponse<Product>

// Station Product type (alias for compatibility)
export type StationProduct = Product