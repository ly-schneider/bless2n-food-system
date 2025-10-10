"use client"
import Link from "next/link"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"

type Invite = {
  id: string
  email: string
  status: string
  invitedBy: string
  expiresAt: string
  createdAt: string
}

export default function AdminInvitesPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<Invite[]>([])
  const [email, setEmail] = useState("")
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void reload()
  }, [])

  async function reload() {
    try {
      const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/invites`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = (await res.json()) as { items: Invite[] }
      setItems(data.items || [])
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to load invites"
      setError(msg)
    }
  }

  async function createInvite() {
    if (!email.trim()) return
    const csrf = getCSRFCookie()
    const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/invites`, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ email: email.trim() }),
    })
    if (!res.ok) {
      setError(await readErrorMessage(res))
      return
    }
    setEmail("")
    await reload()
  }

  async function deleteInvite(id: string) {
    try {
      const csrf = getCSRFCookie()
      const res = await fetchAuth(
        `${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/invites/${encodeURIComponent(id)}`,
        {
          method: "DELETE",
          headers: { "X-CSRF": csrf || "" },
        }
      )
      if (!res.ok && res.status !== 204) throw new Error(await readErrorMessage(res))
      await reload()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Löschen fehlgeschlagen"
      setError(msg)
    }
  }

  return (
    <div className="min-w-0 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Admin‑Einladungen</h1>
      </div>
      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="flex items-center gap-2">
        <Input
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="E‑Mail‑Adresse"
          className="h-8 w-72"
        />
        <Button variant="outline" size="sm" className="h-8" onClick={() => void createInvite()}>
          Einladen
        </Button>
      </div>

      <div className="rounded-md border">
        <div className="relative">
          <div className="from-background pointer-events-none absolute inset-y-0 left-0 w-6 bg-gradient-to-r to-transparent" />
          <div className="from-background pointer-events-none absolute inset-y-0 right-0 w-6 bg-gradient-to-l to-transparent" />
          <div
            className="max-w-full overflow-x-auto overscroll-x-contain rounded-[10px]"
            role="region"
            aria-label="Invites table – scroll horizontally to reveal more columns"
            tabIndex={0}
          >
            <Table className="whitespace-nowrap">
              <TableHeader className="bg-card sticky top-0">
                <TableRow>
                  <TableHead className="whitespace-nowrap">ID</TableHead>
                  <TableHead className="whitespace-nowrap">Eingeladene E‑Mail</TableHead>
                  <TableHead className="whitespace-nowrap">Status</TableHead>
                  <TableHead className="whitespace-nowrap">Eingeladet von</TableHead>
                  <TableHead className="whitespace-nowrap">Erstellt</TableHead>
                  <TableHead className="whitespace-nowrap">Läuft ab</TableHead>
                  <TableHead className="text-right whitespace-nowrap">Aktionen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((i) => {
                  const created = new Date(i.createdAt).toLocaleString("de-CH")
                  const expires = new Date(i.expiresAt).toLocaleString("de-CH")
                  const inviter = (
                    <Link
                      href={`/admin/users/${encodeURIComponent(i.invitedBy)}`}
                      className="text-xs underline underline-offset-2"
                    >
                      {i.invitedBy}
                    </Link>
                  )
                  return (
                    <TableRow key={i.id} className="even:bg-card odd:bg-muted/40">
                      <TableCell className="text-xs">{i.id}</TableCell>
                      <TableCell>{i.email}</TableCell>
                      <TableCell className="uppercase">{i.status}</TableCell>
                      <TableCell>{inviter}</TableCell>
                      <TableCell className="whitespace-nowrap">{created}</TableCell>
                      <TableCell className="whitespace-nowrap">{expires}</TableCell>
                      <TableCell className="text-right">
                        <Button size="sm" variant="destructive" onClick={() => void deleteInvite(i.id)}>
                          Löschen
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>
    </div>
  )
}

function getCSRFCookie(): string | null {
  if (typeof document === "undefined") return null
  const name = (document.location.protocol === "https:" ? "__Host-" : "") + "csrf"
  const m = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/([.$?*|{}()\[\]\\/+^])/g, "\\$1") + "=([^;]*)"))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}
