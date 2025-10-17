"use client"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { API_BASE_URL } from "@/lib/api"

type StationRequest = {
  id: string
  name: string
  model: string
  os: string
  status: string
  createdAt: string
  expiresAt: string
}

type Station = {
  id: string
  name: string
  deviceKey: string
  approved: boolean
  approvedAt?: string
  createdAt: string
}

type Product = { id: string; name: string }

export default function AdminStationRequestsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<StationRequest[]>([])
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
      const res = await fetchAuth(`${API_BASE_URL}/v1/admin/stations/requests?status=pending`)
      const json = (await res.json()) as { items: StationRequest[] }
      setItems(json.items || [])
      const rs = await fetchAuth(`${API_BASE_URL}/v1/admin/stations`)
      const js = (await rs.json()) as { items: Station[] }
      setStations(js.items || [])
      // Load products for selection
      const pr = await fetchAuth(`${API_BASE_URL}/v1/products?limit=500`)
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

  async function act(id: string, action: "approve" | "reject") {
    try {
      const res = await fetchAuth(
        `${API_BASE_URL}/v1/admin/stations/requests/${id}/${action}`,
        { method: "POST" }
      )
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as { detail?: string; message?: string }
        throw new Error(j.detail || j.message || `Error ${res.status}`)
      }
      setItems((prev) => prev.filter((x) => x.id !== id))
    } catch {
      setError(action === "approve" ? "Annahme fehlgeschlagen" : "Ablehnung fehlgeschlagen")
    }
  }

  async function openEditor(st: Station) {
    const next = st.id === editingId ? null : st.id
    setEditingId(next)
    if (next) {
      try {
        const res = await fetchAuth(`${API_BASE_URL}/v1/admin/stations/${st.id}/products`)
        const j = (await res.json()) as { items: { productId: string; name: string }[] }
        setAssigned((prev) => ({ ...prev, [st.id]: j.items || [] }))
        setAddProductId(undefined)
      } catch {}
    }
  }

  async function addProductToStation(stationId: string) {
    if (!addProductId) return
    await fetchAuth(`${API_BASE_URL}/v1/admin/stations/${stationId}/products`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ productIds: [addProductId!] }),
    })
    const res = await fetchAuth(`${API_BASE_URL}/v1/admin/stations/${stationId}/products`)
    const j = (await res.json()) as { items: { productId: string; name: string }[] }
    setAssigned((prev) => ({ ...prev, [stationId]: j.items || [] }))
    setAddProductId(undefined)
  }

  async function removeProductFromStation(stationId: string, productId: string) {
    await fetchAuth(`${API_BASE_URL}/v1/admin/stations/${stationId}/products/${productId}`, {
      method: "DELETE",
    })
    setAssigned((prev) => ({ ...prev, [stationId]: (prev[stationId] || []).filter((x) => x.productId !== productId) }))
  }

  return (
    <div className="pt-4">
      <h1 className="font-primary mb-4 text-2xl">Stationen & Anfragen</h1>
      {error && (
        <div role="alert" className="text-destructive bg-destructive/10 mb-3 rounded px-3 py-2">
          {error}
        </div>
      )}
      {loading && <div className="text-muted-foreground">Lädt…</div>}
      {/* Stations list */}
      <h2 className="mt-2 mb-2 text-lg font-semibold">Stationen</h2>
      {!loading && stations.length === 0 && <div className="text-muted-foreground">Keine Stationen vorhanden.</div>}
      <div className="mb-6 space-y-3">
        {stations.map((st) => {
          const isEditing = editingId === st.id
          const assignedList = assigned[st.id] || []
          const unassigned = allProducts.filter((p) => !assignedList.some((a) => a.productId === p.id))
          return (
            <div key={st.id} className="bg-card border-border rounded-[11px] border p-3">
              <div className="flex items-center justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <div className="truncate font-semibold">{st.name}</div>
                  <div className="text-muted-foreground truncate text-sm">{maskKey(st.deviceKey)}</div>
                  <div className="text-muted-foreground text-xs">
                    Erstellt: {new Date(st.createdAt).toLocaleString()}
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <Button variant="outline" onClick={() => openEditor(st)}>
                    Produkte bearbeiten
                  </Button>
                </div>
              </div>
              {isEditing && (
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

      {/* Requests list */}
      <h2 className="mt-2 mb-2 text-lg font-semibold">Offene Anfragen</h2>
      {!loading && items.length === 0 && <div className="text-muted-foreground">Keine offenen Anfragen.</div>}
      <div className="space-y-3">
        {items.map((it) => (
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
