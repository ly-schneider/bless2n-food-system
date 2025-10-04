"use client"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"
import { useEffect, useState } from "react"

type Row = { userId: string; email: string; familyId: string; device: string; createdAt: string; lastUsedAt: string }

export default function AdminSessionsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<Row[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => { void reload() }, [])

  async function reload() {
    try {
      const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/sessions`)
      if (!res.ok) throw new Error(await readErrorMessage(res))
      const data = await res.json() as { items: Row[] }
      setItems(data.items || [])
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed to load sessions'
      setError(msg)
    }
  }

  async function revoke(userId: string, familyId: string) {
    try {
      const csrf = getCSRFCookie()
      const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/users/${userId}/sessions/revoke`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ familyId })
      })
      if (!res.ok) throw new Error(await readErrorMessage(res))
      await reload()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed to revoke session'
      setError(msg)
    }
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Sessions</h1>
      {error && <div className="text-sm text-red-700">{error}</div>}
      <div className="overflow-x-auto">
        <table className="w-full text-sm border border-gray-100">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-3 py-2">Email</th>
              <th className="text-left px-3 py-2">Device</th>
              <th className="text-left px-3 py-2">Created</th>
              <th className="text-left px-3 py-2">Last used</th>
              <th className="text-right px-3 py-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r) => (
              <tr key={r.userId + ':' + r.familyId} className="border-t border-gray-100">
                <td className="px-3 py-2">{r.email || r.userId}</td>
                <td className="px-3 py-2">{r.device}</td>
                <td className="px-3 py-2">{new Date(r.createdAt).toLocaleString()}</td>
                <td className="px-3 py-2">{new Date(r.lastUsedAt).toLocaleString()}</td>
                <td className="px-3 py-2 text-right"><button className="text-xs text-red-700" onClick={() => void revoke(r.userId, r.familyId)}>Revoke</button></td>
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
