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
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { API_BASE_URL } from "@/lib/api"

type UserDetails = {
  id: string
  email: string
  firstName?: string | null
  lastName?: string | null
  role: string
  isVerified?: boolean
  createdAt?: string
  updatedAt?: string
}

export default function AdminUserDetailPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const fetchAuth = useAuthorizedFetch()
  const [user, setUser] = useState<UserDetails | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [saving, setSaving] = useState<boolean>(false)
  const [editOpen, setEditOpen] = useState(false)

  // Local editable fields
  const [email, setEmail] = useState("")
  const [firstName, setFirstName] = useState("")
  const [lastName, setLastName] = useState("")
  const [role, setRole] = useState("customer")
  const [isVerified, setIsVerified] = useState<boolean | undefined>(undefined)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoading(true)
      setError(null)
      try {
        const res = await fetchAuth(`${API_BASE_URL}/v1/admin/users/${encodeURIComponent(id)}`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { user: UserDetails }
        if (!cancelled) {
          setUser(data.user)
          setEmail(data.user.email || "")
          setFirstName(data.user.firstName || "")
          setLastName(data.user.lastName || "")
          setRole(data.user.role || "customer")
          setIsVerified(typeof data.user.isVerified === "boolean" ? data.user.isVerified : undefined)
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

  const created = user?.createdAt ? new Date(user.createdAt).toLocaleString("de-CH") : "–"
  const updated = user?.updatedAt ? new Date(user.updatedAt).toLocaleString("de-CH") : "–"

  async function save() {
    try {
      setSaving(true)
      setError(null)
      const csrf = getCSRFCookie()
      type UpdateUserPayload = {
        email: string
        firstName?: string
        lastName?: string
        role: string
        isVerified?: boolean
      }
      const body: UpdateUserPayload = {
        email,
        firstName: firstName || undefined,
        lastName: lastName || undefined,
        role,
        isVerified: typeof isVerified === "boolean" ? isVerified : undefined,
      }
      const res = await fetchAuth(`${API_BASE_URL}/v1/admin/users/${encodeURIComponent(id)}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify(body),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = (await res.json()) as { user: UserDetails }
      setUser(data.user)
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
      const csrf = getCSRFCookie()
      const res = await fetchAuth(`${API_BASE_URL}/v1/admin/users/${encodeURIComponent(id)}`, {
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
          <Button variant="outline" onClick={() => setEditOpen(true)} disabled={!user}>
            Benutzer bearbeiten
          </Button>
          <Link href="/admin/users">
            <Button variant="outline">Zurück zur Übersicht</Button>
          </Link>
        </div>
      </div>

      {loading && <div className="text-muted-foreground text-sm">Lade Benutzer…</div>}
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
                <span className="text-muted-foreground">E‑Mail:</span> {user.email}
              </div>
              <div>
                <span className="text-muted-foreground">Name:</span>{" "}
                {(user.firstName || "–") + " " + (user.lastName || "")}
              </div>
              <div>
                <span className="text-muted-foreground">Rolle:</span> <span className="uppercase">{user.role}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Verifiziert:</span> {user.isVerified ? "Ja" : "Nein"}
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

      {/* Edit dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Benutzer bearbeiten</DialogTitle>
            <DialogDescription>Aktualisiere die Profildaten dieses Benutzers.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1">
              <label className="text-muted-foreground text-sm">E‑Mail</label>
              <Input value={email} onChange={(e) => setEmail(e.target.value)} />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <label className="text-muted-foreground text-sm">Vorname</label>
                <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
              </div>
              <div className="space-y-1">
                <label className="text-muted-foreground text-sm">Nachname</label>
                <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <label className="text-muted-foreground text-sm">Rolle</label>
                <Select value={role} onValueChange={setRole}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Rolle" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="customer">Kunde</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1">
                <label className="text-muted-foreground text-sm">Verifiziert</label>
                <div className="flex items-center gap-2">
                  <Button
                    type="button"
                    variant={isVerified ? "default" : "outline"}
                    size="sm"
                    onClick={() => setIsVerified(true)}
                  >
                    Ja
                  </Button>
                  <Button
                    type="button"
                    variant={isVerified === false ? "default" : "outline"}
                    size="sm"
                    onClick={() => setIsVerified(false)}
                  >
                    Nein
                  </Button>
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <div className="flex w-full justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  if (user) {
                    setEmail(user.email || "")
                    setFirstName(user.firstName || "")
                    setLastName(user.lastName || "")
                    setRole(user.role || "customer")
                    setIsVerified(user.isVerified)
                  }
                }}
              >
                Zurücksetzen
              </Button>
              <Button
                onClick={async () => {
                  await save()
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

function getCSRFCookie(): string | null {
  if (typeof document === "undefined") return null
  const name = (document.location.protocol === "https:" ? "__Host-" : "") + "csrf"
  const m = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/([.$?*|{}()\[\]\\/+^])/g, "\\$1") + "=([^;]*)"))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}
