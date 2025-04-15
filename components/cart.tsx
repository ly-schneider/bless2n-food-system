"use client";

import { useState } from "react";
import { useCart, CartItem } from "@/contexts/cart-context";
import { createClient } from "@/utils/supabase/client";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { MinusCircle, PlusCircle, ShoppingCart, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

export function Cart() {
  const { items, removeItem, updateQuantity, clearCart, getTotal } = useCart();
  const [checkoutModalOpen, setCheckoutModalOpen] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const supabase = createClient();

  const handleCheckout = async () => {
    if (items.length === 0) return;

    setIsProcessing(true);

    try {
      // Get current user
      const {
        data: { user },
      } = await supabase.auth.getUser();

      const orderBody = {
        admin_id: user?.id,
        order_date: new Date().toISOString(),
        status: "completed",
        total: getTotal(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      // Create order in database
      const { data: order, error: orderError } = await supabase
        .from("orders")
        .insert(orderBody)
        .select();

      if (orderError) throw orderError;

      // Add order items
      if (order && order.length > 0) {
        const orderId = order[0].id;

        // Prepare order items for insertion
        const orderItems = items.map((item) => ({
          order_id: orderId,
          product_id: item.id,
          quantity: item.quantity,
          price_at_order: item.price,
          created_at: new Date().toISOString(),
        }));

        const { error: orderItemsError } = await supabase
          .from("order_items")
          .insert(orderItems);

        if (orderItemsError) throw orderItemsError;
      }

      setCheckoutModalOpen(false);
      clearCart();
    } catch (error) {
      console.error("Error processing order:", error);
    } finally {
      setIsProcessing(false);
    }
  };

  const handleCloseModal = () => {
    setCheckoutModalOpen(false);
  };

  const handleQuantityChange = (item: CartItem, change: number) => {
    const newQuantity = Math.max(1, item.quantity + change);
    updateQuantity(item.id, newQuantity);
  };

  const subtotal = getTotal();

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 pb-4 pt-2 sticky top-0 z-10">
        <h2 className="text-2xl font-medium flex items-center gap-2">
          Bestellung
        </h2>
      </div>

      <div className="flex-1 overflow-auto px-4 py-3">
        {items.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-40 text-muted-foreground">
            <ShoppingCart className="h-10 w-10 mb-2 opacity-50" />
            <p>Keine Produkte</p>
          </div>
        ) : (
          <ul className="space-y-4">
            {items.map((item) => (
              <li
                key={item.id}
                className="flex gap-3 group relative pb-3 border-b border-border last:border-b-0"
              >
                <div className="flex-1 min-w-0">
                  <div className="flex justify-between">
                    <h3 className="font-medium line-clamp-1">{item.name}</h3>
                    <button
                      onClick={() => removeItem(item.id)}
                      className="text-muted-foreground hover:text-destructive transition-colors"
                      aria-label="Remove item"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </div>

                  <div className="text-sm mt-1">
                    CHF {item.price.toFixed(2)}
                  </div>

                  <div className="flex items-center gap-2 mt-2">
                    <button
                      onClick={() => handleQuantityChange(item, -1)}
                      className="text-primary hover:text-primary/80 transition-colors"
                      aria-label="Decrease quantity"
                    >
                      <MinusCircle className="h-4 w-4" />
                    </button>
                    <span className="w-8 text-center">{item.quantity}</span>
                    <button
                      onClick={() => handleQuantityChange(item, 1)}
                      className="text-primary hover:text-primary/80 transition-colors"
                      aria-label="Increase quantity"
                    >
                      <PlusCircle className="h-4 w-4" />
                    </button>

                    <div className="ml-auto text-sm font-medium">
                      CHF {(item.price * item.quantity).toFixed(2)}
                    </div>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className="border-t p-4 space-y-4 bg-background/95 backdrop-blur-sm">
        <div className="flex justify-between text-sm">
          <span>Subtotal</span>
          <span>CHF {subtotal.toFixed(2)}</span>
        </div>

        <div className="flex justify-between font-medium">
          <span>Total</span>
          <span>CHF {subtotal.toFixed(2)}</span>
        </div>

        <div className="pt-2">
          <Button
            className="w-full"
            size="lg"
            disabled={items.length === 0}
            onClick={() => setCheckoutModalOpen(true)}
          >
            Abschliessen
          </Button>
        </div>
      </div>

      <Dialog open={checkoutModalOpen} onOpenChange={handleCloseModal}>
        <DialogContent>
          <DialogHeader className="mb-4">
            <DialogTitle>Abschliessen</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 max-h-[40vh] overflow-y-auto pr-2 mb-4">
            {items.map((item) => (
              <div
                key={item.id}
                className="flex justify-between items-center text-sm"
              >
                <span>
                  {item.quantity} x {item.name}
                </span>
                <span>CHF {(item.price * item.quantity).toFixed(2)}</span>
              </div>
            ))}

            <Separator />

            <div className="flex justify-between font-bold text-lg">
              <span>Total</span>
              <span>CHF {subtotal.toFixed(2)}</span>
            </div>

            {/* <Separator />
                  Implement Bons */}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleCloseModal}
              className="w-full"
            >
              Abbrechen
            </Button>
            <Button
              type="button"
              disabled={isProcessing}
              onClick={handleCheckout}
              className="w-full"
            >
              {isProcessing ? "Verarbeitung..." : "Abschliessen"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
