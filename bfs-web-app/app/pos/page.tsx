"use client"

import "./bridge.client"

import { Banknote, CreditCard, QrCode } from "lucide-react"
import { useCallback, useEffect, useRef, useState } from "react"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { PairingCodeDisplay } from "@/components/device/pairing-code-display"
import { BasketPanel } from "@/components/pos/basket-panel"
import { POSHeader } from "@/components/pos/pos-header"
import { ProductGrid } from "@/components/pos/product-grid"
import { StuckOrdersDialog } from "@/components/pos/stuck-orders-dialog"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { CartProvider } from "@/contexts/cart-context"
import { useInventoryStream } from "@/hooks/use-inventory-stream"
import { useOrderQueue } from "@/hooks/use-order-queue"
import { listProducts } from "@/lib/api/products"
import { getDeviceToken } from "@/lib/device-auth"
import { formatChf } from "@/lib/utils"
import type { ListResponse, ProductDTO } from "@/types"
import type { PosFulfillmentMode } from "@/types/jeton"
import type { PosPaymentMethod } from "@/types/order-queue"

type PosStatus = {
  exists: boolean
  approved: boolean
  name?: string
  cardCapable?: boolean | null
  mode?: PosFulfillmentMode
}

type PosDeviceResponse = {
  id: string
  name: string
  status: string
  model?: string
  os?: string
  settings?: { posMode?: PosFulfillmentMode }
}

