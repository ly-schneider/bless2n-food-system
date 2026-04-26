"use client"

import { Check, ChevronDown, ChevronRight, Pencil, X } from "lucide-react"
import { useEffect, useMemo, useRef, useState } from "react"
import { PairDeviceCard } from "@/components/admin/pair-device-card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Input } from "@/components/ui/input"
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

function csrfHeaders(method: string, json = true): Record<string, string> {
  const csrf = getCSRFToken()
  const h: Record<string, string> = { "X-CSRF": csrf || "" }
  if (json && method !== "DELETE") h["Content-Type"] = "application/json"
  return h
}

export default function AdminStationRequestsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [stations, setStations] = useState<Station[]>([])
  const [assigned, setAssigned] = useState<Record<string, { productId: string; name: string }[]>>({})
  const [allProducts, setAllProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const [sr, pr] = await Promise.all([fetchAuth(`/api/v1/stations`), fetchAuth(`/api/v1/products`)])
      const sj = (await sr.json()) as { items: Station[] }
      setStations(sj.items || [])
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

  async function renameStation(id: string, name: string) {
    try {
      const res = await fetchAuth(`/api/v1/stations/${id}`, {
        method: "PATCH",
        headers: csrfHeaders("PATCH"),
        body: JSON.stringify({ name }),
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as { detail?: string; message?: string }
        throw new Error(j.detail || j.message || `Error ${res.status}`)
      }
      setStations((prev) => prev.map((s) => (s.id === id ? { ...s, name } : s)))
    } catch {
      setError("Umbenennen fehlgeschlagen")
    }
  }

  async function revokeStation(id: string) {
    try {
      const res = await fetchAuth(`/api/v1/stations/${id}`, {
        method: "DELETE",
        headers: csrfHeaders("DELETE", false),
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as { detail?: string; message?: string }
        throw new Error(j.detail || j.message || `Error ${res.status}`)
      }
      setStations((prev) => prev.map((s) => (s.id === id ? { ...s, status: "revoked", approved: false } : s)))
    } catch {
      setError("Sperrung fehlgeschlagen")
    }
  }

  async function loadStationProducts(stationId: string) {
    const res = await fetchAuth(`/api/v1/stations/${stationId}/products`)
    const j = (await res.json()) as { items: { productId: string; name: string }[] }
    setAssigned((prev) => ({ ...prev, [stationId]: j.items || [] }))
  }

  async function addProduct(stationId: string, productId: string) {
    const current = assigned[stationId] || []
    const allIds = [...current.map((a) => a.productId), productId]
    try {
      const res = await fetchAuth(`/api/v1/stations/${stationId}/products`, {
        method: "PUT",
        headers: csrfHeaders("PUT"),
        body: JSON.stringify({ productIds: allIds }),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      await loadStationProducts(stationId)
    } catch {
      setError("Produkt hinzufügen fehlgeschlagen")
    }
  }

  async function removeProduct(stationId: string, productId: string) {
    try {
      const res = await fetchAuth(`/api/v1/stations/${stationId}/products/${productId}`, {
        method: "DELETE",
        headers: csrfHeaders("DELETE", false),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      setAssigned((prev) => ({
        ...prev,
        [stationId]: (prev[stationId] || []).filter((x) => x.productId !== productId),
      }))
    } catch {
      setError("Produkt entfernen fehlgeschlagen")
    }
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
        {sortedStations.map((st) => (
          <StationCard
            key={st.id}
            station={st}
            allProducts={allProducts}
            assigned={assigned[st.id] || []}
            onLoadProducts={() => loadStationProducts(st.id)}
            onAddProduct={(productId) => addProduct(st.id, productId)}
            onRemoveProduct={(productId) => removeProduct(st.id, productId)}
            onRename={(name) => renameStation(st.id, name)}
            onRevoke={() => revokeStation(st.id)}
          />
        ))}
      </div>
    </div>
  )
}

function StationCard({
  station: st,
  allProducts,
  assigned,
  onLoadProducts,
  onAddProduct,
  onRemoveProduct,
  onRename,
  onRevoke,
}: {
  station: Station
  allProducts: Product[]
  assigned: { productId: string; name: string }[]
  onLoadProducts: () => Promise<void>
  onAddProduct: (productId: string) => Promise<void>
  onRemoveProduct: (productId: string) => Promise<void>
  onRename: (name: string) => Promise<void>
  onRevoke: () => void
}) {
  const [open, setOpen] = useState(false)
  const [loaded, setLoaded] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editName, setEditName] = useState(st.name)
  const inputRef = useRef<HTMLInputElement>(null)

  const unassigned = useMemo(
    () => allProducts.filter((p) => !assigned.some((a) => a.productId === p.id)),
    [allProducts, assigned]
  )

  const handleOpenChange = async (isOpen: boolean) => {
    setOpen(isOpen)
    if (isOpen && !loaded) {
      await onLoadProducts()
      setLoaded(true)
    }
  }

  return (
    <div className="bg-card border-border rounded-[11px] border p-3">
      <div className="flex items-center justify-between gap-4">
        <div className="min-w-0 flex-1">
          {editing ? (
            <form
              className="flex items-center gap-1.5"
              onSubmit={async (e) => {
                e.preventDefault()
                const trimmed = editName.trim()
                if (trimmed && trimmed !== st.name) {
                  await onRename(trimmed)
                }
                setEditing(false)
              }}
            >
              <Input
                ref={inputRef}
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                maxLength={20}
                className="h-7 w-40 text-sm font-semibold"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === "Escape") {
                    setEditName(st.name)
                    setEditing(false)
                  }
                }}
              />
              <Button type="submit" variant="ghost" size="icon" className="size-7">
                <Check className="size-3.5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-7"
                onClick={() => {
                  setEditName(st.name)
                  setEditing(false)
                }}
              >
                <X className="size-3.5" />
              </Button>
            </form>
          ) : (
            <div className="flex items-center gap-1.5">
              <div className="truncate font-semibold">{st.name}</div>
              <button
                className="text-muted-foreground hover:text-foreground shrink-0"
                onClick={() => {
                  setEditName(st.name)
                  setEditing(true)
                }}
                aria-label="Station umbenennen"
              >
                <Pencil className="size-3.5" />
              </button>
            </div>
          )}
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
            <Button variant="outline" size="sm" className="text-destructive hover:text-destructive" onClick={onRevoke}>
              Entfernen
            </Button>
          )}
        </div>
      </div>

      {st.status === "approved" && (
        <Collapsible open={open} onOpenChange={handleOpenChange} className="mt-2">
          <CollapsibleTrigger asChild>
            <Button variant="ghost" size="sm" className="text-muted-foreground -ml-2 gap-1.5 text-xs">
              {open ? <ChevronDown className="size-3.5" /> : <ChevronRight className="size-3.5" />}
              {assigned.length} {assigned.length === 1 ? "Produkt" : "Produkte"}
            </Button>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div className="mt-2 space-y-3">
              <div className="flex flex-wrap gap-1.5">
                {assigned.map((a) => (
                  <Badge key={a.productId} variant="secondary" className="gap-1 pr-1">
                    {a.name}
                    <button
                      className="hover:bg-destructive/20 rounded p-0.5"
                      onClick={() => onRemoveProduct(a.productId)}
                      aria-label={`Produkt ${a.name} entfernen`}
                    >
                      <X className="size-3" />
                    </button>
                  </Badge>
                ))}
                {assigned.length === 0 && (
                  <span className="text-muted-foreground text-sm">Keine Produkte zugewiesen.</span>
                )}
              </div>

              {unassigned.length > 0 && (
                <Select value="" onValueChange={onAddProduct}>
                  <SelectTrigger className="h-8 text-xs">
                    <SelectValue placeholder="Produkt hinzufügen…" />
                  </SelectTrigger>
                  <SelectContent>
                    {unassigned.map((p) => (
                      <SelectItem key={p.id} value={p.id}>
                        {p.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
          </CollapsibleContent>
        </Collapsible>
      )}
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
  return <span className={`inline-flex h-8 items-center rounded-md px-2.5 text-xs font-medium ${tone}`}>{label}</span>
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
