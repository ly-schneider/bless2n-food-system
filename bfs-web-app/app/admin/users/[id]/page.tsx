"use client"
import Link from "next/link"
import { useParams, useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAuth } from "@/contexts/auth-context"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { ALL_ROLES, hasPermission } from "@/lib/auth/rbac"
import { getCSRFToken } from "@/lib/csrf"
import type { UserRole } from "@/types"

type UserDetails = {
  id: string
  email: string
  name?: string | null
  role: string
  emailVerified?: boolean
  createdAt?: string
  updatedAt?: string
}

export default function AdminUserDetailPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const fetchAuth = useAuthorizedFetch()
  const { user: currentUser } = useAuth()
  const [user, setUser] = useState<UserDetails | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [saving, setSaving] = useState<boolean>(false)
  const [editOpen, setEditOpen] = useState(false)

  // Local editable fields
  const [role, setRole] = useState("customer")

  const canSetRole = hasPermission(currentUser?.role as UserRole, "users:role:set")
  const isSelf = currentUser?.id === id

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoading(true)
      setError(null)
      try {
        const res = await fetchAuth(`/api/v1/users/${encodeURIComponent(id)}`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as UserDetails
        if (!cancelled) {
          setUser(data)
          setRole(data.role || "customer")
        }
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Fehler beim Laden des Benutzers"
        if (!cancelled) setError(msg)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [fetchAuth, id])

  const created = user?.createdAt ? new Date(user.createdAt).toLocaleString("de-CH") : "-"
  const updated = user?.updatedAt ? new Date(user.updatedAt).toLocaleString("de-CH") : "-"

  async function saveRole() {
    try {
      setSaving(true)
      setError(null)
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/users/${encodeURIComponent(id)}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({ role }),
      })
      if (!res.ok) {
        const data = (await res.json().catch(() => ({}))) as { detail?: string }
        throw new Error(data.detail || `HTTP ${res.status}`)
      }
      // Update the local user state with the new role
      if (user) {
        setUser({ ...user, role })
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Aktualisierung fehlgeschlagen"
      setError(msg)
    } finally {
      setSaving(false)
    }
  }

  async function deleteUser() {
    if (!confirm("Benutzer wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.")) return
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/users/${encodeURIComponent(id)}`, {
        method: "DELETE",
        headers: { "X-CSRF": csrf || "" },
      })
      if (!res.ok && res.status !== 204) throw new Error(`HTTP ${res.status}`)
      router.push("/admin/users")
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Löschen fehlgeschlagen"
      setError(msg)
    }
  }

  return (
    <div className="min-w-0 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Benutzer</h1>
        <div className="flex items-center gap-2">
          {canSetRole && !isSelf && (
            <Button variant="outline" onClick={() => setEditOpen(true)} disabled={!user}>
              Rolle ändern
            </Button>
          )}
          <Link href="/admin/users">
            <Button variant="outline">Zurück zur Übersicht</Button>
          </Link>
        </div>
      </div>

      {loading && <div className="text-muted-foreground text-sm">Lade Benutzer...</div>}
      {error && <div className="text-sm text-red-600">{error}</div>}

      {user && (
        <div className="grid gap-4 md:grid-cols-2">
          <div className="rounded-md border p-4">
            <h2 className="mb-3 text-base font-semibold">Details</h2>
            <div className="space-y-1 text-sm">
              <div>
                <span className="text-muted-foreground">ID:</span> <span className="text-xs">{user.id}</span>
              </div>
              <div>
                <span className="text-muted-foreground">E-Mail:</span> {user.email}
              </div>
              <div>
                <span className="text-muted-foreground">Name:</span>{" "}
                {user.name || "-"}
              </div>
              <div>
                <span className="text-muted-foreground">Rolle:</span>{" "}
                <span className="uppercase">
                  {ALL_ROLES.find((r) => r.value === user.role)?.label || user.role}
                </span>
              </div>
              <div>
                <span className="text-muted-foreground">Verifiziert:</span> {user.emailVerified ? "Ja" : "Nein"}
              </div>
              <div>
                <span className="text-muted-foreground">Erstellt:</span> {created}
              </div>
              <div>
                <span className="text-muted-foreground">Aktualisiert:</span> {updated}
              </div>
            </div>
          </div>

          <div className="rounded-md border p-4 md:col-span-1">
            <h2 className="mb-3 text-base font-semibold">Gefährlicher Bereich</h2>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Benutzer löschen</p>
                <p className="text-muted-foreground text-xs">Entfernt den Benutzer dauerhaft.</p>
              </div>
              <Button variant="destructive" onClick={() => void deleteUser()}>
                Benutzer löschen
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Role edit dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Rolle ändern</DialogTitle>
            <DialogDescription>Weise diesem Benutzer eine neue Rolle zu.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1">
              <label className="text-muted-foreground text-sm">Rolle</label>
              <Select value={role} onValueChange={setRole}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Rolle" />
                </SelectTrigger>
                <SelectContent>
                  {ALL_ROLES.map((r) => (
                    <SelectItem key={r.value} value={r.value}>
                      {r.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <div className="flex w-full justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  if (user) setRole(user.role || "customer")
                }}
              >
                Zurücksetzen
              </Button>
              <Button
                onClick={async () => {
                  await saveRole()
                  setEditOpen(false)
                }}
                disabled={saving}
              >
                Speichern
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
