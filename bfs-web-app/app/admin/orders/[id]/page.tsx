"use client"
import Image from "next/image"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { API_BASE_URL } from "@/lib/api"
import { formatChf } from "@/lib/utils"

type OrderItem = {
  id: string
  orderId: string
  productId: string
  title: string
  quantity: number
  pricePerUnitCents: number
  parentItemId?: string | null
  menuSlotId?: string | null
  menuSlotName?: string | null
  productImage?: string | null
}

type AdminOrderDetails = {
  id: string
  status: string
  totalCents: number
  createdAt: string
  updatedAt: string
  contactEmail?: string | null
  customerId?: string | null
  paymentIntentId?: string | null
  stripeChargeId?: string | null
  paymentAttemptId?: string | null
  origin?: string | null
  posPayment?: {
    method: string
    amountReceivedCents?: number | null
    changeCents?: number | null
  } | null
  items: OrderItem[]
}

export default function AdminOrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const fetchAuth = useAuthorizedFetch()
  const [order, setOrder] = useState<AdminOrderDetails | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoading(true)
      setError(null)
      try {
        const res = await fetchAuth(`${API_BASE_URL}/v1/admin/orders/${encodeURIComponent(id)}`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { order: AdminOrderDetails }
        if (!cancelled) setOrder(data.order)
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Fehler beim Laden der Bestellung"
        if (!cancelled) setError(msg)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [fetchAuth, id])

  const grouped = useMemo(() => {
    if (!order?.items) return [] as Array<{ parent: OrderItem; children: OrderItem[] }>
    const byId: Record<string, OrderItem> = {}
    for (const it of order.items) byId[it.id] = it
    const childrenByRoot: Record<string, OrderItem[]> = {}
    const roots: OrderItem[] = []
    for (const it of order.items) {
      const hasParent = typeof it.parentItemId === "string" && !!it.parentItemId
      if (!hasParent) {
        roots.push(it)
        continue
      }
      let parentId = it.parentItemId as string
      while (parentId) {
        const node = byId[parentId]
        if (!node) break
        const p = node.parentItemId
        if (typeof p === "string" && p.length > 0) parentId = p
        else break
      }
      const arr = childrenByRoot[parentId] || []
      arr.push(it)
      childrenByRoot[parentId] = arr
    }
    return roots.map((root) => ({ parent: root, children: childrenByRoot[root.id] || [] }))
  }, [order])

  const created = order?.createdAt ? new Date(order.createdAt).toLocaleString("de-CH") : "–"
  const updated = order?.updatedAt ? new Date(order.updatedAt).toLocaleString("de-CH") : "–"
  const origin = order?.origin || "shop"
  const posMethod = order?.posPayment?.method
  const posReceived =
    typeof order?.posPayment?.amountReceivedCents === "number"
      ? formatChf(order!.posPayment!.amountReceivedCents!)
      : null
  const posChange =
    typeof order?.posPayment?.changeCents === "number" ? formatChf(order!.posPayment!.changeCents!) : null

  return (
    <div className="min-w-0 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Bestellung</h1>
        <Link href="/admin/orders">
          <Button variant="outline">Zurück zur Übersicht</Button>
        </Link>
      </div>

      {loading && <div className="text-muted-foreground text-sm">Lade Bestellung…</div>}
      {error && <div className="text-sm text-red-600">{error}</div>}

      {order && (
        <div className="grid gap-4 md:grid-cols-2">
          <div className="rounded-md border p-4">
            <h2 className="mb-3 text-base font-semibold">Details</h2>
            <div className="space-y-1 text-sm">
              <div>
                <span className="text-muted-foreground">ID:</span> <span className="text-xs">{order.id}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Status:</span> <span className="uppercase">{order.status}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Ursprung:</span> <span className="uppercase">{origin}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Summe:</span>{" "}
                <span className="font-medium">{formatChf(order.totalCents)}</span>
              </div>
              <div>
                <span className="text-muted-foreground">E‑Mail:</span> {order.contactEmail || "–"}
              </div>
              <div>
                <span className="text-muted-foreground">Benutzer:</span>{" "}
                {order.customerId ? (
                  <Link
                    href={`/admin/users/${encodeURIComponent(order.customerId)}`}
                    className="text-xs underline underline-offset-2"
                  >
                    {order.customerId}
                  </Link>
                ) : (
                  "–"
                )}
              </div>
              <div>
                <span className="text-muted-foreground">Erstellt:</span> {created}
              </div>
              <div>
                <span className="text-muted-foreground">Aktualisiert:</span> {updated}
              </div>
            </div>
          </div>

          <div className="rounded-md border p-4">
            <h2 className="mb-3 text-base font-semibold">Zahlung</h2>
            <div className="space-y-1 text-sm">
              <div>
                <span className="text-muted-foreground">POS‑Methode:</span>{" "}
                <span className="uppercase">{posMethod || "–"}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Erhalten:</span> <span>{posReceived || "–"}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Rückgeld:</span> <span>{posChange || "–"}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Payment Intent:</span>{" "}
                <span className="text-xs">{order.paymentIntentId || "–"}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Charge:</span>{" "}
                <span className="text-xs">{order.stripeChargeId || "–"}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Payment Attempt:</span>{" "}
                <span className="text-xs">{order.paymentAttemptId || "–"}</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {order && (
        <div>
          <h2 className="mb-3 text-lg font-semibold">Bestellte Artikel</h2>
          {grouped.length === 0 ? (
            <p className="text-muted-foreground text-sm">Keine Artikel gefunden.</p>
          ) : (
            <div className="flex flex-col gap-3">
              {grouped.map(({ parent, children }) => (
                <div key={parent.id} className="rounded-xl border p-3">
                  <div className="flex items-center gap-3">
                    {parent.productImage && (
                      <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
                        <Image
                          src={parent.productImage}
                          alt={"Produktbild von " + parent.title}
                          fill
                          sizes="64px"
                          quality={90}
                          className="h-full w-full rounded-[11px] object-cover"
                        />
                      </div>
                    )}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0">
                          <p className="truncate font-medium">{parent.title}</p>
                          {children.length > 0 && (
                            <div className="mt-1 flex flex-row flex-wrap gap-1.5">
                              {children.map((c) => (
                                <span
                                  key={c.id}
                                  className="text-muted-foreground border-border rounded-lg border px-2 py-0.5 text-xs"
                                >
                                  {c.menuSlotName ?? "Option"}: {c.title}
                                </span>
                              ))}
                            </div>
                          )}
                        </div>
                        <div className="shrink-0 text-right">
                          <p className="text-muted-foreground text-sm">x{parent.quantity}</p>
                          <p className="font-medium">{formatChf(parent.pricePerUnitCents * parent.quantity)}</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
          <div className="mt-4 flex items-center justify-between">
            <span className="text-muted-foreground text-sm">Summe</span>
            <span className="text-base font-semibold">{formatChf(order.totalCents)}</span>
          </div>
        </div>
      )}
    </div>
  )
}
