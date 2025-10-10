"use client"

import { Banknote, Check, CreditCard, Printer, ShoppingCart } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { CartItemDisplay } from "@/components/cart/cart-item-display"
import { InlineMenuGroup } from "@/components/cart/inline-menu-group"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { useBestMenuSuggestion } from "@/components/cart/use-best-menu-suggestion"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useCart } from "@/contexts/cart-context"
import { API_BASE_URL } from "@/lib/api"
import { formatChf } from "@/lib/utils"

type Tender = "cash" | "card" | null
type Receipt = {
  method: "cash" | "card"
  totalCents: number
  orderId?: string
  amountReceivedCents?: number
  changeCents?: number
}

type PosBridge = {
  print?: (s: string) => void
  payWithCard?: (p: { amountCents: number; currency: string; reference: string }) => void
}

const getBridge = () => (globalThis as unknown as { PosBridge?: PosBridge }).PosBridge

export function BasketPanel({ token }: { token: string }) {
  const { cart, updateQuantity, removeFromCart, clearCart } = useCart()
  const { suggestion, contiguous, startIndex, endIndex } = useBestMenuSuggestion()
  const [showCheckout, setShowCheckout] = useState(false)
  const [tender, setTender] = useState<Tender>(null)
  const [received, setReceived] = useState<string>("")
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editingItem, setEditingItem] = useState<import("@/types/cart").CartItem | null>(null)
  const [showReceipt, setShowReceipt] = useState(false)
  const [receipt, setReceipt] = useState<Receipt | null>(null)
  const [printing, setPrinting] = useState(false)
  const [printed, setPrinted] = useState(false)
  const [printError, setPrintError] = useState<string | null>(null)
  const [hasPosBridge, setHasPosBridge] = useState(false)
  const [canPrint, setCanPrint] = useState(false)
  const [canPayWithCard, setCanPayWithCard] = useState(false)
  const total = cart.totalCents
  const receivedCents = useMemo(() => Math.round((parseFloat(received || "0") || 0) * 100), [received])
  const changeCents = Math.max(0, receivedCents - total)
  const cartIsEmpty = cart.items.length === 0

  useEffect(() => {
    const onLock = () => {
      setShowCheckout(false)
      setShowReceipt(false)
      setTender(null)
      setEditingItem(null)
    }
    window.addEventListener("pos:lock", onLock)
    return () => window.removeEventListener("pos:lock", onLock)
  }, [])

  useEffect(() => {
    try {
      type Bridge = {
        print?: (s: string) => void
        payWithCard?: (p: { amountCents: number; currency: string; reference: string }) => void
      }
      const g = globalThis as unknown as { PosBridge?: Bridge }
      const bridge = g.PosBridge
      setHasPosBridge(!!bridge)
      setCanPrint(!!(bridge && typeof bridge.print === "function"))
      setCanPayWithCard(!!(bridge && typeof bridge.payWithCard === "function"))
    } catch {
      setHasPosBridge(false)
      setCanPrint(false)
      setCanPayWithCard(false)
    }
  }, [])

  // Handlers & helpers
  const resetPrintState = useCallback(() => {
    setPrinting(false)
    setPrinted(false)
    setPrintError(null)
  }, [])

  const openReceipt = useCallback(
    (next: Receipt) => {
      setReceipt(next)
      resetPrintState()
      try {
        clearCart()
      } catch {}
      setShowCheckout(false)
      setShowReceipt(true)
    },
    [clearCart, resetPrintState]
  )

  const closeReceipt = useCallback(() => {
    setShowReceipt(false)
    setReceipt(null)
    resetPrintState()
  }, [resetPrintState])

  const closeCheckout = useCallback(() => {
    setShowCheckout(false)
    setTender(null)
    setReceived("")
    setBusy(false)
    setError(null)
  }, [])

  const startCashPayment = useCallback(async () => {
    if (busy) return
    setBusy(true)
    setError(null)
    try {
      const items = cart.items.map((it) => ({
        productId: it.product.id,
        quantity: it.quantity,
        configuration: it.configuration || undefined,
      }))
      const resOrder = await fetch(`${API_BASE_URL}/v1/pos/orders`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Pos-Token": token },
        body: JSON.stringify({ items }),
      })
      type ApiError = { detail?: string }
      type OrderOk = { orderId: string }
      const orderJson = (await resOrder.json()) as OrderOk & ApiError
      if (!resOrder.ok) throw new Error(orderJson.detail || "order failed")
      const orderId = orderJson.orderId

      const resPay = await fetch(`${API_BASE_URL}/v1/pos/orders/${orderId}/pay-cash`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Pos-Token": token },
        body: JSON.stringify({ amountReceivedCents: receivedCents }),
      })
      type PayOk = { changeCents?: number }
      const payJson = (await resPay.json()) as PayOk & ApiError
      if (!resPay.ok) throw new Error(payJson.detail || "payment failed")

      openReceipt({
        method: "cash",
        orderId,
        totalCents: total,
        amountReceivedCents: receivedCents,
        changeCents: payJson.changeCents,
      })
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Bezahlen fehlgeschlagen"
      setError(msg)
    } finally {
      setBusy(false)
    }
  }, [busy, cart.items, token, receivedCents, openReceipt, total])

  const startCardPayment = useCallback(() => {
    const bridge = getBridge()
    if (bridge && typeof bridge.payWithCard === "function") {
      bridge.payWithCard({ amountCents: total, currency: "CHF", reference: `pos_${Date.now()}` })
    }
    openReceipt({ method: "card", totalCents: total })
  }, [openReceipt, total])

  const handlePrint = useCallback(() => {
    setPrinting(true)
    setPrintError(null)
    try {
      if (canPrint) {
        getBridge()?.print?.(JSON.stringify(receipt || {}))
      }
    } catch (e: unknown) {
      setPrintError(e instanceof Error ? e.message : "Drucken fehlgeschlagen")
      setPrinting(false)
      return
    }
  }, [canPrint, receipt])

  const handleSkip = useCallback(() => {
    setPrinting(false)
    setPrinted(true)
  }, [])

  return (
    <>
      <aside className="bg-card top-0 mr-3 flex h-[calc(100dvh-5rem)] min-h-0 flex-col rounded-2xl pb-3 md:sticky md:mr-4">
        <div className="px-4 py-4">
          <h3 className="text-lg font-semibold">Warenkorb</h3>
        </div>
        <div className="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
          {cart.items.length === 0 ? (
            <div className="flex h-full items-start justify-center">
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <ShoppingCart className="text-muted-foreground mb-3 size-8" />
                <p className="text-muted-foreground text-sm font-medium">Warenkorb ist leer</p>
              </div>
            </div>
          ) : (
            (() => {
              const rows: React.ReactNode[] = []
              if (suggestion && contiguous && startIndex >= 0 && endIndex >= startIndex) {
                for (let i = 0; i < startIndex; i++) {
                  const item = cart.items[i]!
                  rows.push(
                    <div key={item.id} className="pb-2">
                      <CartItemDisplay
                        item={item}
                        onUpdateQuantity={(q) => updateQuantity(item.id, q)}
                        onRemove={() => removeFromCart(item.id)}
                        onEdit={() => setEditingItem(item)}
                        isPOS
                      />
                    </div>
                  )
                }
                const grouped = cart.items.slice(startIndex, endIndex + 1)
                rows.push(
                  <div key={`group-${grouped.map((g) => g.id).join("-")}`} className="pb-2">
                    <InlineMenuGroup
                      suggestion={suggestion}
                      items={grouped}
                      onEditItem={(it) => setEditingItem(it)}
                      isPOS
                    />
                  </div>
                )
                for (let i = endIndex + 1; i < cart.items.length; i++) {
                  const item = cart.items[i]!
                  rows.push(
                    <div key={item.id} className="pb-2">
                      <CartItemDisplay
                        item={item}
                        onUpdateQuantity={(q) => updateQuantity(item.id, q)}
                        onRemove={() => removeFromCart(item.id)}
                        onEdit={() => setEditingItem(item)}
                        isPOS
                      />
                    </div>
                  )
                }
              } else {
                for (const item of cart.items) {
                  rows.push(
                    <div key={item.id} className="pb-2">
                      <CartItemDisplay
                        item={item}
                        onUpdateQuantity={(q) => updateQuantity(item.id, q)}
                        onRemove={() => removeFromCart(item.id)}
                        onEdit={() => setEditingItem(item)}
                        isPOS
                      />
                    </div>
                  )
                }
              }
              return rows
            })()
          )}
        </div>
        <div className="border-border flex flex-col gap-3 border-t p-4">
          <div className="flex items-center justify-between pb-2">
            <div className="flex flex-col">
              <p className="text-lg font-semibold">Total</p>
              <span className="text-muted-foreground text-sm">{cart.items.length} Produkte</span>
            </div>
            <p className="text-lg font-semibold">{formatChf(total)}</p>
          </div>
          <div className="flex flex-col gap-2">
            <Button
              className="h-12 w-full rounded-xl text-base"
              disabled={cartIsEmpty}
              onClick={() => {
                setTender(null)
                setReceived("")
                setShowCheckout(true)
              }}
            >
              Jetzt bezahlen
            </Button>
            <Button variant="outline" className="h-9 w-full rounded-xl text-sm" onClick={clearCart}>
              Leeren
            </Button>
          </div>
        </div>
      </aside>

      {editingItem && (
        <ProductConfigurationModal
          product={editingItem.product}
          isOpen={true}
          onClose={() => setEditingItem(null)}
          initialConfiguration={editingItem.configuration}
          editingItemId={editingItem.id}
        />
      )}

      <Dialog
        open={showCheckout}
        onOpenChange={(v) => {
          setShowCheckout(v)
          if (!v) {
            closeCheckout()
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {tender === null ? "Bezahlen" : tender === "cash" ? "Barzahlung" : "Kartenzahlung"}
            </DialogTitle>
          </DialogHeader>

          {tender === null && (
            <div className="grid grid-cols-2 gap-3">
              <Button
                className="flex h-36 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => setTender("cash")}
                aria-label="Bar bezahlen"
              >
                <Banknote className="size-14" />
                <span className="text-lg font-medium">Bar</span>
              </Button>
              <Button
                className="flex h-36 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => setTender("card")}
                aria-label="Mit Karte bezahlen"
              >
                <CreditCard className="size-14" />
                <span className="text-lg font-medium">Karte</span>
              </Button>
            </div>
          )}

          {tender === "cash" && (
            <div className="mt-2 space-y-3">
              <div className="flex items-center justify-between">
                <span>Gesamt</span>
                <span>{formatChf(total)}</span>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="received">Erhaltener Betrag</Label>
                <Input
                  id="received"
                  inputMode="decimal"
                  placeholder="z. B. 20.00"
                  value={received}
                  onChange={(e) => setReceived(e.target.value)}
                />
                <div className="grid grid-cols-3 gap-2">
                  {["7", "8", "9", "4", "5", "6", "1", "2", "3"].map((d) => (
                    <Button
                      key={d}
                      variant="outline"
                      onClick={() => setReceived((v) => (v || "") + d)}
                      className="h-16 text-lg"
                    >
                      {d}
                    </Button>
                  ))}
                  <Button variant="outline" onClick={() => setReceived("")} className="h-16 text-lg">
                    C
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => setReceived((v) => (v || "") + "0")}
                    className="h-16 text-lg"
                  >
                    0
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() =>
                      setReceived((v) => {
                        const cur = v || ""
                        if (cur.includes(".") || cur.includes(",")) return cur
                        return cur + "."
                      })
                    }
                    className="h-16 text-lg"
                  >
                    .
                  </Button>
                </div>
              </div>
              <div className="flex items-center justify-between">
                <span>Rückgeld</span>
                <span>{formatChf(changeCents)}</span>
              </div>
              <div className="grid gap-2">
                <Button
                  className="h-12 w-full rounded-xl text-base"
                  disabled={receivedCents < total || busy}
                  onClick={startCashPayment}
                >
                  Bar bezahlen
                </Button>
                {error && <div className="text-sm text-red-600">{error}</div>}
              </div>
            </div>
          )}

          {tender === "card" && (
            <div className="mt-2 space-y-3">
              <p className="text-muted-foreground text-sm">Zahlung am verbundenen SumUp‑Terminal starten.</p>
              <Button className="w-full" onClick={startCardPayment} disabled={!canPayWithCard}>
                Mit Karte bezahlen
              </Button>
              {!canPayWithCard && (
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => openReceipt({ method: "card", totalCents: total })}
                >
                  Überspringen
                </Button>
              )}
              {!hasPosBridge && (
                <div className="text-muted-foreground text-xs">Kartenzahlung außerhalb der Android‑App deaktiviert</div>
              )}
            </div>
          )}

          <DialogFooter>
            {tender !== null ? (
              <Button
                variant="outline"
                onClick={() => {
                  setTender(null)
                  setReceived("")
                  setError(null)
                }}
              >
                Zurück
              </Button>
            ) : (
              <Button variant="outline" onClick={() => setShowCheckout(false)}>
                Abbrechen
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Receipt printing screen */}
      <Dialog
        open={showReceipt}
        onOpenChange={(v) => {
          setShowReceipt(v)
          if (!v) {
            closeReceipt()
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Beleg drucken</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {printing && (
              <div className="flex flex-col items-center justify-center gap-6 py-4">
                <div className="relative h-72 w-72">
                  <div className="absolute inset-0 rounded-full bg-gradient-to-br from-blue-200/60 to-blue-400/40 blur-sm" />
                  <div className="absolute inset-8 rounded-full border-8 border-blue-300/60" />
                  <div className="absolute inset-16 rounded-full border-8 border-blue-400/50" />
                  <div className="absolute inset-24 flex items-center justify-center rounded-full bg-blue-500/80">
                    <Printer className="h-14 w-14 text-white" />
                  </div>
                </div>
                <div className="text-xl font-semibold">Beleg wird gedruckt</div>
                {!canPrint && (
                  <Button
                    variant="outline"
                    className="h-10"
                    onClick={() => {
                      setPrinting(false)
                      setPrinted(true)
                    }}
                  >
                    Überspringen
                  </Button>
                )}
              </div>
            )}
            {printed && (
              <div className="flex flex-col items-center justify-center gap-6 py-4">
                <div className="relative h-72 w-72">
                  <div className="absolute inset-0 rounded-full bg-gradient-to-br from-green-200/60 to-green-400/40 blur-sm" />
                  <div className="absolute inset-8 rounded-full border-8 border-green-300/60" />
                  <div className="absolute inset-16 rounded-full border-8 border-green-400/50" />
                  <div className="absolute inset-24 flex items-center justify-center rounded-full bg-green-500/80">
                    <Check className="h-14 w-14 text-white" />
                  </div>
                </div>
                <div className="text-xl font-semibold">Beleg gedruckt</div>
              </div>
            )}
            {receipt && (
              <div className="grid gap-1 text-sm">
                <div className="flex items-center justify-between">
                  <span>Gesamt</span>
                  <span className="font-medium">{formatChf(receipt.totalCents)}</span>
                </div>
                {typeof receipt.amountReceivedCents === "number" && (
                  <div className="flex items-center justify-between">
                    <span>Erhalten</span>
                    <span className="font-medium">{formatChf(receipt.amountReceivedCents)}</span>
                  </div>
                )}
                {typeof receipt.changeCents === "number" && (
                  <div className="flex items-center justify-between">
                    <span>Rückgeld</span>
                    <span className="font-medium">{formatChf(receipt.changeCents)}</span>
                  </div>
                )}
                {receipt.orderId && (
                  <div className="flex items-center justify-between">
                    <span>Bestellung</span>
                    <code className="bg-muted rounded px-2 py-0.5 text-xs">{receipt.orderId}</code>
                  </div>
                )}
                <div className="flex items-center justify-between">
                  <span>Zahlart</span>
                  <span className="font-medium">{receipt.method === "cash" ? "Bar" : "Karte"}</span>
                </div>
              </div>
            )}
            {!printing && !printed && (
              <div className="grid gap-2">
                <Button className="h-12 w-full rounded-xl text-base" onClick={handlePrint} disabled={printing}>
                  {printing ? "Drucken…" : "Beleg drucken"}
                </Button>
                <Button variant="outline" className="h-10 w-full" onClick={handleSkip}>
                  Überspringen
                </Button>
                {printError && <div className="text-sm text-red-600">{printError}</div>}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowReceipt(false)}>
              Schliessen
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
