"use client";

import { useState } from "react";
import { useCart, CartItem } from "@/contexts/cart-context";
import { createClient } from "@/utils/supabase/client";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Minus, Plus, ShoppingCart,
  X
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { toast } from "sonner";

const bonConfigUI: Record<string, { bg: string; text: string; label: string }> =
  {
    "2.50": { bg: "bg-purple-300", text: "text-magenta-900", label: "Magenta" },
    "4.00": { bg: "bg-green-300", text: "text-green-900", label: "Grün" },
    "4.50": { bg: "bg-gray-300", text: "text-gray-900", label: "Weiss" },
    "5.00": { bg: "bg-orange-300", text: "text-orange-900", label: "Orange" },
    "7.00": { bg: "bg-yellow-300", text: "text-yellow-900", label: "Gelb" },
  };

export function Cart() {
  const { items, removeItem, updateQuantity, clearCart, getTotal, getBons } =
    useCart();
  const [checkoutModalOpen, setCheckoutModalOpen] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const supabase = createClient();

  const handleCheckout = async () => {
    if (!items.length) return;
    setIsProcessing(true);
    try {
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
      const { data: order, error: orderError } = await supabase
        .from("orders")
        .insert(orderBody)
        .select();
      if (orderError) throw orderError;
      if (order?.length) {
        const orderId = order[0].id;
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
      toast.success("Bestellung abgeschlossen!", { duration: 1000 });
    } catch (error) {
      console.error("Error processing order:", error);
    } finally {
      setIsProcessing(false);
    }
  };

  const handleQuantityChange = (item: CartItem, change: number) => {
    if (item.quantity + change <= 0) {
      removeItem(item.id);
    } else {
      const qty = item.quantity + change;
      updateQuantity(item.id, qty);
    }
  };

  const subtotal = getTotal();
  const bonsCount = getBons();

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
                    <Button
                      variant={"outline"}
                      size="sm"
                      onClick={() => removeItem(item.id)}
                      className="text-primary hover:text-destructive transition-colors"
                      aria-label="Remove item"
                    >
                      <X className="h-2 w-2" />
                    </Button>
                  </div>

                  <div className="text-sm mt-1">
                    CHF {item.price.toFixed(2)}
                  </div>

                  <div className="flex items-center gap-2 mt-2">
                    <Button
                      variant={"outline"}
                      onClick={() => handleQuantityChange(item, -1)}
                      className="text-primary hover:text-primary/80 transition-colors"
                      aria-label="Decrease quantity"
                      size={"sm"}
                    >
                      <Minus className="h-4 w-4" />
                    </Button>
                    <span className="w-8 text-center">{item.quantity}</span>
                    <Button
                      variant={"outline"}
                      onClick={() => handleQuantityChange(item, 1)}
                      className="text-primary hover:text-primary/80 transition-colors"
                      aria-label="Increase quantity"
                      size={"sm"}
                    >
                      <Plus className="h-4 w-4" />
                    </Button>

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

      <Dialog
        open={checkoutModalOpen}
        onOpenChange={() => setCheckoutModalOpen(false)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Abschliessen</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4 overflow-y-auto pr-2 mb-4">
            <div className="flex flex-col gap-2">
              <div className="flex flex-col gap-2">
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
              </div>
              <div className="flex justify-between font-medium text-xl">
                <span>Total</span>
                <span>CHF {subtotal.toFixed(2)}</span>
              </div>
            </div>
            <Separator />
            <div className="flex flex-col">
              <h4 className="font-medium mb-2">Bons</h4>
              <ul className="space-y-2">
                {Object.entries(bonsCount)
                  .sort(([, a], [, b]) => b - a)
                  .map(([price, count]) => {
                    const cfg = bonConfigUI[price];
                    return (
                      <li key={price} className="flex items-center gap-2">
                        <span
                          className={`${cfg.bg} ${cfg.text} inline-block w-6 h-6 rounded-full`}
                        />
                        <span className="text-base">
                          {count} x {cfg.label}
                        </span>
                      </li>
                    );
                  })}
              </ul>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCheckoutModalOpen(false)}
              className="w-full"
            >
              Abbrechen
            </Button>
            <Button
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
