"use client"

import { useEffect, useMemo, useState } from "react"
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
}

export function useBestMenuSuggestion(products?: ProductDTO[]): BestMenuSuggestionState {
  const { cart } = useCart()
  const [allProducts, setAllProducts] = useState<ProductDTO[] | null>(products ?? null)
  const [suggestion, setSuggestion] = useState<MenuSuggestion | null>(null)

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
    setSuggestion(best)
  }, [cart, allProducts])

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

  return { suggestion, allProductsReady: !!allProducts, contiguous, startIndex, endIndex }
}
