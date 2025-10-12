"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { API_BASE_URL } from "@/lib/api"
import { getClientInfo } from "@/lib/client-info"

export function RequestAccess({ token, onRefresh: _onRefresh }: { token: string; onRefresh: () => void }) {
  const [name, setName] = useState("")
  const [requested, setRequested] = useState(false)

  const submitRequest = async () => {
    try {
      const client = await getClientInfo()
      const payload = {
        name: name || "POS‑Terminal",
        model: client.model,
        os: client.os,
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
    <div className="grid min-h-[calc(100dvh-8rem)] place-items-center p-4">
      <div className="bg-background w-full max-w-md rounded-xl border p-5 shadow-sm">
        <h1 className="mb-2 text-2xl font-semibold">POS‑Zugang anfordern</h1>
        <p className="text-muted-foreground mb-4 text-sm">
          Dieses Gerät muss vor dem Verkauf von einem Admin freigegeben werden.
        </p>
        <div className="grid gap-3">
          <div className="grid gap-1">
            <Label htmlFor="name">Gerätename</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} placeholder="z. B. Kasse 1" />
          </div>
          <Button onClick={submitRequest}>Zugang anfordern</Button>
          {requested && (
            <div className="rounded-xl border border-amber-200 bg-amber-50 p-2 text-sm text-amber-700">
              Anfrage gesendet. Ein Admin muss dieses Gerät freigeben.
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
