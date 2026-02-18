"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { getCSRFToken } from "@/lib/csrf"

interface PairDeviceCardProps {
  onPaired?: () => void
}

type PairResult = {
  id: string
  name: string
  type: string
  status: string
}

export function PairDeviceCard({ onPaired }: PairDeviceCardProps) {
  const fetchAuth = useAuthorizedFetch()
  const [code, setCode] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<PairResult | null>(null)

  function formatInput(raw: string): string {
    // Strip non-alphanumeric, uppercase, limit to 6
    const cleaned = raw
      .replace(/[^A-Za-z0-9]/g, "")
      .toUpperCase()
      .slice(0, 6)
    if (cleaned.length <= 3) return cleaned
    return `${cleaned.slice(0, 3)} ${cleaned.slice(3)}`
  }

  function rawCode(): string {
    return code.replace(/\s/g, "").toUpperCase()
  }

  async function handlePair() {
    const c = rawCode()
    if (c.length !== 6) {
      setError("Code muss 6 Zeichen lang sein.")
      return
    }

    setLoading(true)
    setError(null)
    setSuccess(null)

    try {
      const csrf = getCSRFToken()
      const res = await fetchAuth(`/api/v1/devices/pairings/${encodeURIComponent(c)}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF": csrf || "",
        },
        body: JSON.stringify({ code: c }),
      })

      if (res.status === 404) {
        setError("Code ungültig oder abgelaufen.")
        return
      }
      if (res.status === 409) {
        setError("Gerät bereits gekoppelt.")
        return
      }
      if (res.status === 403) {
        setError("Kopplung mit diesem Code nicht möglich. Bitte prüfe den Code und versuche es erneut.")
        return
      }
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as { code?: string; message?: string }
        setError(body.message || `Fehler ${res.status}`)
        return
      }

      const data = (await res.json()) as PairResult
      setSuccess(data)
      setCode("")
      onPaired?.()
    } catch {
      setError("Kopplung fehlgeschlagen.")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="bg-card rounded-xl border p-4 shadow-sm">
      <h2 className="mb-2 text-lg font-semibold">Gerät koppeln</h2>
      <p className="text-muted-foreground mb-3 text-sm">
        Gib den 6-stelligen Code ein, der auf dem Gerät angezeigt wird.
      </p>

      {error && <div className="mb-3 rounded-lg border border-red-200 bg-red-50 p-2 text-sm text-red-700">{error}</div>}

      {success && (
        <div className="mb-3 rounded-lg border border-green-200 bg-green-50 p-2 text-sm text-green-700">
          {success.name} ({success.type}) erfolgreich gekoppelt.
        </div>
      )}

      <div className="flex items-end gap-2">
        <div className="grid flex-1 gap-1">
          <Label htmlFor="pairing-code">Kopplungscode</Label>
          <Input
            id="pairing-code"
            value={code}
            onChange={(e) => setCode(formatInput(e.target.value))}
            placeholder="z. B. X7K 2M9"
            className="mt-1 text-lg"
            maxLength={7}
            onKeyDown={(e) => {
              if (e.key === "Enter") void handlePair()
            }}
          />
        </div>
        <Button onClick={handlePair} disabled={loading || rawCode().length !== 6}>
          {loading ? "Wird gekoppelt..." : "Koppeln"}
        </Button>
      </div>
    </div>
  )
}
