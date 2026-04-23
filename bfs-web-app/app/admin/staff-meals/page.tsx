"use client"

import { ArrowRight, Plus, Ticket, Trash2 } from "lucide-react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useCallback, useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { getCSRFToken } from "@/lib/csrf"
import { readErrorMessage } from "@/lib/http"
import type { ProductSummaryDTO } from "@/types/product"
import type { VolunteerCampaignStatus, VolunteerCampaignSummary } from "@/types/volunteer"

type ProductRow = { productId: string; quantity: number }

const STATUS_LABEL: Record<VolunteerCampaignStatus, string> = {
  draft: "Entwurf",
  active: "Aktiv",
  ended: "Beendet",
}

const STATUS_CLASS: Record<VolunteerCampaignStatus, string> = {
  draft: "bg-amber-100 text-amber-900",
  active: "bg-emerald-100 text-emerald-900",
  ended: "bg-muted text-muted-foreground",
}

export default function AdminStaffMealsPage() {
  const fetchAuth = useAuthorizedFetch()
  const router = useRouter()
  const [meals, setMeals] = useState<VolunteerCampaignSummary[]>([])
  const [products, setProducts] = useState<ProductSummaryDTO[]>([])
  const [firstLoadDone, setFirstLoadDone] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [formName, setFormName] = useState("")
  const [formSlotCount, setFormSlotCount] = useState(20)
  const [formProducts, setFormProducts] = useState<ProductRow[]>([])
  const [formValidUntil, setFormValidUntil] = useState<string>("")
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    try {
      const res = await fetchAuth("/api/v1/staff-meals")
      if (!res.ok) throw new Error(await readErrorMessage(res))
      const j = (await res.json()) as { items?: VolunteerCampaignSummary[] }
      setMeals(j.items || [])
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Laden fehlgeschlagen")
    } finally {
      setFirstLoadDone(true)
    }
  }, [fetchAuth])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetchAuth("/api/v1/products")
        if (!res.ok) return
        const j = (await res.json()) as { items?: ProductSummaryDTO[] }
        if (!cancelled) setProducts(j.items || [])
      } catch {}
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  const resetForm = () => {
    setFormName("")
    setFormSlotCount(20)
    setFormProducts([])
    setFormValidUntil("")
    setFormError(null)
  }

  function openCreate() {
    resetForm()
    setDialogOpen(true)
  }

  async function submitCreate() {
    setFormError(null)
    if (!formName.trim()) {
      setFormError("Name erforderlich.")
      return
    }
    if (formProducts.length === 0) {
      setFormError("Mindestens ein Produkt erforderlich.")
      return
    }
    if (formSlotCount < 1) {
      setFormError("Anzahl muss mindestens 1 sein.")
      return
    }
    setSaving(true)
    try {
      const res = await fetchAuth("/api/v1/staff-meals", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF": getCSRFToken() || "",
        },
        body: JSON.stringify({
          name: formName.trim(),
          validUntil: formValidUntil ? new Date(formValidUntil).toISOString() : undefined,
          slotCount: formSlotCount,
          products: formProducts,
        }),
      })
      if (!res.ok) throw new Error(await readErrorMessage(res))
      const created = (await res.json()) as VolunteerCampaignSummary
      setDialogOpen(false)
      resetForm()
      router.push(`/admin/staff-meals/${created.id}`)
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : "Erstellen fehlgeschlagen")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex flex-col gap-6 px-6 md:px-8 lg:px-10">
      <Card>
        <CardHeader className="flex-col items-start gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex flex-col gap-1">
            <CardTitle>Mitarbeiter-Essen</CardTitle>
            <p className="text-muted-foreground text-sm">
              Gratis-Mahlzeiten, die Mitarbeiter per QR-Code an der Station einlösen.
            </p>
          </div>
          <Button onClick={openCreate} className="shrink-0">
            <Plus className="size-4" aria-hidden />
            Neues Essen
          </Button>
        </CardHeader>
        <CardContent>
          {error && <div className="bg-destructive/10 text-destructive mb-4 rounded-xl p-3 text-sm">{error}</div>}

          {!firstLoadDone ? (
            <LoadingList />
          ) : meals.length === 0 ? (
            <EmptyState onCreate={openCreate} />
          ) : (
            <>
              {/* Desktop table */}
              <div className="hidden sm:block">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="text-right">Eingelöst</TableHead>
                      <TableHead className="text-right">Reserviert</TableHead>
                      <TableHead className="text-right">Gesamt</TableHead>
                      <TableHead>Code</TableHead>
                      <TableHead />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {meals.map((m) => (
                      <TableRow
                        key={m.id}
                        className="hover:bg-accent/40 cursor-pointer"
                        onClick={() => router.push(`/admin/staff-meals/${m.id}`)}
                      >
                        <TableCell className="font-medium">{m.name}</TableCell>
                        <TableCell>
                          <StatusBadge status={m.status} />
                        </TableCell>
                        <TableCell className="text-right tabular-nums">{m.redeemedSlots}</TableCell>
                        <TableCell className="text-right tabular-nums">{m.reservedSlots}</TableCell>
                        <TableCell className="text-right tabular-nums">{m.totalSlots}</TableCell>
                        <TableCell className="font-mono text-sm tracking-wider">{m.accessCode}</TableCell>
                        <TableCell className="text-right">
                          <ArrowRight className="text-muted-foreground inline size-4" aria-hidden />
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Mobile cards */}
              <div className="flex flex-col gap-2 sm:hidden">
                {meals.map((m) => (
                  <Link
                    key={m.id}
                    href={`/admin/staff-meals/${m.id}`}
                    className="hover:bg-accent/40 focus-visible:ring-ring/40 flex items-center justify-between gap-3 rounded-xl border p-4 transition focus-visible:ring-2 focus-visible:outline-none"
                  >
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{m.name}</span>
                        <StatusBadge status={m.status} />
                      </div>
                      <div className="text-muted-foreground mt-1 flex items-center gap-3 text-xs tabular-nums">
                        <span>
                          {m.redeemedSlots}/{m.totalSlots} eingelöst
                        </span>
                        {m.reservedSlots > 0 && <span>· {m.reservedSlots} reserviert</span>}
                      </div>
                    </div>
                    <div className="flex flex-col items-end gap-1">
                      <code className="font-mono text-sm tracking-wider">{m.accessCode}</code>
                      <ArrowRight className="text-muted-foreground size-4" aria-hidden />
                    </div>
                  </Link>
                ))}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Neues Mitarbeiter-Essen</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4">
            <div className="grid gap-2">
              <Label htmlFor="meal-name">Name</Label>
              <Input
                id="meal-name"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder="z.B. April BlessThun"
                autoFocus
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="grid gap-2">
                <Label htmlFor="meal-count">Anzahl Mitarbeiter</Label>
                <Input
                  id="meal-count"
                  type="number"
                  value={formSlotCount}
                  min={1}
                  onChange={(e) => setFormSlotCount(Math.max(1, parseInt(e.target.value || "0", 10) || 0))}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="meal-until">Gültig bis (optional)</Label>
                <Input
                  id="meal-until"
                  type="datetime-local"
                  value={formValidUntil}
                  onChange={(e) => setFormValidUntil(e.target.value)}
                />
              </div>
            </div>
            <div className="grid gap-2">
              <Label>Produkte pro Mitarbeiter</Label>
              <div className="flex flex-col gap-2">
                {formProducts.map((row, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    <Select
                      value={row.productId}
                      onValueChange={(v) =>
                        setFormProducts((rs) => rs.map((r, i) => (i === idx ? { ...r, productId: v } : r)))
                      }
                    >
                      <SelectTrigger className="flex-1">
                        <SelectValue placeholder="Produkt wählen …" />
                      </SelectTrigger>
                      <SelectContent>
                        {products.map((p) => (
                          <SelectItem key={p.id} value={p.id}>
                            {p.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Input
                      type="number"
                      min={1}
                      className="w-20"
                      value={row.quantity}
                      aria-label="Menge"
                      onChange={(e) =>
                        setFormProducts((rs) =>
                          rs.map((r, i) =>
                            i === idx ? { ...r, quantity: Math.max(1, parseInt(e.target.value || "0", 10) || 0) } : r
                          )
                        )
                      }
                    />
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setFormProducts((rs) => rs.filter((_, i) => i !== idx))}
                      aria-label="Produkt entfernen"
                    >
                      <Trash2 className="size-4" aria-hidden />
                    </Button>
                  </div>
                ))}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFormProducts((rs) => [...rs, { productId: products[0]?.id || "", quantity: 1 }])}
                  disabled={products.length === 0}
                >
                  <Plus className="size-4" aria-hidden />
                  Produkt hinzufügen
                </Button>
              </div>
            </div>
            {formError && (
              <div className="bg-destructive/10 text-destructive rounded-xl p-3 text-sm" role="alert">
                {formError}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)} disabled={saving}>
              Abbrechen
            </Button>
            <Button onClick={submitCreate} disabled={saving}>
              {saving ? "Erstellen …" : "Erstellen"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function StatusBadge({ status }: { status: VolunteerCampaignStatus }) {
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_CLASS[status]}`}>
      {STATUS_LABEL[status]}
    </span>
  )
}

function LoadingList() {
  return (
    <div className="flex flex-col gap-2">
      {[0, 1, 2].map((i) => (
        <Skeleton key={i} className="h-14 w-full rounded-xl" />
      ))}
    </div>
  )
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="border-border/60 flex flex-col items-center gap-4 rounded-xl border border-dashed py-12 text-center">
      <div className="bg-muted text-muted-foreground flex h-12 w-12 items-center justify-center rounded-full">
        <Ticket className="size-5" aria-hidden />
      </div>
      <div className="flex flex-col gap-1">
        <p className="text-sm font-medium">Noch keine Mitarbeiter-Essen angelegt</p>
        <p className="text-muted-foreground text-xs">Lege das erste an und verteile den Link an dein Team.</p>
      </div>
      <Button variant="outline" size="sm" onClick={onCreate}>
        <Plus className="size-4" aria-hidden />
        Erstes Essen anlegen
      </Button>
    </div>
  )
}
