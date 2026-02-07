"use client"
import { useEffect, useState } from "react"
import { PairDeviceCard } from "@/components/admin/pair-device-card"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { getCSRFToken } from "@/lib/csrf"

type Station = {
  id: string
  name: string
  model?: string
  os?: string
  deviceKey: string
  approved: boolean
  status: string
  approvedAt?: string
  createdAt: string
}

type Product = { id: string; name: string }

export default function AdminStationRequestsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [stations, setStations] = useState<Station[]>([])
  const [editingId, setEditingId] = useState<string | null>(null)
  const [assigned, setAssigned] = useState<Record<string, { productId: string; name: string }[]>>({})
  const [allProducts, setAllProducts] = useState<Product[]>([])
  const [addProductId, setAddProductId] = useState<string | undefined>(undefined)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const res = await fetchAuth(`/api/v1/stations`)
      const json = (await res.json()) as { items: Station[] }
      setStations(json.items || [])
      // Load products for selection
      const pr = await fetchAuth(`/api/v1/products`)
      const pj = (await pr.json()) as { items?: { id: string; name: string }[] }
      if (pj && Array.isArray(pj.items)) setAllProducts(pj.items.map((x) => ({ id: x.id, name: x.name })))
    } catch {
      setError("Laden fehlgeschlagen")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const sortedStations = [...stations].sort(
    (a, b) =>
      statusOrder(a.status) - statusOrder(b.status) || new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  )

  async function revokeStation(id: string) {
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/stations/${id}`, {
        method: "DELETE",
        headers: { "X-CSRF": csrf || "" },
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as { detail?: string; message?: string }
        throw new Error(j.detail || j.message || `Error ${res.status}`)
      }
      setStations((prev) =>
        prev.map((s) => (s.id === id ? { ...s, status: "revoked", approved: false } : s))
      )
      if (editingId === id) {
        setEditingId(null)
      }
    } catch {
      setError("Sperrung fehlgeschlagen")
    }
  }

  async function openEditor(st: Station) {
    if (st.status !== "approved") {
      return
    }
    const next = st.id === editingId ? null : st.id
    setEditingId(next)
    if (next) {
      try {
        const res = await fetchAuth(`/api/v1/stations/${st.id}/products`)
        const j = (await res.json()) as { items: { productId: string; name: string }[] }
        setAssigned((prev) => ({ ...prev, [st.id]: j.items || [] }))
        setAddProductId(undefined)
      } catch {}
    }
  }

  async function addProductToStation(stationId: string) {
    if (!addProductId) return
    await fetchAuth(`/api/v1/stations/${stationId}/products`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ productIds: [addProductId!] }),
    })
    const res = await fetchAuth(`/api/v1/stations/${stationId}/products`)
    const j = (await res.json()) as { items: { productId: string; name: string }[] }
    setAssigned((prev) => ({ ...prev, [stationId]: j.items || [] }))
    setAddProductId(undefined)
  }

  async function removeProductFromStation(stationId: string, productId: string) {
    await fetchAuth(`/api/v1/stations/${stationId}/products/${productId}`, {
      method: "DELETE",
    })
    setAssigned((prev) => ({ ...prev, [stationId]: (prev[stationId] || []).filter((x) => x.productId !== productId) }))
  }

  return (
    <div className="pt-4">
      <h1 className="font-primary mb-4 text-2xl">Stationen</h1>

      <div className="mb-6">
        <PairDeviceCard onPaired={load} />
      </div>

      {error && (
        <div role="alert" className="text-destructive bg-destructive/10 mb-3 rounded px-3 py-2">
          {error}
        </div>
      )}
      {loading && <div className="text-muted-foreground">Lädt…</div>}
      <h2 className="mt-2 mb-2 text-lg font-semibold">Übersicht</h2>
      {!loading && sortedStations.length === 0 && (
        <div className="text-muted-foreground">Keine Stationen vorhanden.</div>
      )}
      <div className="space-y-3">
        {sortedStations.map((st) => {
          const isEditing = editingId === st.id
          const assignedList = assigned[st.id] || []
          const unassigned = allProducts.filter((p) => !assignedList.some((a) => a.productId === p.id))
          return (
            <div key={st.id} className="bg-card border-border rounded-[11px] border p-3">
              <div className="flex items-center justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <div className="truncate font-semibold">{st.name}</div>
                  <div className="text-muted-foreground truncate text-sm">
                    {st.model || "Unbekanntes Modell"} • {st.os || "Unbekanntes OS"}
                  </div>
                  <div className="text-muted-foreground truncate text-sm">{maskKey(st.deviceKey)}</div>
                  <div className="text-muted-foreground text-xs">
                    Erstellt: {new Date(st.createdAt).toLocaleString()}
                    {st.approvedAt && st.status === "approved" && (
                      <> • Freigegeben: {new Date(st.approvedAt).toLocaleString()}</>
                    )}
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <StatusBadge status={st.status} />
                  {st.status === "approved" && (
                    <>
                      <Button variant="outline" onClick={() => openEditor(st)}>
                        Produkte bearbeiten
                      </Button>
                      <Button variant="destructive" onClick={() => revokeStation(st.id)}>
                        Entfernen
                      </Button>
                    </>
                  )}
                </div>
              </div>
              {isEditing && st.status === "approved" && (
                <div className="border-border mt-3 rounded-lg border p-3">
                  <div className="flex flex-col gap-2">
                    <label className="text-sm">Produkt hinzufügen</label>
                    <div className="flex items-center gap-2">
                      <Select
                        key={`${st.id}-${assignedList.length}-${addProductId ?? "none"}`}
                        value={addProductId}
                        onValueChange={setAddProductId}
                      >
                        <SelectTrigger className="flex-1">
                          <SelectValue placeholder="Produkt auswählen" />
                        </SelectTrigger>
                        <SelectContent>
                          {unassigned.map((p) => (
                            <SelectItem key={p.id} value={p.id}>
                              {p.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <Button disabled={!addProductId} onClick={() => addProductToStation(st.id)}>
                        Hinzufügen
                      </Button>
                    </div>
                  </div>
                  <div className="mt-3">
                    <div className="mb-2 text-sm font-medium">Zugewiesene Produkte</div>
                    {assignedList.length === 0 ? (
                      <div className="text-muted-foreground text-sm">Keine Produkte zugewiesen.</div>
                    ) : (
                      <div className="flex flex-wrap gap-2">
                        {assignedList.map((a) => (
                          <span
                            key={a.productId}
                            className="border-border inline-flex items-center gap-2 rounded-full border px-3 py-1 text-sm"
                          >
                            {a.name}
                            <Button
                              aria-label="Entfernen"
                              variant="ghost"
                              size="sm"
                              className="text-muted-foreground hover:text-destructive h-6 px-2"
                              onClick={() => removeProductFromStation(st.id, a.productId)}
                            >
                              ×
                            </Button>
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )
        })}
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
