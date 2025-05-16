import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { MenuItem, Product } from "@/types"

interface OrderDetailsDialogProps {
  isOpen: boolean;
  onClose: () => void;
  order: Record<string, { baseId: string; selections?: Record<string, string>; quantity: number }>;
  menuItems: MenuItem[];
  products: Product[];
  menuSelections: Record<string, Record<string, string>>;
}

export function OrderDetailsDialog({ 
  isOpen, 
  onClose,
  order,
  menuItems,
  products,
  menuSelections,
}: OrderDetailsDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-[800px] max-h-[80vh]">
        <DialogHeader>
          <DialogTitle className="text-xl">📋 Bestellte Produkte</DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4 overflow-y-auto max-h-[calc(80vh-100px)]">
          {Object.entries(order).map(([_, item]) => {
            const menuItem = menuItems.find(m => m.id === item.baseId);
            const product = products.find(p => p.id === item.baseId);
            
            if (menuItem) {
              return (
                <div 
                  key={item.baseId} 
                  style={{ borderLeftColor: menuItem.color || '#e5e7eb' }}
                  className="border-l-4 pl-4 py-4 bg-white rounded-lg shadow-sm"
                >
                  <div className="flex justify-between items-start w-full">
                    <div className="flex-1">
                      <h3 className="font-bold">
                        {menuItem.emoji ? `${menuItem.emoji} ` : ''}{menuItem.name}
                        <span className="text-sm text-gray-600 ml-2">× {item.quantity}</span>
                      </h3>
                      {item.selections && (
                        <div className="text-sm text-gray-600 mt-2">
                          {Object.entries(item.selections).map(([groupId, productId]) => {
                            const product = products.find(p => p.id === productId);
                            return (
                              <div 
                                key={groupId}
                                className="border-l-2 pl-2 ml-2 my-1"
                                style={{ borderLeftColor: product?.color || '#e5e7eb' }}
                              >
                                {product?.emoji ? `${product.emoji} ` : ''}{product?.name}
                              </div>
                            );
                          })}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              );
            }
            
            if (product) {
              return (
                <div 
                  key={item.baseId} 
                  style={{ borderLeftColor: product.color || '#e5e7eb' }}
                  className="border-l-4 pl-4 py-4 bg-white rounded-lg shadow-sm"
                >
                  <div className="flex justify-between items-start w-full">
                    <h3 className="font-bold">
                      {product.emoji ? `${product.emoji} ` : ''}{product.name}
                      <span className="text-sm text-gray-600 ml-2">× {item.quantity}</span>
                    </h3>
                  </div>
                </div>
              );
            }
            
            return null;
          })}
        </div>
      </DialogContent>
    </Dialog>
  );
}
