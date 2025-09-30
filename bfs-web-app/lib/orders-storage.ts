"use client"

export type StoredOrder = {
  id: string
  createdAt: string // ISO date
}

const STORAGE_KEY = "bfs-orders"

function read(): StoredOrder[] {
  if (typeof window === "undefined") return []
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as StoredOrder[]
    if (!Array.isArray(parsed)) return []
    return parsed.filter(o => o && typeof o.id === "string")
  } catch {
    return []
  }
}

function write(list: StoredOrder[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(list))
  } catch {
    // ignore
  }
}

export function addOrder(id: string) {
  const list = read()
  const exists = list.some(o => o.id === id)
  if (exists) return
  list.unshift({ id, createdAt: new Date().toISOString() })
  write(list)
}

export function getOrders(): StoredOrder[] {
  return read()
}

export function clearOrders() {
  write([])
}

export function removeOrder(id: string) {
  const list = read().filter(o => o.id !== id)
  write(list)
}

