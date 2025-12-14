"use client"

import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { getCSRFToken } from "@/lib/csrf"
import { readErrorMessage } from "@/lib/http"
import type { Jeton } from "@/types/jeton"
import type { ProductSummaryDTO } from "@/types/product"

const palette = [
  { value: "yellow", label: "Gelb", hex: "#FACC15" },
  { value: "blue", label: "Blau", hex: "#3B82F6" },
  { value: "red", label: "Rot", hex: "#EF4444" },
  { value: "green", label: "Grün", hex: "#22C55E" },
  { value: "purple", label: "Lila", hex: "#A855F7" },
  { value: "orange", label: "Orange", hex: "#F97316" },
  { value: "gray", label: "Grau", hex: "#6B7280" },
]

const hexPattern = /^#?[0-9a-fA-F]{6}$/
const NO_JETON_VALUE = "__none__"

function resolveColor(paletteColor: string, hexColor?: string | null) {
  if (hexColor && hexPattern.test(hexColor)) {
    return hexColor.startsWith("#") ? hexColor : `#${hexColor}`
  }
  const p = palette.find((p) => p.value === paletteColor)
  return p?.hex ?? "#9CA3AF"
}

export default function AdminJetonsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [jetons, setJetons] = useState<Jeton[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [formName, setFormName] = useState("")
  const [formPalette, setFormPalette] = useState("yellow")
  const [formHex, setFormHex] = useState("")
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState<Jeton | null>(null)
  const [products, setProducts] = useState<ProductSummaryDTO[]>([])
  const [assignDraft, setAssignDraft] = useState<Record<string, string>>({})
  const [assignSaving, setAssignSaving] = useState(false)
  const [assignError, setAssignError] = useState<string | null>(null)
  const [productQuery, setProductQuery] = useState("")
  const [jetonFilter, setJetonFilter] = useState("all")
  const [productLoading, setProductLoading] = useState(false)

  const previewColor = useMemo(() => resolveColor(formPalette, formHex || undefined), [formPalette, formHex])

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const res = await fetchAuth(`/api/v1/admin/pos/jetons`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const j = (await res.json()) as { items?: Jeton[] }
      setJetons(j.items || [])
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Jetons laden fehlgeschlagen"
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [fetchAuth])

  useEffect(() => {
    let cancelled = false
    const loadProducts = async () => {
      setProductLoading(true)
      setAssignError(null)
      try {
        const res = await fetchAuth(`/api/v1/products?limit=500&offset=0`)
        if (!res.ok) throw new Error(await readErrorMessage(res))
        const data = (await res.json()) as { items: ProductSummaryDTO[] }
        if (cancelled) return
        setProducts(data.items || [])
        const draft: Record<string, string> = {}
        for (const p of data.items || []) {
          draft[p.id] = p.jeton?.id ?? ""
        }
        setAssignDraft(draft)
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Produkte laden fehlgeschlagen"
        if (!cancelled) setAssignError(msg)
      } finally {
        if (!cancelled) setProductLoading(false)
      }
    }
    loadProducts()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  function openCreate() {
    setEditing(null)
    setFormName("")
    setFormPalette("yellow")
    setFormHex("")
    setFormError(null)
    setDialogOpen(true)
  }

  function openEdit(j: Jeton) {
    setEditing(j)
    setFormName(j.name)
    setFormPalette(j.paletteColor || "yellow")
    setFormHex(j.hexColor || "")
    setFormError(null)
    setDialogOpen(true)
  }

  async function saveJeton() {
    if (!formName.trim()) {
      setFormError("Name darf nicht leer sein.")
      return
    }
    if (!formPalette && !formHex.trim()) {
      setFormError("Bitte eine Palettenfarbe oder einen HEX-Wert wählen.")
      return
    }
    if (formHex && !hexPattern.test(formHex)) {
      setFormError("HEX-Farbe muss sechsstellig sein.")
      return
    }
    setSaving(true)
    setFormError(null)
    try {
      const csrf = getCSRFToken()
      const payload = { name: formName.trim(), paletteColor: formPalette, hexColor: formHex ? formHex : undefined }
      if (editing) {
        const res = await fetchAuth(`/api/v1/admin/pos/jetons/${editing.id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
          body: JSON.stringify(payload),
        })
        if (!res.ok) throw new Error(`Speichern fehlgeschlagen (${res.status})`)
      } else {
        const res = await fetchAuth(`/api/v1/admin/pos/jetons`, {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
          body: JSON.stringify(payload),
        })
        if (!res.ok) throw new Error(`Erstellen fehlgeschlagen (${res.status})`)
      }
      await load()
      setDialogOpen(false)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Speichern fehlgeschlagen"
      setFormError(msg)
    } finally {
      setSaving(false)
    }
  }

  async function deleteJeton(id: string) {
    const csrf = getCSRFToken()
    try {
      const res = await fetchAuth(`/api/v1/admin/pos/jetons/${id}`, {
        method: "DELETE",
        headers: { "X-CSRF": csrf || "" },
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as { error?: string; usage?: number }
        if (j.error === "jeton_in_use") {
          throw new Error(
            `Jeton ist noch ${typeof j.usage === "number" ? j.usage : "bei Produkten"} zugewiesen und kann nicht gelöscht werden.`
          )
        }
        throw new Error(`Löschen fehlgeschlagen (${res.status})`)
      }
      await load()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Löschen fehlgeschlagen"
      setError(msg)
    }
  }

  const filteredProducts = useMemo(() => {
    const q = productQuery.trim().toLowerCase()
    return products.filter((p) => {
      const matchesQuery = !q || p.name.toLowerCase().includes(q) || p.category?.name?.toLowerCase().includes(q)
      const currentJeton = assignDraft[p.id] ?? p.jeton?.id ?? ""
      const matchesJeton =
        jetonFilter === "all" ||
        (jetonFilter === "none" && currentJeton === "") ||
        (jetonFilter !== "none" && jetonFilter === currentJeton)
      return matchesQuery && matchesJeton
    })
  }, [products, productQuery, jetonFilter, assignDraft])

  async function saveAssignments() {
    const changes = Object.entries(assignDraft).filter(([pid, jid]) => {
      const current = products.find((p) => p.id === pid)?.jeton?.id ?? ""
      return current !== jid
    })
    if (changes.length === 0) return

    setAssignSaving(true)
    setAssignError(null)
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

      const updateMap = new Map(changes.map(([pid, jid]) => [pid, jid]))
      setProducts((prev) =>
        prev.map((p) => {
          if (!updateMap.has(p.id)) return p
          const nextId = updateMap.get(p.id) || ""
          const nextJeton = jetons.find((j) => j.id === nextId)
          return { ...p, jeton: nextJeton }
        })
      )
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Speichern fehlgeschlagen"
      setAssignError(msg)
    } finally {
      setAssignSaving(false)
    }
  }

  return (
    <div className="space-y-4 pt-4">
      <div className="flex items-center justify-between">
        <h1 className="font-primary text-2xl">Jetons</h1>
        <Button onClick={openCreate}>Jeton hinzufügen</Button>
      </div>
      {error && <div className="text-destructive bg-destructive/10 rounded px-3 py-2 text-sm">{error}</div>}
      {loading && <div className="text-muted-foreground text-sm">Lädt…</div>}
      {!loading && jetons.length === 0 && (
        <div className="text-muted-foreground text-sm">Noch keine Jetons angelegt.</div>
      )}
      <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
        {jetons.map((j) => (
          <div key={j.id} className="border-border bg-card flex flex-col gap-3 rounded-[11px] border p-3">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-full border" style={{ backgroundColor: j.colorHex }} aria-hidden />
                <div>
                  <div className="font-semibold">{j.name}</div>
                  <div className="text-muted-foreground text-xs">
                    Palette: {j.paletteColor || "—"}
                    {j.hexColor ? ` • ${j.hexColor}` : ""}
                  </div>
                </div>
              </div>
              {typeof j.usageCount === "number" && (
                <span className="text-muted-foreground text-xs">{j.usageCount} Produkte</span>
              )}
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => openEdit(j)}>
                Bearbeiten
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="text-destructive"
                disabled={(j.usageCount || 0) > 0}
                onClick={() => {
                  if (!confirm("Diesen Jeton wirklich löschen?")) return
                  deleteJeton(j.id)
                }}
              >
                Löschen
              </Button>
            </div>
          </div>
        ))}
      </div>

      <Card className="rounded-2xl">
        <CardHeader className="flex flex-col gap-2">
          <CardTitle>Jeton-Zuweisungen verwalten</CardTitle>
          <p className="text-muted-foreground text-sm">
            Weise Jetons den Produkten zu oder ändere bestehende Zuweisungen. Änderungen wirken sofort, der POS bleibt
            standardmäßig im QR-Code-Modus.
          </p>
          <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
            <div className="flex flex-wrap items-center gap-2">
              <Input
                placeholder="Produkte oder Kategorien suchen…"
                value={productQuery}
                onChange={(e) => setProductQuery(e.target.value)}
                className="w-64"
              />
              <Select value={jetonFilter} onValueChange={setJetonFilter}>
                <SelectTrigger className="w-44">
                  <SelectValue placeholder="Filter" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Alle Jetons</SelectItem>
                  <SelectItem value="none">Ohne Jeton</SelectItem>
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
            </div>
            <Button onClick={saveAssignments} disabled={assignSaving || productLoading || jetons.length === 0}>
              {assignSaving ? "Speichert…" : "Zuweisungen speichern"}
            </Button>
          </div>
          {assignError && <div className="text-destructive text-sm">{assignError}</div>}
        </CardHeader>
        <CardContent className="space-y-3">
          {productLoading && <div className="text-muted-foreground text-sm">Produkte werden geladen…</div>}
          {!productLoading && filteredProducts.length === 0 && (
            <div className="text-muted-foreground text-sm">Keine Produkte gefunden.</div>
          )}
          {filteredProducts.length > 0 && (
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
                  {filteredProducts.map((p) => (
                    <TableRow key={p.id}>
                      <TableCell className="font-medium">{p.name}</TableCell>
                      <TableCell>{p.category?.name ?? "–"}</TableCell>
                      <TableCell className="text-xs uppercase">{p.isActive ? "Aktiv" : "Inaktiv"}</TableCell>
                      <TableCell>
                        <Select
                          value={(assignDraft[p.id] ?? p.jeton?.id ?? "") || NO_JETON_VALUE}
                          onValueChange={(jid) =>
                            setAssignDraft((prev) => ({ ...prev, [p.id]: jid === NO_JETON_VALUE ? "" : jid }))
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
          )}
        </CardContent>
      </Card>

      <Dialog
        open={dialogOpen}
        onOpenChange={(v) => {
          setDialogOpen(v)
          if (!v) setFormError(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? "Jeton bearbeiten" : "Jeton erstellen"}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-3">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={formName} onChange={(e) => setFormName(e.target.value)} />
            </div>
            <div className="grid gap-2">
              <Label>Farbpalette</Label>
              <div className="grid grid-cols-3 gap-2">
                {palette.map((p) => {
                  const active = formPalette === p.value
                  return (
                    <button
                      key={p.value}
                      type="button"
                      onClick={() => setFormPalette(p.value)}
                      className={`flex items-center gap-2 rounded-md border px-2 py-2 text-sm ${
                        active ? "border-foreground" : "border-border"
                      }`}
                    >
                      <span className="h-6 w-6 rounded-full border" style={{ backgroundColor: p.hex }} aria-hidden />
                      {p.label}
                    </button>
                  )
                })}
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="hex">Eigene HEX-Farbe (optional)</Label>
              <Input id="hex" placeholder="#FFAA00" value={formHex} onChange={(e) => setFormHex(e.target.value)} />
              <p className="text-muted-foreground text-xs">Hex überschreibt die Palette, wenn gesetzt.</p>
            </div>
            <div className="grid gap-2">
              <Label>Vorschau</Label>
              <div className="flex items-center gap-3 rounded-md border px-3 py-3">
                <div className="h-10 w-10 rounded-full border" style={{ backgroundColor: previewColor }} />
                <span className="text-sm">{previewColor}</span>
              </div>
            </div>
            {formError && <div className="text-destructive text-sm">{formError}</div>}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Abbrechen
            </Button>
            <Button onClick={saveJeton} disabled={saving}>
              Speichern
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
