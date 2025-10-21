"use client"

import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"

type PosRequest = {
  id: string
  name: string
  model: string
  os: string
  status: string
  createdAt: string
  expiresAt: string
}

type PosDevice = {
  id: string
  name: string
  token: string
  approved: boolean
  approvedAt?: string
  cardCapable?: boolean | null
  printerMac?: string | null
  printerUuid?: string | null
}

export default function AdminPOSPage() {
  const fetchAuth = useAuthorizedFetch()
  const [requests, setRequests] = useState<PosRequest[]>([])
  const [devices, setDevices] = useState<PosDevice[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const r1 = await fetchAuth(`/api/v1/admin/pos/requests?status=pending`)
      const j1 = (await r1.json()) as { items: PosRequest[] }
      setRequests(j1.items || [])
      const r2 = await fetchAuth(`/api/v1/admin/pos/devices`)
      const j2 = (await r2.json()) as { items: PosDevice[] }
      setDevices(j2.items || [])
    } catch {
      setError("Laden fehlgeschlagen")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function act(id: string, action: "approve" | "reject") {
    try {
      const res = await fetchAuth(`/api/v1/admin/pos/requests/${id}/${action}`, {
        method: "POST",
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as { detail?: string; message?: string }
        throw new Error(j.detail || j.message || `Error ${res.status}`)
      }
      setRequests((prev) => prev.filter((x) => x.id !== id))
      if (action === "approve") {
        try {
          const r = await fetchAuth(`/api/v1/admin/pos/devices`)
          const j = (await r.json()) as { items: PosDevice[] }
          setDevices(j.items || [])
        } catch {}
      }
    } catch {
      setError(action === "approve" ? "Annahme fehlgeschlagen" : "Ablehnung fehlgeschlagen")
    }
  }

  return (
    <div className="pt-4">
      <h1 className="font-primary mb-4 text-2xl">POS & Anfragen</h1>
      {error && (
        <div role="alert" className="text-destructive bg-destructive/10 mb-3 rounded px-3 py-2">
          {error}
        </div>
      )}
      {loading && <div className="text-muted-foreground">Lädt…</div>}

      {/* Devices list */}
      <h2 className="mt-2 mb-2 text-lg font-semibold">POS‑Geräte</h2>
      {!loading && devices.length === 0 && <div className="text-muted-foreground">Keine POS‑Geräte vorhanden.</div>}
      <div className="mb-6 space-y-3">
        {devices.map((d) => (
          <div key={d.id} className="bg-card border-border rounded-[11px] border p-3">
            <div className="flex items-center justify-between gap-4">
              <div className="min-w-0 flex-1">
                <div className="truncate font-semibold">{d.name}</div>
                <div className="text-muted-foreground truncate text-sm">{maskKey(d.token)}</div>
                <div className="text-muted-foreground flex flex-wrap gap-2 text-xs">
                  <span>Status: {d.approved ? "Angenommen" : "Ausstehend"}</span>
                  {d.approvedAt && <span>• Freigegeben: {new Date(d.approvedAt).toLocaleString()}</span>}
                  {typeof d.cardCapable === "boolean" && <span>• Karte: {d.cardCapable ? "Ja" : "Nein"}</span>}
                  {d.printerMac && <span>• Drucker MAC: {d.printerMac}</span>}
                  {d.printerUuid && <span>• Drucker UUID: {d.printerUuid}</span>}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Requests list */}
      <h2 className="mt-2 mb-2 text-lg font-semibold">Offene Anfragen</h2>
      {!loading && requests.length === 0 && <div className="text-muted-foreground">Keine offenen Anfragen.</div>}
      <div className="space-y-3">
        {requests.map((it) => (
          <div
            key={it.id}
            className="bg-card border-border flex items-center justify-between gap-4 rounded-[11px] border p-3"
          >
            <div className="min-w-0 flex-1">
              <div className="truncate font-semibold">{it.name}</div>
              <div className="text-muted-foreground truncate text-sm">
                {it.model} • {it.os}
              </div>
              <div className="text-muted-foreground text-xs">Erstellt: {new Date(it.createdAt).toLocaleString()}</div>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <Button variant="destructive" onClick={() => act(it.id, "reject")}>
                Ablehnen
              </Button>
              <Button variant="success" onClick={() => act(it.id, "approve")}>
                Annehmen
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function maskKey(k: string) {
  if (!k) return ""
  if (k.length <= 6) return "•••"
  return `${k.slice(0, 3)}•••${k.slice(-3)}`
}
