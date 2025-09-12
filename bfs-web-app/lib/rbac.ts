import type { NextRequest } from "next/server"
import { AuthService } from "./auth"
import { Role } from "../types/auth"
import type { Session, User } from "../types/auth"

export enum Permission {
  // Menu permissions
  MENU_READ = "menu:read",
  MENU_WRITE = "menu:write",
  MENU_DELETE = "menu:delete",

  // Order permissions
  ORDER_READ = "order:read",
  ORDER_WRITE = "order:write",
  ORDER_UPDATE_STATUS = "order:update_status",
  ORDER_CANCEL = "order:cancel",

  // User permissions
  USER_READ = "user:read",
  USER_WRITE = "user:write",
  USER_DELETE = "user:delete",

  // Admin permissions
  ADMIN_ANALYTICS = "admin:analytics",
  ADMIN_SETTINGS = "admin:settings",
  ADMIN_AUDIT_LOGS = "admin:audit_logs",

  // POS permissions
  POS_ACCESS = "pos:access",
  POS_PROCESS_ORDERS = "pos:process_orders",
  POS_PRINT_RECEIPTS = "pos:print_receipts",

  // Cart permissions
  CART_READ = "cart:read",
  CART_WRITE = "cart:write",
}

const ROLE_PERMISSIONS: Record<Role, Permission[]> = {
  [Role.ADMIN]: [
    // All permissions for admin
    Permission.MENU_READ,
    Permission.MENU_WRITE,
    Permission.MENU_DELETE,
    Permission.ORDER_READ,
    Permission.ORDER_WRITE,
    Permission.ORDER_UPDATE_STATUS,
    Permission.ORDER_CANCEL,
    Permission.USER_READ,
    Permission.USER_WRITE,
    Permission.USER_DELETE,
    Permission.ADMIN_ANALYTICS,
    Permission.ADMIN_SETTINGS,
    Permission.ADMIN_AUDIT_LOGS,
    Permission.POS_ACCESS,
    Permission.POS_PROCESS_ORDERS,
    Permission.POS_PRINT_RECEIPTS,
    Permission.CART_READ,
    Permission.CART_WRITE,
  ],
  [Role.CUSTOMER]: [
    // Limited permissions for customers
    Permission.MENU_READ,
    Permission.ORDER_READ, // Only their own orders
    Permission.ORDER_WRITE, // Create orders
    Permission.CART_READ,
    Permission.CART_WRITE,
  ],
}

export class RBACService {
  static hasPermission(role: Role, permission: Permission): boolean {
    const rolePermissions = ROLE_PERMISSIONS[role] || []
    return rolePermissions.includes(permission)
  }

  static hasAnyPermission(role: Role, permissions: Permission[]): boolean {
    return permissions.some((permission) => this.hasPermission(role, permission))
  }

  static hasAllPermissions(role: Role, permissions: Permission[]): boolean {
    return permissions.every((permission) => this.hasPermission(role, permission))
  }

  static requirePermission(role: Role, permission: Permission): void {
    if (!this.hasPermission(role, permission)) {
      throw new Error(`Access denied. Required permission: ${permission}`)
    }
  }

  static requireRole(requiredRole: Role, userRole: Role): void {
    // Simple role hierarchy: ADMIN > CUSTOMER
    const roleHierarchy = {
      [Role.ADMIN]: 2,
      [Role.CUSTOMER]: 1,
    }

    if (roleHierarchy[userRole] < roleHierarchy[requiredRole]) {
      throw new Error(`Access denied. Required role: ${requiredRole}`)
    }
  }

  static canAccessRoute(role: Role, pathname: string): boolean {
    // Public routes - accessible to all
    const publicRoutes = ["/", "/menu", "/cart", "/checkout", "/login"]
    if (publicRoutes.some((route) => pathname === route || pathname.startsWith(route))) {
      return true
    }

    // Admin routes - admin only
    if (pathname.startsWith("/admin")) {
      return this.hasPermission(role, Permission.ADMIN_ANALYTICS) // Basic admin permission check
    }

    // POS routes - admin only (since POS is admin functionality)
    if (pathname.startsWith("/pos")) {
      return this.hasPermission(role, Permission.POS_ACCESS)
    }

    return false
  }

