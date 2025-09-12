export interface ApiResponse<T = unknown> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
  totalPages: number
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, unknown>
}

export interface ValidationError {
  field: string
  message: string
}

export interface ApiValidationError extends ApiError {
  code: "VALIDATION_ERROR"
  details: {
    errors: ValidationError[]
  }
}

export interface ApiAuthError extends ApiError {
  code: "UNAUTHORIZED" | "FORBIDDEN" | "TOKEN_EXPIRED"
}

export interface WebSocketMessage<T = unknown> {
  type: string
  payload: T
  timestamp: string
}

export interface OrderUpdateMessage {
  orderId: string
  status: string
  estimatedTime?: number
}

export interface CartSyncMessage {
  cartId: string
  items: unknown[]
  total: number
}

export interface AuditLogEntry {
  id: string
  userId: string
  action: string
  resource: string
  resourceId: string
  details: Record<string, unknown>
  ipAddress: string
  userAgent: string
  timestamp: Date
}
