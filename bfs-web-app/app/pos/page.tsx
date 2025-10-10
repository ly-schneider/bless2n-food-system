"use client"

import { useEffect, useState } from "react"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { BasketPanel } from "@/components/pos/basket-panel"
import { POSHeader } from "@/components/pos/pos-header"
import { ProductGrid } from "@/components/pos/product-grid"
import { RequestAccess } from "@/components/pos/request-access"
import { usePosToken } from "@/components/pos/use-pos-token"
import { CartProvider } from "@/contexts/cart-context"
import { API_BASE_URL } from "@/lib/api"
import { listProducts } from "@/lib/api/products"
import type { ListResponse, ProductDTO } from "@/types"

type PosStatus = { exists: boolean; approved: boolean; name?: string; cardCapable?: boolean | null }

function PosInner() {
  const token = usePosToken()
  const [status, setStatus] = useState<PosStatus | null>(null)
  const [products, setProducts] = useState<ListResponse<ProductDTO>>({ items: [], count: 0 })
  const [configOpen, setConfigOpen] = useState(false)
  const [configProduct, setConfigProduct] = useState<ProductDTO | null>(null)

  // Close product configuration on POS lock
  useEffect(() => {
    const onLock = () => {
      setConfigOpen(false)
      setConfigProduct(null)
    }
    window.addEventListener("pos:lock", onLock)
    return () => window.removeEventListener("pos:lock", onLock)
  }, [])

  useEffect(() => {
    ;(async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/v1/pos/me`, { headers: { "X-Pos-Token": token } })
        const json = await res.json()
        setStatus(json as PosStatus)
      } catch {
        setStatus({ exists: false, approved: false })
      }
    })()
  }, [token])

  useEffect(() => {
    ;(async () => {
      try {
        setProducts(await listProducts())
      } catch {}
    })()
  }, [])

  // Support AdminMainHeader "Aktualisieren" button to refresh POS data
  useEffect(() => {
    const onRefresh = () => {
      ;(async () => {
        try {
          setProducts(await listProducts())
        } catch {}
        try {
          const res = await fetch(`${API_BASE_URL}/v1/pos/me`, { headers: { "X-Pos-Token": token } })
          const json = await res.json()
          setStatus(json as PosStatus)
        } catch {}
      })()
    }
    window.addEventListener("admin:refresh", onRefresh)
    return () => window.removeEventListener("admin:refresh", onRefresh)
  }, [token])

  if (!status?.approved) {
    return (
      <>
        <POSHeader />
        <RequestAccess
          token={token}
          onRefresh={async () => {
            try {
              const res = await fetch(`${API_BASE_URL}/v1/pos/me`, { headers: { "X-Pos-Token": token } })
              const json = await res.json()
              setStatus(json as PosStatus)
            } catch {}
          }}
        />
      </>
    )
  }

  return (
    <>
      <POSHeader />
      <div className="grid h-[calc(100dvh-4rem)] grid-cols-1 overflow-hidden md:grid-cols-[1fr_450px]">
        <div className="min-h-0 overflow-hidden">
          <ProductGrid
            products={products}
            onConfigure={(p) => {
              setConfigProduct(p)
              setConfigOpen(true)
            }}
          />
        </div>
        <BasketPanel token={token} />

        {configProduct && (
          <ProductConfigurationModal
            product={configProduct}
            isOpen={configOpen}
            onClose={() => {
              setConfigOpen(false)
              setConfigProduct(null)
            }}
          />
        )}
      </div>
    </>
  )
}

export default function POSPage() {
  return (
    <CartProvider>
      <PosInner />
    </CartProvider>
  )
}
