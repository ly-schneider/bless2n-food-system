"use client"
import { useCallback, useEffect, useMemo, useState } from "react"
import Image from "next/image"
import { Pencil, Plus, RefreshCw } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"

type Category = { id: string; name: string; isActive: boolean }
type Product = {
  id: string
  name: string
  priceCents: number
  isActive: boolean
  image?: string | null
  category?: { id: string; name: string } | null
  availableQuantity?: number | null
  isLowStock?: boolean
}

type DirtyProduct = {
  id: string
  name: string
  priceCents: number
  categoryId: string | null
  stock: number | null
}

const CATEGORY_KEYS = ["Alles", "Menus", "Burgers", "Beilagen", "Getränke"] as const
type CatKey = (typeof CATEGORY_KEYS)[number]

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
  const [items, setItems] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeCat, setActiveCat] = useState<CatKey>("Alles")
  const [edit, setEdit] = useState<DirtyProduct | null>(null)
  const [saving, setSaving] = useState(false)

  const totalCount = items.length

  const chips = useMemo(() => {
    const counts: Record<CatKey, number> = { Alles: items.length, Menus: 0, Burgers: 0, Beilagen: 0, Getränke: 0 }
    for (const it of items) {
      const n = it.category?.name || ""
      if (n.toLowerCase().includes("menu")) counts.Menus++
      else if (n.toLowerCase().includes("burger")) counts.Burgers++
      else if (n.toLowerCase().includes("beilagen")) counts.Beilagen++
      else if (n.toLowerCase().includes("getränke") || n.toLowerCase().includes("getraenke")) counts.Getränke++
    }
    return counts
  }, [items])

  const filtered = useMemo(() => {
    if (activeCat === "Alles") return items
    return items.filter((it) => {
      const n = it.category?.name?.toLowerCase() || ""
      if (activeCat === "Menus") return n.includes("menu")
      if (activeCat === "Burgers") return n.includes("burger")
      if (activeCat === "Beilagen") return n.includes("beilagen")
      if (activeCat === "Getränke") return n.includes("getränke") || n.includes("getraenke")
      return true
    })
  }, [items, activeCat])

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
        setCats(d.items || [])
      }
      if (pr.ok) {
        const d = (await pr.json()) as { items: Product[] }
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

  // Initialize and persist category filter via query param
  useEffect(() => {
    const sp = new URLSearchParams(window.location.search)
    const cat = sp.get("cat")
    if (cat && CATEGORY_KEYS.includes(cat as CatKey)) setActiveCat(cat as CatKey)
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
      {/* Category/filter chips row */}
      <h1 className="text-xl font-semibold">Menu</h1>
      <div className="flex flex-wrap items-center gap-2">
        {/* Alles with red numeric counter badge anchored on right */}
        <button
          className={`relative h-9 rounded-full border px-4 text-sm ${
            activeCat === "Alles" ? "bg-muted font-semibold" : "bg-card"
          }`}
          onClick={() => setActiveCat("Alles")}
          aria-pressed={activeCat === "Alles"}
        >
          Alles
          <span
            className="bg-destructive absolute -top-2 -right-2 flex h-5 min-w-5 items-center justify-center rounded-full px-1 text-xs text-white"
            aria-label={`${totalCount} Produkte`}
          >
            {totalCount}
          </span>
        </button>
        {(["Menus", "Burgers", "Beilagen", "Getränke"] as CatKey[]).map((ck) => (
          <button
            key={ck}
            className={`flex h-9 items-center gap-1 rounded-full border px-4 text-sm ${
              activeCat === ck ? "bg-muted font-semibold" : "bg-card"
            }`}
            onClick={() => setActiveCat(ck)}
            aria-pressed={activeCat === ck}
          >
            {ck} <span className="text-muted-foreground">{chips[ck]}</span>
          </button>
        ))}
        <span className="mx-1" />
        <Button
          variant="secondary"
          size="sm"
          className="rounded-full"
          onClick={() => {
            /* could toggle category edit */
          }}
        >
          <Pencil className="size-4" /> Bearbeiten
        </Button>
        <Button
          variant="primary"
          size="sm"
          className="rounded-full"
          onClick={() => setEdit({ id: "new", name: "", priceCents: 0, categoryId: null, stock: null })}
        >
          <Plus className="size-4" /> Erstellen
        </Button>
      </div>

      {/* Grid */}
      {loading ? (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
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
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {filtered.map((p) => (
            <ProductCard
              key={p.id}
              p={p}
              onEdit={() =>
                setEdit({
                  id: p.id,
                  name: p.name,
                  priceCents: p.priceCents,
                  categoryId: p.category?.id ?? null,
                  stock: p.availableQuantity ?? null,
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

function ProductCard({ p, onEdit }: { p: Product; onEdit: () => void }) {
  const unavailable = !p.isActive
  const stockBadge = typeof p.availableQuantity === "number" ? `${p.availableQuantity} verfügbar` : null
  return (
    <Card
      className={`overflow-hidden rounded-xl border transition-shadow ${!unavailable ? "hover:shadow" : "opacity-60"}`}
    >
      <div className="relative">
        <div
          className={`bg-muted/60 flex aspect-[4/3] w-full items-center justify-center overflow-hidden ${
            unavailable ? "grayscale" : ""
          }`}
        >
          {p.image ? (
            <Image src={p.image} alt={p.name} width={400} height={300} className="h-full w-full object-cover" />
          ) : (
            <div className="text-muted-foreground text-xs">Kein Bild</div>
          )}
        </div>
        {stockBadge && (
          <span
            className="bg-primary text-primary-foreground absolute top-2 left-2 rounded-full px-2 py-0.5 text-[11px]"
            aria-label={`${p.availableQuantity ?? 0} Stück verfügbar`}
          >
            {stockBadge}
          </span>
        )}
        {!unavailable && (
          <button
            className="absolute right-2 bottom-2 rounded-md bg-black/80 p-1.5 text-white"
            aria-label={`Bearbeiten ${p.name}`}
            onClick={onEdit}
          >
            <Pencil className="size-4" />
          </button>
        )}
      </div>
      <div className="p-3">
        <div className="text-sm font-medium">{p.name}</div>
        <div className="text-muted-foreground mt-1 text-sm">{formatPriceLabel(p.priceCents)}</div>
      </div>
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

async function saveChanges(fetchAuth: ReturnType<typeof useAuthorizedFetch>, items: Product[], p: DirtyProduct) {
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
