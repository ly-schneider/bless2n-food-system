import { createAccessControl } from "better-auth/plugins/access"

/**
 * Better Auth access control definitions.
 *
 * These satisfy the admin plugin's requirement that adminRoles map to
 * defined roles. Actual permission enforcement happens in the Go backend
 * via the RBAC module (auth/rbac.go). This file is intentionally minimal.
 */
const statement = {
  admin: ["access"],
} as const

export const ac = createAccessControl(statement)

// Every role must be declared so the admin plugin accepts it.
export const customerRole = ac.newRole({ admin: [] })
export const adminRole = ac.newRole({ admin: ["access"] })
