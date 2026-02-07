"use client"
import { useEffect, useState } from "react"
import { PairDeviceCard } from "@/components/admin/pair-device-card"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuth } from "@/contexts/auth-context"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { hasPermission } from "@/lib/auth/rbac"
import { getCSRFToken } from "@/lib/csrf"
import type { UserRole } from "@/types"

type DeviceBinding = {
  id: string
  device_type: string
  name: string
  created_at: string
  last_seen_at: string
  created_by_user_id: string
  revoked_at?: string | null
  station_id?: string | null
}

export default function AdminDevicesPage() {
  const fetchAuth = useAuthorizedFetch()
  const { user: currentUser } = useAuth()
  const [items, setItems] = useState<DeviceBinding[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [revoking, setRevoking] = useState<string | null>(null)
  const [confirmRevoke, setConfirmRevoke] = useState<{ id: string; name: string } | null>(null)

  const canRevoke = hasPermission(currentUser?.role as UserRole, "devices:revoke")

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const res = await fetchAuth(`/api/v1/devices`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = (await res.json()) as { items?: DeviceBinding[] } | DeviceBinding[]
      setItems(Array.isArray(data) ? data : data.items || [])
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Geräte konnten nicht geladen werden"
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  async function handleRevoke(id: string) {
    setConfirmRevoke(null)
    setRevoking(id)
    setError(null)
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/devices/${encodeURIComponent(id)}`, {
        method: "DELETE",
        headers: { "X-CSRF": csrf || "" },
      })
      if (!res.ok && res.status !== 204) throw new Error(`HTTP ${res.status}`)
      // Remove the revoked device from the list
      setItems((prev) => prev.filter((d) => d.id !== id))
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Gerät konnte nicht gesperrt werden"
      setError(msg)
    } finally {
      setRevoking(null)
    }
  }

  function formatDeviceType(type: string) {
    switch (type) {
      case "POS":
        return "POS"
      case "STATION":
        return "Station"
      default:
        return type
    }
  }

  return (
    <div className="min-w-0 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Geräte</h1>
      </div>

      <PairDeviceCard onPaired={() => void load()} />

      {error && <div className="text-sm text-red-600">{error}</div>}
      {loading && <div className="text-muted-foreground text-sm">Lädt...</div>}

      <div className="rounded-md border">
        <div className="relative">
          <div className="from-background pointer-events-none absolute inset-y-0 left-0 w-6 bg-gradient-to-r to-transparent" />
          <div className="from-background pointer-events-none absolute inset-y-0 right-0 w-6 bg-gradient-to-l to-transparent" />
          <div
            className="max-w-full overflow-x-auto overscroll-x-contain rounded-[10px]"
            role="region"
            aria-label="Devices table - scroll horizontally to reveal more columns"
            tabIndex={0}
          >
            <Table className="whitespace-nowrap">
              <TableHeader className="bg-card sticky top-0">
                <TableRow>
                  <TableHead className="whitespace-nowrap">Name</TableHead>
                  <TableHead className="whitespace-nowrap">Typ</TableHead>
                  <TableHead className="whitespace-nowrap">Zuletzt gesehen</TableHead>
                  <TableHead className="whitespace-nowrap">Erstellt</TableHead>
                  <TableHead className="whitespace-nowrap">Erstellt von</TableHead>
                  {canRevoke && <TableHead className="text-right whitespace-nowrap">Aktionen</TableHead>}
                </TableRow>
              </TableHeader>
              <TableBody>
                {!loading && items.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={canRevoke ? 6 : 5} className="text-muted-foreground text-center">
                      Keine aktiven Geräte.
                    </TableCell>
                  </TableRow>
                )}
                {items.map((d) => {
                  const lastSeen = d.last_seen_at ? new Date(d.last_seen_at).toLocaleString("de-CH") : "-"
                  const created = d.created_at ? new Date(d.created_at).toLocaleString("de-CH") : "-"
                  return (
                    <TableRow key={d.id} className="even:bg-card odd:bg-muted/40">
                      <TableCell>{d.name || "-"}</TableCell>
                      <TableCell>{formatDeviceType(d.device_type)}</TableCell>
                      <TableCell className="whitespace-nowrap">{lastSeen}</TableCell>
                      <TableCell className="whitespace-nowrap">{created}</TableCell>
                      <TableCell className="text-xs">{d.created_by_user_id}</TableCell>
                      {canRevoke && (
                        <TableCell className="text-right">
                          <Button
                            size="sm"
                            variant="destructive"
                            disabled={revoking === d.id}
                            onClick={() => setConfirmRevoke({ id: d.id, name: d.name || d.id })}
                          >
                            {revoking === d.id ? "Wird gesperrt..." : "Sperren"}
                          </Button>
                        </TableCell>
                      )}
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>

      <AlertDialog open={!!confirmRevoke} onOpenChange={(open) => { if (!open) setConfirmRevoke(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Gerät sperren</AlertDialogTitle>
            <AlertDialogDescription>
              Möchtest du &quot;{confirmRevoke?.name}&quot; wirklich sperren? Das Gerät muss danach erneut registriert werden.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction onClick={() => { if (confirmRevoke) void handleRevoke(confirmRevoke.id) }}>
              Sperren
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
