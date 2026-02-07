"use client"
import { useEffect, useState } from "react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"

import { getCSRFToken } from "@/lib/csrf"
import { readErrorMessage } from "@/lib/http"

type Category = { id: string; name: string; isActive: boolean; position: number }

export default function AdminCategoriesPage() {
  const fetchAuth = useAuthorizedFetch()
  const [items, setItems] = useState<Category[]>([])
  const [name, setName] = useState("")
  const [position, setPosition] = useState<number>(0)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void reload()
  }, [])

  async function reload() {
    try {
      const res = await fetchAuth(`/api/v1/categories`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = (await res.json()) as { items: Category[] }
      const sorted = (data.items || []).slice().sort((a, b) => a.position - b.position || a.name.localeCompare(b.name))
      setItems(sorted)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to load"
      setError(msg)
    }
  }

  async function createCategory() {
    if (!name.trim()) return
    if (!Number.isFinite(position) || position < 0) {
      setError("Position muss >= 0 sein")
      return
    }
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/categories`, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ name: name.trim(), position }),
    })
    if (!res.ok) {
      setError(await readErrorMessage(res))
      return
    }
    setName("")
    setPosition(0)
    await reload()
  }

  // removed unused rename() helper; name+active are updated together

  async function updatePosition(id: string, pos: number) {
    if (!Number.isFinite(pos) || pos < 0) {
      setError("Position muss >= 0 sein")
      return
    }
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/categories/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ position: pos }),
    })
    if (!res.ok) {
      setError(await readErrorMessage(res))
      return
    }
    await reload()
  }

  async function toggle(id: string, isActive: boolean) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/categories/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ isActive }),
    })
    if (!res.ok) {
      setError(await readErrorMessage(res))
      return
    }
    await reload()
  }

  async function remove(id: string) {
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/categories/${id}`, {
      method: "DELETE",
      headers: { "X-CSRF": csrf || "" },
    })
    if (!res.ok) {
      setError(await readErrorMessage(res))
      return
    }
    await reload()
  }

  async function rename(id: string, newName: string) {
    const trimmed = newName.trim()
    if (!trimmed) return
    const csrf = getCSRFToken()
    const res = await fetchAuth(`/api/v1/categories/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", "X-CSRF": csrf || "" },
      body: JSON.stringify({ name: trimmed }),
    })
    if (!res.ok) {
      setError(await readErrorMessage(res))
      return
    }
    await reload()
  }

  return (
    <div className="min-w-0 space-y-4">
      <h1 className="text-xl font-semibold">Kategorien</h1>
      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="flex items-center gap-2">
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Neue Kategorie"
          className="h-8 w-64"
        />
        <Input
          type="number"
          value={position}
          onChange={(e) => setPosition(Number(e.target.value))}
          placeholder="Position"
          className="h-8 w-24"
        />
        <Button variant="outline" size="sm" className="h-8" onClick={() => void createCategory()}>
          Erstellen
        </Button>
      </div>

      <div className="rounded-md border">
        <div className="relative">
          <div className="from-background pointer-events-none absolute inset-y-0 left-0 w-6 bg-gradient-to-r to-transparent" />
          <div className="from-background pointer-events-none absolute inset-y-0 right-0 w-6 bg-gradient-to-l to-transparent" />
          <div
            className="max-w-full overflow-x-auto overscroll-x-contain rounded-[10px]"
            role="region"
            aria-label="Categories table – scroll horizontally to reveal more columns"
            tabIndex={0}
          >
            <Table className="whitespace-nowrap">
              <TableHeader className="bg-card sticky top-0">
                <TableRow>
                  <TableHead className="whitespace-nowrap">Position</TableHead>
                  <TableHead className="whitespace-nowrap">Name</TableHead>
                  <TableHead className="whitespace-nowrap">Status</TableHead>
                  <TableHead className="text-right whitespace-nowrap">Aktionen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((c) => (
                  <TableRow key={c.id} className="even:bg-card odd:bg-muted/40">
                    <TableCell className="w-24">
                      <Input
                        type="number"
                        value={c.position}
                        onChange={(e) => {
                          const v = Number(e.target.value)
                          setItems((prev) => prev.map((it) => (it.id === c.id ? { ...it, position: v } : it)))
                        }}
                        onBlur={(e) => void updatePosition(c.id, Number(e.target.value))}
                        className="h-7"
                      />
                    </TableCell>
                    <TableCell>
                      <Input
                        value={c.name}
                        onChange={(e) => {
                          setItems((prev) => prev.map((it) => (it.id === c.id ? { ...it, name: e.target.value } : it)))
                        }}
                        onBlur={(e) => void rename(c.id, e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            e.preventDefault()
                            e.currentTarget.blur()
                          }
                        }}
                        maxLength={20}
                        className="h-7 w-40"
                      />
                    </TableCell>
                    <TableCell>
                      <label className="inline-flex items-center gap-2">
                        <Switch checked={c.isActive} onCheckedChange={(v) => void toggle(c.id, v)} />
                        <span>{c.isActive ? "Aktiv" : "Inaktiv"}</span>
                      </label>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="inline-flex items-center gap-1">
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-7 text-red-700">
                              Löschen
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Kategorie löschen?</AlertDialogTitle>
                              <AlertDialogDescription>
                                Diese Aktion kann nicht rückgängig gemacht werden. Kategorie "{c.name}" dauerhaft
                                löschen?
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Abbrechen</AlertDialogCancel>
                              <AlertDialogAction onClick={() => void remove(c.id)}>Löschen</AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>

    </div>
  )
}

// CSRF helper now centralized in lib/csrf
