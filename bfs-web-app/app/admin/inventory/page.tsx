"use client"
import Link from "next/link"
import { useCallback, useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { formatChf } from "@/lib/utils"

type Product = {
  id: string
  name: string
  priceCents: number
}

type LedgerEntry = {
  id: string
  productId: string
  delta: number
  reason: string
  createdAt: string
  orderId?: string | null
  orderLineId?: string | null
  deviceId?: string | null
  createdBy?: string | null
}

const REASON_LABELS: Record<string, string> = {
  opening_balance: "Eröffnung",
  sale: "Verkauf",
  refund: "Erstattung",
  cancellation: "Stornierung",
  manual_adjust: "Manuelle Anpassung",
  correction: "Korrektur",
}

const PAGE_SIZE = 50

export default function InventoryHistoryPage() {
  const fetchAuth = useAuthorizedFetch()
  const [products, setProducts] = useState<Product[]>([])
  const [selectedProductId, setSelectedProductId] = useState<string>("")
  const [entries, setEntries] = useState<LedgerEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [offset, setOffset] = useState(0)
  const [hasMore, setHasMore] = useState(false)

  // Load products for the selector
  useEffect(() => {
    let cancelled = false
    async function loadProducts() {
      try {
        const res = await fetchAuth("/api/v1/products")
        if (!res.ok) return
        const data = (await res.json()) as { items?: Product[] }
        const items = data.items ?? []
        if (!cancelled) setProducts(items)
      } catch {
        // ignore
      }
    }
    void loadProducts()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  const loadEntries = useCallback(
    async (productId: string, newOffset: number) => {
      if (!productId) return
      setLoading(true)
      setError(null)
      try {
        const params = new URLSearchParams({
          limit: String(PAGE_SIZE),
          offset: String(newOffset),
        })
        const res = await fetchAuth(`/api/v1/products/${encodeURIComponent(productId)}/inventory/history?${params}`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items?: LedgerEntry[] }
        const items = data.items ?? []
        if (newOffset === 0) {
          setEntries(items)
        } else {
          setEntries((prev) => [...prev, ...items])
        }
        setHasMore(items.length === PAGE_SIZE)
        setOffset(newOffset)
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Fehler beim Laden"
        setError(msg)
      } finally {
        setLoading(false)
      }
    },
    [fetchAuth]
  )

  function handleProductChange(productId: string) {
    setSelectedProductId(productId)
    setEntries([])
    setOffset(0)
    setHasMore(false)
    if (productId) {
      void loadEntries(productId, 0)
    }
  }

  return (
    <div className="min-w-0 space-y-6">
      <h1 className="text-xl font-semibold">Inventarverlauf</h1>

      {/* Product selector */}
      <div className="max-w-sm space-y-1">
        <Label htmlFor="product-select">Produkt auswählen</Label>
        <Select value={selectedProductId} onValueChange={handleProductChange}>
          <SelectTrigger id="product-select">
            <SelectValue placeholder="Produkt wählen…" />
          </SelectTrigger>
          <SelectContent>
            {products.map((p) => (
              <SelectItem key={p.id} value={p.id}>
                {p.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {error && <div className="text-sm text-red-600">{error}</div>}

      {selectedProductId && entries.length === 0 && !loading && (
        <p className="text-muted-foreground text-sm">Keine Einträge vorhanden.</p>
      )}

      {entries.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left">
                <th className="pr-4 pb-2 font-medium">Datum</th>
                <th className="pr-4 pb-2 font-medium">Delta</th>
                <th className="pr-4 pb-2 font-medium">Grund</th>
                <th className="pr-4 pb-2 font-medium">Bestellung</th>
                <th className="pb-2 font-medium">Erstellt von</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((e) => (
                <tr key={e.id} className="border-b last:border-0">
                  <td className="text-muted-foreground py-2 pr-4 whitespace-nowrap">
                    {new Date(e.createdAt).toLocaleString("de-CH")}
                  </td>
                  <td className="py-2 pr-4">
                    <span className={e.delta >= 0 ? "font-medium text-green-700" : "font-medium text-red-600"}>
                      {e.delta >= 0 ? `+${e.delta}` : e.delta}
                    </span>
                  </td>
                  <td className="py-2 pr-4">{REASON_LABELS[e.reason] ?? e.reason}</td>
                  <td className="py-2 pr-4">
                    {e.orderId ? (
                      <Link
                        href={`/admin/orders/${encodeURIComponent(e.orderId)}`}
                        className="text-xs underline underline-offset-2"
                      >
                        {e.orderId.slice(0, 8)}…
                      </Link>
                    ) : (
                      <span className="text-muted-foreground">–</span>
                    )}
                  </td>
                  <td className="py-2">
                    {e.createdBy ? (
                      <Link
                        href={`/admin/users/${encodeURIComponent(e.createdBy)}`}
                        className="text-xs underline underline-offset-2"
                      >
                        {e.createdBy.slice(0, 8)}…
                      </Link>
                    ) : (
                      <span className="text-muted-foreground">–</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {loading && <p className="text-muted-foreground text-sm">Lade…</p>}

      {hasMore && !loading && (
        <Button variant="outline" size="sm" onClick={() => void loadEntries(selectedProductId, offset + PAGE_SIZE)}>
          Mehr laden
        </Button>
      )}
    </div>
  )
}