function PosInner() {
  const [sessionToken] = useState<string | null>(() => getDeviceToken())
  const [status, setStatus] = useState<PosStatus | null>(null)
  const [systemDisabled, setSystemDisabled] = useState(false)
  const [products, setProducts] = useState<ListResponse<ProductDTO>>({ items: [], count: 0 })
  const [configOpen, setConfigOpen] = useState(false)
  const [configProduct, setConfigProduct] = useState<ProductDTO | null>(null)
  const [showStuckOrders, setShowStuckOrders] = useState(false)

  const deviceId = sessionToken ? sessionToken.slice(0, 16) : "unknown"
  const [paymentErrorOrder, setPaymentErrorOrder] = useState<{
    order: import("@/types/order-queue").QueuedOrder
    errorCode: string
    errorMessage: string
  } | null>(null)

  const handlePaymentError = useCallback(
    (order: import("@/types/order-queue").QueuedOrder, errorCode: string, errorMessage: string) => {
      setPaymentErrorOrder({ order, errorCode, errorMessage })
    },
    []
  )

  const { isOnline, pendingCount, failedOrders, submitOrder, retryOrder, deleteOrder, changePaymentMethod } =
    useOrderQueue({
      token: sessionToken || "",
      deviceId,
      onPaymentError: handlePaymentError,
    })

  const handleFailedClick = useCallback(() => setShowStuckOrders(true), [])

  // Close product configuration on POS lock
  useEffect(() => {
    const onLock = () => {
      setConfigOpen(false)
      setConfigProduct(null)
    }
    window.addEventListener("pos:lock", onLock)
    return () => window.removeEventListener("pos:lock", onLock)
  }, [])

  // Fetch POS status using the session token (if we have one)
  useEffect(() => {
    if (!sessionToken) {
      setStatus({ exists: false, approved: false })
      return
    }
    ;(async () => {
      try {
        const res = await fetch(`/api/v1/pos/me`, {
          headers: { Authorization: `Bearer ${sessionToken}` },
        })
        if (res.status === 503) {
          setSystemDisabled(true)
          return
        }
        if (res.status === 401 || res.status === 403) {
          setStatus({ exists: false, approved: false })
          return
        }
        setSystemDisabled(false)
        const device = (await res.json()) as PosDeviceResponse
        setStatus({
          exists: true,
          approved: device.status === "approved",
          name: device.name,
          cardCapable: null,
          mode: device.settings?.posMode,
        })
      } catch {
        setStatus({ exists: false, approved: false })
      }
    })()
  }, [sessionToken])

  // Store latest stock snapshot from SSE to handle race condition with product loading
  const stockSnapshotRef = useRef<Map<string, number>>(new Map())

  // Helper to apply stock to products (including menu slot options)
  const applyStockToProducts = useCallback((items: ProductDTO[], stockMap: Map<string, number>): ProductDTO[] => {
    return items.map((p) => {
      // Update menu slot options with stock data
      if (p.type === "menu" && p.menu?.slots) {
        return {
          ...p,
          menu: {
            ...p.menu,
            slots: p.menu.slots.map((slot) => ({
              ...slot,
              options:
                slot.options?.map((opt) => {
                  const optStock = stockMap.get(opt.id)
                  if (optStock === undefined) return opt
                  return {
                    ...opt,
                    availableQuantity: optStock,
                    isAvailable: optStock > 0,
                    isLowStock: optStock > 0 && optStock <= 5,
                  }
                }) ?? null,
            })),
          },
        }
      }
      // Update simple product stock
      const newStock = stockMap.get(p.id)
      if (newStock === undefined) return p
      return {
        ...p,
        availableQuantity: newStock,
        isAvailable: newStock > 0,
        isLowStock: newStock > 0 && newStock <= 5,
      }
    })
  }, [])

  // Load products and apply any cached stock snapshot
  useEffect(() => {
    ;(async () => {
      try {
        const loaded = await listProducts()
        const itemsWithStock = applyStockToProducts(loaded.items, stockSnapshotRef.current)
        setProducts({ ...loaded, items: itemsWithStock })
      } catch {}
    })()
  }, [applyStockToProducts])

  // Handle SSE stock updates - store in ref and apply to current products
  const handleStockUpdate = useCallback(
    (stockMap: Map<string, number>) => {
      // Merge into our cached snapshot
      stockMap.forEach((value, key) => {
        stockSnapshotRef.current.set(key, value)
      })
      // Apply to current products
      setProducts((prev) => ({
        ...prev,
        items: applyStockToProducts(prev.items, stockMap),
      }))
    },
    [applyStockToProducts]
  )

  useInventoryStream({ onUpdate: handleStockUpdate, enabled: status?.approved ?? false })

  // Listen for local optimistic decrements
  useEffect(() => {
    const onDecrement = (e: Event) => {
      const ce = e as CustomEvent<Map<string, number>>
      // Apply deltas to cached snapshot
      ce.detail.forEach((delta, productId) => {
        const current = stockSnapshotRef.current.get(productId) ?? 0
        stockSnapshotRef.current.set(productId, Math.max(0, current + delta))
      })
      // Apply to products
      setProducts((prev) => ({
        ...prev,
        items: prev.items.map((p) => {
          const delta = ce.detail.get(p.id)
          if (!delta) return p
          const newQty = Math.max(0, (p.availableQuantity ?? 0) + delta)
          return {
            ...p,
            availableQuantity: newQty,
            isAvailable: newQty > 0,
            isLowStock: newQty > 0 && newQty <= 5,
          }
        }),
      }))
    }
    window.addEventListener("pos:inventory-decrement", onDecrement as EventListener)
    return () => window.removeEventListener("pos:inventory-decrement", onDecrement as EventListener)
  }, [])

  // Support AdminMainHeader "Aktualisieren" button to refresh POS data
  useEffect(() => {
    const onRefresh = () => {
      ;(async () => {
        try {
          const loaded = await listProducts()
          const itemsWithStock = applyStockToProducts(loaded.items, stockSnapshotRef.current)
          setProducts({ ...loaded, items: itemsWithStock })
        } catch {}
        if (sessionToken) {
          try {
            const res = await fetch(`/api/v1/pos/me`, {
              headers: { Authorization: `Bearer ${sessionToken}` },
            })
            const device = (await res.json()) as PosDeviceResponse
            setStatus({
              exists: true,
              approved: device.status === "approved",
              name: device.name,
              cardCapable: null,
              mode: device.settings?.posMode,
            })
          } catch {}
        }
      })()
    }
    window.addEventListener("admin:refresh", onRefresh)
    return () => window.removeEventListener("admin:refresh", onRefresh)
  }, [sessionToken, applyStockToProducts])

  useEffect(() => {
    if (!systemDisabled) return
    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/v1/system/status`)
        if (!res.ok) return
        const data = (await res.json()) as { enabled: boolean }
        if (data.enabled) {
          setSystemDisabled(false)
          window.location.reload()
        }
      } catch {}
    }, 30_000)
    return () => clearInterval(interval)
  }, [systemDisabled])

  const syncStatus = {
    isOnline,
    pendingCount,
    failedCount: failedOrders.length,
    onFailedClick: handleFailedClick,
  }

  if (systemDisabled) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-semibold">System geschlossen</h2>
          <p className="text-muted-foreground mt-2">Das System ist momentan nicht verfügbar.</p>
        </div>
      </div>
    )
  }

  if (!status?.approved) {
    return (
      <>
        <POSHeader mode={status?.mode} syncStatus={syncStatus} />
        <PairingCodeDisplay deviceType="POS" />
      </>
    )
  }

  return (
    <>
      <POSHeader mode={status?.mode} syncStatus={syncStatus} />
      <div className="grid h-[calc(100dvh-4rem)] grid-cols-1 overflow-hidden md:grid-cols-[1fr_450px]">
        <div className="min-h-0 overflow-hidden">
          <ProductGrid
            products={products}
            onConfigure={(p) => {
              setConfigProduct(p)
              setConfigOpen(true)
            }}
          />
        </div>
        <BasketPanel
          token={sessionToken || ""}
          mode={status?.mode}
          submitOrder={submitOrder}
          stockMap={stockSnapshotRef.current}
        />

        {configProduct && (
          <ProductConfigurationModal
            product={configProduct}
            isOpen={configOpen}
            onClose={() => {
              setConfigOpen(false)
              setConfigProduct(null)
            }}
          />
        )}
      </div>

      <StuckOrdersDialog
        open={showStuckOrders}
        onOpenChange={setShowStuckOrders}
        orders={failedOrders}
        onRetry={retryOrder}
        onDelete={deleteOrder}
      />

      {/* Payment error dialog - shown when Club100 payment fails due to ineligible products */}
      <Dialog
        open={!!paymentErrorOrder}
        onOpenChange={(open) => {
          if (!open) setPaymentErrorOrder(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Zahlung fehlgeschlagen</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm">
              {paymentErrorOrder?.errorCode === "product_not_free"
                ? "Die Produkte in dieser Bestellung sind nicht für 100 Club Gratis-Einlösung konfiguriert. Bitte wähle eine andere Zahlungsmethode."
                : paymentErrorOrder?.errorMessage || "Zahlung konnte nicht abgeschlossen werden."}
            </p>
            {paymentErrorOrder && (
              <div className="bg-muted rounded-lg p-3">
                <div className="text-muted-foreground mb-1 text-xs">Bestellsumme</div>
                <div className="font-semibold">{formatChf(paymentErrorOrder.order.totalCents)}</div>
              </div>
            )}
            <div className="grid grid-cols-3 gap-3">
              <Button
                className="flex h-24 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => {
                  if (paymentErrorOrder) {
                    changePaymentMethod(paymentErrorOrder.order.localId, "cash" as PosPaymentMethod)
                    setPaymentErrorOrder(null)
                  }
                }}
              >
                <Banknote className="size-8" />
                <span className="text-sm font-medium">Bar</span>
              </Button>
              <Button
                className="flex h-24 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => {
                  if (paymentErrorOrder) {
                    changePaymentMethod(paymentErrorOrder.order.localId, "card" as PosPaymentMethod)
                    setPaymentErrorOrder(null)
                  }
                }}
              >
                <CreditCard className="size-8" />
                <span className="text-sm font-medium">Karte</span>
              </Button>
              <Button
                className="flex h-24 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => {
                  if (paymentErrorOrder) {
                    changePaymentMethod(paymentErrorOrder.order.localId, "twint" as PosPaymentMethod)
                    setPaymentErrorOrder(null)
                  }
                }}
              >
                <QrCode className="size-8" />
                <span className="text-sm font-medium">TWINT</span>
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="ghost"
              onClick={() => {
                if (paymentErrorOrder) {
                  deleteOrder(paymentErrorOrder.order.localId)
                  setPaymentErrorOrder(null)
                }
              }}
            >
              Bestellung löschen
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

export default function POSPage() {
  return (
    <CartProvider>
      <PosInner />
    </CartProvider>
  )
}
