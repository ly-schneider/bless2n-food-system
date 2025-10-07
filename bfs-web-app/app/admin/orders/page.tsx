"use client"
import Link from "next/link"
import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { formatChf } from "@/lib/utils"

// Minimal shape expected from the admin orders endpoint
type Order = {
  id: string
  status: string
  totalCents?: number | null
  createdAt: string
  updatedAt?: string
  contactEmail?: string | null
  customerId?: string | null
  paymentIntentId?: string | null
  stripeChargeId?: string | null
  paymentAttemptId?: string | null
}

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
    <div className="min-w-0 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Bestellungen</h1>
      </div>

      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="rounded-md border">
        <div className="relative">
          <div className="from-background pointer-events-none absolute inset-y-0 left-0 w-6 bg-gradient-to-r to-transparent" />
          <div className="from-background pointer-events-none absolute inset-y-0 right-0 w-6 bg-gradient-to-l to-transparent" />
          <div
            className="max-w-full overflow-x-auto overscroll-x-contain rounded-[10px]"
            role="region"
            aria-label="Orders table – scroll horizontally to reveal more columns"
            tabIndex={0}
          >
            <Table className="whitespace-nowrap">
              <TableHeader className="bg-card sticky top-0">
                <TableRow>
                  <TableHead className="whitespace-nowrap">ID</TableHead>
                  <TableHead className="whitespace-nowrap">Status</TableHead>
                  <TableHead className="whitespace-nowrap">Preis</TableHead>
                  <TableHead className="whitespace-nowrap">Kontakt E-Mail</TableHead>
                  <TableHead className="whitespace-nowrap">Benutzer</TableHead>
                  <TableHead className="whitespace-nowrap">Stripe PI</TableHead>
                  <TableHead className="whitespace-nowrap">Stripe Charge</TableHead>
                  <TableHead className="whitespace-nowrap">Payment Attempt</TableHead>
                  <TableHead className="whitespace-nowrap">Erstellt</TableHead>
                  <TableHead className="whitespace-nowrap">Aktualisiert</TableHead>
                  <TableHead className="text-right whitespace-nowrap">Aktionen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((o) => {
                  const price = typeof o.totalCents === "number" ? formatChf(o.totalCents) : "–"
                  const created = new Date(o.createdAt).toLocaleString("de-CH")
                  const updated = o.updatedAt ? new Date(o.updatedAt).toLocaleString("de-CH") : "–"
                  const userLink = o.customerId ? (
                    <Link
                      href={`/admin/users/${encodeURIComponent(o.customerId)}`}
                      className="underline underline-offset-2 text-xs"
                    >
                      {o.customerId}
                    </Link>
                  ) : (
                    <span className="text-muted-foreground">–</span>
                  )
                  return (
                    <TableRow key={o.id} className="even:bg-card odd:bg-muted/40">
                      <TableCell className="text-xs">{o.id}</TableCell>
                      <TableCell className="uppercase">{o.status}</TableCell>
                      <TableCell>{price}</TableCell>
                      <TableCell>{o.contactEmail || "–"}</TableCell>
                      <TableCell>{userLink}</TableCell>
                      <TableCell className="text-xs">{o.paymentIntentId || "–"}</TableCell>
                      <TableCell className="text-xs">{o.stripeChargeId || "–"}</TableCell>
                      <TableCell className="text-xs">{o.paymentAttemptId || "–"}</TableCell>
                      <TableCell className="whitespace-nowrap">{created}</TableCell>
                      <TableCell className="whitespace-nowrap">{updated}</TableCell>
                      <TableCell className="text-right">
                        <Link href={`/admin/orders/${encodeURIComponent(o.id)}`}>
                          <Button size="sm" variant="outline">
                            Details
                          </Button>
                        </Link>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>
    </div>
  )
}
