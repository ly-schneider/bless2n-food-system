"use client"
import { useEffect, useState } from "react"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Button } from "@/components/ui/button"

type Order = { id: string; status: string; totalCents?: number; createdAt: string }

export default function AdminOrdersPage() {
  const fetchAuth = useAuthorizedFetch()
  const [status, setStatus] = useState<string>("all")
  const [items, setItems] = useState<Order[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const url = new URL(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/orders`)
        if (status && status !== "all") url.searchParams.set("status", status)
        const res = await fetchAuth(url.toString())
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: Order[] }
        if (!cancelled) setItems(data.items || [])
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Failed to load orders"
        if (!cancelled) setError(msg)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth, status])

  async function exportCSV() {
    const url = new URL(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/orders/export.csv`)
    if (status && status !== "all") url.searchParams.set("status", status)
    window.location.href = url.toString()
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Bestellungen</h1>
      {error && <div className="text-sm text-red-600">{error}</div>}
      <div className="overflow-x-auto">
        <table className="w-full border border-gray-100 text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-3 py-2 text-left">ID</th>
              <th className="px-3 py-2 text-left">Status</th>
              <th className="px-3 py-2 text-left">Erstellt</th>
            </tr>
          </thead>
          <tbody>
            {items.map((o) => (
              <tr key={o.id} className="border-t border-gray-100">
                <td className="px-3 py-2">{o.id}</td>
                <td className="px-3 py-2">{o.status}</td>
                <td className="px-3 py-2">{new Date(o.createdAt).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
