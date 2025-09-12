"use client"

import { DollarSign, Edit, Loader2, Package, Plus, Trash2 } from "lucide-react"
import { useEffect, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import ProductAPI from "@/lib/api/products"
import { CreateProductBundleRequest, Product, ProductBundle, UpdateProductBundleRequest } from "@/types/product"

export default function BundlesPage() {
  const [bundles, setBundles] = useState<ProductBundle[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [editingBundle, setEditingBundle] = useState<ProductBundle | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      setIsLoading(true)

      // Load products for bundle creation
      const productsResponse = await ProductAPI.listProducts({ limit: 100 })
      setProducts(productsResponse.products)

      // TODO: Load existing bundles when backend endpoint is ready
      setBundles([])
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateBundle = async (
    bundleData: CreateProductBundleRequest | UpdateProductBundleRequest
  ) => {
    try {
      const newBundle = await ProductAPI.createProductBundle(
        bundleData as CreateProductBundleRequest
      )
      setBundles((prev) => [...prev, newBundle])
      setShowCreateDialog(false)
    } catch (error) {
      console.error("Failed to create bundle:", error)
      alert("Failed to create bundle. Please try again.")
    }
  }

  const handleUpdateBundle = async (id: string, bundleData: UpdateProductBundleRequest) => {
    try {
      const updatedBundle = await ProductAPI.updateProductBundle(id, bundleData)
      setBundles((prev) => prev.map((bundle) => (bundle.id === id ? updatedBundle : bundle)))
      setEditingBundle(null)
    } catch (error) {
      console.error("Failed to update bundle:", error)
      alert("Failed to update bundle. Please try again.")
    }
  }


  const calculateBundleValue = (productIds: string[]) => {
    return productIds.map((id) => products.find((p) => p.id === id)?.price || 0).reduce((sum, price) => sum + price, 0)
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mx-auto max-w-6xl">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Product Bundles</h1>
            <p className="text-gray-600">Create and manage product bundles (menus) for your stations</p>
          </div>

          <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                Create Bundle
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <BundleForm
                products={products}
                onSubmit={handleCreateBundle}
                onCancel={() => setShowCreateDialog(false)}
              />
            </DialogContent>
          </Dialog>
        </div>

        {bundles.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <Package className="mx-auto mb-4 h-12 w-12 text-gray-400" />
              <h3 className="mb-2 text-lg font-medium text-gray-900">No bundles yet</h3>
              <p className="mb-4 text-gray-600">
                Create your first product bundle to offer curated menu combinations to customers.
              </p>
              <Button onClick={() => setShowCreateDialog(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create First Bundle
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-6">
            {bundles.map((bundle) => (
              <Card key={bundle.id}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="flex items-center gap-2">
                        {bundle.name}
                        {!bundle.isActive && <Badge variant="secondary">Inactive</Badge>}
                      </CardTitle>
                      <CardDescription>{bundle.description}</CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button variant="outline" size="sm" onClick={() => setEditingBundle(bundle)}>
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="outline" size="sm" className="text-red-600">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-6 md:grid-cols-2">
                    <div>
                      <h4 className="mb-2 font-medium">Products in Bundle:</h4>
                      <div className="space-y-2">
                        {bundle.productIds.map((productId) => {
                          const product = products.find((p) => p.id === productId)
                          return product ? (
                            <div key={productId} className="flex items-center justify-between text-sm">
                              <span>{product.name}</span>
                              <span className="text-gray-600">${product.price.toFixed(2)}</span>
                            </div>
                          ) : null
                        })}
                      </div>
                    </div>

                    <div>
                      <h4 className="mb-2 font-medium">Pricing:</h4>
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span>Individual Total:</span>
                          <span>${calculateBundleValue(bundle.productIds).toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between font-medium text-green-600">
                          <span>Bundle Price:</span>
                          <span>${bundle.price.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between text-green-600">
                          <span>Savings:</span>
                          <span>${(calculateBundleValue(bundle.productIds) - bundle.price).toFixed(2)}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Edit Bundle Dialog */}
        <Dialog open={!!editingBundle} onOpenChange={(open) => !open && setEditingBundle(null)}>
          <DialogContent className="max-w-2xl">
            {editingBundle && (
              <BundleForm
                products={products}
                bundle={editingBundle}
                onSubmit={(data) => handleUpdateBundle(editingBundle.id, data)}
                onCancel={() => setEditingBundle(null)}
                isEditing
              />
            )}
          </DialogContent>
        </Dialog>
      </div>
    </div>
  )
}

interface BundleFormProps {
  products: Product[]
  bundle?: ProductBundle
  onSubmit: (data: CreateProductBundleRequest | UpdateProductBundleRequest) => void
  onCancel: () => void
  isEditing?: boolean
}

function BundleForm({ products, bundle, onSubmit, onCancel, isEditing }: BundleFormProps) {
  const [formData, setFormData] = useState({
    name: bundle?.name || "",
    description: bundle?.description || "",
    price: bundle?.price || 0,
    productIds: bundle?.productIds || [],
  })
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleProductToggle = (productId: string, checked: boolean) => {
    setFormData((prev) => ({
      ...prev,
      productIds: checked ? [...prev.productIds, productId] : prev.productIds.filter((id) => id !== productId),
    }))
  }

  const calculateTotalValue = () => {
    return formData.productIds
      .map((id) => products.find((p) => p.id === id)?.price || 0)
      .reduce((sum, price) => sum + price, 0)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (formData.productIds.length === 0) {
      alert("Please select at least one product for the bundle.")
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(formData)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>{isEditing ? "Edit Bundle" : "Create New Bundle"}</DialogTitle>
        <DialogDescription>
          {isEditing
            ? "Update the bundle details below."
            : "Create a new product bundle by selecting products and setting a price."}
        </DialogDescription>
      </DialogHeader>

      <form onSubmit={handleSubmit}>
        <div className="space-y-6 py-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="name">Bundle Name *</Label>
              <Input
                id="name"
                required
                value={formData.name}
                onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
                placeholder="e.g., Lunch Combo, Breakfast Special"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="price">Bundle Price *</Label>
              <div className="relative">
                <DollarSign className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform text-gray-400" />
                <Input
                  id="price"
                  type="number"
                  step="0.01"
                  min="0"
                  required
                  value={formData.price}
                  onChange={(e) => setFormData((prev) => ({ ...prev, price: parseFloat(e.target.value) || 0 }))}
                  placeholder="0.00"
                  className="pl-10"
                />
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
              placeholder="Describe what's included in this bundle"
              rows={3}
            />
          </div>

          <div className="space-y-4">
            <Label>Products in Bundle *</Label>
            <div className="max-h-64 overflow-y-auto rounded-lg border p-3">
              <div className="space-y-3">
                {products
                  .filter((p) => p.isActive)
                  .map((product) => (
                    <div key={product.id} className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <Checkbox
                          id={`product-${product.id}`}
                          checked={formData.productIds.includes(product.id)}
                          onCheckedChange={(checked) => handleProductToggle(product.id, checked as boolean)}
                        />
                        <label htmlFor={`product-${product.id}`} className="text-sm font-medium">
                          {product.name}
                        </label>
                      </div>
                      <span className="text-sm text-gray-600">${product.price.toFixed(2)}</span>
                    </div>
                  ))}
              </div>
            </div>

            {formData.productIds.length > 0 && (
              <div className="space-y-2 rounded-lg bg-blue-50 p-3">
                <div className="flex justify-between text-sm">
                  <span>Individual Total:</span>
                  <span>${calculateTotalValue().toFixed(2)}</span>
                </div>
                <div className="flex justify-between text-sm font-medium">
                  <span>Bundle Price:</span>
                  <span>${formData.price.toFixed(2)}</span>
                </div>
                <div className="flex justify-between text-sm text-green-600">
                  <span>Customer Savings:</span>
                  <span>${Math.max(0, calculateTotalValue() - formData.price).toFixed(2)}</span>
                </div>
              </div>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting || formData.productIds.length === 0}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {isEditing ? "Updating..." : "Creating..."}
              </>
            ) : isEditing ? (
              "Update Bundle"
            ) : (
              "Create Bundle"
            )}
          </Button>
        </DialogFooter>
      </form>
    </>
  )
}
