"use client"
import Image from "next/image"
import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"

import { getCSRFToken } from "@/lib/csrf"
import { readErrorMessage } from "@/lib/http"
import type { Jeton, PosFulfillmentMode } from "@/types/jeton"

type Product = {
  id: string
  name: string
  priceCents: number
  isActive: boolean
  image?: string | null
  category?: { id: string; name: string }
  availableQuantity?: number | null
  isLowStock?: boolean
  jeton?: Jeton
}
type Category = { id: string; name: string; isActive: boolean }
const NO_JETON_VALUE = "__none__"

export default function AdminProductsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [q, setQ] = useState("")
  const [debouncedQ, setDebouncedQ] = useState("")
  const [page, setPage] = useState(0)
  const [items, setItems] = useState<Product[]>([])
  const [count, setCount] = useState(0)
  const [, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [cats, setCats] = useState<Category[]>([])
  const [jetons, setJetons] = useState<Jeton[]>([])
  const [posMode, setPosMode] = useState<PosFulfillmentMode>("QR_CODE")
  const [bulkDraft, setBulkDraft] = useState<Record<string, string>>({})
  const [bulkSaving, setBulkSaving] = useState(false)
  const missingActiveJetons = useMemo(() => items.filter((p) => p.isActive && !p.jeton).length, [items])

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
        const res = await fetchAuth(`/api/v1/products?limit=${limit}&offset=${offset}`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: Product[]; count: number }
        if (cancelled) return
        const filtered = debouncedQ
          ? data.items.filter((p) => p.name.toLowerCase().includes(debouncedQ.toLowerCase()))
          : data.items
        setItems(filtered)
        setCount(data.count)
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Produkte laden fehlgeschlagen"
        if (!cancelled) setError(msg)
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth, page, debouncedQ])

  useEffect(() => {
    ;(async () => {
      try {
        const res = await fetchAuth(`/api/v1/admin/pos/jetons`)
        if (res.ok) {
          const j = (await res.json()) as { items?: Jeton[] }
          setJetons(j.items || [])
        }
      } catch {
        // ignore jeton load errors here
      }
    })()
  }, [fetchAuth])

  useEffect(() => {
    ;(async () => {
      try {
        const res = await fetchAuth(`/api/v1/admin/pos/settings`)
        if (res.ok) {
          const j = (await res.json()) as { mode?: PosFulfillmentMode }
          setPosMode((j.mode as PosFulfillmentMode) || "QR_CODE")
        }
      } catch {
        // ignore
      }
    })()
  }, [fetchAuth])

  // Load categories once
  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const cr = await fetchAuth(`/api/v1/admin/categories`)
        if (cr.ok) {
          const d = (await cr.json()) as { items: Category[] }
          if (!cancelled) setCats(d.items || [])
        }
      } catch {}
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  useEffect(() => {
    const draft: Record<string, string> = {}
    for (const p of items) {
      draft[p.id] = p.jeton?.id ?? ""
    }
    setBulkDraft(draft)
  }, [items])

  async function updatePrice(id: string, priceCents: number) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/admin/products/${encodeURIComponent(id)}/price`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ priceCents }),
    })
    if (!res.ok) throw new Error(await readErrorMessage(res))
  }
  async function moveCategory(id: string, categoryId: string) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/admin/products/${encodeURIComponent(id)}/category`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ categoryId }),
    })
    if (!res.ok) throw new Error("Kategorie verschieben fehlgeschlagen")
  }

  async function setActive(id: string, isActive: boolean) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/admin/products/${encodeURIComponent(id)}/active`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ isActive }),
    })
    if (!res.ok) {
      const msg = await readErrorMessage(res)
      if (msg === "jeton_required") throw new Error("Bitte zuerst einen Jeton zuweisen.")
      throw new Error(msg)
    }
  }

  async function deleteHard(id: string) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/admin/products/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: { "X-CSRF": csrf || "" },
    })
    if (!res.ok) throw new Error("Produkt löschen fehlgeschlagen")
  }

  async function adjustInventory(id: string, delta: number) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/admin/products/${encodeURIComponent(id)}/inventory-adjust`, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ delta, reason: "manual_adjust" }),
    })
    if (!res.ok) throw new Error(await readErrorMessage(res))
  }

  async function updateJeton(id: string, jetonId: string | null) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/admin/products/${encodeURIComponent(id)}/jeton`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ jetonId }),
    })
    if (!res.ok) {
      const msg = await readErrorMessage(res)
      if (msg === "jeton_required") throw new Error("Im Jeton-Modus ist ein Jeton Pflicht.")
      throw new Error(msg)
    }
  }

  async function saveBulkJetons() {
    const changes = Object.entries(bulkDraft).filter(([pid, jetonId]) => {
      const current = items.find((p) => p.id === pid)
      const currentId = current?.jeton?.id ?? ""
      return currentId !== jetonId
    })
    if (changes.length === 0) return
    setBulkSaving(true)
    setError(null)
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/admin/products/jetons/bulk`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({
          assignments: changes.map(([productId, jetonId]) => ({ productId, jetonId: jetonId || null })),
        }),
      })
      if (!res.ok) throw new Error(await readErrorMessage(res))
      setItems((prev) =>
        prev.map((p) => {
          const nextId = bulkDraft[p.id]
          if (nextId === undefined) return p
          const nextJeton = jetons.find((j) => j.id === nextId)
          return { ...p, jeton: nextJeton }
        })
      )
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Bulk-Zuweisung fehlgeschlagen"
      setError(msg)
    } finally {
      setBulkSaving(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-2">
        <h1 className="text-xl font-semibold">Produkte</h1>
        <div>
          <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Produkte suchen…" className="h-8 w-56" />
        </div>
      </div>
      {error && <div className="text-sm text-red-600">{error}</div>}
      <div className="grid grid-cols-2 gap-3 md:gap-5 lg:grid-cols-3 xl:grid-cols-4">
        {items.map((p) => {
          const isAvailable = p.availableQuantity == null || (p.availableQuantity ?? 0) > 0
          return (
            <Card key={p.id} className="gap-0 overflow-hidden rounded-[11px] p-0">
              <CardHeader className="p-2">
                <div className="relative aspect-video rounded-[11px] rounded-t-lg bg-[#cec9c6]">
                  {p.image ? (
                    <Image
                      src={p.image}
                      alt={"Produktbild von " + p.name}
                      fill
                      sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
                      quality={90}
                      className="h-full w-full rounded-[11px] object-cover"
                    />
                  ) : (
                    <div className="absolute inset-0 flex items-center justify-center text-zinc-500">Kein Bild</div>
                  )}
                  {p.isLowStock && isAvailable && (
                    <div className="absolute top-1 left-2 z-10">
                      <span className="rounded-full bg-amber-600 px-2 py-0.5 text-xs font-medium text-white">
                        {typeof p.availableQuantity === "number"
                          ? `Nur ${p.availableQuantity} übrig`
                          : "Geringer Bestand"}
                      </span>
                    </div>
                  )}
                  {!p.isActive && (
                    <div className="absolute inset-0 z-10 grid place-items-center rounded-[11px] bg-black/55">
                      <span className="rounded-full bg-zinc-700 px-3 py-1 text-sm font-medium text-white">
                        Nicht verfügbar
                      </span>
                    </div>
                  )}
                </div>
              </CardHeader>
              <CardContent className="space-y-2 px-2 pt-0 pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <h3 className="text-base font-medium">{p.name}</h3>
                    <InlinePrice
                      value={p.priceCents}
                      onSave={async (v) => {
                        await updatePrice(p.id, v)
                        p.priceCents = v
                        setItems([...items])
                      }}
                    />
                  </div>
                  <label className="inline-flex items-center gap-2 text-sm">
                    <Switch
                      checked={p.isActive}
                      onCheckedChange={async (v) => {
                        try {
                          if (v && posMode === "JETON" && !p.jeton) {
                            setError("Bitte zuerst einen Jeton zuweisen.")
                            return
                          }
                          await setActive(p.id, v)
                          p.isActive = v
                          setItems([...items])
                        } catch (e: unknown) {
                          const msg = e instanceof Error ? e.message : "Aktualisierung fehlgeschlagen"
                          setError(msg)
                        }
                      }}
                    />
                    <span>{p.isActive ? "Aktiv" : "Inaktiv"}</span>
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Select
                    value={p.category?.id ?? undefined}
                    onValueChange={async (cid) => {
                      await moveCategory(p.id, cid)
                      const c = cats.find((c) => c.id === cid)
                      p.category = c ? { id: c.id, name: c.name } : undefined
                      setItems([...items])
                    }}
                  >
                    <SelectTrigger className="h-8 w-48">
                      <SelectValue placeholder={p.category ? undefined : "Keine Kategorie"} />
                    </SelectTrigger>
                    <SelectContent>
                      {cats.map((c) => (
                        <SelectItem key={c.id} value={c.id}>
                          {c.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <InlineInventory
                    value={p.availableQuantity ?? null}
                    onAdjust={async (d) => {
                      await adjustInventory(p.id, d)
                    }}
                  />
                </div>
                <div className="flex items-center gap-2">
                  <Select
                    value={p.jeton?.id ?? NO_JETON_VALUE}
                    onValueChange={async (jid) => {
                      const nextId = jid === NO_JETON_VALUE ? "" : jid
                      try {
                        if (nextId === "" && posMode === "JETON" && p.isActive) {
                          setError("Im Jeton-Modus benötigen aktive Produkte einen Jeton.")
                          return
                        }
                        await updateJeton(p.id, nextId || null)
                        const selected = nextId ? jetons.find((j) => j.id === nextId) : undefined
                        p.jeton = selected || undefined
                        setItems([...items])
                        setBulkDraft((prev) => ({ ...prev, [p.id]: nextId }))
                      } catch (e: unknown) {
                        const msg = e instanceof Error ? e.message : "Zuweisung fehlgeschlagen"
                        setError(msg)
                      }
                    }}
                  >
                    <SelectTrigger className="h-8 w-56">
                      <SelectValue placeholder="Jeton wählen" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={NO_JETON_VALUE}>Kein Jeton</SelectItem>
                      {jetons.map((j) => (
                        <SelectItem key={j.id} value={j.id}>
                          <span
                            aria-hidden
                            className="mr-2 inline-block h-3 w-3 rounded-full"
                            style={{ backgroundColor: j.colorHex }}
                          />
                          {j.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {posMode === "JETON" && p.isActive && !p.jeton && (
                    <span className="text-destructive text-xs">Pflicht im Jeton-Modus</span>
                  )}
                </div>
                <div className="flex items-center justify-end gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-7 px-2"
                    onClick={async () => {
                      try {
                        if (!confirm("Dieses Produkt deaktivieren?")) return
                        await setActive(p.id, false)
                        p.isActive = false
                        setItems([...items])
                      } catch (e: unknown) {
                        const msg = e instanceof Error ? e.message : "Deaktivieren fehlgeschlagen"
                        setError(msg)
                      }
                    }}
                  >
                    Deaktivieren
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 px-2 text-red-700"
                    onClick={async () => {
                      try {
                        if (
                          !confirm(
                            "Dieses Produkt dauerhaft löschen? Dieser Vorgang kann nicht rückgängig gemacht werden."
                          )
                        )
                          return
                        await deleteHard(p.id)
                        setItems(items.filter((i) => i.id !== p.id))
                      } catch (e: unknown) {
                        const msg = e instanceof Error ? e.message : "Löschen fehlgeschlagen"
                        setError(msg)
                      }
                    }}
                  >
                    Dauerhaft löschen
                  </Button>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>
      <div className="flex items-center justify-between text-sm">
        <div>{count} insgesamt</div>
        <div className="flex gap-2">
          <button
            disabled={page === 0}
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            className="rounded border px-2 py-1 disabled:opacity-50"
          >
            Zurück
          </button>
          <button onClick={() => setPage((p) => p + 1)} className="rounded border px-2 py-1">
            Weiter
          </button>
        </div>
      </div>
      <div className="mt-6 space-y-3">
        <div className="flex items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold">Jetons zuweisen</h2>
            <p className="text-muted-foreground text-sm">
              Bulk-Zuweisung für schnelle Änderungen. Aktueller Modus: {posMode === "JETON" ? "Jetons" : "QR-Code"}.
            </p>
          </div>
          <Button size="sm" onClick={saveBulkJetons} disabled={bulkSaving || jetons.length === 0}>
            {bulkSaving ? "Speichern…" : "Bulk speichern"}
          </Button>
        </div>
        {posMode === "JETON" && missingActiveJetons > 0 && (
          <div className="text-destructive text-sm">
            {missingActiveJetons} aktive Produkte benötigen noch einen Jeton.
          </div>
        )}
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Produkt</TableHead>
                <TableHead>Kategorie</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Jeton</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell>{p.category?.name ?? "–"}</TableCell>
                  <TableCell className="text-xs uppercase">{p.isActive ? "Aktiv" : "Inaktiv"}</TableCell>
                  <TableCell>
                    <Select
                      value={(bulkDraft[p.id] ?? p.jeton?.id ?? "") || NO_JETON_VALUE}
                      onValueChange={(jid) =>
                        setBulkDraft((prev) => ({ ...prev, [p.id]: jid === NO_JETON_VALUE ? "" : jid }))
                      }
                    >
                      <SelectTrigger className="h-9 w-56">
                        <SelectValue placeholder="Jeton wählen" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value={NO_JETON_VALUE}>Kein Jeton</SelectItem>
                        {jetons.map((j) => (
                          <SelectItem key={j.id} value={j.id}>
                            <span
                              aria-hidden
                              className="mr-2 inline-block h-3 w-3 rounded-full"
                              style={{ backgroundColor: j.colorHex }}
                            />
                            {j.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  )
}

function InlinePrice({ value, onSave }: { value: number; onSave: (v: number) => Promise<void> }) {
  const [editing, setEditing] = useState(false)
  const [val, setVal] = useState((value / 100).toFixed(2))
  useEffect(() => setVal((value / 100).toFixed(2)), [value])
  if (!editing)
    return (
      <button className="underline decoration-dotted" onClick={() => setEditing(true)}>
        {new Intl.NumberFormat("de-CH", { style: "currency", currency: "CHF" }).format(value / 100)}
      </button>
    )
  return (
    <form
      className="flex items-center gap-1"
      onSubmit={async (e) => {
        e.preventDefault()
        const cents = Math.round(parseFloat(val) * 100)
        await onSave(cents)
        setEditing(false)
      }}
    >
      <Input className="h-8 w-24" value={val} onChange={(e) => setVal(e.target.value)} />
      <Button variant="outline" size="sm" className="h-7" type="submit">
        Speichern
      </Button>
      <Button variant="ghost" size="sm" className="h-7 text-gray-600" type="button" onClick={() => setEditing(false)}>
        Abbrechen
      </Button>
    </form>
  )
}

function InlineInventory({ value, onAdjust }: { value: number | null; onAdjust: (delta: number) => Promise<void> }) {
  const [delta, setDelta] = useState(0)
  const label = typeof value === "number" ? `${value}` : "—"
  return (
    <div className="flex items-center gap-2">
      <span>{label}</span>
      <form
        className="flex items-center gap-1"
        onSubmit={async (e) => {
          e.preventDefault()
          if (!Number.isFinite(delta) || delta === 0) return
          await onAdjust(delta)
          setDelta(0)
        }}
      >
        <Input
          type="number"
          className="h-8 w-20"
          value={String(delta)}
          onChange={(e) => setDelta(parseInt(e.target.value || "0", 10))}
        />
        <Button variant="outline" size="sm" className="h-7" type="submit">
          Anwenden
        </Button>
      </form>
    </div>
  )
}

// CSRF helper now centralized in lib/csrf
