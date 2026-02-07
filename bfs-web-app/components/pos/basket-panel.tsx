"use client"

import { Banknote, Check, CreditCard, Gift, Printer, QrCode, ShoppingCart, XCircle } from "lucide-react"
import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { CartItemDisplay } from "@/components/cart/cart-item-display"
import { InlineMenuGroup } from "@/components/cart/inline-menu-group"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { useBestMenuSuggestion } from "@/components/cart/use-best-menu-suggestion"
import { Club100PickerDialog } from "@/components/pos/club100-picker-dialog"
import { GratisTypeDialog } from "@/components/pos/gratis-type-dialog"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useCart } from "@/contexts/cart-context"

import type { Club100Person } from "@/lib/api/club100"
import { formatChf } from "@/lib/utils"
import type { CartItem } from "@/types/cart"
import type { PosFulfillmentMode } from "@/types/jeton"
import type {
  Club100Discount,
  GratisInfo,
  PosPaymentMethod,
  QueuedOrder,
  QueuedOrderItem,
  ReceiptItem,
} from "@/types/order-queue"
import type { ProductSummaryDTO } from "@/types/product"

type Tender = PosPaymentMethod | null
type Receipt = {
  method: PosPaymentMethod
  totalCents: number
  orderId?: string
  items?: Array<{
    title: string
    quantity: number
    unitPriceCents: number
    configuration?: Array<{ slot: string; choice: string }>
  }>
  pickupQr?: string | null
}

type PosBridge = {
  print?: (s: string) => void
  payWithCard?: (p: { amountCents: number; currency: string; reference: string } | string) => void
}

const getBridge = () => (globalThis as unknown as { PosBridge?: PosBridge }).PosBridge

type JetonTotal = { id: string; name: string; color: string; count: number }

function toMenuSelections(config?: Record<string, string>) {
  if (!config) return undefined
  return Object.entries(config).map(([slotId, productId]) => ({ slotId, productId }))
}

function textColorForBg(hex: string) {
  const h = (hex || "").replace("#", "")
  if (h.length !== 6) return "#0f172a"
  const r = parseInt(h.slice(0, 2), 16) / 255
  const g = parseInt(h.slice(2, 4), 16) / 255
  const b = parseInt(h.slice(4, 6), 16) / 255
  const luminance = 0.299 * r + 0.587 * g + 0.114 * b
  return luminance > 0.55 ? "#0f172a" : "#ffffff"
}

interface BasketPanelProps {
  token: string
  mode?: PosFulfillmentMode
  submitOrder: (
    items: QueuedOrderItem[],
    totalCents: number,
    paymentMethod: PosPaymentMethod,
    receiptData?: { items: ReceiptItem[]; pickupQr: string | null },
    gratisInfo?: GratisInfo
  ) => QueuedOrder | null
  stockMap?: Map<string, number>
}

