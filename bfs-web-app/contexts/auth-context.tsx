"use client"

import React, { createContext, useCallback, useContext, useMemo } from "react"
import { authClient, signOut as authSignOut } from "@/lib/auth/client"
import { clearOrders } from "@/lib/orders-storage"
import type { User, UserRole } from "@/types"

type AuthContextType = {
  accessToken: string | null
  user: User | null
  isLoading: boolean
  signOut: () => Promise<void>
  getToken: () => string | null
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

/** Valid user roles for the system */
const VALID_ROLES: UserRole[] = ["customer", "admin"]

/**
 * Maps Better Auth session user to our User type.
 * Better Auth provides: id, email, name, image, emailVerified, role, etc.
 */
function mapBetterAuthUserToUser(betterAuthUser: Record<string, unknown> | null | undefined): User | null {
  if (!betterAuthUser) return null

  const id = String(betterAuthUser.id || "")
  const email = String(betterAuthUser.email || "")

  if (!id || !email) return null

  // Parse name into firstName/lastName if available
  const name = typeof betterAuthUser.name === "string" ? betterAuthUser.name : ""
  const nameParts = name.split(" ")
  const firstName = nameParts[0] || undefined
  const lastName = nameParts.slice(1).join(" ") || undefined

  // emailVerified can be boolean or Date depending on Better Auth config
  const emailVerified = betterAuthUser.emailVerified
  const isVerified = emailVerified === true || emailVerified instanceof Date

  // createdAt/updatedAt can be Date objects or strings
  const createdAt =
    betterAuthUser.createdAt instanceof Date
      ? betterAuthUser.createdAt.toISOString()
      : typeof betterAuthUser.createdAt === "string"
        ? betterAuthUser.createdAt
        : new Date().toISOString()

  const updatedAt =
    betterAuthUser.updatedAt instanceof Date
      ? betterAuthUser.updatedAt.toISOString()
      : typeof betterAuthUser.updatedAt === "string"
        ? betterAuthUser.updatedAt
        : new Date().toISOString()

  // Parse role, defaulting to "customer" if not a valid role
  const rawRole = typeof betterAuthUser.role === "string" ? betterAuthUser.role : ""
  const role: UserRole = VALID_ROLES.includes(rawRole as UserRole) ? (rawRole as UserRole) : "customer"

  return {
    id,
    email,
    firstName,
    lastName,
    role,
    isVerified,
    isDisabled: betterAuthUser.banned === true,
    disabledReason: typeof betterAuthUser.banReason === "string" ? betterAuthUser.banReason : null,
    isClub100: betterAuthUser.isClub100 === true,
    createdAt,
    updatedAt,
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: session, isPending, error } = authClient.useSession()

  // Better Auth manages tokens internally via cookies
  // The session token is available for API calls that need explicit auth
  const accessToken = session?.session?.token || null
  const user = mapBetterAuthUserToUser(session?.user || null)

  const signOut = useCallback(async () => {
    clearOrders()
    await authSignOut()
  }, [])

  // getToken provides current token for useAuthorizedFetch compatibility
  const getToken = useCallback(() => {
    return accessToken
  }, [accessToken])

  const value = useMemo(
    () => ({
      accessToken,
      user,
      isLoading: isPending,
      signOut,
      getToken,
    }),
    [accessToken, user, isPending, signOut, getToken]
  )

  // Log error in development for debugging
  if (error && process.env.NODE_ENV === "development") {
    console.error("[AuthProvider] Session error:", error)
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}
