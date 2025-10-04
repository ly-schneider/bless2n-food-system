"use client"
import { useEffect, useState } from "react"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"
import { Button } from "@/components/ui/button"

type User = { id: string; email: string; role: string; isDisabled?: boolean }

export default function AdminUsersPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<User[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/users`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: User[] }
        if (!cancelled) setItems(data.items || [])
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Failed to load users"
        if (!cancelled) setError(msg)
      }
    })()
    return () => { cancelled = true }
  }, [fetchAuth])

  async function promote(id: string) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/users/${id}/promote`, {
      method: 'POST', headers: { 'X-CSRF': csrf || '' }
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    const r = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/users`)
    if (r.ok) { const d = await r.json() as { items: User[] }; setItems(d.items || []) }
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Benutzer</h1>
      {error && <div className="text-red-600 text-sm">{error}</div>}
      <div className="overflow-x-auto">
        <table className="w-full text-sm border border-gray-100">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-3 py-2">E‑Mail</th>
              <th className="text-left px-3 py-2">Rolle</th>
              <th className="text-left px-3 py-2">Status</th>
              <th className="text-right px-3 py-2">Aktionen</th>
            </tr>
          </thead>
          <tbody>
            {items.map((u) => (
              <tr key={u.id} className="border-t border-gray-100">
                <td className="px-3 py-2">{u.email}</td>
                <td className="px-3 py-2 uppercase">{u.role}</td>
                <td className="px-3 py-2">{u.isDisabled ? "Gesperrt" : "Aktiv"}</td>
                <td className="px-3 py-2 text-right">
                  {u.role !== 'admin' && (
                    <Button variant="outline" size="sm" className="h-7" onClick={() => void promote(u.id)}>Als Admin befördern</Button>
                  )}
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
