"use client"

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
}

export function ProductConfigurationModal({ product, isOpen, onClose, initialConfiguration }: ProductConfigurationModalProps) {
  const { addToCart } = useCart();
  const [selectedConfiguration, setSelectedConfiguration] = useState<CartItemConfiguration>(initialConfiguration || {});
  
  const handleSlotSelection = (slotId: string, productId: string) => {
    setSelectedConfiguration(prev => ({
      ...prev,
      [slotId]: productId,
    }));
  };
  
  const handleAddToCart = () => {
    addToCart(product, selectedConfiguration);
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
    
    return product.menu.slots.every(slot => 
      selectedConfiguration[slot.id] !== undefined
    );
  };
  
  if (!product.menu?.slots) {
    return null;
  }
  
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-family-secondary">
            {product.name} konfigurieren
          </DialogTitle>
          <DialogDescription>
            Wählen Sie Ihre gewünschten Optionen aus.
          </DialogDescription>
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
              variant="outline"
              onClick={onClose}
              className="flex-1"
            >
              Abbrechen
            </Button>
            <Button
              onClick={handleAddToCart}
              disabled={!isConfigurationComplete()}
              className="flex-1"
            >
              Zum Warenkorb hinzufügen
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
        {slot.menuSlotItems.map((item) => (
          <Card
            key={item.id}
            className={`cursor-pointer transition-all hover:shadow-md ${
              selectedProductId === item.id
                ? 'ring-2 ring-primary bg-primary/5'
                : 'hover:bg-gray-50'
            }`}
            onClick={() => onSelect(item.id)}
          >
            <CardContent className="p-3">
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-3">
                  {item.image && (
                    <div className="w-12 h-12 bg-gray-200 rounded-lg overflow-hidden">
                      <img
                        src={item.image}
                        alt={item.name}
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
        ))}
      </div>
    </div>
  );
}