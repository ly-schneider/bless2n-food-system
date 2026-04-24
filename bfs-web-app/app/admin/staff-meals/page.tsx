"use client"

import { ArrowRight, Plus, Ticket, Trash2 } from "lucide-react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useCallback, useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
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
  const [formMaxRedemptions, setFormMaxRedemptions] = useState(20)
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
    setFormMaxRedemptions(20)
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
    if (formMaxRedemptions < 1) {
      setFormError("Maximale Einlösungen muss mindestens 1 sein.")
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
          maxRedemptions: formMaxRedemptions,
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
    <div className="space-y-4 pt-4">
      <div className="flex items-center justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-1">
          <h1 className="font-primary text-2xl">Mitarbeiter-Essen</h1>
          <p className="text-muted-foreground text-sm">
            Gratis-Mahlzeiten, die Mitarbeiter per QR-Code an der Station einlösen.
          </p>
        </div>
        <Button onClick={openCreate} className="shrink-0">
          <Plus className="size-4" aria-hidden />
          Neues Essen
        </Button>
      </div>

      {error && <div className="bg-destructive/10 text-destructive rounded-xl px-3 py-2 text-sm">{error}</div>}

      {!firstLoadDone ? (
        <LoadingList />
      ) : meals.length === 0 ? (
        <EmptyState onCreate={openCreate} />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {meals.map((m) => {
            const progress =
              m.maxRedemptions > 0 ? Math.min(100, Math.round((m.redemptionCount / m.maxRedemptions) * 100)) : 0
            return (
              <Link
                key={m.id}
                href={`/admin/staff-meals/${m.id}`}
                className="border-border bg-card hover:bg-accent/40 focus-visible:ring-ring/40 group flex flex-col gap-4 rounded-2xl border p-5 transition focus-visible:ring-2 focus-visible:outline-none"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="truncate font-semibold">{m.name}</span>
                      <StatusBadge status={m.status} />
                    </div>
                    <p className="text-muted-foreground mt-1 text-xs tabular-nums">
                      {m.redemptionCount} / {m.maxRedemptions} eingelöst
                    </p>
                  </div>
                  <ArrowRight
                    className="text-muted-foreground group-hover:text-foreground mt-0.5 size-4 shrink-0 transition-transform group-hover:translate-x-0.5"
                    aria-hidden
                  />
                </div>
                <div className="bg-muted-foreground/20 relative h-1.5 overflow-hidden rounded-full">
                  <div
                    className="bg-primary absolute inset-y-0 left-0 rounded-full transition-[width] duration-300"
                    style={{ width: `${progress}%` }}
                  />
                </div>
                <div className="flex items-center justify-between gap-2">
                  <span className="text-muted-foreground text-xs">Zugangscode</span>
                  <code className="font-mono text-sm tracking-wider">{m.accessCode}</code>
                </div>
              </Link>
            )
          })}
        </div>
      )}

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
                <Label htmlFor="meal-count">Maximale Einlösungen</Label>
                <Input
                  id="meal-count"
                  type="number"
                  value={formMaxRedemptions}
                  min={1}
                  onChange={(e) => setFormMaxRedemptions(Math.max(1, parseInt(e.target.value || "0", 10) || 0))}
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
              <Label>Produkte pro Einlösung</Label>
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
    <div className="border-border/60 flex flex-col items-center gap-4 rounded-2xl border border-dashed py-12 text-center">
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
