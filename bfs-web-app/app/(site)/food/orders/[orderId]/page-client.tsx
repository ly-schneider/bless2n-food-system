"use client"

import { ArrowLeft } from "lucide-react"
import Image from "next/image"
import { useParams, useRouter, useSearchParams } from "next/navigation"
import { useEffect, useMemo, useRef, useState } from "react"
import QRCode from "@/components/qrcode"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"

import { getOrderPublicById, type OrderLineDTO, type PublicOrderDetailsDTO } from "@/lib/api/orders"
import { addOrder } from "@/lib/orders-storage"
import { formatChf } from "@/lib/utils"

export default function OrderPageClient() {
  const sp = useSearchParams()
  const { orderId } = useParams<{ orderId: string }>()
  const from = sp.get("from")
  const { clearCart } = useCart()
  const clearedRef = useRef(false)
  const footerRef = useRef<HTMLDivElement>(null)
  const [footerHeight, setFooterHeight] = useState(0)
  const [mounted, setMounted] = useState(false)
  const [serverOrder, setServerOrder] = useState<PublicOrderDetailsDTO | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [apiError, setApiError] = useState<string | null>(null)
  const [qrReady, setQrReady] = useState<boolean>(false)

  useEffect(() => {
    if (orderId) addOrder(orderId)
  }, [orderId])

  // Load order details from API (public) for any order id
  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!orderId) return
      setQrReady(false)
      setLoading(true)
      setApiError(null)
      try {
        const res = await getOrderPublicById(orderId)
        if (!cancelled) setServerOrder(res)
      } catch (e: unknown) {
        const msg =
          typeof e === "object" && e && "message" in e ? String((e as { message?: unknown }).message ?? "") : undefined
        if (!cancelled) setApiError(msg || "Fehler beim Laden der Bestellung")
      } finally {
        if (!cancelled) setQrReady(true)
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
  useEffect(() => {
    setMounted(true)
  }, [])
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

  const groupedItems = useMemo(() => {
    if (!serverOrder?.lines || serverOrder.lines.length === 0)
      return [] as Array<{ parent: OrderLineDTO; children: OrderLineDTO[] }>

    const byId: Record<string, OrderLineDTO> = {}
    for (const it of serverOrder.lines) byId[it.id] = it

    const childrenByRoot: Record<string, OrderLineDTO[]> = {}
    const roots: OrderLineDTO[] = []

    for (const it of serverOrder.lines) {
      const hasParent = typeof it.parentLineId === "string" && it.parentLineId.length > 0
      if (!hasParent) {
        roots.push(it)
        continue
      }
      let parentId = it.parentLineId as string
      while (parentId) {
        const node = byId[parentId]
        if (!node) break
        const p = node.parentLineId
        if (typeof p === "string" && p.length > 0) {
          parentId = p
        } else {
          break
        }
      }
      const arr = childrenByRoot[parentId] || []
      arr.push(it)
      childrenByRoot[parentId] = arr
    }

    return roots.map((root) => ({ parent: root, children: childrenByRoot[root.id] || [] }))
  }, [serverOrder])

  const qrDisplay = (() => {
    if (!mounted) {
      return <div className="mx-auto h-[260px] w-[260px] animate-pulse rounded-[11px] border-2 bg-gray-100" />
    }
    if (!orderId) {
      return <p className="text-red-600">Bestellnummer fehlt.</p>
    }
    if (qrReady) {
      return <QRCode value={orderId ?? ""} size={260} className="mx-auto rounded-[11px] border-2 p-1" />
    }
    return <div className="mx-auto h-[260px] w-[260px] animate-pulse rounded-[11px] border-2 bg-gray-100" />
  })()

  return (
    <div className="flex flex-col p-4" style={{ paddingBottom: footerHeight ? footerHeight + 16 : 16 }}>
      <h1 className="mb-2 text-2xl font-semibold">Dein Abhol-QR-Code</h1>
      <p className={`text-muted-foreground text-sm ${from === "success" ? "mb-2" : "mb-6"}`}>
        Zeigen Sie diesen QR-Code bei der Abholung vor.
      </p>
      {from === "success" && (
        <p className="text-muted-foreground mb-6 text-sm">
          Du kannst diesen QR-Code jederzeit in deinen Bestellungen finden.
        </p>
      )}
      {qrDisplay}

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
                        sizes="128px"
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
                        <p className="font-medium">{formatChf(parent.unitPriceCents * parent.quantity)}</p>
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

      <div ref={footerRef} className="fixed inset-x-0 bottom-0 z-50 mx-auto max-w-xl p-4">
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
