"use client"

import Image from "next/image";
import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useCart } from "@/contexts/cart-context";
import { CartItemConfiguration, MenuSlotDTO, ProductDTO } from "@/types";

interface ProductConfigurationModalProps {
  product: ProductDTO;
  isOpen: boolean;
  onClose: () => void;
  initialConfiguration?: CartItemConfiguration;
  editingItemId?: string; // when provided, modal acts in edit mode
}

export function ProductConfigurationModal({ product, isOpen, onClose, initialConfiguration, editingItemId }: ProductConfigurationModalProps) {
  const { addToCart, updateItemConfiguration } = useCart();
  const [selectedConfiguration, setSelectedConfiguration] = useState<CartItemConfiguration>(initialConfiguration || {});
  
  const handleSlotSelection = (slotId: string, productId: string) => {
    setSelectedConfiguration(prev => ({
      ...prev,
      [slotId]: productId,
    }));
  };
  
  const handleSave = () => {
    if (editingItemId) {
      updateItemConfiguration(editingItemId, product, selectedConfiguration);
    } else {
      addToCart(product, selectedConfiguration);
    }
    setSelectedConfiguration(initialConfiguration || {});
    onClose();
  };

  // Reset configuration when modal opens/closes
  React.useEffect(() => {
    if (isOpen) {
      setSelectedConfiguration(initialConfiguration || {});
    }
  }, [isOpen, initialConfiguration]);
  
  const isConfigurationComplete = () => {
    if (!product.menu?.slots) return true;
    
    return product.menu.slots.every(slot => selectedConfiguration[slot.id] !== undefined);
  };

  const isConfigurationValid = () => {
    if (!product.menu?.slots) return true;
    // Ensure selected products are still active
    return product.menu.slots.every(slot => {
      const sel = selectedConfiguration[slot.id];
      if (!sel) return false;
      const item = slot.menuSlotItems?.find(it => it.id === sel);
      if (!item) return false;
      // must be active and available
      const isActive = item.isActive !== false;
      const isAvailable = item.isAvailable !== false;
      return isActive && isAvailable;
    });
  }
  
  if (!product.menu?.slots) {
    return null;
  }
  
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-family-primary">
            Konfigurieren
          </DialogTitle>
        </DialogHeader>
        
        <div className="space-y-6">
          {product.menu.slots.map((slot) => (
            <MenuSlotSelector
              key={slot.id}
              slot={slot}
              selectedProductId={selectedConfiguration[slot.id]}
              onSelect={(productId) => handleSlotSelection(slot.id, productId)}
            />
          ))}
        </div>
        
        <DialogFooter className="flex-col gap-4 sm:flex-col">
          <div className="flex gap-2 w-full">
            <Button
              onClick={handleSave}
              disabled={!isConfigurationComplete() || !isConfigurationValid()}
              className="flex-1 rounded-pill h-12 text-base font-medium"
            >
              {editingItemId ? 'Im Warenkorb aktualisieren' : 'Zum Warenkorb hinzufügen'}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface MenuSlotSelectorProps {
  slot: MenuSlotDTO;
  selectedProductId?: string;
  onSelect: (productId: string) => void;
}

function MenuSlotSelector({ slot, selectedProductId, onSelect }: MenuSlotSelectorProps) {
  if (!slot.menuSlotItems) {
    return null;
  }
  
  return (
    <div className="space-y-3">
      <h3 className="font-family-secondary font-medium text-lg">
        {slot.name}
      </h3>
      
      <div className="grid gap-2">
        {slot.menuSlotItems.map((item) => {
          const isActive = item.isActive !== false;
          const isAvailable = item.isAvailable !== false;
          const isSelected = selectedProductId === item.id;
          const isLowStock = item.isLowStock === true;
          const qty = typeof item.availableQuantity === 'number' ? item.availableQuantity : null;
          return (
            <Card
              key={item.id}
              className={`relative transition-all hover:shadow-md ${
                isSelected ? 'ring-2 ring-primary bg-primary/5' : 'hover:bg-gray-50'
              } ${(!isActive || !isAvailable) ? 'opacity-60 grayscale pointer-events-none' : 'cursor-pointer'}`}
              onClick={() => { if (isActive && isAvailable) onSelect(item.id) }}
              aria-disabled={!isActive || !isAvailable}
            >
              {!isActive && (
                <span className="absolute right-2 top-2 z-10 px-2 py-0.5 text-xs font-medium text-white bg-zinc-700 rounded-full">
                  Nicht verfügbar
                </span>
              )}
              {isActive && !isAvailable && (
                <span className="absolute right-2 top-2 z-10 px-2 py-0.5 text-xs font-medium text-white bg-red-600 rounded-full">
                  Ausverkauft
                </span>
              )}
              {isActive && isAvailable && isLowStock && (
                <span className="absolute right-2 top-2 z-10 px-2 py-0.5 text-xs font-medium text-white bg-amber-600 rounded-full">
                  {qty !== null ? `Nur ${qty} übrig` : 'Geringer Bestand'}
                </span>
              )}
              <CardContent className="p-3">
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    {item.image && (
                      <div className="w-12 h-12 bg-gray-200 rounded-lg overflow-hidden relative">
                        <Image
                          src={item.image}
                          alt={item.name}
                          fill
                          sizes="48px"
                          quality={90}
                          className="w-full h-full object-cover"
                        />
                      </div>
                    )}
                    <div>
                      <h4 className="font-family-secondary font-medium">
                        {item.name}
                      </h4>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  );
}
