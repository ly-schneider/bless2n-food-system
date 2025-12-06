"use client"

import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"

type PosDevice = {
  id: string
  name: string
  model?: string
  os?: string
  token: string
  approved: boolean
  status: string
  approvedAt?: string
  createdAt?: string
  cardCapable?: boolean | null
  printerMac?: string | null
  printerUuid?: string | null
}

export default function AdminPOSPage() {
  const fetchAuth = useAuthorizedFetch()
  const [devices, setDevices] = useState<PosDevice[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const r = await fetchAuth(`/api/v1/admin/pos/devices`)
      const j = (await r.json()) as { items: PosDevice[] }
      setDevices(j.items || [])
    } catch {
      setError("Laden fehlgeschlagen")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const sortedDevices = [...devices].sort(
    (a, b) =>
      statusOrder(a.status) - statusOrder(b.status) ||
      (new Date(b.approvedAt ?? b.createdAt ?? "").getTime() || 0) -
        (new Date(a.approvedAt ?? a.createdAt ?? "").getTime() || 0)
  )

  async function act(id: string, action: "approve" | "reject" | "revoke") {
    try {
      const res = await fetchAuth(`/api/v1/admin/pos/requests/${id}/${action}`, {
        method: "POST",
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as { detail?: string; message?: string }
        throw new Error(j.detail || j.message || `Error ${res.status}`)
      }
      const nextStatus = action === "approve" ? "approved" : action === "revoke" ? "revoked" : "rejected"
      const approvedAt = action === "approve" ? new Date().toISOString() : undefined
      setDevices((prev) =>
        prev.map((d) => (d.id === id ? { ...d, status: nextStatus, approved: action === "approve", approvedAt } : d))
      )
    } catch {
      setError(
        action === "approve"
          ? "Annahme fehlgeschlagen"
          : action === "revoke"
            ? "Sperrung fehlgeschlagen"
            : "Ablehnung fehlgeschlagen"
      )
    }
  }

  return (
    <div className="pt-4">
      <h1 className="font-primary mb-4 text-2xl">POS-Geräte</h1>
      {error && (
        <div role="alert" className="text-destructive bg-destructive/10 mb-3 rounded px-3 py-2">
          {error}
        </div>
      )}
      {loading && <div className="text-muted-foreground">Lädt…</div>}

      <h2 className="mt-2 mb-2 text-lg font-semibold">Übersicht</h2>
      {!loading && sortedDevices.length === 0 && (
        <div className="text-muted-foreground">Keine POS-Geräte vorhanden.</div>
      )}
      <div className="space-y-3">
        {sortedDevices.map((d) => (
          <div key={d.id} className="bg-card border-border rounded-[11px] border p-3">
            <div className="flex items-center justify-between gap-4">
              <div className="min-w-0 flex-1">
                <div className="truncate font-semibold">{d.name}</div>
                <div className="text-muted-foreground truncate text-sm">
                  {d.model || "Unbekanntes Modell"} • {d.os || "Unbekanntes OS"}
                </div>
                <div className="text-muted-foreground truncate text-sm">{maskKey(d.token)}</div>
                <div className="text-muted-foreground flex flex-wrap gap-2 text-xs">
                  {d.createdAt && <span>Erstellt: {new Date(d.createdAt).toLocaleString()}</span>}
                  {d.approvedAt && d.status === "approved" && (
                    <span>• Freigegeben: {new Date(d.approvedAt).toLocaleString()}</span>
                  )}
                  {typeof d.cardCapable === "boolean" && <span>• Karte: {d.cardCapable ? "Ja" : "Nein"}</span>}
                  {d.printerMac && <span>• Drucker MAC: {d.printerMac}</span>}
                  {d.printerUuid && <span>• Drucker UUID: {d.printerUuid}</span>}
                </div>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <StatusBadge status={d.status} />
                {d.status === "pending" && (
                  <>
                    <Button variant="destructive" onClick={() => act(d.id, "reject")}>
                      Ablehnen
                    </Button>
                    <Button variant="success" onClick={() => act(d.id, "approve")}>
                      Annehmen
                    </Button>
                  </>
                )}
                {d.status === "approved" && (
                  <Button variant="destructive" onClick={() => act(d.id, "revoke")}>
                    Entfernen
                  </Button>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const label =
    status === "approved"
      ? "Angenommen"
      : status === "rejected"
        ? "Abgelehnt"
        : status === "revoked"
          ? "Gesperrt"
          : "Ausstehend"
  const tone =
    status === "approved"
      ? "bg-emerald-100 text-emerald-800"
      : status === "rejected"
        ? "bg-destructive/10 text-destructive"
        : status === "revoked"
          ? "bg-slate-200 text-slate-700"
          : "bg-amber-100 text-amber-800"
  return (
    <span className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold ${tone}`}>{label}</span>
  )
}

function statusOrder(status: string) {
  if (status === "pending") return 0
  if (status === "approved") return 1
  if (status === "revoked") return 2
  return 2
}

function maskKey(k: string) {
  if (!k) return ""
  if (k.length <= 6) return "•••"
  return `${k.slice(0, 3)}•••${k.slice(-3)}`
}
