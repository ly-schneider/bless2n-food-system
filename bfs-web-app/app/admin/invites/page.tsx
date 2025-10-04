"use client"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"
import { useEffect, useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

type Invite = { id: string; email: string; status: string; expiresAt: string; createdAt: string }

export default function AdminInvitesPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<Invite[]>([])
  const [email, setEmail] = useState("")
  const [error, setError] = useState<string| null>(null)

  useEffect(() => { void reload() }, [])

  async function reload() {
    try {
      const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/invites`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json() as { items: Invite[] }
      setItems(data.items || [])
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed to load invites'
      setError(msg)
    }
  }

  async function createInvite() {
    if (!email.trim()) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/invites`, {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ email: email.trim() })
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setEmail(""); await reload()
  }

  async function revoke(id: string) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/invites/${id}/revoke`, { method: 'POST', headers: { 'X-CSRF': csrf || '' } })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    await reload()
  }

  async function resend(id: string) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/invites/${id}/resend`, { method: 'POST', headers: { 'X-CSRF': csrf || '' } })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    await reload()
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Admin‑Einladungen</h1>
      {error && <div className="text-sm text-red-600">{error}</div>}
      <div className="flex items-center gap-2">
        <Input value={email} onChange={(e)=>setEmail(e.target.value)} placeholder="E‑Mail‑Adresse" className="h-8 w-72" />
        <Button variant="outline" size="sm" className="h-8" onClick={() => void createInvite()}>Einladen</Button>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm border border-gray-100">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-3 py-2">E‑Mail</th>
              <th className="text-left px-3 py-2">Status</th>
              <th className="text-left px-3 py-2">Ablauf</th>
              <th className="text-right px-3 py-2">Aktionen</th>
            </tr>
          </thead>
          <tbody>
            {items.map((i) => (
              <tr key={i.id} className="border-t border-gray-100">
                <td className="px-3 py-2">{i.email}</td>
                <td className="px-3 py-2">{i.status}</td>
                <td className="px-3 py-2">{new Date(i.expiresAt).toLocaleString()}</td>
                <td className="px-3 py-2 text-right space-x-2">
                  {i.status === 'pending' && <>
                    <Button variant="outline" size="sm" className="h-7 px-2" onClick={() => void resend(i.id)}>Erneut senden</Button>
                    <Button variant="ghost" size="sm" className="h-7 px-2 text-red-700" onClick={() => void revoke(i.id)}>Widerrufen</Button>
                  </>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function getCSRFCookie(): string | null {
  if (typeof document === 'undefined') return null
  const name = (document.location.protocol === 'https:' ? '__Host-' : '') + 'csrf'
  const m = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()\[\]\\/+^])/g, '\\$1') + '=([^;]*)'))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}
