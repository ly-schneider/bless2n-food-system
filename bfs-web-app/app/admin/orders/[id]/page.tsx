"use client"
import Image from "next/image"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useEffect, useMemo, useState } from "react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { getCSRFToken } from "@/lib/csrf"

import { formatChf } from "@/lib/utils"

type OrderLineRedemption = {
  id: string
  orderLineId: string
  redeemedAt: string
}

type OrderItem = {
  id: string
  orderId: string
  productId: string
  title: string
  quantity: number
  unitPriceCents: number
  pricePerUnitCents?: number
  lineType?: string | null
  parentLineId?: string | null
  menuSlotId?: string | null
  menuSlotName?: string | null
  productImage?: string | null
  redemption?: OrderLineRedemption | null
  childLines?: OrderItem[] | null
}

type OrderPaymentSummary = {
  id?: string
  method?: string
  amountCents?: number
  paidAt?: string
}

type AdminOrderDetails = {
  id: string
  status: string
  totalCents: number
  createdAt: string
  updatedAt: string
  contactEmail?: string | null
  customerId?: string | null
  paymentAttemptId?: string | null
  payrexxGatewayId?: number | null
  payrexxTransactionId?: number | null
  origin?: string | null
  payments?: OrderPaymentSummary[] | null
  lines?: OrderItem[] | null
  payment?: {
    method: string
    amountCents?: number | null
  } | null
  items?: OrderItem[]
}

const STATUS_LABELS: Record<string, string> = {
  pending: "Ausstehend",
  paid: "Bezahlt",
  cancelled: "Storniert",
  refunded: "Erstattet",
}

const STATUS_COLORS: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  paid: "bg-green-100 text-green-800",
  cancelled: "bg-red-100 text-red-800",
  refunded: "bg-blue-100 text-blue-800",
}

const STATUS_TRANSITIONS: Record<string, string[]> = {
  pending: ["paid", "cancelled"],
  paid: ["cancelled", "refunded"],
  cancelled: [],
  refunded: [],
}

