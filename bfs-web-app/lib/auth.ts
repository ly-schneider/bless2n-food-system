import { cookies } from "next/headers"
import { NextRequest } from "next/server"
import {
  LoginRequest,
  LoginResponse,
  LogoutResponse,
  RefreshTokenResponse,
  RegisterCustomerRequest,
  RegisterCustomerResponse,
  RequestOTPRequest,
  RequestOTPResponse,
  Session,
  User,
} from "@/types"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"
const ACCESS_TOKEN_COOKIE = "access_token"
const REFRESH_TOKEN_COOKIE = "refresh_token"

export class AuthService {
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

  static async registerCustomer(request: RegisterCustomerRequest): Promise<RegisterCustomerResponse> {
    return this.apiRequest<RegisterCustomerResponse>("/v1/auth/register/customer", {
      method: "POST",
      body: JSON.stringify(request),
    })
  }

  static async requestOTP(request: RequestOTPRequest): Promise<RequestOTPResponse> {
    return this.apiRequest<RequestOTPResponse>("/v1/auth/request-otp", {
      method: "POST",
      body: JSON.stringify(request),
    })
  }

  static async login(request: LoginRequest): Promise<LoginResponse> {
    const response = await this.apiRequest<LoginResponse>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(request),
    })

    // Store tokens in secure cookies
    await this.setTokens(response.accessToken, response.refreshToken)

    return response
  }

  static async refreshToken(): Promise<RefreshTokenResponse> {
    const cookieStore = await cookies()
    const refreshToken = cookieStore.get(REFRESH_TOKEN_COOKIE)?.value

    if (!refreshToken) {
      throw new Error("No refresh token available")
    }

    const response = await this.apiRequest<RefreshTokenResponse>("/v1/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refreshToken }),
    })

    // Update tokens in cookies
    await this.setTokens(response.accessToken, response.refreshToken)

    return response
  }

  static async logout(): Promise<void> {
    const cookieStore = await cookies()
    const refreshToken = cookieStore.get(REFRESH_TOKEN_COOKIE)?.value

    if (refreshToken) {
      try {
        await this.apiRequest<LogoutResponse>("/v1/auth/logout", {
          method: "POST",
          body: JSON.stringify({ refreshToken }),
        })
      } catch {
        // Continue with logout even if API call fails
      }
    }

    await this.clearTokens()
  }

  private static async setTokens(accessToken: string, refreshToken: string): Promise<void> {
    const cookieStore = await cookies()

    // Set access token (shorter expiry)
    cookieStore.set(ACCESS_TOKEN_COOKIE, accessToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax", // WebView compatibility
      maxAge: 15 * 60, // 15 minutes
      path: "/",
    })

    // Set refresh token (longer expiry)
    cookieStore.set(REFRESH_TOKEN_COOKIE, refreshToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax", // WebView compatibility
      maxAge: 7 * 24 * 60 * 60, // 7 days
      path: "/",
    })
  }

  static async getAccessToken(): Promise<string | null> {
    const cookieStore = await cookies()
    return cookieStore.get(ACCESS_TOKEN_COOKIE)?.value || null
  }

  static async getRefreshToken(): Promise<string | null> {
    const cookieStore = await cookies()
    return cookieStore.get(REFRESH_TOKEN_COOKIE)?.value || null
  }

  static async getTokensFromRequest(
    request: NextRequest
  ): Promise<{ accessToken: string | null; refreshToken: string | null }> {
    return {
      accessToken: request.cookies.get(ACCESS_TOKEN_COOKIE)?.value || null,
      refreshToken: request.cookies.get(REFRESH_TOKEN_COOKIE)?.value || null,
    }
  }

  private static async clearTokens(): Promise<void> {
    const cookieStore = await cookies()
    cookieStore.delete(ACCESS_TOKEN_COOKIE)
    cookieStore.delete(REFRESH_TOKEN_COOKIE)
  }

  // Helper: get user profile using a provided access token
  static async getUserFromToken(token: string): Promise<User | null> {
    try {
      return await this.apiRequest<User>(
        "/v1/users/profile",
        {
          method: "GET",
        },
        token
      )
    } catch {
      return null
    }
  }

  // For compatibility with existing call sites
  static async getSessionFromCookies(): Promise<Session | null> {
    const accessToken = await this.getAccessToken()
    const refreshToken = await this.getRefreshToken()

    if (!accessToken) {
      return this.createGuestSession()
    }

    const user = await this.getUserFromToken(accessToken)
    return {
      id: crypto.randomUUID(),
      userId: user?.id ?? null,
      accessToken,
      refreshToken: refreshToken || "",
      expiresAt: new Date(Date.now() + 15 * 60 * 1000),
      isGuest: !user,
      createdAt: new Date(),
    }
  }

  static async getSessionFromRequest(
    request: NextRequest
  ): Promise<Session | null> {
    const { accessToken, refreshToken } = await this.getTokensFromRequest(request)
    if (!accessToken) {
      return this.createGuestSession()
    }
    const user = await this.getUserFromToken(accessToken)
    return {
      id: crypto.randomUUID(),
      userId: user?.id ?? null,
      accessToken,
      refreshToken: refreshToken || "",
      expiresAt: new Date(Date.now() + 15 * 60 * 1000),
      isGuest: !user,
      createdAt: new Date(),
    }
  }

  static async getUserById(_userId: string): Promise<User | null> {
    // Backend provides profile endpoint; return current user
    return this.getCurrentUser()
  }

  static async getCurrentUser(): Promise<User | null> {
    const accessToken = await this.getAccessToken()
    if (!accessToken) return null

    try {
      return await this.apiRequest<User>(
        "/v1/users/profile",
        {
          method: "GET",
        },
        accessToken
      )
    } catch {
      return null
    }
  }

  static async createGuestSession(): Promise<Session> {
    const sessionId = crypto.randomUUID()
    const expiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000) // 7 days

    return {
      id: sessionId,
      userId: null,
      accessToken: "",
      refreshToken: "",
      expiresAt,
      isGuest: true,
      guestId: `guest_${sessionId}`,
      createdAt: new Date(),
    }
  }

  static async mergeGuestCart(guestId: string, userId: string): Promise<void> {
    // TODO: Implement cart merging logic with backend API
    console.log(`Merging cart from guest ${guestId} to user ${userId}`)
  }
}

// Cookie configuration optimized for WebView
export const COOKIE_OPTIONS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "lax" as const, // Best for WebView compatibility
  path: "/",
}
