import { MessageResponse } from './common'

export enum StationStatus {
  PENDING = "pending",
  APPROVED = "approved", 
  REJECTED = "rejected",
  SUSPENDED = "suspended",
}

export interface StationPermission {
  action: "redeem" | "view_products" | "view_orders"
  resourceId?: string
}

export interface StationSession {
  stationId: string
  stationName: string
  location: string
  status: StationStatus
  permissions: StationPermission[]
  isAuthenticated: boolean
  expiresAt: Date
}

export interface StationRequestForm {
  businessName: string
  contactEmail: string
  contactName: string
  location: string
  description?: string
  businessType: string
  operatingHours: string
}

export interface StationLoginRequest {
  stationId: string
  accessCode: string
}

export interface StationLoginResponse extends MessageResponse {
  stationId: string
  stationName: string
  location: string
  status: StationStatus
  accessToken: string
  permissions: StationPermission[]
}

export interface StationContext {
  station: StationSession | null
  isStationAuthenticated: boolean
  requestStationAccess: (request: StationRequestForm) => Promise<void>
  loginAsStation: (stationId: string, accessCode: string) => Promise<void>
  logoutStation: () => Promise<void>
}