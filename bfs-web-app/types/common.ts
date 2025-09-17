export interface BaseEntity {
  id: string
  createdAt: string
  updatedAt: string
}

export interface ActivatableEntity extends BaseEntity {
  isActive: boolean
}

export interface NamedEntity extends BaseEntity {
  name: string
}

export interface ActivatableNamedEntity extends NamedEntity {
  isActive: boolean
}

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

export interface ListResponse<T> {
  items: T[]
  total: number
  limit?: number
  offset?: number
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

export interface MessageResponse {
  message: string
}

export interface StatusUpdateResponse extends MessageResponse {
  id: string
  isActive: boolean
}

export interface WebSocketMessage<T = unknown> {
  type: string
  payload: T
  timestamp: string
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