"use client"

import { ArrowLeft } from "lucide-react"
import Image from "next/image"
import { useParams, useRouter, useSearchParams } from "next/navigation"
import { useEffect, useMemo, useRef, useState } from "react"
import QRCode from "@/components/qrcode"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { API_BASE_URL } from "@/lib/api"
import { getOrderPublicById, type PublicOrderDetailsDTO } from "@/lib/api/orders"
import { addOrder, getOrder } from "@/lib/orders-storage"
import { formatChf } from "@/lib/utils"

export default function CheckoutQRPage() {
  const sp = useSearchParams()
  const { orderId } = useParams<{ orderId: string }>()
  const from = sp.get("from")
  const { clearCart } = useCart()
  const clearedRef = useRef(false)
  const footerRef = useRef<HTMLDivElement>(null)
  const [footerHeight, setFooterHeight] = useState(0)
  const [mounted, setMounted] = useState(false)
  const [serverOrder, setServerOrder] = useState<PublicOrderDetailsDTO | null>(null)
  const [pickupCode, setPickupCode] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [apiError, setApiError] = useState<string | null>(null)

  useEffect(() => {
    if (orderId) addOrder(orderId)
  }, [orderId])

  // Load order details from API (public) for any order id
  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!orderId) return
      setLoading(true)
      setApiError(null)
      try {
        const res = await getOrderPublicById(orderId)
        if (!cancelled) setServerOrder(res)
        // Fetch signed pickup QR payload for stations
        const q = await fetch(`${API_BASE_URL}/v1/orders/${encodeURIComponent(orderId)}/pickup-qr`)
        if (q.ok) {
          const data = (await q.json()) as { code?: string }
          if (!cancelled) setPickupCode(data.code || null)
        }
      } catch (e: unknown) {
        const msg = typeof e === 'object' && e && 'message' in e ? String((e as { message?: unknown }).message ?? '') : undefined
        if (!cancelled) setApiError(msg || "Fehler beim Laden der Bestellung")
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [orderId])

  // Ensure cart is cleared when landing on QR page as a fallback
  useEffect(() => {
    if (!clearedRef.current) {
      clearedRef.current = true
      clearCart()
    }
  }, [clearCart])
  // Avoid hydration mismatches by rendering dynamic content only after mount
  useEffect(() => { setMounted(true) }, [])
  const router = useRouter()

  // Measure footer height to ensure content (e.g., Summe) is not obscured
  useEffect(() => {
    const el = footerRef.current
    if (!el) return
    const update = () => setFooterHeight(el.offsetHeight || 0)
    update()
    let ro: ResizeObserver | null = null
    if (typeof ResizeObserver !== "undefined") {
      ro = new ResizeObserver(() => update())
      ro.observe(el)
    }
    const onResize = () => update()
    window.addEventListener("resize", onResize)
    return () => {
      window.removeEventListener("resize", onResize)
      if (ro) ro.disconnect()
    }
  }, [from, mounted])

  const order = orderId ? getOrder(orderId) : undefined

  const groupedItems = useMemo(() => {
    if (!serverOrder?.items)
      return [] as Array<{ parent: PublicOrderDetailsDTO["items"][number]; children: PublicOrderDetailsDTO["items"] }>

    // Build quick lookup maps
    const byId: Record<string, PublicOrderDetailsDTO["items"][number]> = {}
    for (const it of serverOrder.items) byId[it.id] = it

    // Group children by their ultimate root parent (walk up until no parent)
    const childrenByRoot: Record<string, PublicOrderDetailsDTO["items"]> = {}
    const roots: Array<PublicOrderDetailsDTO["items"][number]> = []

    for (const it of serverOrder.items) {
      const hasParent = typeof it.parentItemId === 'string' && it.parentItemId.length > 0
      if (!hasParent) {
        roots.push(it)
        continue
      }
      // follow chain to root
      let parentId = it.parentItemId as string
      while (parentId) {
        const node = byId[parentId]
        if (!node) break
        const p = node.parentItemId
        if (typeof p === 'string' && p.length > 0) {
          parentId = p
        } else {
          break
        }
      }
      const arr = childrenByRoot[parentId] || []
      arr.push(it)
      childrenByRoot[parentId] = arr
    }

    // Preserve original order of roots
    return roots.map((root) => ({ parent: root, children: childrenByRoot[root.id] || [] }))
  }, [serverOrder])

  return (
    <div
      className="flex flex-col p-4"
      style={{ paddingBottom: footerHeight ? footerHeight + 16 : 16 }}
    >
      <h1 className="mb-2 text-2xl font-semibold">Dein Abhol-QR-Code</h1>
      <p className={`text-muted-foreground text-sm ${from === "success" ? "mb-2" : "mb-6"}`}>
        Zeigen Sie diesen QR-Code bei der Abholung vor.
      </p>
      {from === "success" && (
        <p className="text-muted-foreground mb-6 text-sm">
          Du kannst diesen QR-Code jederzeit in deinen Bestellungen finden.
        </p>
      )}

      {mounted ? (
        orderId ? (
          <QRCode value={pickupCode ?? orderId ?? ''} size={260} className="mx-auto rounded-[11px] border-2 p-1" />
        ) : (
          <p className="text-red-600">Bestellnummer fehlt.</p>
        )
      ) : (
        <div className="mx-auto h-[260px] w-[260px] animate-pulse rounded-[11px] border-2 bg-gray-100" />
      )}

      {/* Order items summary */}
      {mounted && serverOrder && groupedItems.length > 0 ? (
        <div className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Bestellte Artikel</h2>
          <div className="flex flex-col gap-3">
            {groupedItems.map(({ parent, children }) => (
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
          {typeof serverOrder.totalCents === "number" && (
            <div className="mt-4 flex items-center justify-between">
              <span className="text-muted-foreground text-sm">Summe</span>
              <span className="text-base font-semibold">{formatChf(serverOrder.totalCents)}</span>
            </div>
          )}
        </div>
      ) : mounted && order && order.items && order.items.length > 0 ? (
        <div className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Bestellte Artikel</h2>
          <div className="flex flex-col gap-3">
            {order.items.map((item) => {
              const isMenuProduct = item.product.type === "menu"
              const hasConfiguration = item.configuration && Object.keys(item.configuration).length > 0
              return (
                <div key={item.id} className="rounded-xl border p-3">
                  <div className="flex items-center gap-3">
                    {item.product.image && (
                      <div className="relative h-20 w-20 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
                        <Image
                          src={item.product.image}
                          alt={"Produktbild von " + item.product.name}
                          fill
                          sizes="80px"
                          quality={90}
                          className="h-full w-full rounded-[11px] object-cover"
                        />
                      </div>
                    )}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0">
                          <p className="truncate font-medium">{item.product.name}</p>
                          {isMenuProduct && hasConfiguration && (
                            <div className="mt-1 flex flex-row flex-wrap gap-1.5">
                              {Object.entries(item.configuration || {}).map(([slotId, productId]) => {
                                const slot = item.product.menu?.slots?.find((s) => s.id === slotId)
                                const slotItem = slot?.menuSlotItems?.find((si) => si.id === productId)
                                if (slot && slotItem) {
                                  return (
                                    <span
                                      key={slotId}
                                      className="text-muted-foreground border-border rounded-lg border px-2 py-0.5 text-xs"
                                    >
                                      {slot.name}: {slotItem.name}
                                    </span>
                                  )
                                }
                                return null
                              })}
                            </div>
                          )}
                        </div>
                        <div className="shrink-0 text-right">
                          <p className="text-muted-foreground text-sm">x{item.quantity}</p>
                          <p className="font-medium">{formatChf(item.totalPriceCents)}</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
          {typeof order.totalCents === "number" && (
            <div className="mt-4 flex items-center justify-between">
              <span className="text-muted-foreground text-sm">Summe</span>
              <span className="text-base font-semibold">{formatChf(order.totalCents)}</span>
            </div>
          )}
        </div>
      ) : mounted && orderId ? (
        <div className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Bestellte Artikel</h2>
          {loading ? (
            <p className="text-muted-foreground text-sm">Lade Bestellung…</p>
          ) : apiError ? (
            <p className="text-sm text-red-600">{apiError}</p>
          ) : (
            <p className="text-muted-foreground text-sm">
              Keine Artikelliste gespeichert. Der QR-Code identifiziert die Bestellung beim Abholen.
            </p>
          )}
        </div>
      ) : null}

      <div ref={footerRef} className="max-w-xl mx-auto fixed inset-x-0 bottom-0 z-50 p-4">
        <div className="flex flex-col gap-2">
          <Button
            variant="outline"
            className="rounded-pill h-12 w-full bg-white text-base font-medium text-black"
            onClick={() => router.push(from === "orders" ? "/food/orders" : "/")}
          >
            <ArrowLeft style={{ width: 20, height: 20 }} /> {from === "orders" ? "Zurück" : "Zum Menü"}
          </Button>

          {from === "success" && (
            <Button
              variant="selected"
              className="rounded-pill h-12 w-full text-base"
              onClick={() => router.push("/food/orders")}
            >
              Alle Bestellungen
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
