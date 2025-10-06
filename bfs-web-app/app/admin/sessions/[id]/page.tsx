"use client"
import Link from "next/link"
import { useParams, useRouter } from "next/navigation"
import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"

type Row = {
  userId: string
  email: string
  familyId: string
  device: string
  createdAt: string
  lastUsedAt: string
}

export default function AdminSessionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const fetchAuth = useAuthorizedFetch()
  const [item, setItem] = useState<Row | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [revoking, setRevoking] = useState(false)

  const { userId, familyId } = useMemo(() => {
    const [u, f] = String(id).split(":")
    return { userId: decodeURIComponent(u || ""), familyId: decodeURIComponent(f || "") }
  }, [id])

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoading(true)
      setError(null)
      try {
        const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/sessions`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: Row[] }
        const found = (data.items || []).find((r) => r.userId === userId && r.familyId === familyId) || null
        if (!cancelled) setItem(found)
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Fehler beim Laden der Session"
        if (!cancelled) setError(msg)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    if (userId && familyId) void load()
    return () => { cancelled = true }
  }, [fetchAuth, userId, familyId])

  async function revoke() {
    if (!userId || !familyId) return
    if (!confirm("Diese Sitzung wirklich widerrufen?")) return
    try {
      setRevoking(true)
      const csrf = getCSRFCookie()
      const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/users/${encodeURIComponent(userId)}/sessions/revoke`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
        body: JSON.stringify({ familyId }),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      router.push("/admin/sessions")
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Widerruf fehlgeschlagen"
      setError(msg)
    } finally {
      setRevoking(false)
    }
  }

  const created = item?.createdAt ? new Date(item.createdAt).toLocaleString("de-CH") : "–"
  const lastUsed = item?.lastUsedAt ? new Date(item.lastUsedAt).toLocaleString("de-CH") : "–"

  return (
    <div className="min-w-0 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Session</h1>
        <div className="flex items-center gap-2">
          <Button variant="destructive" onClick={() => void revoke()} disabled={!item || revoking}>
            {revoking ? "Widerrufe…" : "Sitzung widerrufen"}
          </Button>
          <Link href="/admin/sessions">
            <Button variant="outline">Zurück</Button>
          </Link>
        </div>
      </div>

      {loading && <div className="text-sm text-muted-foreground">Lade Session…</div>}
      {error && <div className="text-sm text-red-600">{error}</div>}

      {item ? (
        <div className="rounded-md border p-4">
          <h2 className="mb-3 text-base font-semibold">Details</h2>
          <div className="space-y-1 text-sm">
            <div><span className="text-muted-foreground">Benutzer ID:</span> <Link href={`/admin/users/${encodeURIComponent(item.userId)}`} className="underline underline-offset-2 text-xs">{item.userId}</Link></div>
            <div><span className="text-muted-foreground">E‑Mail:</span> {item.email || "–"}</div>
            <div><span className="text-muted-foreground">Gerät:</span> {item.device}</div>
            <div><span className="text-muted-foreground">Family ID:</span> <span className="font-mono text-xs">{item.familyId}</span></div>
            <div><span className="text-muted-foreground">Erstellt:</span> {created}</div>
            <div><span className="text-muted-foreground">Zuletzt genutzt:</span> {lastUsed}</div>
          </div>
        </div>
      ) : (
        !loading && <div className="text-sm text-muted-foreground">Keine Session gefunden.</div>
      )}
    </div>
  )
}

function getCSRFCookie(): string | null {
  if (typeof document === "undefined") return null
  const name = (document.location.protocol === "https:" ? "__Host-" : "") + "csrf"
  const m = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/([.$?*|{}()\[\]\\/+^])/g, "\\$1") + "=([^;]*)"))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}

