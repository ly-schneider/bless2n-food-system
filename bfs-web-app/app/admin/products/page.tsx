"use client"
import Image from "next/image"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"

type Product = {
  id: string
  name: string
  priceCents: number
  isActive: boolean
  image?: string | null
  category?: { id: string; name: string }
  availableQuantity?: number | null
  isLowStock?: boolean
}
type Category = { id: string; name: string; isActive: boolean }

export default function AdminProductsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [q, setQ] = useState("")
  const [debouncedQ, setDebouncedQ] = useState("")
  const [page, setPage] = useState(0)
  const [items, setItems] = useState<Product[]>([])
  const [count, setCount] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [cats, setCats] = useState<Category[]>([])

  const limit = 50
  const offset = page * limit

  // Debounce search input to reduce requests
  useEffect(() => {
    const t = setTimeout(() => setDebouncedQ(q.trim()), 300)
    return () => clearTimeout(t)
  }, [q])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    ;(async () => {
      try {
        const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/products?limit=${limit}&offset=${offset}`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: Product[]; count: number }
        if (cancelled) return
        const filtered = debouncedQ ? data.items.filter((p) => p.name.toLowerCase().includes(debouncedQ.toLowerCase())) : data.items
        setItems(filtered)
        setCount(data.count)
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Produkte laden fehlgeschlagen"
        if (!cancelled) setError(msg)
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => { cancelled = true }
  }, [fetchAuth, page, debouncedQ])

  // Load categories once
  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const cr = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories`)
        if (cr.ok) { const d = await cr.json() as { items: Category[] }; if (!cancelled) setCats(d.items || []) }
      } catch {}
    })()
    return () => { cancelled = true }
  }, [fetchAuth])

  async function updatePrice(id: string, priceCents: number) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(id)}/price`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ priceCents }),
    })
    if (!res.ok) throw new Error(await readErrorMessage(res))
  }
  async function moveCategory(id: string, categoryId: string) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(id)}/category`, {
      method: 'PATCH', headers: { 'Content-Type':'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ categoryId })
    })
    if (!res.ok) throw new Error('Kategorie verschieben fehlgeschlagen')
  }

  async function setActive(id: string, isActive: boolean) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(id)}/active`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ isActive }),
    })
    if (!res.ok) throw new Error(await readErrorMessage(res))
  }

  async function deleteHard(id: string) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(id)}`, {
      method: 'DELETE', headers: { 'X-CSRF': csrf || '' }
    })
    if (!res.ok) throw new Error('Produkt löschen fehlgeschlagen')
  }

  async function adjustInventory(id: string, delta: number) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/products/${encodeURIComponent(id)}/inventory-adjust`, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ delta, reason: "manual_adjust" }),
    })
    if (!res.ok) throw new Error(await readErrorMessage(res))
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-2">
        <h1 className="text-xl font-semibold">Produkte</h1>
        <div>
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Produkte suchen…"
            className="h-8 w-56"
          />
        </div>
      </div>
      {error && <div className="text-red-600 text-sm">{error}</div>}
      <div className="grid gap-3 md:gap-5 grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {items.map((p) => {
          const isAvailable = p.availableQuantity == null || (p.availableQuantity ?? 0) > 0
          return (
            <Card key={p.id} className="overflow-hidden p-0 rounded-[11px] gap-0">
              <CardHeader className="p-2">
                <div className="relative aspect-video rounded-t-lg bg-[#cec9c6] rounded-[11px]">
                  {p.image ? (
                    <Image
                      src={p.image}
                      alt={"Produktbild von " + p.name}
                      fill
                      sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
                      quality={90}
                      className="w-full h-full object-cover rounded-[11px]"
                    />
                  ) : (
                    <div className="absolute inset-0 flex items-center justify-center text-zinc-500">Kein Bild</div>
                  )}
                  {p.isLowStock && isAvailable && (
                    <div className="absolute top-1 left-2 z-10">
                      <span className="px-2 py-0.5 text-xs font-medium text-white bg-amber-600 rounded-full">
                        {typeof p.availableQuantity === 'number' ? `Nur ${p.availableQuantity} übrig` : 'Geringer Bestand'}
                      </span>
                    </div>
                  )}
                  {!p.isActive && (
                    <div className="absolute inset-0 z-10 grid place-items-center bg-black/55 rounded-[11px]">
                      <span className="px-3 py-1 text-sm font-medium text-white bg-zinc-700 rounded-full">Nicht verfügbar</span>
                    </div>
                  )}
                </div>
              </CardHeader>
              <CardContent className="px-2 pt-0 pb-3 space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <h3 className="text-base font-medium">{p.name}</h3>
                    <InlinePrice value={p.priceCents} onSave={async (v) => { await updatePrice(p.id, v); p.priceCents = v; setItems([...items]) }} />
                  </div>
                  <label className="inline-flex items-center gap-2 text-sm">
                    <Switch checked={p.isActive} onCheckedChange={async (v) => { await setActive(p.id, v); p.isActive = v; setItems([...items]) }} />
                    <span>{p.isActive ? "Aktiv" : "Inaktiv"}</span>
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Select
                    value={p.category?.id ?? undefined}
                    onValueChange={async (cid) => {
                      await moveCategory(p.id, cid)
                      const c = cats.find(c => c.id === cid)
                      p.category = c ? { id: c.id, name: c.name } : undefined
                      setItems([...items])
                    }}
                  >
                    <SelectTrigger className="h-8 w-48">
                      <SelectValue placeholder={p.category ? undefined : "Keine Kategorie"} />
                    </SelectTrigger>
                    <SelectContent>
                      {cats.map(c => (
                        <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <InlineInventory value={p.availableQuantity ?? null} onAdjust={async (d) => { await adjustInventory(p.id, d) }} />
                </div>
                <div className="flex items-center justify-end gap-2">
                  <Button variant="outline" size="sm" className="h-7 px-2" onClick={async () => { try { if (!confirm('Dieses Produkt deaktivieren?')) return; await setActive(p.id, false); p.isActive = false; setItems([...items]) } catch(e: unknown) { const msg = e instanceof Error ? e.message : 'Deaktivieren fehlgeschlagen'; setError(msg) } }}>Deaktivieren</Button>
                  <Button variant="ghost" size="sm" className="h-7 px-2 text-red-700" onClick={async () => { try { if (!confirm('Dieses Produkt dauerhaft löschen? Dieser Vorgang kann nicht rückgängig gemacht werden.')) return; await deleteHard(p.id); setItems(items.filter(i => i.id !== p.id)) } catch(e: unknown) { const msg = e instanceof Error ? e.message : 'Löschen fehlgeschlagen'; setError(msg) } }}>Dauerhaft löschen</Button>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>
      <div className="flex items-center justify-between text-sm">
        <div>{count} insgesamt</div>
        <div className="flex gap-2">
          <button disabled={page === 0} onClick={() => setPage((p) => Math.max(0, p - 1))} className="border px-2 py-1 rounded disabled:opacity-50">Zurück</button>
          <button onClick={() => setPage((p) => p + 1)} className="border px-2 py-1 rounded">Weiter</button>
        </div>
      </div>
    </div>
  )
}

function InlinePrice({ value, onSave }: { value: number; onSave: (v: number) => Promise<void> }) {
  const [editing, setEditing] = useState(false)
  const [val, setVal] = useState((value / 100).toFixed(2))
  useEffect(() => setVal((value / 100).toFixed(2)), [value])
  if (!editing) return <button className="underline decoration-dotted" onClick={() => setEditing(true)}>{new Intl.NumberFormat("de-CH", { style: "currency", currency: "CHF" }).format(value / 100)}</button>
  return (
    <form className="flex items-center gap-1" onSubmit={async (e) => { e.preventDefault(); const cents = Math.round(parseFloat(val) * 100); await onSave(cents); setEditing(false) }}>
      <Input className="h-8 w-24" value={val} onChange={(e) => setVal(e.target.value)} />
      <Button variant="outline" size="sm" className="h-7" type="submit">Speichern</Button>
      <Button variant="ghost" size="sm" className="h-7 text-gray-600" type="button" onClick={() => setEditing(false)}>Abbrechen</Button>
    </form>
  )
}

function InlineInventory({ value, onAdjust }: { value: number | null; onAdjust: (delta: number) => Promise<void> }) {
  const [delta, setDelta] = useState(0)
  const label = typeof value === "number" ? `${value}` : "—"
  return (
    <div className="flex items-center gap-2">
      <span>{label}</span>
      <form className="flex items-center gap-1" onSubmit={async (e) => { e.preventDefault(); if (!Number.isFinite(delta) || delta === 0) return; await onAdjust(delta); setDelta(0) }}>
        <Input type="number" className="h-8 w-20" value={String(delta)} onChange={(e) => setDelta(parseInt(e.target.value || "0", 10))} />
        <Button variant="outline" size="sm" className="h-7" type="submit">Anwenden</Button>
      </form>
    </div>
  )
}

function getCSRFCookie(): string | null {
  if (typeof document === 'undefined') return null
  const name = (document.location.protocol === 'https:' ? '__Host-' : '') + 'csrf'
  const m = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()\[\]\\/+^])/g, '\\$1') + '=([^;]*)'))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}
