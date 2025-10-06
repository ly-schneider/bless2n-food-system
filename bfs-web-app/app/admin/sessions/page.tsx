"use client"
import Link from "next/link"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

type Row = {
  userId: string
  email: string
  familyId: string
  device: string
  createdAt: string
  lastUsedAt: string
}

export default function AdminSessionsPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<Row[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void reload()
  }, [])

  async function reload() {
    try {
      const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/sessions`)
      if (!res.ok) throw new Error(await readErrorMessage(res))
      const data = (await res.json()) as { items: Row[] }
      setItems(data.items || [])
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to load sessions"
      setError(msg)
    }
  }

  return (
    <div className="min-w-0 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Sessions</h1>
      </div>

      {error && <div className="text-sm text-red-700">{error}</div>}

      <div className="rounded-md border">
        <div className="relative">
          <div className="from-background pointer-events-none absolute inset-y-0 left-0 w-6 bg-gradient-to-r to-transparent" />
          <div className="from-background pointer-events-none absolute inset-y-0 right-0 w-6 bg-gradient-to-l to-transparent" />
          <div
            className="max-w-full overflow-x-auto overscroll-x-contain rounded-[10px]"
            role="region"
            aria-label="Sessions table – scroll horizontally to reveal more columns"
            tabIndex={0}
          >
            <Table className="whitespace-nowrap">
              <TableHeader className="bg-card sticky top-0">
                <TableRow>
                  <TableHead className="whitespace-nowrap">E‑Mail</TableHead>
                  <TableHead className="whitespace-nowrap">Gerät</TableHead>
                  <TableHead className="whitespace-nowrap">Benutzer ID</TableHead>
                  <TableHead className="whitespace-nowrap">Erstellt</TableHead>
                  <TableHead className="whitespace-nowrap">Zuletzt genutzt</TableHead>
                  <TableHead className="text-right whitespace-nowrap">Aktionen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((r) => {
                  const created = new Date(r.createdAt).toLocaleString("de-CH")
                  const last = new Date(r.lastUsedAt).toLocaleString("de-CH")
                  const userLink = (
                    <Link
                      href={`/admin/users/${encodeURIComponent(r.userId)}`}
                      className="underline underline-offset-2"
                    >
                      {r.userId}
                    </Link>
                  )
                  return (
                    <TableRow key={r.userId + ":" + r.familyId} className="even:bg-card odd:bg-muted/40">
                      <TableCell>{r.email || r.userId}</TableCell>
                      <TableCell>{r.device}</TableCell>
                      <TableCell className="text-xs">{userLink}</TableCell>
                      <TableCell className="whitespace-nowrap">{created}</TableCell>
                      <TableCell className="whitespace-nowrap">{last}</TableCell>
                      <TableCell className="text-right">
                        <Link href={`/admin/sessions/${encodeURIComponent(r.userId)}:${encodeURIComponent(r.familyId)}`}>
                          <Button size="sm" variant="outline">Details</Button>
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
