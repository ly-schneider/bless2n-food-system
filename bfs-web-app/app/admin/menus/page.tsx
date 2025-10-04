"use client"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"
import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

type MenuSummary = { id: string; name: string; priceCents: number; isActive: boolean; image?: string | null }
type SlotItem = { id: string; type: "simple"|"menu"; name: string; image?: string|null; priceCents: number; isActive: boolean }
type Slot = { id: string; name: string; sequence: number; menuSlotItems: SlotItem[] }
type MenuDetail = { id: string; name: string; priceCents: number; isActive: boolean; image?: string|null; slots: Slot[] }
type ProductSummary = { id: string; type: "simple"|"menu"; name: string; image?: string|null; priceCents: number; isActive: boolean }
type Category = { id: string; name: string; isActive: boolean }

export default function AdminMenusPage() {
  const fetchAuth = useAuthorizedFetch()
  const [menus, setMenus] = useState<MenuSummary[]>([])
  const [selected, setSelected] = useState<string | null>(null)
  const [detail, setDetail] = useState<MenuDetail | null>(null)
  const [allProducts, setAllProducts] = useState<ProductSummary[]>([])
  const [q, setQ] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [cats, setCats] = useState<Category[]>([])
  const [newMenu, setNewMenu] = useState<{ name: string; price: string; categoryId: string }>({ name: "", price: "", categoryId: "" })

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus?limit=100`)
        if (res.ok) {
          const data = await res.json() as { items: MenuSummary[] }
          if (!cancelled) setMenus(data.items || [])
        }
        // Load up to 200 products for attach search
        const pr = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/products?limit=200`)
        if (pr.ok) {
          const d = await pr.json() as { items: ProductSummary[] }
          if (!cancelled) setAllProducts((d.items || []).filter(p => p.type === 'simple'))
        }
        // Load categories for creation flow
        const cr = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/categories`)
        if (cr.ok) {
          const d = await cr.json() as { items: Category[] }
          if (!cancelled) setCats(d.items || [])
        }
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : 'Failed to load'
        if (!cancelled) setError(msg)
      }
    })()
    return () => { cancelled = true }
  }, [fetchAuth])

  useEffect(() => {
    if (!selected) { setDetail(null); return }
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${encodeURIComponent(selected)}`)
        if (res.ok) {
          const d = await res.json() as MenuDetail
          if (!cancelled) setDetail(d)
        }
      } catch {}
    })()
    return () => { cancelled = true }
  }, [fetchAuth, selected])

  const filteredAttachables = useMemo(() => {
    const term = q.trim().toLowerCase()
    const list = term ? allProducts.filter(p => p.name.toLowerCase().includes(term)) : allProducts
    return list.slice(0, 20)
  }, [q, allProducts])

  async function createSlot(name: string) {
    if (!detail) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}/slots`, {
      method: "POST",
      headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' },
      body: JSON.stringify({ name }),
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setSelected(detail.id)
  }

  async function renameSlot(slotId: string, name: string) {
    if (!detail) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}/slots/${slotId}`, {
      method: "PATCH",
      headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' },
      body: JSON.stringify({ name }),
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setSelected(detail.id)
  }

  async function deleteSlot(slotId: string) {
    if (!detail) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}/slots/${slotId}`, {
      method: "DELETE",
      headers: { 'X-CSRF': csrf || '' },
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setSelected(detail.id)
  }

  async function moveSlot(slotId: string, dir: -1|1) {
    if (!detail) return
    const slots = [...detail.slots]
    const idx = slots.findIndex(s => s.id === slotId)
    const j = idx + dir
    if (idx < 0 || j < 0 || j >= slots.length) return
    const tmp: Slot = slots[idx]!
    slots[idx] = slots[j]!
    slots[j] = tmp
    const body = { order: slots.map((s, i) => ({ slotId: s.id, sequence: i+1 })) }
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}/slots/reorder`, {
      method: 'PATCH', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify(body)
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setSelected(detail.id)
  }

  async function attach(slotId: string, productId: string) {
    if (!detail) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}/slots/${slotId}/items`, {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ productId })
    })
    if (res.ok) setSelected(detail.id)
  }

  async function createMenu() {
    const name = newMenu.name.trim()
    const price = Math.round(parseFloat(newMenu.price || '0') * 100)
    const categoryId = newMenu.categoryId
    if (!name || !categoryId || !(price >= 0)) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus`, {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'X-CSRF': csrf || '' }, body: JSON.stringify({ name, priceCents: price, categoryId })
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setNewMenu({ name: '', price: '', categoryId: '' }); await refreshMenus()
  }

  async function refreshMenus() {
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus?limit=100`)
    if (res.ok) { const data = await res.json() as { items: MenuSummary[] }; setMenus(data.items || []) }
  }

  async function detach(slotId: string, productId: string) {
    if (!detail) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}/slots/${slotId}/items/${productId}`, {
      method: 'DELETE', headers: { 'X-CSRF': csrf || '' }
    })
    if (!res.ok) { setError(await readErrorMessage(res)); return }
    setSelected(detail.id)
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div className="md:col-span-1 border rounded p-3 space-y-3">
        <div className="flex items-center justify-between mb-2">
          <h2 className="font-semibold">Menus</h2>
        </div>
        <div className="border rounded p-2">
          <div className="text-sm font-medium mb-2">Create Menu</div>
          <div className="flex flex-col gap-2">
            <input className="border rounded px-2 py-1 text-sm" placeholder="Name" value={newMenu.name} onChange={(e)=>setNewMenu(v=>({...v, name: e.target.value}))} />
            <input className="border rounded px-2 py-1 text-sm" placeholder="Price (CHF)" value={newMenu.price} onChange={(e)=>setNewMenu(v=>({...v, price: e.target.value}))} />
            <select className="border rounded px-2 py-1 text-sm" value={newMenu.categoryId} onChange={(e)=>setNewMenu(v=>({...v, categoryId: e.target.value}))}>
              <option value="">Select category</option>
              {cats.map(c => (<option key={c.id} value={c.id}>{c.name}</option>))}
            </select>
            <button className="border rounded px-2 py-1 text-sm" onClick={() => void createMenu()}>Create</button>
          </div>
        </div>
        <ul className="text-sm divide-y">
          {menus.map(m => (
            <li key={m.id}>
              <button className={`w-full text-left px-2 py-2 ${selected===m.id? 'bg-gray-100' : ''}`} onClick={() => setSelected(m.id)}>
                <div className="flex items-center justify-between"><span>{m.name}</span><span className="text-xs text-gray-600">{formatCHF(m.priceCents)}</span></div>
              </button>
            </li>
          ))}
        </ul>
      </div>
      <div className="md:col-span-2">
        {!detail ? (
          <div className="text-sm text-gray-600">Wähle ein Menü, um den Inhalt zu bearbeiten.</div>
        ) : (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-xl font-semibold">{detail.name}</h1>
                <div className="text-sm text-gray-600">{formatCHF(detail.priceCents)} {detail.isActive ? '' : '(inactive)'}</div>
              </div>
              <div className="flex items-center gap-2">
                <button className="text-xs border rounded px-2 py-1" onClick={async ()=>{
                  const csrf = getCSRFCookie(); if (!csrf) return
                  const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}/active`, { method:'PATCH', headers:{'Content-Type':'application/json','X-CSRF':csrf}, body: JSON.stringify({ isActive: false }) })
                  if (!res.ok) { setError(await readErrorMessage(res)); return }
                  setSelected(null); await refreshMenus()
                }}>Deaktivieren</button>
                <button className="text-xs text-red-700 border rounded px-2 py-1" onClick={async ()=>{
                  if (!confirm('Dieses Menü dauerhaft löschen? Dieser Vorgang kann nicht rückgängig gemacht werden.')) return
                  const csrf = getCSRFCookie(); const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/menus/${detail.id}`, { method:'DELETE', headers:{ 'X-CSRF': csrf || '' } })
                  if (!res.ok) { setError(await readErrorMessage(res)); return }
                  setSelected(null); setDetail(null); await refreshMenus()
                }}>Delete Permanently</button>
              </div>
            </div>
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Input placeholder="Neuer Slot‑Name" id="new-slot" className="h-8 w-56" />
                <Button variant="outline" size="sm" className="h-8" onClick={() => {
                  const el = document.getElementById('new-slot') as HTMLInputElement | null; const val = el?.value?.trim(); if (val) { void createSlot(val); if (el) el.value=''; }
                }}>Slot hinzufügen</Button>
              </div>
              {detail.slots.map((s) => (
                <div key={s.id} className="border rounded p-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <strong>{s.name}</strong>
                      <button className="text-xs text-gray-600 underline" onClick={() => {
                        const v = prompt('Slot umbenennen', s.name); if (v && v.trim()) void renameSlot(s.id, v.trim())
                      }}>Umbenennen</button>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button variant="outline" size="sm" className="h-7" onClick={() => void moveSlot(s.id, -1)}>Hoch</Button>
                      <Button variant="outline" size="sm" className="h-7" onClick={() => void moveSlot(s.id, +1)}>Runter</Button>
                      <Button variant="ghost" size="sm" className="h-7 text-red-700" onClick={() => void deleteSlot(s.id)}>Löschen</Button>
                    </div>
                  </div>
                  <div className="mt-2">
                    <div className="text-xs text-gray-600 mb-1">Zugeordnete Artikel</div>
                    <ul className="text-sm flex flex-wrap gap-2">
                      {s.menuSlotItems.map((it) => (
                        <li key={it.id} className="border rounded px-2 py-1 flex items-center gap-2">
                          <span>{it.name}</span>
                          <button className="text-xs text-red-700" onClick={() => void detach(s.id, it.id)}>Entfernen</button>
                        </li>
                      ))}
                      {s.menuSlotItems.length === 0 && <li className="text-gray-500">Keine Einträge.</li>}
                    </ul>
                  </div>
                  <div className="mt-3">
                    <div className="text-xs text-gray-600 mb-1">Produkt zuordnen</div>
                    <Input value={q} onChange={(e)=>setQ(e.target.value)} placeholder="Produkte suchen…" className="h-8 w-full" />
                    <div className="mt-2 max-h-48 overflow-auto border rounded">
                      <ul className="text-sm divide-y">
                        {filteredAttachables.map(p => (
                          <li key={p.id} className="flex items-center justify-between px-2 py-1">
                            <span>{p.name}</span>
                            <Button variant="outline" size="sm" className="h-7 px-2" onClick={() => void attach(s.id, p.id)}>Hinzufügen</Button>
                          </li>
                        ))}
                        {filteredAttachables.length === 0 && <li className="px-2 py-1 text-gray-500">Keine Treffer.</li>}
                      </ul>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function formatCHF(cents: number) {
  return new Intl.NumberFormat('de-CH', { style: 'currency', currency: 'CHF' }).format((cents||0)/100)
}

function getCSRFCookie(): string | null {
  if (typeof document === 'undefined') return null
  const name = (document.location.protocol === 'https:' ? '__Host-' : '') + 'csrf'
  const m = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()\[\]\\/+^])/g, '\\$1') + '=([^;]*)'))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}