  static getRoutePermissions(pathname: string): Permission[] {
    if (pathname.startsWith("/admin/menu")) {
      return [Permission.MENU_READ, Permission.MENU_WRITE]
    }

    if (pathname.startsWith("/admin/orders")) {
      return [Permission.ORDER_READ, Permission.ORDER_UPDATE_STATUS]
    }

    if (pathname.startsWith("/admin/users")) {
      return [Permission.USER_READ, Permission.USER_WRITE]
    }

    if (pathname.startsWith("/admin/analytics")) {
      return [Permission.ADMIN_ANALYTICS]
    }

    if (pathname.startsWith("/admin/settings")) {
      return [Permission.ADMIN_SETTINGS]
    }

    if (pathname.startsWith("/pos")) {
      return [Permission.POS_ACCESS, Permission.POS_PROCESS_ORDERS]
    }

    if (pathname.startsWith("/api/admin")) {
      return [Permission.ADMIN_ANALYTICS] // Basic admin API access
    }

    if (pathname.startsWith("/api/pos")) {
      return [Permission.POS_ACCESS]
    }

    return []
  }
}

// Higher-order function for API route protection
export function withRoleProtection(requiredRole: Role) {
  return function (
    handler: (request: Request, context: { user: User; session: Session }) => Promise<Response> | Response
  ) {
    return async function (request: Request) {
      try {
        const session = await AuthService.getSessionFromRequest(request as unknown as NextRequest)

        if (!session) {
          return new Response(JSON.stringify({ error: "Authentication required" }), {
            status: 401,
            headers: { "Content-Type": "application/json" },
          })
        }

        if (session.isGuest) {
          return new Response(JSON.stringify({ error: "Access denied for guest users" }), {
            status: 403,
            headers: { "Content-Type": "application/json" },
          })
        }

        const user = session.userId ? await AuthService.getUserById(session.userId) : null
        if (!user) {
          return new Response(JSON.stringify({ error: "User not found" }), {
            status: 404,
            headers: { "Content-Type": "application/json" },
          })
        }

        RBACService.requireRole(requiredRole, user.role)

        return handler(request, { user, session })
      } catch (error) {
        const message = error instanceof Error ? error.message : "Access denied"
        return new Response(JSON.stringify({ error: message }), {
          status: 403,
          headers: { "Content-Type": "application/json" },
        })
      }
    }
  }
}

// Higher-order function for permission-based protection
export function withPermissionProtection(requiredPermissions: Permission[]) {
  return function (
    handler: (request: Request, context: { user: User; session: Session }) => Promise<Response> | Response
  ) {
    return async function (request: Request) {
      try {
        const session = await AuthService.getSessionFromRequest(request as unknown as NextRequest)

        if (!session) {
          return new Response(JSON.stringify({ error: "Authentication required" }), {
            status: 401,
            headers: { "Content-Type": "application/json" },
          })
        }

        if (session.isGuest) {
          return new Response(JSON.stringify({ error: "Access denied for guest users" }), {
            status: 403,
            headers: { "Content-Type": "application/json" },
          })
        }

        const user = session.userId ? await AuthService.getUserById(session.userId) : null
        if (!user) {
          return new Response(JSON.stringify({ error: "User not found" }), {
            status: 404,
            headers: { "Content-Type": "application/json" },
          })
        }

        for (const permission of requiredPermissions) {
          RBACService.requirePermission(user.role, permission)
        }

        return handler(request, { user, session })
      } catch (error) {
        const message = error instanceof Error ? error.message : "Access denied"
        return new Response(JSON.stringify({ error: message }), {
          status: 403,
          headers: { "Content-Type": "application/json" },
        })
      }
    }
  }
}
