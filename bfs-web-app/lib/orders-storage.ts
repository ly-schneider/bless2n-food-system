"use client"

import type { CartItem } from "@/types/cart"

export type StoredOrder = {
  id: string
  createdAt: string // ISO date
  items?: CartItem[]
  totalCents?: number
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

export function addOrder(id: string, items?: CartItem[], totalCents?: number) {
  const list = read()
  const idx = list.findIndex(o => o.id === id)
  if (idx >= 0) {
    // If already stored but without items, enrich the record.
    const existing = list[idx]!
    if ((!existing.items || existing.items.length === 0) && items && items.length > 0) {
      list[idx] = { ...existing, items, totalCents }
      write(list)
    }
    return
  }
  list.unshift({ id, createdAt: new Date().toISOString(), items, totalCents })
  write(list)
}

export function getOrders(): StoredOrder[] {
  return read()
}

export function getOrder(id: string): StoredOrder | undefined {
  return read().find(o => o.id === id)
}

export function clearOrders() {
  write([])
}

export function removeOrder(id: string) {
  const list = read().filter(o => o.id !== id)
  write(list)
}
