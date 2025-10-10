// Public base used by the browser. Default to localhost:8080 to avoid
// accidentally calling the web origin (localhost:3000) when unset.
const PUBLIC_API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"
// Internal base used by SSR/RSC inside the container to reach backend service name.
const INTERNAL_API_BASE = process.env.BACKEND_INTERNAL_URL || process.env.INTERNAL_API_BASE_URL

function resolveApiBase(): string {
  // On the server (SSR/RSC), prefer the internal URL on the Docker network
  if (typeof window === "undefined") {
    return INTERNAL_API_BASE || PUBLIC_API_BASE || "http://backend:8080"
  }
  // In the browser, use the public/base URL reachable from the user’s machine
  return PUBLIC_API_BASE
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
  const url = `${API_BASE_URL}${endpoint}`

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
