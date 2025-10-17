"use client"
import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { API_BASE_URL } from "@/lib/api"

export function AdminGuard({ children }: { children: React.ReactNode }) {
  const fetchAuth = useAuthorizedFetch()
  const router = useRouter()
  const [ok, setOk] = useState<boolean | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetchAuth(`${API_BASE_URL}/v1/users/me`, { method: "GET" })
        if (!res.ok) throw new Error("Unauthorized")
        const data = (await res.json()) as { user?: { role?: string } }
        if (!cancelled) setOk(data?.user?.role === "admin")
      } catch {
        if (!cancelled) setOk(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  useEffect(() => {
    if (ok === false) router.replace("/")
  }, [ok, router])

  if (ok === null)
    return (
      <div className="items-center justify-center text-center text-sm font-semibold text-gray-600">
        Überprüfe den Zugriff...
      </div>
    )
  if (!ok) return null
  return <>{children}</>
}
