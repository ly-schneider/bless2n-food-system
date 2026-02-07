"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { useCart } from "@/contexts/cart-context"
import { listProducts } from "@/lib/api/products"
import { findBestMenuSuggestion, MenuSuggestion } from "@/lib/menu-suggestions"
import type { ProductDTO } from "@/types"

export interface BestMenuSuggestionState {
  suggestion: MenuSuggestion | null
  allProductsReady: boolean
  contiguous: boolean
  startIndex: number
  endIndex: number
  dismissSuggestion: () => void
}

export function useBestMenuSuggestion(products?: ProductDTO[]): BestMenuSuggestionState {
  const { cart } = useCart()
  const [allProducts, setAllProducts] = useState<ProductDTO[] | null>(products ?? null)
  const [rawSuggestion, setRawSuggestion] = useState<MenuSuggestion | null>(null)
  const [dismissedMenuId, setDismissedMenuId] = useState<string | null>(null)
  const [dismissedCartKey, setDismissedCartKey] = useState<string | null>(null)

  const cartKey = useMemo(
    () => cart.items.map((i) => `${i.product.id}:${i.quantity}`).join(","),
    [cart.items]
  )

  useEffect(() => {
    if (dismissedMenuId && cartKey !== dismissedCartKey) {
      setDismissedMenuId(null)
      setDismissedCartKey(null)
    }
  }, [cartKey, dismissedMenuId, dismissedCartKey])

  useEffect(() => {
    let cancelled = false
    async function ensureProducts() {
      if (allProducts) return
      try {
        const res = await listProducts()
        if (!cancelled) setAllProducts(res.items)
      } catch {
        if (!cancelled) setAllProducts([])
      }
    }
    ensureProducts()
    return () => {
      cancelled = true
    }
  }, [allProducts])

  useEffect(() => {
    if (!allProducts) return
    const best = findBestMenuSuggestion(cart, allProducts)
    setRawSuggestion(best)
  }, [cart, allProducts])

  const suggestion = useMemo(() => {
    if (!rawSuggestion) return null
    if (dismissedMenuId && rawSuggestion.menuProduct.id === dismissedMenuId) return null
    return rawSuggestion
  }, [rawSuggestion, dismissedMenuId])

  const dismissSuggestion = useCallback(() => {
    if (rawSuggestion) {
      setDismissedMenuId(rawSuggestion.menuProduct.id)
      setDismissedCartKey(cartKey)
    }
  }, [rawSuggestion, cartKey])

  const { contiguous, startIndex, endIndex } = useMemo(() => {
    if (!suggestion) return { contiguous: false, startIndex: -1, endIndex: -1 }
    const involvedIds = new Set(suggestion.sourceItems.map((s) => s.cartItem.id))
    const indices: number[] = []
    cart.items.forEach((it, idx) => {
      if (involvedIds.has(it.id)) indices.push(idx)
    })
    if (indices.length !== suggestion.sourceItems.length) return { contiguous: false, startIndex: -1, endIndex: -1 }
    const min = Math.min(...indices)
    const max = Math.max(...indices)
    for (let i = min; i <= max; i++) {
      const id = cart.items[i]?.id
      if (!id || !involvedIds.has(id)) return { contiguous: false, startIndex: -1, endIndex: -1 }
    }
    return { contiguous: true, startIndex: min, endIndex: max }
  }, [cart.items, suggestion])

  return { suggestion, allProductsReady: !!allProducts, contiguous, startIndex, endIndex, dismissSuggestion }
}
