import { BaseEntity, MessageResponse } from './common'

export enum Role {
  ADMIN = "admin",
  CUSTOMER = "customer",
  STATION = "station",
}

export interface User extends BaseEntity {
  email: string
  name: string
  role: Role
  isActive: boolean
  isEmailVerified: boolean
}

export interface Session {
  id: string
  userId: string | null
  accessToken: string
  refreshToken: string
  expiresAt: Date
  isGuest: boolean
  guestId?: string
  createdAt: Date
}

export interface RegisterCustomerRequest {
  email: string
  name: string
}

export interface RequestOTPRequest {
  email: string
}

export interface LoginRequest {
  email: string
  otp: string
}

export interface LoginResponse {
  accessToken: string
  refreshToken: string
  user: User
}

export interface RefreshTokenRequest {
  refreshToken: string
}

export interface RefreshTokenResponse {
  accessToken: string
  refreshToken: string
}

export interface LogoutRequest {
  refreshToken: string
}

export interface RegisterCustomerResponse extends MessageResponse {
  userId: string
}

export type RequestOTPResponse = MessageResponse
export type LogoutResponse = MessageResponse

export interface AuthContext {
  user: User | null
  isAuthenticated: boolean
  isGuest: boolean
  registerCustomer: (email: string, name: string) => Promise<void>
  requestOTP: (email: string) => Promise<void>
  login: (email: string, otp: string) => Promise<void>
  logout: () => Promise<void>
  refreshToken: () => Promise<void>
}