export function BasketPanel({ token, mode = "QR_CODE", submitOrder, stockMap }: BasketPanelProps) {
  const { cart, updateQuantity, removeFromCart, clearCart } = useCart()

  const getMaxQuantity = useCallback(
    (item: CartItem): number | null => {
      if (!stockMap || item.product.type === "menu") return null
      return stockMap.get(item.product.id) ?? null
    },
    [stockMap]
  )
  const { suggestion, contiguous, startIndex, endIndex, dismissSuggestion } = useBestMenuSuggestion()
  const [showCheckout, setShowCheckout] = useState(false)
  const [tender, setTender] = useState<Tender>(null)
  const [received, setReceived] = useState<string>("")
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editingItem, setEditingItem] = useState<CartItem | null>(null)
  const [showReceipt, setShowReceipt] = useState(false)
  const [receipt, setReceipt] = useState<Receipt | null>(null)
  const [printing, setPrinting] = useState(false)
  const [printed, setPrinted] = useState(false)
  const [printError, setPrintError] = useState<string | null>(null)
  const [hasPosBridge, setHasPosBridge] = useState(false)
  const [canPrint, setCanPrint] = useState(false)
  const [canPayWithCard, setCanPayWithCard] = useState(false)
  const [showCard, setShowCard] = useState(false)
  const [cardProcessing, setCardProcessing] = useState(false)
  const [cardSuccess, setCardSuccess] = useState(false)
  const [cardError, setCardError] = useState<string | null>(null)
  const [cardRef, setCardRef] = useState<string | null>(null)
  const [cardPrintInProgress, setCardPrintInProgress] = useState(false)
  const [cardPrintDone, setCardPrintDone] = useState(false)
  const [printErrorDialog, setPrintErrorDialog] = useState<string | null>(null)
  const [jetonSummary, setJetonSummary] = useState<{
    items: JetonTotal[]
    orderId?: string
    payment?: PosPaymentMethod
  } | null>(null)
  const [showGratisDialog, setShowGratisDialog] = useState(false)
  const [showClub100Picker, setShowClub100Picker] = useState(false)
  const [selectedGratisInfo, setSelectedGratisInfo] = useState<GratisInfo | null>(null)
  const [club100Discount, setClub100Discount] = useState<Club100Discount | null>(null)
  const [club100FreeProductIds, setClub100FreeProductIds] = useState<string[]>([])
  const originalTotal = cart.totalCents
  const discountAmount = club100Discount?.totalDiscountCents ?? 0
  const total = originalTotal - discountAmount
  const receivedCents = useMemo(() => Math.round((parseFloat(received || "0") || 0) * 100), [received])
  const changeCents = Math.max(0, receivedCents - total)
  const scrollRef = useRef<HTMLDivElement>(null)
  const prevItemCount = useRef(cart.items.length)
  const cartIsEmpty = cart.items.length === 0
  const jetonMode = mode === "JETON"
  const resolveMenuSelections = useCallback((item: CartItem) => {
    if (item.product.type !== "menu" || !item.configuration) return []
    const slots = item.product.menu?.slots || []
    const selections: ProductSummaryDTO[] = []
    for (const [slotId, productId] of Object.entries(item.configuration)) {
      const slot = slots.find((s) => s.id === slotId)
      const choice = slot?.options?.find((p) => p.id === productId)
      if (choice) selections.push(choice)
    }
    return selections
  }, [])
  const hasMissingJeton = useMemo(() => {
    for (const it of cart.items) {
      if (it.product.type === "menu") {
        const selections = resolveMenuSelections(it)
        const configuredSlots = Object.keys(it.configuration || {})
        if (configuredSlots.length === 0 || selections.length < configuredSlots.length) return true
        for (const choice of selections) {
          if (!choice.jeton) return true
        }
      } else if (!it.product?.jeton) {
        return true
      }
    }
    return false
  }, [cart.items, resolveMenuSelections])
  const computeJetonTotals = useCallback((): JetonTotal[] => {
    const totals = new Map<string, JetonTotal>()
    for (const it of cart.items) {
      if (it.product.type === "menu") {
        const selections = resolveMenuSelections(it)
        for (const choice of selections) {
          const jeton = choice.jeton
          if (!jeton) continue
          const existing = totals.get(jeton.id)
          if (existing) existing.count += it.quantity
          else totals.set(jeton.id, { id: jeton.id, name: jeton.name, color: jeton.color, count: it.quantity })
        }
        continue
      }
      const jeton = it.product.jeton
      if (!jeton) continue
      const existing = totals.get(jeton.id)
      if (existing) existing.count += it.quantity
      else totals.set(jeton.id, { id: jeton.id, name: jeton.name, color: jeton.color, count: it.quantity })
    }
    return Array.from(totals.values()).sort((a, b) => a.name.localeCompare(b.name))
  }, [cart.items, resolveMenuSelections])
  const buildPrintItems = useCallback((): NonNullable<Receipt["items"]> => {
    return cart.items.map((it) => {
      const cfg: Array<{ slot: string; choice: string }> = []
      if (it.product.type === "menu" && it.configuration && it.product.menu?.slots) {
        for (const [slotId, productId] of Object.entries(it.configuration)) {
          const slot = it.product.menu.slots.find((s) => s.id === slotId)
          const choice = slot?.options?.find((p) => p.id === productId)
          if (slot && choice) cfg.push({ slot: slot.name, choice: choice.name })
        }
      }
      return {
        title: it.product.name,
        quantity: it.quantity,
        unitPriceCents: it.totalPriceCents,
        configuration: cfg.length ? cfg : undefined,
      }
    })
  }, [cart.items])
  const generatePickupQr = useCallback((orderId: string) => {
    return orderId
  }, [])

  const emitInventoryDecrement = useCallback((items: typeof cart.items) => {
    const decrements = new Map<string, number>()
    for (const item of items) {
      decrements.set(item.product.id, (decrements.get(item.product.id) || 0) - item.quantity)
      if (item.product.type === "menu" && item.configuration && item.product.menu?.slots) {
        for (const [slotId, productId] of Object.entries(item.configuration)) {
          const slot = item.product.menu.slots.find((s) => s.id === slotId)
          const choice = slot?.options?.find((p) => p.id === productId)
          if (choice) {
            decrements.set(productId, (decrements.get(productId) || 0) - item.quantity)
          }
        }
      }
    }
    window.dispatchEvent(new CustomEvent("pos:inventory-decrement", { detail: decrements }))
  }, [])

  useEffect(() => {
    const onLock = () => {
      setShowCheckout(false)
      setShowReceipt(false)
      setTender(null)
      setEditingItem(null)
      setClub100Discount(null)
    }
    window.addEventListener("pos:lock", onLock)
    return () => window.removeEventListener("pos:lock", onLock)
  }, [])

  useEffect(() => {
    if (!token) return
    ;(async () => {
      try {
        const res = await fetch("/api/v1/settings", {
          headers: { Authorization: `Bearer ${token}` },
        })
        if (res.ok) {
          const data = (await res.json()) as { club100FreeProductIds?: string[] }
          setClub100FreeProductIds(data.club100FreeProductIds ?? [])
        }
      } catch {}
    })()
  }, [token])

  const clearClub100Discount = useCallback(() => {
    setClub100Discount(null)
    setSelectedGratisInfo(null)
  }, [])

  // Auto-scroll to bottom when new items are added
  useEffect(() => {
    if (cart.items.length > prevItemCount.current && scrollRef.current) {
      scrollRef.current.scrollTo({ top: scrollRef.current.scrollHeight, behavior: "smooth" })
    }
    prevItemCount.current = cart.items.length
  }, [cart.items.length])

  // Clear Club100 discount when cart items change (quantities, items added/removed)
  const cartItemsKey = useMemo(() => cart.items.map((i) => `${i.id}:${i.quantity}`).join(","), [cart.items])
  useEffect(() => {
    if (club100Discount) {
      clearClub100Discount()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cartItemsKey])

  // Listen for native print results from the Android WebView bridge
  useEffect(() => {
    const onPrintResult = (ev: Event) => {
      try {
        const ce = ev as CustomEvent<{ success?: boolean; error?: string }>
        const detail = ce.detail || {}
        if (showReceipt) {
          // Manual print dialog path
          setPrinting(false)
          if (detail.success) setPrinted(true)
          else setPrintError(detail.error || "Drucken fehlgeschlagen")
        } else if (showCard) {
          // Card success screen background printing progress
          setCardPrintInProgress(false)
          if (detail.success) {
            setCardPrintDone(true)
            setTimeout(() => setShowCard(false), 1200)
          } else {
            setPrintErrorDialog(detail.error || "Drucken fehlgeschlagen")
          }
        } else {
          // Background printing path — only surface errors
          if (!detail.success) setPrintErrorDialog(detail.error || "Drucken fehlgeschlagen")
        }
      } catch {
        if (showReceipt) {
          setPrinting(false)
          setPrintError("Drucken fehlgeschlagen")
        } else if (showCard) {
          setCardPrintInProgress(false)
          setPrintErrorDialog("Drucken fehlgeschlagen")
        } else {
          setPrintErrorDialog("Drucken fehlgeschlagen")
        }
      }
    }
    window.addEventListener("bfs:print:result", onPrintResult as EventListener)
    return () => window.removeEventListener("bfs:print:result", onPrintResult as EventListener)
  }, [showReceipt, showCard])

  // Listen for SumUp results from the Android WebView bridge
  useEffect(() => {
    const onSumup = async (ev: Event) => {
      try {
        const ce = ev as CustomEvent<{ success?: boolean; error?: string; txId?: string; correlationId?: string }>
        const d = ce.detail || {}
        if (cardRef && d.correlationId && d.correlationId !== cardRef) return
        if (!showCard) return

        if (d.success) {
          const itemsBody = cart.items.map((it) => ({
            productId: it.product.id,
            quantity: it.quantity,
            menuSelections: toMenuSelections(it.configuration),
          }))
          const printItems = buildPrintItems()
          const localOrderId = `local_${Date.now()}`
          const pickupQr = generatePickupQr(localOrderId)

          const nextReceipt: Receipt & { orderTimestamp?: number } = {
            method: "card",
            totalCents: total,
            items: printItems,
            pickupQr,
            orderTimestamp: Date.now(),
          }
          const jetons = computeJetonTotals()

          submitOrder(
            itemsBody,
            originalTotal,
            "card",
            { items: printItems, pickupQr },
            selectedGratisInfo ?? undefined
          )
          emitInventoryDecrement(cart.items)

          if (jetonMode) {
            setCardProcessing(false)
            setCardSuccess(true)
            setShowCard(false)
            setJetonSummary({ items: jetons, orderId: localOrderId, payment: "card" })
            clearCart()
            clearClub100Discount()
            return
          }

          setCardProcessing(false)
          setCardSuccess(true)
          setCardPrintInProgress(true)
          try {
            if (canPrint) getBridge()?.print?.(JSON.stringify(nextReceipt))
          } catch {}
          clearCart()
          clearClub100Discount()
        } else {
          setCardProcessing(false)
          setCardSuccess(false)
          setCardError(d.error || "Kartenzahlung fehlgeschlagen")
        }
      } catch (e) {
        setCardProcessing(false)
        setCardSuccess(false)
        setCardError(e instanceof Error ? e.message : "Kartenzahlung fehlgeschlagen")
      }
    }
    window.addEventListener("bfs:sumup:result", onSumup as EventListener)
    return () => window.removeEventListener("bfs:sumup:result", onSumup as EventListener)
  }, [
    total,
    originalTotal,
    showCard,
    cardRef,
    clearCart,
    clearClub100Discount,
    jetonMode,
    computeJetonTotals,
    buildPrintItems,
    generatePickupQr,
    cart.items,
    submitOrder,
    canPrint,
    emitInventoryDecrement,
    selectedGratisInfo,
  ])

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
    if (jetonMode && hasMissingJeton) {
      setError("Im Jeton-Modus benötigt jedes Produkt einen Jeton.")
      return
    }
    setBusy(true)
    setError(null)

    const items = cart.items.map((it) => ({
      productId: it.product.id,
      quantity: it.quantity,
      menuSelections: toMenuSelections(it.configuration),
    }))
    const printItems = buildPrintItems()
    const jetons = computeJetonTotals()
    const localOrderId = `local_${Date.now()}`
    const pickupQr = generatePickupQr(localOrderId)

    const receiptPayload: Receipt & { orderTimestamp?: number } = {
      method: "cash",
      totalCents: total,
      items: printItems,
      pickupQr: pickupQr || undefined,
      orderTimestamp: Date.now(),
    }

    submitOrder(items, originalTotal, "cash", { items: printItems, pickupQr }, selectedGratisInfo ?? undefined)
    emitInventoryDecrement(cart.items)

    if (jetonMode) {
      setJetonSummary({ items: jetons, orderId: localOrderId, payment: "cash" })
    } else {
      try {
        if (canPrint) getBridge()?.print?.(JSON.stringify(receiptPayload))
      } catch {}
    }

    clearCart()
    clearClub100Discount()
    setShowCheckout(false)
    setTender(null)
    setBusy(false)
  }, [
    busy,
    cart.items,
    originalTotal,
    total,
    canPrint,
    jetonMode,
    hasMissingJeton,
    computeJetonTotals,
    clearCart,
    clearClub100Discount,
    buildPrintItems,
    generatePickupQr,
    submitOrder,
    emitInventoryDecrement,
    selectedGratisInfo,
  ])

  const startTwintPayment = useCallback(async () => {
    if (busy) return
    if (jetonMode && hasMissingJeton) {
      setError("Im Jeton-Modus benötigt jedes Produkt einen Jeton.")
      return
    }
    setBusy(true)
    setError(null)

    const items = cart.items.map((it) => ({
      productId: it.product.id,
      quantity: it.quantity,
      menuSelections: toMenuSelections(it.configuration),
    }))
    const printItems = buildPrintItems()
    const jetons = computeJetonTotals()
    const localOrderId = `local_${Date.now()}`
    const pickupQr = generatePickupQr(localOrderId)

    const receiptPayload: Receipt & { orderTimestamp?: number } = {
      method: "twint",
      totalCents: total,
      items: printItems,
      pickupQr: pickupQr || undefined,
      orderTimestamp: Date.now(),
    }

    submitOrder(items, originalTotal, "twint", { items: printItems, pickupQr }, selectedGratisInfo ?? undefined)
    emitInventoryDecrement(cart.items)

    if (jetonMode) {
      setJetonSummary({ items: jetons, orderId: localOrderId, payment: "twint" })
    } else {
      try {
        if (canPrint) getBridge()?.print?.(JSON.stringify(receiptPayload))
      } catch {}
    }

    clearCart()
    clearClub100Discount()
    setShowCheckout(false)
    setTender(null)
    setBusy(false)
  }, [
    busy,
    jetonMode,
    hasMissingJeton,
    cart.items,
    computeJetonTotals,
    buildPrintItems,
    generatePickupQr,
    originalTotal,
    total,
    canPrint,
    clearCart,
    clearClub100Discount,
    submitOrder,
    emitInventoryDecrement,
    selectedGratisInfo,
  ])

  const startCardPayment = useCallback(() => {
    if (jetonMode && hasMissingJeton) {
      setError("Im Jeton-Modus benötigt jedes Produkt einen Jeton.")
      return
    }
    const bridge = getBridge()
    if (bridge && typeof bridge.payWithCard === "function") {
      const ref = `pos_${Date.now()}`
      const payload = { amountCents: total, currency: "CHF", reference: ref }
      setCardRef(ref)
      setCardError(null)
      setCardSuccess(false)
      setCardProcessing(true)
      setShowCheckout(false)
      setShowCard(true)
      try {
        // Prefer JSON string for Android JS bridge
        bridge.payWithCard(JSON.stringify(payload))
      } catch {
        // Fallback to object for other environments
        bridge.payWithCard(payload)
      }
    }
  }, [total, jetonMode, hasMissingJeton])

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

  const tenderTitle =
    tender === null
      ? "Bezahlen"
      : tender === "cash"
        ? "Barzahlung"
        : tender === "card"
          ? "Kartenzahlung"
          : tender === "twint"
            ? "TWINT-Zahlung"
            : "Gratis"
  const paymentMethodLabel = (method: PosPaymentMethod) => {
    switch (method) {
      case "cash":
        return "Bar"
      case "card":
        return "Karte"
      case "twint":
        return "TWINT"
      case "gratis_guest":
        return "Gratis (Gast)"
      case "gratis_vip":
        return "Gratis (VIP)"
      case "gratis_staff":
        return "Gratis (Mitarbeiter)"
      case "gratis_100club":
        return "Gratis (100 Club)"
      default:
        return method
    }
  }

  const completeGratisPayment = useCallback(
    (paymentMethod: PosPaymentMethod, gratisInfo: GratisInfo) => {
      if (jetonMode && hasMissingJeton) {
        setError("Im Jeton-Modus benötigt jedes Produkt einen Jeton.")
        return
      }

      const items = cart.items.map((it) => ({
        productId: it.product.id,
        quantity: it.quantity,
        menuSelections: toMenuSelections(it.configuration),
      }))
      const printItems = buildPrintItems()
      const jetons = computeJetonTotals()
      const localOrderId = `local_${Date.now()}`
      const pickupQr = generatePickupQr(localOrderId)

      submitOrder(items, total, paymentMethod, { items: printItems, pickupQr }, gratisInfo)
      emitInventoryDecrement(cart.items)

      if (jetonMode) {
        setJetonSummary({ items: jetons, orderId: localOrderId, payment: paymentMethod })
      }

      clearCart()
      setShowCheckout(false)
      setTender(null)
    },
    [
      cart.items,
      total,
      jetonMode,
      hasMissingJeton,
      computeJetonTotals,
      buildPrintItems,
      generatePickupQr,
      submitOrder,
      emitInventoryDecrement,
      clearCart,
    ]
  )

  const handleGratisTypeSelect = useCallback(
    (type: "guest" | "vip" | "staff" | "100club") => {
      setShowGratisDialog(false)
      if (type === "100club") {
        setShowClub100Picker(true)
      } else {
        setSelectedGratisInfo({ type })
        const paymentMethodMap: Record<"guest" | "vip" | "staff", PosPaymentMethod> = {
          guest: "gratis_guest",
          vip: "gratis_vip",
          staff: "gratis_staff",
        }
        completeGratisPayment(paymentMethodMap[type], { type })
      }
    },
    [completeGratisPayment]
  )

  const handleClub100Select = useCallback((person: Club100Person, discount: Club100Discount | null) => {
    setShowClub100Picker(false)

    if (!discount || discount.discountedItems.length === 0) {
      setClub100Discount(null)
      setSelectedGratisInfo(null)
      return
    }

    setClub100Discount(discount)

    const totalDiscountedQty = discount.discountedItems.reduce((sum, d) => sum + d.discountedQuantity, 0)
    const gratisInfo: GratisInfo = {
      type: "100club",
      elvantoPersonId: person.id,
      elvantoPersonName: `${person.firstName} ${person.lastName}`,
      freeQuantity: totalDiscountedQty,
    }
    setSelectedGratisInfo(gratisInfo)
  }, [])

  const getItemDiscountInfo = useCallback(
    (itemId: string) => {
      if (!club100Discount) return undefined
      const discountItem = club100Discount.discountedItems.find((d) => d.cartItemId === itemId)
      if (!discountItem) return undefined
      return {
        discountedQuantity: discountItem.discountedQuantity,
        unitPriceCents: discountItem.unitPriceCents,
      }
    },
    [club100Discount]
  )

  return (
    <>
      <aside className="bg-card top-0 mr-3 flex h-[calc(100dvh-5rem)] min-h-0 flex-col rounded-2xl md:sticky md:mr-4">
        <div className="px-4 py-4">
          <h3 className="text-lg font-semibold">Warenkorb</h3>
        </div>
        <div ref={scrollRef} className="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
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
                        discountInfo={getItemDiscountInfo(item.id)}
                        maxQuantity={getMaxQuantity(item)}
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
                      onDismiss={dismissSuggestion}
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
                        discountInfo={getItemDiscountInfo(item.id)}
                        maxQuantity={getMaxQuantity(item)}
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
                        discountInfo={getItemDiscountInfo(item.id)}
                        maxQuantity={getMaxQuantity(item)}
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
          {club100Discount && (
            <div className="flex items-center justify-between rounded-lg bg-green-50 px-3 py-2 dark:bg-green-950/30">
              <div className="flex flex-col">
                <span className="text-sm font-medium text-green-700 dark:text-green-400">
                  100 Club: {club100Discount.person.firstName} {club100Discount.person.lastName}
                </span>
                <span className="text-xs text-green-600 dark:text-green-500">
                  {club100Discount.discountedItems.reduce((sum, d) => sum + d.discountedQuantity, 0)} Produkt(e) gratis
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearClub100Discount}
                className="h-7 px-2 text-xs text-green-700 hover:text-green-900 dark:text-green-400"
              >
                Entfernen
              </Button>
            </div>
          )}
          <div className="flex items-center justify-between pb-2">
            <div className="flex flex-col">
              <p className="text-base font-semibold">Total</p>
              <span className="text-muted-foreground text-xs">{cart.items.length} Produkte</span>
            </div>
            <div className="flex flex-col items-end">
              {discountAmount > 0 && (
                <span className="text-muted-foreground text-sm line-through">{formatChf(originalTotal)}</span>
              )}
              <p className="text-base font-semibold">{formatChf(total)}</p>
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <Button
              className="h-12 w-full rounded-xl text-sm"
              disabled={cartIsEmpty}
              onClick={() => {
                if (jetonMode && hasMissingJeton) {
                  setError("Bitte weise allen Produkten einen Jeton zu.")
                  return
                }
                if (club100Discount && selectedGratisInfo && total <= 0) {
                  completeGratisPayment("gratis_100club", selectedGratisInfo)
                  return
                }
                setTender(null)
                setReceived("")
                setShowCheckout(true)
              }}
            >
              {club100Discount && total <= 0 ? "Gratis abschliessen" : "Jetzt bezahlen"}
            </Button>
            <Button
              variant="outline"
              className="h-10 w-full rounded-xl text-xs"
              onClick={clearCart}
              disabled={cartIsEmpty}
            >
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
            <DialogTitle>{tenderTitle}</DialogTitle>
          </DialogHeader>

          {tender === null && (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
              <Button
                className="flex h-36 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => {
                  setError(null)
                  setReceived("")
                  setTender("cash")
                }}
                aria-label="Bar bezahlen"
              >
                <Banknote className="size-12" />
                <span className="text-base font-medium">Bar</span>
              </Button>
              <Button
                className="flex h-36 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => {
                  setError(null)
                  if (canPayWithCard) {
                    startCardPayment()
                  } else {
                    setTender("card")
                  }
                }}
                aria-label="Mit Karte bezahlen"
              >
                <CreditCard className="size-12" />
                <span className="text-base font-medium">Karte</span>
              </Button>
              <Button
                className="flex h-36 flex-col items-center justify-center gap-2 rounded-xl"
                variant="outline"
                onClick={() => {
                  setError(null)
                  setTender("twint")
                }}
                aria-label="Mit TWINT bezahlen"
              >
                <QrCode className="size-12" />
                <span className="text-base font-medium">TWINT</span>
              </Button>
            </div>
          )}

          {tender === null && (
            <div className="mt-2 flex justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setShowCheckout(false)
                  setShowGratisDialog(true)
                }}
              >
                <Gift className="mr-2 size-4" />
                Gratis
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
                <span className="font-semibold">{formatChf(changeCents)}</span>
              </div>
              <div className="flex flex-col gap-2">
                <Button
                  className="h-12 w-full rounded-xl text-base"
                  disabled={receivedCents < total || busy}
                  onClick={startCashPayment}
                >
                  Barzahlung abschliessen
                </Button>
                <Button
                  variant="outline"
                  className="h-10 w-full rounded-xl text-xs"
                  disabled={busy}
                  onClick={startCashPayment}
                >
                  Überspringen
                </Button>
                {error && <div className="text-sm text-red-600">{error}</div>}
              </div>
            </div>
          )}

          {tender === "card" && (
            <div className="mt-2 space-y-3">
              <p className="text-muted-foreground text-sm">Zahlung am verbundenen SumUp-Terminal starten.</p>
              <Button className="w-full" onClick={startCardPayment} disabled={!canPayWithCard}>
                Mit Karte bezahlen
              </Button>
              {!canPayWithCard &&
                (jetonMode ? (
                  <div className="text-muted-foreground text-xs">
                    Kartenzahlung ohne Terminal ist im Jeton-Modus deaktiviert.
                  </div>
                ) : (
                  <Button
                    variant="outline"
                    className="w-full"
                    onClick={() => openReceipt({ method: "card", totalCents: total })}
                  >
                    Überspringen
                  </Button>
                ))}
              {!hasPosBridge && (
                <div className="text-muted-foreground text-xs">Kartenzahlung außerhalb der Android-App deaktiviert</div>
              )}
            </div>
          )}

          {tender === "twint" && (
            <div className="mt-2 space-y-3">
              <div className="flex items-center justify-between">
                <span>Gesamt</span>
                <span>{formatChf(total)}</span>
              </div>
              <div className="grid gap-2">
                <Button className="h-12 w-full rounded-xl text-base" disabled={busy} onClick={startTwintPayment}>
                  TWINT-Zahlung abschliessen
                </Button>
                {error && <div className="text-sm text-red-600">{error}</div>}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Card payment screen */}
      <Dialog
        open={showCard}
        onOpenChange={(v) => {
          setShowCard(v)
          if (!v) {
            setCardProcessing(false)
            setCardSuccess(false)
            setCardError(null)
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Kartenzahlung</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {cardProcessing && (
              <div className="flex flex-col items-center justify-center gap-6 py-4">
                <div className="relative h-72 w-72">
                  <div className="absolute inset-0 rounded-full bg-gradient-to-br from-blue-200/60 to-blue-400/40 blur-sm" />
                  <div className="absolute inset-8 rounded-full border-8 border-blue-300/60" />
                  <div className="absolute inset-16 rounded-full border-8 border-blue-400/50" />
                  <div className="absolute inset-24 flex items-center justify-center rounded-full bg-blue-500/80">
                    <CreditCard className="h-14 w-14 text-white" />
                  </div>
                </div>
                <div className="text-xl font-semibold">Kartenzahlung läuft</div>
              </div>
            )}
            {cardSuccess && (
              <div className="flex flex-col items-center justify-center gap-6 py-4">
                <div className="relative h-72 w-72">
                  <div className="absolute inset-0 rounded-full bg-gradient-to-br from-green-200/60 to-green-400/40 blur-sm" />
                  <div className="absolute inset-8 rounded-full border-8 border-green-300/60" />
                  <div className="absolute inset-16 rounded-full border-8 border-green-400/50" />
                  <div className="absolute inset-24 flex items-center justify-center rounded-full bg-green-500/80">
                    <Check className="h-14 w-14 text-white" />
                  </div>
                </div>
                <div className="text-xl font-semibold">Kartenzahlung erfolgreich</div>
                {cardPrintInProgress && (
                  <div className="text-muted-foreground -mt-2 text-sm">Beleg wird im Hintergrund gedruckt…</div>
                )}
                {cardPrintDone && <div className="text-muted-foreground -mt-2 text-sm">Beleg gedruckt</div>}
              </div>
            )}
            {!cardProcessing && cardError && (
              <div className="flex flex-col items-center justify-center gap-6 py-4">
                <div className="relative h-72 w-72">
                  <div className="absolute inset-0 rounded-full bg-gradient-to-br from-red-200/60 to-red-400/40 blur-sm" />
                  <div className="absolute inset-8 rounded-full border-8 border-red-300/60" />
                  <div className="absolute inset-16 rounded-full border-8 border-red-400/50" />
                  <div className="absolute inset-24 flex items-center justify-center rounded-full bg-red-500/80">
                    <XCircle className="h-14 w-14 text-white" />
                  </div>
                </div>
                <div className="text-xl font-semibold">Kartenzahlung fehlgeschlagen</div>
                <div className="text-muted-foreground -mt-4 text-sm">
                  {cardError || "Bitte versuchen Sie es erneut."}
                </div>
                <div className="mt-2 grid w-full grid-cols-2 gap-3">
                  <Button
                    className="h-12 rounded-xl text-base"
                    onClick={() => {
                      setCardError(null)
                      startCardPayment()
                    }}
                  >
                    Erneut versuchen
                  </Button>
                  <Button variant="outline" className="h-12 rounded-xl text-base" onClick={() => setShowCard(false)}>
                    Schliessen
                  </Button>
                </div>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCard(false)}>
              Schliessen
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Background print error dialog (cash and card) */}
      <Dialog
        open={!!printErrorDialog}
        onOpenChange={(v) => {
          if (!v) setPrintErrorDialog(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Drucken fehlgeschlagen</DialogTitle>
          </DialogHeader>
          <div className="text-sm">{printErrorDialog}</div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPrintErrorDialog(null)}>
              OK
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Jeton summary dialog */}
      <Dialog
        open={!!jetonSummary}
        onOpenChange={(v) => {
          if (!v) setJetonSummary(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Jetons ausgeben</DialogTitle>
          </DialogHeader>
          {jetonSummary && (
            <div className="mb-8 space-y-3">
              <div className="grid gap-3 sm:grid-cols-2">
                {jetonSummary.items.map((j) => (
                  <div
                    key={j.id}
                    className="flex items-center justify-between rounded-xl border px-4 py-3"
                    style={{ backgroundColor: j.color, color: textColorForBg(j.color) }}
                  >
                    <div className="text-base font-semibold">{j.name}</div>
                    <div className="text-lg font-bold">{j.count}</div>
                  </div>
                ))}
                {jetonSummary.items.length === 0 && (
                  <div className="text-muted-foreground text-sm">Keine Jetons berechnet.</div>
                )}
              </div>
            </div>
          )}
          <DialogFooter className="w-full">
            <Button className="h-12 w-full rounded-xl text-base" onClick={() => setJetonSummary(null)}>
              Fertig
            </Button>
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
                {receipt.orderId && (
                  <div className="flex items-center justify-between">
                    <span>Bestellung</span>
                    <code className="bg-muted rounded px-2 py-0.5 text-xs">{receipt.orderId}</code>
                  </div>
                )}
                <div className="flex items-center justify-between">
                  <span>Zahlart</span>
                  <span className="font-medium">{paymentMethodLabel(receipt.method)}</span>
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

      {/* Gratis type selection dialog */}
      <GratisTypeDialog open={showGratisDialog} onOpenChange={setShowGratisDialog} onSelect={handleGratisTypeSelect} />

      {/* 100 Club member picker dialog */}
      <Club100PickerDialog
        open={showClub100Picker}
        onOpenChange={setShowClub100Picker}
        onSelect={handleClub100Select}
        token={token}
        cartItems={cart.items}
        freeProductIds={club100FreeProductIds}
      />
    </>
  )
}
