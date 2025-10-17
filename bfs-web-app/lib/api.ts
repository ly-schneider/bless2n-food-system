// Public base used by the browser. Always use relative "/api" so that
// container runtime env can decide the backend target via API routes.
const PUBLIC_API_BASE = "/api"
// Internal base used by SSR/RSC inside the container to reach backend service name.
const INTERNAL_API_BASE = process.env.BACKEND_INTERNAL_URL || process.env.INTERNAL_API_BASE_URL

function resolveApiBase(): string {
  // On the server (SSR/RSC), prefer the internal URL on the Docker network
  if (typeof window === "undefined") {
    return INTERNAL_API_BASE || PUBLIC_API_BASE || "http://backend:8080"
  }
  // In the browser, use the public/base URL reachable from the user's machine
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

  const response = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  })

  if (!response.ok) {
    const errorData = (await response.json().catch(() => ({}))) as { message?: string; detail?: string }
    const error = createApiError(
      response.status,
      errorData.message || errorData.detail || `HTTP ${response.status}: ${response.statusText}`
    )
    throw error
  }

  return response.json() as Promise<T>
}

export {}
