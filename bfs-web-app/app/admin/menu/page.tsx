"use client"
import { useCallback, useEffect, useMemo, useState } from "react"
import Image from "next/image"
import { Pencil, PencilIcon, Plus, RefreshCw } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { Category, ProductDTO } from "@/types"

type DirtyProduct = {
  id: string
  name: string
  priceCents: number
  categoryId: string | null
  stock: number | null
}

function formatPriceLabel(cents: number): string {
  const francs = Math.floor(cents / 100)
  const rappen = cents % 100
  if (rappen === 0) return `CHF ${francs}.-`
  const v = (cents / 100).toFixed(2)
  return `CHF ${v}`
}

function parsePriceInputToCents(input: string): number | null {
  if (!input) return null
  const norm = input.replace(",", ".").trim()
  const val = Number(norm)
  if (!isFinite(val)) return null
  return Math.round(val * 100)
}

export default function AdminMenuPage() {
  const fetchAuth = useAuthorizedFetch()
  const [cats, setCats] = useState<Category[]>([])
  const [items, setItems] = useState<ProductDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  // activeCat: "all" for all items or category.id for a specific category
  const [activeCat, setActiveCat] = useState<string>("all")
  const [edit, setEdit] = useState<DirtyProduct | null>(null)
  const [saving, setSaving] = useState(false)

  const totalCount = items.length

  // Count items by category id for dynamic chips
  const countsByCatId = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const it of items) {
      const id = it.category?.id
      if (!id) continue
      counts[id] = (counts[id] || 0) + 1
    }
    return counts
  }, [items])

  // Helper to normalize category position (undefined -> large number)
  const getCatPos = useCallback((c?: { position?: number | null } | null) => {
    const p = c?.position
    return typeof p === "number" && isFinite(p) ? p : 1_000_000
  }, [])

  // Categories sorted by position then name
  const sortedCats = useMemo(() => {
    return [...cats].sort((a, b) => {
      const pa = getCatPos(a)
      const pb = getCatPos(b)
      if (pa !== pb) return pa - pb
      return a.name.localeCompare(b.name)
    })
  }, [cats, getCatPos])

  const filtered = useMemo(() => {
    if (activeCat === "all") {
      return [...items].sort((a, b) => {
        const pa = getCatPos(a.category)
        const pb = getCatPos(b.category)
        if (pa !== pb) return pa - pb
        return a.name.localeCompare(b.name)
      })
    }
    return items
      .filter((it) => it.category?.id === activeCat)
      .sort((a, b) => a.name.localeCompare(b.name))
  }, [items, activeCat, getCatPos])

  const refetch = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [pr, cr] = await Promise.all([
        fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/products?limit=200`),
        fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories`),
      ])
      if (cr.ok) {
        const d = (await cr.json()) as { items: Category[] }
        const sorted = (d.items || []).sort((a, b) => {
          const pa = getCatPos(a)
          const pb = getCatPos(b)
          if (pa !== pb) return pa - pb
          return a.name.localeCompare(b.name)
        })
        setCats(sorted)
      }
      if (pr.ok) {
        const d = (await pr.json()) as { items: ProductDTO[] }
        setItems((d.items || []).filter((p) => p.isActive))
      } else {
        throw new Error(`HTTP ${pr.status}`)
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Fehler beim Laden"
      setError(msg)
    } finally {
      setLoading(false)
    }
  }, [fetchAuth])

  useEffect(() => {
    refetch()
  }, [refetch])

  // Initialize category filter via query param; allow "all" or a category id
  useEffect(() => {
    const sp = new URLSearchParams(window.location.search)
    const cat = sp.get("cat")
    if (!cat) return
    if (cat === "all") setActiveCat("all")
    else setActiveCat(cat)
  }, [])
  useEffect(() => {
    const url = new URL(window.location.href)
    url.searchParams.set("cat", activeCat)
    window.history.replaceState({}, "", url.toString())
  }, [activeCat])

  // Listen to global refresh dispatched by header
  useEffect(() => {
    const onR = () => refetch()
    window.addEventListener("admin:refresh", onR)
    return () => window.removeEventListener("admin:refresh", onR)
  }, [refetch])

  return (
    <div className="space-y-5">
      <h1 className="mb-2 text-xl font-semibold">Menu</h1>
      <div className="flex flex-wrap items-center gap-2">
        <Button
          className={`border-border hover:bg-card hover:text-foreground group flex h-10 items-center justify-between gap-2 rounded-[10px] border px-1.5 text-sm ${
            activeCat === "all" ? "bg-card" : "text-muted-foreground bg-transparent"
          }`}
          onClick={() => setActiveCat("all")}
          aria-pressed={activeCat === "all"}
        >
          Alles
          <span
            className={`text-foreground group-hover:text-foreground flex h-7 min-w-7 items-center justify-center rounded-[6px] px-1 text-sm transition-all duration-300 group-hover:bg-[#FFBBBB] ${
              activeCat === "all" ? "text-foreground bg-[#FFBBBB]" : "text-muted-foreground bg-[#D9D9D9]"
            }`}
            aria-label={`${totalCount} Produkte`}
          >
            {totalCount}
          </span>
        </Button>
        {sortedCats.map((c) => (
          <Button
            key={c.id}
            className={`border-border hover:bg-card hover:text-foreground group flex h-10 items-center justify-between gap-2 rounded-[10px] border px-1.5 text-sm ${
              activeCat === c.id ? "bg-card" : "text-muted-foreground bg-transparent"
            }`}
            onClick={() => setActiveCat(c.id)}
            aria-pressed={activeCat === c.id}
          >
            {c.name}{" "}
            <span
              className={`text-foreground group-hover:text-foreground flex h-7 min-w-7 items-center justify-center rounded-[6px] px-1 text-sm transition-all duration-300 group-hover:bg-[#FFBBBB] ${
                activeCat === c.id ? "text-foreground bg-[#FFBBBB]" : "text-muted-foreground bg-[#D9D9D9]"
              }`}
            >
              {countsByCatId[c.id] ?? 0}
            </span>
          </Button>
        ))}
        <Button
          className="border-border text-foreground hover:bg-card flex h-10 items-center justify-between gap-2 rounded-[10px] border bg-transparent px-1.5 text-sm"
          onClick={() => {
            /* could toggle category edit */
          }}
        >
          Bearbeiten
          <span className="bg-foreground flex h-7 min-w-7 items-center justify-center rounded-[6px] px-1 text-sm text-white">
            <Pencil className="size-3.5" />
          </span>
        </Button>
        <Button
          className="border-border text-foreground hover:bg-card flex h-10 items-center justify-between gap-2 rounded-[10px] border bg-transparent px-1.5 text-sm"
          onClick={() => setEdit({ id: "new", name: "", priceCents: 0, categoryId: null, stock: null })}
        >
          Erstellen
          <span className="bg-foreground flex h-7 min-w-7 items-center justify-center rounded-[6px] px-1 text-sm text-white">
            <Plus className="size-4" />
          </span>
        </Button>
      </div>

      {/* Grid */}
      {loading ? (
        <div className="grid grid-cols-1 gap-5 sm:gap-3 xl:gap-5 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Card key={i} className="rounded-xl border">
              <div className="p-3">
                <Skeleton className="aspect-[4/3] w-full rounded-lg" />
                <div className="mt-3 space-y-2">
                  <Skeleton className="h-4 w-2/3" />
                  <Skeleton className="h-4 w-1/3" />
                </div>
              </div>
            </Card>
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="bg-card rounded-xl border p-10 text-center">
          <div className="text-muted-foreground mb-4 text-sm">Keine Produkte in dieser Kategorie</div>
          <Button
            variant="primary"
            onClick={() => setEdit({ id: "new", name: "", priceCents: 0, categoryId: null, stock: null })}
          >
            Produkt hinzufügen
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-5 sm:gap-3 xl:gap-5 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {filtered.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              onEdit={() =>
                setEdit({
                  id: product.id,
                  name: product.name,
                  priceCents: product.priceCents,
                  categoryId: product.category?.id ?? null,
                  stock: product.availableQuantity ?? null,
                })
              }
            />
          ))}
        </div>
      )}

      {/* Edit/Create dialog */}
      <EditDialog
        cats={cats}
        product={edit}
        onClose={() => setEdit(null)}
        onSaved={async (updated) => {
          setSaving(true)
          try {
            await saveChanges(fetchAuth, items, updated)
            await refetch()
          } catch (e: unknown) {
            setError(e instanceof Error ? e.message : "Speichern fehlgeschlagen")
          } finally {
            setSaving(false)
            setEdit(null)
          }
        }}
        saving={saving}
      />

      {error && <div className="text-destructive text-sm">{error}</div>}
    </div>
  )
}

