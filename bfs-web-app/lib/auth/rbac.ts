// RBAC definitions - mirrors bfs-backend/internal/auth/rbac.go
// This is used for UI gating only. Go is the authoritative enforcement point.

export type UserRole = "customer" | "admin"

export const ROLES = {
  CUSTOMER: "customer" as UserRole,
  ADMIN: "admin" as UserRole,
} as const

export type Permission =
  | "orders:create"
  | "orders:read:own"
  | "pos:access"
  | "pos:orders:create"
  | "pos:payments"
  | "station:access"
  | "station:verify"
  | "station:redeem"
  | "admin:access"
  | "products:write"
  | "orders:read:all"
  | "orders:export"
  | "orders:status:write"
  | "categories:write"
  | "menus:write"
  | "stations:manage"
  | "pos:manage"
  | "users:read"
  | "users:role:set"
  | "devices:approve"
  | "devices:revoke"
  | "invites:manage"

const ROLE_PERMISSIONS: Record<UserRole, Set<Permission>> = {
  customer: new Set<Permission>(["orders:create", "orders:read:own"]),
  admin: new Set<Permission>([
    "orders:create",
    "orders:read:own",
    "pos:access",
    "pos:orders:create",
    "pos:payments",
    "station:access",
    "station:verify",
    "station:redeem",
    "admin:access",
    "products:write",
    "orders:read:all",
    "orders:export",
    "orders:status:write",
    "categories:write",
    "menus:write",
    "stations:manage",
    "pos:manage",
    "users:read",
    "users:role:set",
    "devices:approve",
    "devices:revoke",
    "invites:manage",
  ]),
}

/** Check if a role has a specific permission (for UI gating). */
export function hasPermission(role: UserRole | undefined | null, perm: Permission): boolean {
  if (!role) return false
  return ROLE_PERMISSIONS[role]?.has(perm) ?? false
}

/** Get all permissions for a role. */
export function getPermissions(role: UserRole): Permission[] {
  return Array.from(ROLE_PERMISSIONS[role] ?? [])
}

/** Check if a role can access admin features. */
export function canAccessAdmin(role: UserRole | undefined | null): boolean {
  return hasPermission(role, "admin:access")
}

/** All valid roles for dropdowns/selects. */
export const ALL_ROLES: { value: UserRole; label: string }[] = [
  { value: "customer", label: "Customer" },
  { value: "admin", label: "Administrator" },
]
