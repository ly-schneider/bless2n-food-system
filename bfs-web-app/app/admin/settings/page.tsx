"use client"

import Link from "next/link"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { getCSRFToken } from "@/lib/csrf"
import type { Product } from "@/types"

type PosFulfillmentMode = "QR_CODE" | "JETON"

type SettingsResponse = {
  posMode?: PosFulfillmentMode
  missingJetons?: number
  club100FreeProductIds?: string[]
  club100MaxRedemptions?: number
}

export default function AdminSettingsPage() {
  const fetchAuth = useAuthorizedFetch()

  const [posMode, setPosMode] = useState<PosFulfillmentMode>("QR_CODE")
  const [missingJetons, setMissingJetons] = useState(0)
  const [club100FreeProductIds, setClub100FreeProductIds] = useState<string[]>([])
  const [club100MaxRedemptions, setClub100MaxRedemptions] = useState(2)

  const [products, setProducts] = useState<Product[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    ;(async () => {
      const [settingsRes, productsRes] = await Promise.all([
        fetchAuth(`/api/v1/settings`),
        fetchAuth(`/api/v1/products?type=simple&active=true`),
      ])

      if (settingsRes.ok) {
        const s = (await settingsRes.json()) as SettingsResponse
        setPosMode(s.posMode ?? "QR_CODE")
        setMissingJetons(s.missingJetons ?? 0)
        setClub100FreeProductIds(s.club100FreeProductIds ?? [])
        setClub100MaxRedemptions(s.club100MaxRedemptions ?? 2)
      }

      if (productsRes.ok) {
        const p = (await productsRes.json()) as { items: Product[] }
        setProducts(p.items ?? [])
      }
    })()
  }, [fetchAuth])

  async function updateSettings(updates: Partial<SettingsResponse>) {
    setSaving(true)
    setError(null)
    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/settings`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify(updates),
      })
      if (!res.ok) {
        const j = (await res.json().catch(() => ({}))) as {
          error?: string
          missing?: number
          message?: string
          detail?: string
        }
        if (j.error === "missing_jetons") {
          setMissingJetons(j.missing ?? missingJetons)
          throw new Error(`Jeton-Modus kann nicht aktiviert werden. ${j.missing} aktive Produkte haben keinen Jeton.`)
        }
        throw new Error(j.detail || j.message || `Error ${res.status}`)
      }
      const updated = (await res.json()) as SettingsResponse
      setPosMode(updated.posMode ?? posMode)
      setMissingJetons(updated.missingJetons ?? 0)
      setClub100FreeProductIds(updated.club100FreeProductIds ?? club100FreeProductIds)
      setClub100MaxRedemptions(updated.club100MaxRedemptions ?? club100MaxRedemptions)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Speichern fehlgeschlagen")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6 pt-4">
      <h1 className="font-primary text-2xl">Einstellungen</h1>

      {error && (
        <div role="alert" className="text-destructive bg-destructive/10 rounded px-3 py-2">
          {error}
        </div>
      )}

      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle>POS Modus</CardTitle>
          <p className="text-muted-foreground text-sm">
            Bestimme, ob das POS Belege mit QR-Code druckt oder Jetons zum Ausgeben berechnet.
          </p>
        </CardHeader>
        <CardContent className="max-w-lg space-y-3">
          <div className="grid gap-3 md:grid-cols-2">
            <Button
              variant={posMode === "QR_CODE" ? "default" : "outline"}
              className="h-12 justify-start rounded-xl"
              onClick={() => updateSettings({ posMode: "QR_CODE" })}
              disabled={saving}
            >
              QR-Code
            </Button>
            <Button
              variant={posMode === "JETON" ? "default" : "outline"}
              className="h-12 justify-start rounded-xl"
              onClick={() => updateSettings({ posMode: "JETON" })}
              disabled={saving}
            >
              Jetons
            </Button>
          </div>
          {missingJetons > 0 && (
            <div className="text-sm text-amber-700">
              {missingJetons} aktive Produkte haben keinen Jeton.{" "}
              <Link href="/admin/jetons" className="underline">
                Jetons verwalten
              </Link>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle>100 Club Einstellungen</CardTitle>
          <p className="text-muted-foreground text-sm">Wähle die Produkte, die 100 Club Mitglieder gratis erhalten.</p>
        </CardHeader>
        <CardContent className="max-w-lg space-y-4">
          <div className="space-y-2">
            <Label>Gratis-Produkte</Label>
            <div className="max-h-48 space-y-1 overflow-y-auto rounded-xl border p-2">
              {products.map((p) => (
                <label key={p.id} className="hover:bg-muted flex cursor-pointer items-center gap-2 rounded-lg p-2">
                  <Checkbox
                    checked={club100FreeProductIds.includes(p.id)}
                    onCheckedChange={(checked) => {
                      const newIds = checked
                        ? [...club100FreeProductIds, p.id]
                        : club100FreeProductIds.filter((id) => id !== p.id)
                      setClub100FreeProductIds(newIds)
                      updateSettings({ club100FreeProductIds: newIds })
                    }}
                    disabled={saving}
                  />
                  <span>{p.name}</span>
                </label>
              ))}
            </div>
            {club100FreeProductIds.length === 0 && (
              <p className="text-muted-foreground text-sm">Keine Produkte ausgewählt</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="club100-max">Max. Einlösungen pro Person</Label>
            <Input
              id="club100-max"
              type="number"
              min={1}
              value={club100MaxRedemptions}
              onChange={(e) => {
                const val = parseInt(e.target.value, 10)
                if (!isNaN(val) && val > 0) {
                  setClub100MaxRedemptions(val)
                }
              }}
              onBlur={() => {
                updateSettings({ club100MaxRedemptions })
              }}
              disabled={saving}
              className="max-w-24 rounded-xl"
            />
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