function ProductCard({ product, onEdit }: { product: ProductDTO; onEdit: () => void }) {
  const [isConfigModalOpen, setIsConfigModalOpen] = useState(false)
  const isAvailable = product.isAvailable !== false // default true
  const isLowStock = product.isLowStock === true
  const availableQty = product.availableQuantity ?? null
  const isActive = product.isActive !== false
  const composition = useMemo(() => {
    if (product.type !== "menu" || !product.menu?.slots || product.menu.slots.length === 0) return null
    const counts = new Map<string, number>()
    for (const slot of product.menu.slots) {
      const name = slot.name?.trim() || "Slot"
      counts.set(name, (counts.get(name) || 0) + 1)
    }
    // Build minimal description like: "Burger + Beilage + 2× Getränk"
    const parts = Array.from(counts.entries()).map(([name, count]) => (count > 1 ? `${count}× ${name}` : name))
    return parts.join(" + ")
  }, [product])

  return (
    <Card className="gap-0 overflow-hidden rounded-[11px] p-0 transition-shadow hover:shadow-lg">
      <CardHeader className="p-2">
        <div className="relative aspect-video rounded-[11px] rounded-t-lg bg-[#cec9c6]">
          {product.image ? (
            <Image
              src={product.image}
              alt={"Produktbild von " + product.name}
              fill
              sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
              quality={90}
              className="h-full w-full rounded-[11px] object-cover"
            />
          ) : (
            <div className="absolute inset-0 flex items-center justify-center text-zinc-500">Kein Bild</div>
          )}
          {!isAvailable && (
            <div className="absolute inset-0 z-10 grid place-items-center rounded-[11px] bg-black/55">
              <span className="rounded-full bg-red-400 px-3 py-1 text-sm font-medium text-white">Ausverkauft</span>
            </div>
          )}
          {isAvailable && !isActive && (
            <div className="absolute inset-0 z-10 grid place-items-center rounded-[11px] bg-black/55">
              <span className="rounded-full bg-zinc-700 px-3 py-1 text-sm font-medium text-white">Nicht verfügbar</span>
            </div>
          )}
          {isLowStock && isAvailable && isActive && (
            <div className="absolute top-1 left-2 z-10">
              <span className="rounded-full bg-amber-600 px-2 py-0.5 text-xs font-medium text-white">
                {availableQty !== null ? `Nur ${availableQty} übrig` : "Geringer Bestand"}
              </span>
            </div>
          )}
        </div>
      </CardHeader>

      <CardContent className="px-2 pt-0 pb-4">
        <div className="flex items-center justify-between">
          <div className="flex flex-col">
            <h3 className="font-family-secondary text-lg">{product.name}</h3>
            <p className="font-family-secondary text-base">{formatPriceLabel(product.priceCents)}</p>
          </div>
          <div className="flex items-center">
            <Button size="icon" onClick={onEdit} aria-label={`Produkt ${product.name} bearbeiten`} className="bg-foreground text-white rounded-[10px] hover:bg-foreground/90">
              <PencilIcon className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function EditDialog({
  cats,
  product,
  onClose,
  onSaved,
  saving,
}: {
  cats: Category[]
  product: DirtyProduct | null
  onClose: () => void
  onSaved: (p: DirtyProduct) => void
  saving: boolean
}) {
  const [local, setLocal] = useState<DirtyProduct | null>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    setLocal(product)
    setErrors({})
  }, [product])

  const open = !!product
  const dirty = useMemo(() => {
    return JSON.stringify(product) !== JSON.stringify(local)
  }, [product, local])

  function validate(p: DirtyProduct | null): boolean {
    if (!p) return false
    const errs: Record<string, string> = {}
    if (!p.name.trim()) errs.name = "Name erforderlich"
    if (p.priceCents <= 0) errs.price = "Preis ungültig"
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!o) onClose()
      }}
    >
      <DialogContent className="max-w-[560px] rounded-xl" aria-labelledby="edit-title">
        <DialogHeader>
          <DialogTitle id="edit-title" className="text-sm tracking-wider">
            {product?.id === "new" ? "ERSTELLEN" : "BEARBEITEN"}
          </DialogTitle>
        </DialogHeader>
        {local && (
          <form
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault()
              if (validate(local)) onSaved(local)
            }}
          >
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={local.name}
                onChange={(e) => setLocal({ ...local, name: e.target.value })}
                aria-invalid={!!errors.name}
              />
              {errors.name && <p className="text-destructive text-xs">{errors.name}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="price">Preis</Label>
              <div className="relative">
                <Input
                  id="price"
                  inputMode="decimal"
                  defaultValue={(local.priceCents / 100).toString()}
                  onBlur={(e) => {
                    const cents = parsePriceInputToCents(e.target.value)
                    if (cents != null) setLocal({ ...local, priceCents: cents })
                  }}
                  aria-invalid={!!errors.price}
                />
                <span className="text-muted-foreground absolute top-1/2 right-3 -translate-y-1/2 text-xs">
                  {local.priceCents % 100 === 0 ? ".- CHF" : "CHF"}
                </span>
              </div>
              {errors.price && <p className="text-destructive text-xs">{errors.price}</p>}
            </div>
            <div className="space-y-2">
              <Label>Kategorie</Label>
              <Select
                value={local.categoryId ?? undefined}
                onValueChange={(v) => setLocal({ ...local, categoryId: v })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Kategorie wählen" />
                </SelectTrigger>
                <SelectContent>
                  {cats.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1">
              <Label htmlFor="stock">Bestand</Label>
              <div className="flex items-center gap-3">
                <Input
                  id="stock"
                  type="number"
                  value={local.stock ?? 0}
                  onChange={(e) => setLocal({ ...local, stock: Number(e.target.value) })}
                />
                <span className="text-muted-foreground text-xs">
                  {local.stock == null ? "Unbegrenzt" : local.stock > 0 ? "Verfügbar" : "Ausverkauft"}
                </span>
              </div>
            </div>
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button type="button" variant="outline" onClick={onClose} disabled={saving}>
                Abbrechen
              </Button>
              <Button type="submit" variant="primary" disabled={!dirty || saving}>
                {" "}
                {saving && <RefreshCw className="size-4 animate-spin" />}{" "}
                {product?.id === "new" ? "Erstellen" : "Speichern"}
              </Button>
            </div>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}

async function saveChanges(fetchAuth: ReturnType<typeof useAuthorizedFetch>, items: ProductDTO[], p: DirtyProduct) {
  const csrf = getCSRFCookie()
  const original = items.find((it) => it.id === p.id)

  // Create not implemented yet (depends on backend). Placeholder.
  if (p.id === "new") return

  // Price
  if (original && original.priceCents !== p.priceCents) {
    const res = await fetchAuth(
      `${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(p.id)}/price`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({ priceCents: p.priceCents }),
      }
    )
    if (!res.ok) throw new Error("Preis aktualisieren fehlgeschlagen")
  }

  // Category
  if (original && (original.category?.id ?? null) !== (p.categoryId ?? null)) {
    const res = await fetchAuth(
      `${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(p.id)}/category`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({ categoryId: p.categoryId }),
      }
    )
    if (!res.ok) throw new Error("Kategorie aktualisieren fehlgeschlagen")
  }

  // Inventory (adjust delta)
  const currentStock = original?.availableQuantity ?? null
  if (currentStock != null && p.stock != null && currentStock !== p.stock) {
    const delta = p.stock - currentStock
    if (delta !== 0) {
      const res = await fetchAuth(
        `${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(p.id)}/inventory-adjust`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
          body: JSON.stringify({ delta, reason: "manual_adjust" }),
        }
      )
      if (!res.ok) throw new Error("Bestand aktualisieren fehlgeschlagen")
    }
  }
}

function getCSRFCookie(): string | null {
  if (typeof document === "undefined") return null
  const name = (document.location.protocol === "https:" ? "__Host-" : "") + "csrf"
  const m = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/([.$?*|{}()\[\]\\/+^])/g, "\\$1") + "=([^;]*)"))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}
