"use client"
import { useEffect, useMemo, useState } from "react"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"

type KPI = { label: string; value: string }

export default function AdminDashboard() {
  const fetchAuth = useAuthorizedFetch()
  const [kpis, setKpis] = useState<KPI[]>([])
  const [lowStock, setLowStock] = useState<{ id: string; name: string; qty?: number | null }[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        // Today orders summary (admin endpoint; fallback to 0 on failure)
        const now = new Date()
        const from = new Date(now.getFullYear(), now.getMonth(), now.getDate()).toISOString()
        const to = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1).toISOString()
        const [ordersRes, productsRes] = await Promise.all([
          fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/orders?date_from=${encodeURIComponent(from)}&date_to=${encodeURIComponent(to)}&limit=1`),
          fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/products?limit=200`),
        ])

        let todayCount = 0
        let todayRevenueCents = 0
        if (ordersRes.ok) {
          const data = (await ordersRes.json()) as { items: Array<{ id: string; totalCents: number }>; count: number; totals?: { revenueCents?: number } }
          todayCount = data.count || 0
          todayRevenueCents = data.totals?.revenueCents ?? 0
        }

        let low: { id: string; name: string; qty?: number | null }[] = []
        if (productsRes.ok) {
          const pr = (await productsRes.json()) as { items?: Array<{ id: string; name: string; isLowStock?: boolean; availableQuantity?: number | null }> }
          const items = pr.items || []
          low = items.filter((p) => p.isLowStock).map((p) => ({ id: p.id, name: p.name, qty: p.availableQuantity ?? null }))
        }

        if (!cancelled) {
          setKpis([
            { label: "Bestellungen heute", value: String(todayCount) },
            { label: "Umsatz heute", value: new Intl.NumberFormat("de-CH", { style: "currency", currency: "CHF" }).format((todayRevenueCents || 0) / 100) },
            { label: "Artikel mit niedrigem Bestand", value: String(low.length) },
          ])
          setLowStock(low)
        }
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Dashboard laden fehlgeschlagen"
        if (!cancelled) setError(msg)
      }
    })()
    return () => { cancelled = true }
  }, [fetchAuth])

  const lowList = useMemo(() => lowStock.slice(0, 5), [lowStock])

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Adminbereich</h1>
      {error && <div className="text-red-600 text-sm">{error}</div>}
      <section className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {kpis.map((k) => (
          <div key={k.label} className="rounded-lg border border-gray-200 p-4">
            <div className="text-sm text-gray-500">{k.label}</div>
            <div className="mt-1 text-xl font-semibold">{k.value}</div>
          </div>
        ))}
      </section>
      <section>
        <h2 className="text-lg font-medium mb-2">Artikel mit niedrigem Bestand</h2>
        {lowList.length === 0 ? (
          <div className="text-sm text-gray-500">Keine Artikel mit niedrigem Bestand.</div>
        ) : (
          <ul className="text-sm divide-y divide-gray-100 border border-gray-100 rounded-md">
            {lowList.map((it) => (
              <li key={it.id} className="px-3 py-2 flex items-center justify-between">
                <span>{it.name}</span>
                <span className="text-gray-600">{typeof it.qty === "number" ? `${it.qty} übrig` : "—"}</span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
