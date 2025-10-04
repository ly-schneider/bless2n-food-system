"use client"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"
import { useEffect, useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"

type Category = { id: string; name: string; isActive: boolean }

export default function AdminCategoriesPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<Category[]>([])
  const [name, setName] = useState("")
  const [error, setError] = useState<string | null>(null)

  useEffect(() => { void reload() }, [])

  async function reload() {
    try {
      const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json() as { items: Category[] }
      setItems(data.items || [])
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed to load'
      setError(msg)
    }
  }

  async function createCategory() {
    if (!name.trim()) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories`, {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ name: name.trim() })
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setName(""); await reload()
  }

  async function rename(id: string, newName: string) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories/${id}`, {
      method: 'PATCH', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ name: newName })
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    await reload()
  }

  async function toggle(id: string, isActive: boolean) {
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories/${id}`, {
      method: 'PATCH', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ isActive })
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    await reload()
  }

  async function remove(id: string) {
    if (!confirm('Delete category?')) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories/${id}`, { method: 'DELETE', headers: { 'X-CSRF': csrf || '' } })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    await reload()
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Kategorien</h1>
      {error && <div className="text-sm text-red-600">{error}</div>}
      <div className="flex items-center gap-2">
        <Input value={name} onChange={(e)=>setName(e.target.value)} placeholder="Neue Kategorie" className="h-8 w-64" />
        <Button variant="outline" size="sm" className="h-8" onClick={() => void createCategory()}>Erstellen</Button>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm border border-gray-100">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-3 py-2">Name</th>
              <th className="text-left px-3 py-2">Status</th>
              <th className="text-right px-3 py-2">Aktionen</th>
            </tr>
          </thead>
          <tbody>
            {items.map((c) => (
              <tr key={c.id} className="border-t border-gray-100">
                <td className="px-3 py-2">
                  <button className="underline decoration-dotted" onClick={() => {
                    const v = prompt('Kategorie umbenennen', c.name); if (v && v.trim()) void rename(c.id, v.trim())
                  }}>{c.name}</button>
                </td>
                <td className="px-3 py-2">
                  <label className="inline-flex items-center gap-2">
                    <Switch checked={c.isActive} onCheckedChange={(v) => void toggle(c.id, v)} />
                    <span>{c.isActive ? 'Aktiv' : 'Inaktiv'}</span>
                  </label>
                </td>
                <td className="px-3 py-2 text-right">
                  <Button variant="ghost" size="sm" className="h-7 text-red-700" onClick={() => void remove(c.id)}>Löschen</Button>
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
