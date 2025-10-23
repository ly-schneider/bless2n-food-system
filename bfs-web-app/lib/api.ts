const PUBLIC_API_BASE = "/api"
const INTERNAL_API_BASE = process.env.BACKEND_INTERNAL_URL || process.env.INTERNAL_API_BASE_URL

function resolveApiBase(): string {
  if (typeof window === "undefined") {
    return INTERNAL_API_BASE || "http://backend:8080"
  }
  return PUBLIC_API_BASE || "/api"
}

export const API_BASE_URL = resolveApiBase()

export interface ApiError {
  status: number
  message: string
}

export function createApiError(status: number, message: string): ApiError {
  return { status, message }
}

export async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const baseUrl = API_BASE_URL
  if (!baseUrl) {
    throw new Error("API_BASE_URL is not configured")
  }

  const url = `${baseUrl}${endpoint}`
  const isBrowser = typeof window !== "undefined"
  const method = (options.method || "GET").toUpperCase()
  const mutating = !(method === "GET" || method === "HEAD" || method === "OPTIONS")
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string> | undefined),
  }

  if (isBrowser && mutating && String(baseUrl).startsWith("/api") && !("X-CSRF" in headers) && !("x-csrf" in headers)) {
    try {
      const { getCSRFToken } = await import("./csrf")
      const csrf = getCSRFToken()
      if (csrf) headers["X-CSRF"] = csrf
    } catch {}
  }

  const response = await fetch(url, { ...options, headers })

  if (!response.ok) {
    const errorData = (await response.json().catch(() => ({}))) as { message?: string; detail?: string }
    try {
      console.error("[api] request failed", {
        method,
        url,
        status: response.status,
        statusText: response.statusText,
        message: errorData.message || errorData.detail,
      })
    } catch {}
    const error = createApiError(
      response.status,
      errorData.message || errorData.detail || `HTTP ${response.status}: ${response.statusText}`
    )
    throw error
  }

  return response.json() as Promise<T>
}

export {}
