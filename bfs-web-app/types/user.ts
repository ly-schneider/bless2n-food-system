export type UserRole = "admin" | "customer"

export interface User {
  id: string
  email: string
  firstName?: string // Only for admins
  lastName?: string // Only for admins
  role: UserRole
  isVerified: boolean
  isDisabled: boolean
  disabledReason: string | null
  createdAt: string // ISO date
  updatedAt: string // ISO date
}
