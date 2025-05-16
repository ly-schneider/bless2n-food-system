'use client'

import React, { createContext, useState, useContext, ReactNode, useMemo } from 'react'

export type CartItem = {
  id: number
  name: string
  price: number
  quantity: number
}

interface CartContextType {
  items: CartItem[]
  addItem: (item: CartItem) => void
  removeItem: (id: number) => void
  updateQuantity: (id: number, quantity: number) => void
  clearCart: () => void
  getTotal: () => number
  /** Liefert die aktuellen Bons als Mapping von Preis zu Anzahl */
  getBons: () => Record<string, number>
}

const CartContext = createContext<CartContextType | undefined>(undefined)

export function CartProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<CartItem[]>([])

  const addItem = (newItem: CartItem) => {
    setItems(currentItems => {
      const idx = currentItems.findIndex(item => item.id === newItem.id)
      if (idx > -1) {
        const updated = [...currentItems]
        updated[idx] = { ...updated[idx], quantity: updated[idx].quantity + 1 }
        return updated
      }
      return [...currentItems, newItem]
    })
  }

  const removeItem = (id: number) => {
    setItems(currentItems => currentItems.filter(item => item.id !== id))
  }

  const updateQuantity = (id: number, quantity: number) => {
    setItems(currentItems =>
      currentItems.map(item =>
        item.id === id ? { ...item, quantity } : item
      )
    )
  }

  const clearCart = () => setItems([])

  const getTotal = () => items.reduce((sum, i) => sum + i.price * i.quantity, 0)

  // Bon-Konfiguration
  const bonConfig: Record<string, { label: string }> = useMemo(() => ({
    '2.50': { label: 'Magenta' },
    '4.00': { label: 'Green' },
    '4.50': { label: 'White' },
    '5.00': { label: 'Orange' },
    '7.00': { label: 'Yellow' },
  }), [])

  // Berechnet Bons anhand aktueller Items (nur perf. Realtime im Provider)
  const bonsMap = useMemo(() => {
    const counts: Record<string, number> = {}
    items.forEach(item => {
      const key = item.price.toFixed(2)
      if (bonConfig[key]) {
        counts[key] = (counts[key] || 0) + item.quantity
      }
    })
    return counts
  }, [items, bonConfig])

  const getBons = () => bonsMap

  return (
    <CartContext.Provider value={{ items, addItem, removeItem, updateQuantity, clearCart, getTotal, getBons }}>
      {children}
    </CartContext.Provider>
  )
}

export function useCart() {
  const context = useContext(CartContext)
  if (!context) throw new Error('useCart must be used within a CartProvider')
  return context
}