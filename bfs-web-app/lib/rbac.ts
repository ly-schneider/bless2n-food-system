import type { NextRequest } from "next/server"

import { Role } from "@/types"
import type { Session, User } from "@/types"

import { AuthService } from "./auth"

export class RBACService {
  static isAdmin(role: Role): boolean {
    return role === Role.ADMIN
  }

  static isStation(role: Role): boolean {
    return role === Role.STATION
  }

  static isCustomer(role: Role): boolean {
    return role === Role.CUSTOMER
  }

  static canAccessAdmin(role: Role): boolean {
    return this.isAdmin(role)
  }

  static canAccessStation(role: Role): boolean {
    return this.isAdmin(role) || this.isStation(role)
  }

  static canManageOrders(role: Role): boolean {
    return this.isAdmin(role) || this.isStation(role)
  }

  static canManageMenu(role: Role): boolean {
    return this.isAdmin(role)
  }

  static canManageUsers(role: Role): boolean {
    return this.isAdmin(role)
  }

  static canViewAnalytics(role: Role): boolean {
    return this.isAdmin(role)
  }

  static canPlaceOrders(_role: Role): boolean {
    return true // All roles can place orders
  }

  static canViewMenu(_role: Role): boolean {
    return true // All roles can view menu
  }

  static requireRole(requiredRole: Role, userRole: Role): void {
    // Simple role hierarchy: ADMIN > STATION > CUSTOMER
    const roleHierarchy = {
      [Role.ADMIN]: 3,
      [Role.STATION]: 2,
      [Role.CUSTOMER]: 1,
    }

    if (roleHierarchy[userRole] < roleHierarchy[requiredRole]) {
      throw new Error(`Access denied. Required role: ${requiredRole}`)
    }
  }

  static canAccessRoute(role: Role, pathname: string): boolean {
    // Public routes - accessible to all authenticated users
    const publicRoutes = ["/", "/menu", "/cart", "/checkout", "/orders"]
    if (publicRoutes.some((route) => pathname === route || pathname.startsWith(route))) {
      return true
    }

    // Admin routes - admin only
    if (pathname.startsWith("/admin")) {
      return this.canAccessAdmin(role)
    }

    // Station routes - admin and station only
    if (pathname.startsWith("/station")) {
      return this.canAccessStation(role)
    }

    return false
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

// Higher-order function for admin-only protection
export function withAdminProtection() {
  return withRoleProtection(Role.ADMIN)
}

// Higher-order function for station+ protection (admin or station)
export function withStationProtection() {
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

        if (!RBACService.canAccessStation(user.role)) {
          throw new Error("Access denied. Admin or station role required")
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
