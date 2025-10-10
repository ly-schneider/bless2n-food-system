"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { API_BASE_URL } from "@/lib/api"

export function RequestAccess({ token, onRefresh }: { token: string; onRefresh: () => void }) {
  const [name, setName] = useState("")
  const [requested, setRequested] = useState(false)

  const submitRequest = async () => {
    try {
      const payload = {
        name: name || "POS‑Terminal",
        model: navigator.userAgent,
        os: navigator.platform,
        deviceToken: token,
      }
      const res = await fetch(`${API_BASE_URL}/v1/pos/requests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      await res.json().catch(() => ({}))
      setRequested(true)
    } catch {}
  }

  return (
    <div className="grid min-h-[calc(100dvh-1.5rem)] place-items-center p-4">
      <div className="bg-background w-full max-w-md rounded-xl border p-5 shadow-sm">
        <h1 className="mb-2 text-2xl font-semibold">POS‑Zugang anfordern</h1>
        <p className="text-muted-foreground mb-4 text-sm">
          Dieses Gerät muss vor dem Verkauf von einem Admin freigegeben werden.
        </p>
        <div className="grid gap-3">
          <div className="grid gap-1">
            <Label htmlFor="name">Gerätename</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} placeholder="z. B. Frontkasse" />
          </div>
          <Button onClick={submitRequest}>Zugang anfordern</Button>
          {requested && (
            <div className="rounded-md border border-amber-200 bg-amber-50 p-2 text-sm text-amber-700">
              Anfrage gesendet. Ein Admin muss dieses Gerät freigeben.
            </div>
          )}
          <div className="flex items-center justify-between gap-2 pt-2">
            <span className="text-muted-foreground text-xs">Geräte‑Token</span>
            <code className="bg-muted rounded px-2 py-0.5 text-[11px]">{token}</code>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" className="flex-1" onClick={onRefresh}>
              Status prüfen
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
