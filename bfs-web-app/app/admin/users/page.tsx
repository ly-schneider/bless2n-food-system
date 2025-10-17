"use client"
import Link from "next/link"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { API_BASE_URL } from "@/lib/api"

type User = {
  id: string
  email: string
  firstName?: string | null
  lastName?: string | null
  role: string
  isVerified?: boolean
  createdAt?: string
  updatedAt?: string
}

export default function AdminUsersPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<User[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetchAuth(`${API_BASE_URL}/v1/admin/users`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: User[] }
        if (!cancelled) setItems(data.items || [])
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Failed to load users"
        if (!cancelled) setError(msg)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  return (
    <div className="min-w-0 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Benutzer</h1>
      </div>

      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="rounded-md border">
        <div className="relative">
          <div className="from-background pointer-events-none absolute inset-y-0 left-0 w-6 bg-gradient-to-r to-transparent" />
          <div className="from-background pointer-events-none absolute inset-y-0 right-0 w-6 bg-gradient-to-l to-transparent" />
          <div
            className="max-w-full overflow-x-auto overscroll-x-contain rounded-[10px]"
            role="region"
            aria-label="Users table – scroll horizontally to reveal more columns"
            tabIndex={0}
          >
            <Table className="whitespace-nowrap">
              <TableHeader className="bg-card sticky top-0">
                <TableRow>
                  <TableHead className="whitespace-nowrap">ID</TableHead>
                  <TableHead className="whitespace-nowrap">E‑Mail</TableHead>
                  <TableHead className="whitespace-nowrap">Vorname</TableHead>
                  <TableHead className="whitespace-nowrap">Nachname</TableHead>
                  <TableHead className="whitespace-nowrap">Rolle</TableHead>
                  <TableHead className="whitespace-nowrap">Verifiziert</TableHead>
                  <TableHead className="whitespace-nowrap">Erstellt</TableHead>
                  <TableHead className="whitespace-nowrap">Aktualisiert</TableHead>
                  <TableHead className="text-right whitespace-nowrap">Aktionen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((u) => {
                  const created = u.createdAt ? new Date(u.createdAt).toLocaleString("de-CH") : "–"
                  const updated = u.updatedAt ? new Date(u.updatedAt).toLocaleString("de-CH") : "–"
                  return (
                    <TableRow key={u.id} className="even:bg-card odd:bg-muted/40">
                      <TableCell className="text-xs">{u.id}</TableCell>
                      <TableCell>{u.email}</TableCell>
                      <TableCell>{u.firstName || "–"}</TableCell>
                      <TableCell>{u.lastName || "–"}</TableCell>
                      <TableCell className="uppercase">{u.role}</TableCell>
                      <TableCell>{u.isVerified ? "Ja" : "Nein"}</TableCell>
                      <TableCell className="whitespace-nowrap">{created}</TableCell>
                      <TableCell className="whitespace-nowrap">{updated}</TableCell>
                      <TableCell className="text-right">
                        <Link href={`/admin/users/${encodeURIComponent(u.id)}`}>
                          <Button size="sm" variant="outline">
                            Details
                          </Button>
                        </Link>
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
