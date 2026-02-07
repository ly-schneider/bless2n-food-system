"use client"

import { useEffect, useState } from "react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
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
  { label: "Gelb", hex: "#FACC15" },
  { label: "Blau", hex: "#3B82F6" },
  { label: "Rot", hex: "#EF4444" },
  { label: "Grün", hex: "#22C55E" },
  { label: "Lila", hex: "#A855F7" },
  { label: "Orange", hex: "#F97316" },
  { label: "Grau", hex: "#6B7280" },
]

const hexPattern = /^#?[0-9a-fA-F]{6}$/
const NO_JETON_VALUE = "__none__"

export default function AdminJetonsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [jetons, setJetons] = useState<Jeton[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [formName, setFormName] = useState("")
  const [formColor, setFormColor] = useState("#FACC15")
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState<Jeton | null>(null)
  const [products, setProducts] = useState<ProductSummaryDTO[]>([])
  const [assignError, setAssignError] = useState<string | null>(null)
  const [productLoading, setProductLoading] = useState(false)
  const [savingProducts, setSavingProducts] = useState<Set<string>>(new Set())
  const [deleteTarget, setDeleteTarget] = useState<Jeton | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const res = await fetchAuth(`/api/v1/jetons`)
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
        const res = await fetchAuth(`/api/v1/products`)
        if (!res.ok) throw new Error(await readErrorMessage(res))
        const data = (await res.json()) as { items: ProductSummaryDTO[] }
        if (cancelled) return
        setProducts((data.items || []).filter((p) => p.type === "simple"))
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
    setFormColor("#FACC15")
    setFormError(null)
    setDialogOpen(true)
  }

  function openEdit(j: Jeton) {
    setEditing(j)
    setFormName(j.name)
    setFormColor(j.color)
    setFormError(null)
    setDialogOpen(true)
  }

  async function saveJeton() {
    if (!formName.trim()) {
      setFormError("Name darf nicht leer sein.")
      return
    }
    if (!formColor.trim()) {
      setFormError("Bitte eine Farbe wählen oder einen HEX-Wert eingeben.")
      return
    }
    if (!hexPattern.test(formColor)) {
      setFormError("HEX-Farbe muss sechsstellig sein.")
      return
    }
    setSaving(true)
    setFormError(null)
    try {
      const csrf = getCSRFToken()
      const color = formColor.startsWith("#") ? formColor : `#${formColor}`
      const payload = { name: formName.trim(), color }
      if (editing) {
        const res = await fetchAuth(`/api/v1/jetons/${editing.id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
          body: JSON.stringify(payload),
        })
        if (!res.ok) throw new Error(`Speichern fehlgeschlagen (${res.status})`)
      } else {
        const res = await fetchAuth(`/api/v1/jetons`, {
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
      const res = await fetchAuth(`/api/v1/jetons/${id}`, {
        method: "DELETE",
        headers: { "X-CSRF": csrf || "" },
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as {
          code?: string
          message?: string
          details?: { usage?: number }
        }
        if (j.code === "jeton_in_use") {
          const count = j.details?.usage
          throw new Error(
            typeof count === "number"
              ? `Dieser Jeton ist noch ${count} Produkt${count === 1 ? "" : "en"} zugewiesen. Bitte entferne zuerst die Zuweisung${count === 1 ? "" : "en"}.`
              : "Dieser Jeton ist noch Produkten zugewiesen. Bitte entferne zuerst die Zuweisungen."
          )
        }
        throw new Error(j.message || `Löschen fehlgeschlagen (${res.status})`)
      }
      await load()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Löschen fehlgeschlagen"
      setError(msg)
    }
  }

  async function assignJeton(productId: string, jetonId: string | null) {
    setSavingProducts((prev) => new Set(prev).add(productId))
    setAssignError(null)
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/products/${encodeURIComponent(productId)}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({ jetonId }),
      })
      if (!res.ok) throw new Error(await readErrorMessage(res))
      setProducts((prev) =>
        prev.map((p) => {
          if (p.id !== productId) return p
          const nextJeton = jetonId ? jetons.find((j) => j.id === jetonId) : undefined
          return { ...p, jeton: nextJeton }
        })
      )
      void load()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Speichern fehlgeschlagen"
      setAssignError(msg)
    } finally {
      setSavingProducts((prev) => {
        const next = new Set(prev)
        next.delete(productId)
        return next
      })
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
                <div className="h-10 w-10 rounded-full border" style={{ backgroundColor: j.color }} aria-hidden />
                <div>
                  <div className="font-semibold">{j.name}</div>
                  <div className="text-muted-foreground text-xs">{j.color}</div>
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
                onClick={() => setDeleteTarget(j)}
              >
                Löschen
              </Button>
            </div>
          </div>
        ))}
      </div>

      <Card className="rounded-2xl">
        <CardHeader className="flex flex-col gap-2">
          <CardTitle>Jeton-Zuweisungen</CardTitle>
          <p className="text-muted-foreground text-sm">
            Änderungen werden sofort gespeichert. Menüs verwenden automatisch die Jetons ihrer Einzelprodukte und
            erscheinen hier nicht.
          </p>
          {assignError && <div className="text-destructive text-sm">{assignError}</div>}
        </CardHeader>
        <CardContent className="space-y-3">
          {productLoading && <div className="text-muted-foreground text-sm">Produkte werden geladen…</div>}
          {!productLoading && products.length === 0 && (
            <div className="text-muted-foreground text-sm">Keine Produkte gefunden.</div>
          )}
          {products.length > 0 && (
            <div className="rounded-xl border">
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
                  {products.map((p) => (
                    <TableRow key={p.id}>
                      <TableCell className="font-medium">{p.name}</TableCell>
                      <TableCell>{p.category?.name ?? "–"}</TableCell>
                      <TableCell className="text-xs uppercase">{p.isActive ? "Aktiv" : "Inaktiv"}</TableCell>
                      <TableCell>
                        <Select
                          value={(p.jeton?.id ?? "") || NO_JETON_VALUE}
                          disabled={savingProducts.has(p.id)}
                          onValueChange={(jid) => {
                            void assignJeton(p.id, jid === NO_JETON_VALUE ? null : jid)
                          }}
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
                                  style={{ backgroundColor: j.color }}
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
                  const active = formColor.toUpperCase() === p.hex.toUpperCase()
                  return (
                    <button
                      key={p.hex}
                      type="button"
                      onClick={() => setFormColor(p.hex)}
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
              <Label htmlFor="hex">Eigene HEX-Farbe</Label>
              <Input id="hex" placeholder="#FFAA00" value={formColor} onChange={(e) => setFormColor(e.target.value)} />
              <p className="text-muted-foreground text-xs">Palette-Auswahl oder manuell eingeben.</p>
            </div>
            <div className="grid gap-2">
              <Label>Vorschau</Label>
              <div className="flex items-center gap-3 rounded-md border px-3 py-3">
                <div className="h-10 w-10 rounded-full border" style={{ backgroundColor: formColor }} />
                <span className="text-sm">{formColor}</span>
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

      <AlertDialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Jeton löschen</AlertDialogTitle>
            <AlertDialogDescription>
              Möchtest du den Jeton &ldquo;{deleteTarget?.name}&rdquo; wirklich löschen? Diese Aktion kann nicht
              rückgängig gemacht werden.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={() => {
                if (deleteTarget) void deleteJeton(deleteTarget.id)
                setDeleteTarget(null)
              }}
            >
              Löschen
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
