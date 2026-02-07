"use client"
import { useRouter } from "next/navigation"
import { useEffect } from "react"
import { useAuth } from "@/contexts/auth-context"
import { canAccessAdmin } from "@/lib/auth/rbac"
import type { UserRole } from "@/types"

export function AdminGuard({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth()
  const router = useRouter()

  const hasAccess = user ? canAccessAdmin(user.role as UserRole) : false

  useEffect(() => {
    if (!isLoading && !hasAccess) {
      router.replace("/")
    }
  }, [isLoading, hasAccess, router])

  if (isLoading) {
    return (
      <div className="items-center justify-center text-center text-sm font-semibold text-gray-600">
        Zugriff wird geprüft...
      </div>
    )
  }

  if (!hasAccess) return null

  return <>{children}</>
}
