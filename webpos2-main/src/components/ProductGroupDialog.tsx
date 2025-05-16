import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { ProductGroup, ProductGroupItem } from "@/types"
import { cn } from "@/lib/utils"

interface ProductGroupDialogProps {
  productGroup: ProductGroup;
  onSelect: (selectedProduct: ProductGroupItem) => void;
  onClose: () => void;
  stock: Record<string, number>;
}

export function ProductGroupDialog({
  productGroup,
  onSelect,
  onClose,
  stock,
}: ProductGroupDialogProps) {
  return (
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{`${productGroup.name} auswählen`}</DialogTitle>
        </DialogHeader>
        <div className="grid grid-cols-1 gap-4 py-4">
          {productGroup.products.map((product) => {
            const stockCount = stock[product.id] || 0;
            const isOutOfStock = stockCount === 0;
            
            return (
              <Button
                key={product.id}
                variant="outline"
                className={cn(
                  "flex items-center justify-between gap-3 p-6 border-2",
                  product.color && "hover:bg-opacity-10 hover:bg-black",
                  isOutOfStock && "opacity-50 cursor-not-allowed"
                )}
                style={{
                  borderColor: product.color || undefined,
                  backgroundColor: product.color ? `${product.color}10` : undefined
                }}
                onClick={() => !isOutOfStock && onSelect(product)}
                disabled={isOutOfStock}
              >
                <div className="flex items-center gap-3">
                  {product.emoji && (
                    <span className="text-xl">{product.emoji}</span>
                  )}
                  <span>{product.name}</span>
                </div>
                <span className={cn(
                  "text-sm",
                  isOutOfStock ? "text-red-500" : "text-gray-500"
                )}>
                  {stockCount} verfügbar
                </span>
              </Button>
            );
          })}
        </div>
      </DialogContent>
    </Dialog>
  )
}
