"use client"
import { useEffect, useState } from "react"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuth } from "@/contexts/auth-context"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { ALL_ROLES, hasPermission } from "@/lib/auth/rbac"
import { getCSRFToken } from "@/lib/csrf"
import type { UserRole } from "@/types"

type AdminUser = {
  id: string
  name: string
  email: string
  emailVerified: boolean
  role: string
  createdAt: string
  updatedAt: string
}

export default function AdminUsersPage() {
  const fetchAuth = useAuthorizedFetch()
  const { user: currentUser } = useAuth()
  const [items, setItems] = useState<AdminUser[]>([])
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState<string | null>(null)

  const canSetRole = hasPermission(currentUser?.role as UserRole, "users:role:set")

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetchAuth(`/api/v1/users`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: AdminUser[]; count: number }
        if (!cancelled) setItems(data.items || [])
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Benutzer konnten nicht geladen werden"
        if (!cancelled) setError(msg)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  async function handleRoleChange(userId: string, newRole: string) {
    if (userId === currentUser?.id) return
    setSaving(userId)
    setError(null)
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/users/${encodeURIComponent(userId)}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({ role: newRole }),
      })
      if (!res.ok) {
        const data = (await res.json().catch(() => ({}))) as { detail?: string }
        throw new Error(data.detail || `HTTP ${res.status}`)
      }
      const updated = (await res.json()) as AdminUser
      setItems((prev) => prev.map((u) => (u.id === userId ? updated : u)))
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Rolle konnte nicht aktualisiert werden"
      setError(msg)
    } finally {
      setSaving(null)
    }
  }

  return (
    <div className="min-w-0 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Benutzer</h1>
      </div>

      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="rounded-md border">
        <div className="relative">
          <div className="from-background pointer-events-none absolute inset-y-0 left-0 w-6 bg-gradient-to-r to-transparent" />
          <div className="from-background pointer-events-none absolute inset-y-0 right-0 w-6 bg-gradient-to-l to-transparent" />
          <div
            className="max-w-full overflow-x-auto overscroll-x-contain rounded-[10px]"
            role="region"
            aria-label="Users table - scroll horizontally to reveal more columns"
            tabIndex={0}
          >
            <Table className="whitespace-nowrap">
              <TableHeader className="bg-card sticky top-0">
                <TableRow>
                  <TableHead className="whitespace-nowrap">Name</TableHead>
                  <TableHead className="whitespace-nowrap">E-Mail</TableHead>
                  <TableHead className="whitespace-nowrap">Rolle</TableHead>
                  <TableHead className="whitespace-nowrap">Verifiziert</TableHead>
                  <TableHead className="whitespace-nowrap">Erstellt</TableHead>
                  <TableHead className="whitespace-nowrap">Aktualisiert</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((u) => {
                  const created = u.createdAt ? new Date(u.createdAt).toLocaleString("de-CH") : "-"
                  const updated = u.updatedAt ? new Date(u.updatedAt).toLocaleString("de-CH") : "-"
                  const isSelf = u.id === currentUser?.id
                  return (
                    <TableRow key={u.id} className="even:bg-card odd:bg-muted/40">
                      <TableCell>{u.name || "-"}</TableCell>
                      <TableCell>{u.email}</TableCell>
                      <TableCell>
                        {canSetRole && !isSelf ? (
                          <Select
                            value={u.role}
                            onValueChange={(v) => void handleRoleChange(u.id, v)}
                            disabled={saving === u.id}
                          >
                            <SelectTrigger className="h-8 w-[160px]">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {ALL_ROLES.map((r) => (
                                <SelectItem key={r.value} value={r.value}>
                                  {r.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        ) : (
                          <span className="text-sm uppercase">
                            {ALL_ROLES.find((r) => r.value === u.role)?.label || u.role}
                            {isSelf && <span className="text-muted-foreground ml-1 text-xs normal-case">(du)</span>}
                          </span>
                        )}
                      </TableCell>
                      <TableCell>{u.emailVerified ? "Ja" : "Nein"}</TableCell>
                      <TableCell className="whitespace-nowrap">{created}</TableCell>
                      <TableCell className="whitespace-nowrap">{updated}</TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>
    </div>
  )
}
