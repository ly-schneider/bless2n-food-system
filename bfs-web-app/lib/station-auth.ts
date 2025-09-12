import { cookies } from "next/headers"
import { NextRequest } from "next/server"
import { RequestStationRequest, RequestStationResponse } from "../types/station"
import {
  StationLoginRequest,
  StationLoginResponse,
  StationRequestForm,
  StationSession,
  StationStatus,
} from "../types/station-auth"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"
const STATION_TOKEN_COOKIE = "station_access_token"

export class StationAuthService {
  private static getAuthHeaders(token?: string): HeadersInit {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
    }

    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    return headers
  }

  private static async apiRequest<T>(endpoint: string, options: RequestInit = {}, token?: string): Promise<T> {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        ...this.getAuthHeaders(token),
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

  // Station request flow (public - no auth required)
  static async requestStationAccess(request: StationRequestForm): Promise<RequestStationResponse> {
    const stationRequest: RequestStationRequest = {
      businessName: request.businessName,
      contactEmail: request.contactEmail,
      contactName: request.contactName,
      location: request.location,
      description: request.description,
    }

    return this.apiRequest<RequestStationResponse>("/v1/stations/request", {
      method: "POST",
      body: JSON.stringify(stationRequest),
    })
  }

  // Station login (when admin approves and provides access code)
  static async loginAsStation(request: StationLoginRequest): Promise<StationLoginResponse> {
    // This endpoint would need to be added to backend
    // For now, we'll simulate the response structure
    const response = await this.apiRequest<StationLoginResponse>("/v1/stations/auth/login", {
      method: "POST",
      body: JSON.stringify(request),
    })

    // Store station token in secure cookie
    await this.setStationToken(response.accessToken)

    return response
  }

  static async logoutStation(): Promise<void> {
    try {
      const token = await this.getStationToken()
      if (token) {
        await this.apiRequest(
          "/v1/stations/auth/logout",
          {
            method: "POST",
          },
          token || undefined
        )
      }
    } catch {
      // Continue with logout even if API call fails
    }

    await this.clearStationToken()
  }

  private static async setStationToken(token: string): Promise<void> {
    const cookieStore = await cookies()

    cookieStore.set(STATION_TOKEN_COOKIE, token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax", // WebView compatibility
      maxAge: 8 * 60 * 60, // 8 hours (station shifts)
      path: "/",
    })
  }

  static async getStationToken(): Promise<string | null> {
    const cookieStore = await cookies()
    return cookieStore.get(STATION_TOKEN_COOKIE)?.value || null
  }

  static async getStationTokenFromRequest(request: NextRequest): Promise<string | null> {
    return request.cookies.get(STATION_TOKEN_COOKIE)?.value || null
  }

  private static async clearStationToken(): Promise<void> {
    const cookieStore = await cookies()
    cookieStore.delete(STATION_TOKEN_COOKIE)
  }

  static async getCurrentStation(): Promise<StationSession | null> {
    const token = await this.getStationToken()
    if (!token) return null

    try {
      // This endpoint would need to be added to backend
      const stationInfo = await this.apiRequest<{
        id: string
        name: string
        location: string
        status: string
        permissions: string[]
      }>(
        "/v1/stations/auth/profile",
        {
          method: "GET",
        },
        token
      )

      return {
        stationId: stationInfo.id,
        stationName: stationInfo.name,
        location: stationInfo.location,
        status: stationInfo.status as StationStatus,
        permissions: stationInfo.permissions.map((p) => ({ action: p as "redeem" | "view_products" | "view_orders" })),
        isAuthenticated: true,
        expiresAt: new Date(Date.now() + 8 * 60 * 60 * 1000), // 8 hours
      }
    } catch {
      return null
    }
  }

  // Station-specific API methods with station token
  static async getStationProducts(stationId: string): Promise<unknown> {
    const token = await this.getStationToken()
    return this.apiRequest(
      `/v1/stations/${stationId}/products`,
      {
        method: "GET",
      },
      token || undefined
    )
  }

  static async redeemOrder(orderId: string, items: unknown[]): Promise<unknown> {
    // Use public redemption API (no auth required for station devices)
    return this.apiRequest("/v1/redemption/redeem", {
      method: "POST",
      body: JSON.stringify({ orderId, items }),
    })
  }

  static async getOrderForRedemption(orderId: string): Promise<unknown> {
    // Use public redemption API (no auth required for station devices)
    return this.apiRequest(`/v1/redemption/orders/${orderId}`)
  }
}