export default function AdminOrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const fetchAuth = useAuthorizedFetch()
  const [order, setOrder] = useState<AdminOrderDetails | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [statusUpdating, setStatusUpdating] = useState(false)
  const [confirmDialog, setConfirmDialog] = useState<{ open: boolean; status: string }>({
    open: false,
    status: "",
  })

  async function loadOrder() {
    setLoading(true)
    setError(null)
    try {
      const res = await fetchAuth(`/api/v1/orders/${encodeURIComponent(id)}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = (await res.json()) as Record<string, unknown>
      // Support both wrapped ({ order: ... }) and unwrapped responses
      setOrder((data.order ?? data) as AdminOrderDetails)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Fehler beim Laden der Bestellung"
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let cancelled = false
    void loadOrder().then(() => {
      if (cancelled) setOrder(null)
    })
    return () => {
      cancelled = true
    }
     
  }, [fetchAuth, id])

  async function handleStatusChange(newStatus: string) {
    setStatusUpdating(true)
    setError(null)
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/orders/${encodeURIComponent(id)}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({ status: newStatus }),
      })
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as Record<string, string>
        throw new Error(body.message || `HTTP ${res.status}`)
      }
      await loadOrder()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Statusänderung fehlgeschlagen"
      setError(msg)
    } finally {
      setStatusUpdating(false)
      setConfirmDialog({ open: false, status: "" })
    }
  }

  // Normalize items: support both new `lines` field and legacy `items`
  const orderItems = order?.lines ?? order?.items ?? []

  const grouped = useMemo(() => {
    if (!orderItems.length) return [] as Array<{ parent: OrderItem; children: OrderItem[] }>
    const byId: Record<string, OrderItem> = {}
    for (const it of orderItems) byId[it.id] = it

    // If the API returned childLines embedded, use those directly
    const hasChildLines = orderItems.some((it) => it.childLines && it.childLines.length > 0)
    if (hasChildLines) {
      return orderItems
        .filter((it) => !it.parentLineId)
        .map((root) => ({ parent: root, children: root.childLines ?? [] }))
    }

    // Legacy flat-list grouping
    const childrenByRoot: Record<string, OrderItem[]> = {}
    const roots: OrderItem[] = []
    for (const it of orderItems) {
      const parentId = it.parentLineId ?? ((it as Record<string, unknown>).parentItemId as string | null | undefined)
      const hasParent = typeof parentId === "string" && !!parentId
      if (!hasParent) {
        roots.push(it)
        continue
      }
      let rootId = parentId as string
      while (rootId) {
        const node = byId[rootId]
        if (!node) break
        const p = node.parentLineId ?? ((node as Record<string, unknown>).parentItemId as string | null | undefined)
        if (typeof p === "string" && p.length > 0) rootId = p
        else break
      }
      const arr = childrenByRoot[rootId] || []
      arr.push(it)
      childrenByRoot[rootId] = arr
    }
    return roots.map((root) => ({ parent: root, children: childrenByRoot[root.id] || [] }))
  }, [orderItems])

  const created = order?.createdAt ? new Date(order.createdAt).toLocaleString("de-CH") : "–"
  const updated = order?.updatedAt ? new Date(order.updatedAt).toLocaleString("de-CH") : "–"
  const origin = order?.origin || "shop"

  // Payment info: prefer new payments array, fall back to legacy payment object
  const primaryPayment = order?.payments?.[0]
  const posMethod = primaryPayment?.method ?? order?.payment?.method

  const nextStatuses = order ? (STATUS_TRANSITIONS[order.status] ?? []) : []

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
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground">Status:</span>
                <span
                  className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[order.status] || "bg-gray-100 text-gray-800"}`}
                >
                  {STATUS_LABELS[order.status] || order.status}
                </span>
              </div>
              {nextStatuses.length > 0 && (
                <div className="flex flex-wrap gap-2 pt-1">
                  {nextStatuses.map((s) => (
                    <Button
                      key={s}
                      size="sm"
                      variant={s === "cancelled" || s === "refunded" ? "destructive" : "default"}
                      disabled={statusUpdating}
                      onClick={() => setConfirmDialog({ open: true, status: s })}
                    >
                      {STATUS_LABELS[s] || s}
                    </Button>
                  ))}
                </div>
              )}
              <div>
                <span className="text-muted-foreground">Ursprung:</span> <span className="uppercase">{origin}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Summe:</span>{" "}
                <span className="font-medium">{formatChf(order.totalCents)}</span>
              </div>
              <div>
                <span className="text-muted-foreground">E-Mail:</span> {order.contactEmail || "–"}
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
                <span className="text-muted-foreground">POS-Methode:</span>{" "}
                <span className="uppercase">{posMethod || "–"}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Payrexx Gateway:</span>{" "}
                <span className="text-xs">{order.payrexxGatewayId ?? "–"}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Payrexx Txn:</span>{" "}
                <span className="text-xs">{order.payrexxTransactionId ?? "–"}</span>
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
              {grouped.map(({ parent, children }) => {
                const unitPrice = parent.unitPriceCents ?? parent.pricePerUnitCents ?? 0
                const isBundle = parent.lineType === "bundle" || children.length > 0
                return (
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
                            unoptimized={parent.productImage.includes("localhost:10000")}
                            className="h-full w-full rounded-[11px] object-cover"
                          />
                        </div>
                      )}
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center justify-between gap-3">
                          <div className="min-w-0">
                            <div className="flex items-center gap-2">
                              <p className="truncate font-medium">{parent.title}</p>
                              {!isBundle && <RedemptionBadge redemption={parent.redemption} />}
                            </div>
                          </div>
                          <div className="shrink-0 text-right">
                            <p className="text-muted-foreground text-sm">x{parent.quantity}</p>
                            <p className="font-medium">{formatChf(unitPrice * parent.quantity)}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                    {children.length > 0 && (
                      <div className="mt-2 space-y-1">
                        {children.map((c) => (
                          <div key={c.id} className="flex items-center justify-between gap-2 pl-6 text-sm">
                            <div className="flex min-w-0 items-center gap-2">
                              <span className="truncate">
                                {c.menuSlotName ? `${c.menuSlotName}: ` : ""}
                                {c.title}
                              </span>
                              <RedemptionBadge redemption={c.redemption} />
                            </div>
                            <span className="text-muted-foreground shrink-0 text-xs">x{c.quantity}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
          <div className="mt-4 flex items-center justify-between">
            <span className="text-muted-foreground text-sm">Summe</span>
            <span className="text-base font-semibold">{formatChf(order.totalCents)}</span>
          </div>
        </div>
      )}

      <AlertDialog
        open={confirmDialog.open}
        onOpenChange={(open) => {
          if (!open) setConfirmDialog({ open: false, status: "" })
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Status ändern</AlertDialogTitle>
            <AlertDialogDescription>
              Bestellung wirklich auf &quot;{STATUS_LABELS[confirmDialog.status] || confirmDialog.status}&quot; setzen?
              Diese Aktion kann nicht rückgängig gemacht werden.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={statusUpdating}>Abbrechen</AlertDialogCancel>
            <AlertDialogAction disabled={statusUpdating} onClick={() => void handleStatusChange(confirmDialog.status)}>
              {statusUpdating ? "Wird geändert…" : "Bestätigen"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function RedemptionBadge({ redemption }: { redemption?: OrderLineRedemption | null }) {
  if (!redemption) {
    return <span className="inline-flex rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-600">Offen</span>
  }
  const at = new Date(redemption.redeemedAt).toLocaleString("de-CH")
  return (
    <span className="inline-flex rounded-full bg-green-100 px-2 py-0.5 text-xs text-green-700" title={at}>
      Eingelöst
    </span>
  )
}
