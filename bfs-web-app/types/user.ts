export type UserRole = "customer" | "admin"

export interface User {
  id: string
  email: string
  firstName?: string
  lastName?: string
  role: UserRole
  isVerified: boolean
  isDisabled: boolean
  disabledReason: string | null
  isClub100: boolean
  createdAt: string // ISO date
  updatedAt: string // ISO date
}
