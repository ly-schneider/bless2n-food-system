import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { MenuItem, Product } from "@/types"
import { cn } from "@/lib/utils"
import { CreditCard, Banknote, QrCode, UserCircle2, Star, Heart, Gift, Coffee } from "lucide-react"
import { useState, useMemo } from "react"
import { Input } from "@/components/ui/input"
import { OrderDetailsDialog } from "./OrderDetailsDialog"

interface OrderSummaryDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (donationAmount: number) => void;
  order: Record<string, { baseId: string; selections?: Record<string, string>; quantity: number }>;
  menuItems: MenuItem[];
  products: Product[];
  menuSelections: Record<string, Record<string, string>>;
  total: number;
  donationAmount: number;
  finalTotal: number;
  paymentMethod: string;
}

interface CoinCount {
  color: string;
  count: number;
}

function formatPrice(price: number) {
  return price.toFixed(2);
}

function PaymentIcon({ method }: { method: string }) {
  switch (method) {
    case 'CRE': return <CreditCard className="w-6 h-6" />;
    case 'CSH': return <Banknote className="w-6 h-6" />;
    case 'TWI': return <QrCode className="w-6 h-6" />;
    case 'EMP': return <UserCircle2 className="w-6 h-6" />;
    case 'VIP': return <Star className="w-6 h-6" />;
    case 'KUL': return <Heart className="w-6 h-6" />;
    case 'GUT': return <Gift className="w-6 h-6" />;
    case 'DIV': return <Coffee className="w-6 h-6" />;
    default: return null;
  }
}

function getPaymentMethodName(method: string) {
  switch (method) {
    case 'CRE': return 'Karte';
    case 'CSH': return 'Bar';
    case 'TWI': return 'Twint';
    case 'EMP': return 'Mitarbeiter';
    case 'VIP': return 'VIP';
    case 'KUL': return 'Kultur';
    case 'GUT': return 'Gutschein';
    case 'DIV': return 'Diverses';
    default: return method;
  }
}

function calculateCoins(
  order: Record<string, { baseId: string; selections?: Record<string, string>; quantity: number }>,
  menuItems: MenuItem[],
  products: Product[],
  menuSelections: Record<string, Record<string, string>>
): CoinCount[] {
  console.log('Calculating coins with:', { order, menuItems, products, menuSelections });
  const coins: CoinCount[] = [];
  
  // Helper function to add a coin
  const addCoin = (color: string, quantity: number) => {
    const existingCoin = coins.find(c => c.color === color);
    if (existingCoin) {
      existingCoin.count += quantity;
    } else {
      coins.push({ color, count: quantity });
    }
  };

  // Process order items
  Object.entries(order).forEach(([_, item]) => {
    console.log('Processing order item:', item);

    // Try to find the item in menuItems first
    const menuItem = menuItems.find(m => m.id === item.baseId);
    
    if (menuItem) {
      console.log('Found menu item:', menuItem);
      
      // Process each product in the menu
      menuItem.products?.forEach(menuProduct => {
        if (menuProduct.type === 'Product') {
          // Regular product in menu
          if (menuProduct.color) {
            console.log('Adding coin for menu product:', menuProduct.color);
            addCoin(menuProduct.color, item.quantity);
          }
        } else if (menuProduct.type === 'Product Group' && item.selections) {
          // Get the selected product from the group
          const selectedProductId = item.selections[menuProduct.id];
          if (selectedProductId && menuProduct.products) {
            const selectedProduct = menuProduct.products.find(p => p.id === selectedProductId);
            if (selectedProduct?.color) {
              console.log('Adding coin for selected product in group:', selectedProduct.color);
              addCoin(selectedProduct.color, item.quantity);
            }
          }
        }
      });
    } else {
      // Try to find it in products (single product)
      const product = products.find(p => p.id === item.baseId);
      if (product?.color) {
        console.log('Adding coin for single product:', product.color);
        addCoin(product.color, item.quantity);
      }
    }
  });

  console.log('Final coins array:', coins);
  return coins.sort((a, b) => b.count - a.count);
}

export function OrderSummaryDialog({ 
  isOpen, 
  onClose, 
  onConfirm,
  order,
  menuItems,
  products,
  menuSelections,
  total,
  donationAmount,
  finalTotal,
  paymentMethod
}: OrderSummaryDialogProps) {
  console.log('OrderSummaryDialog props:', { order, menuItems, products, menuSelections });
  const coins = calculateCoins(order, menuItems, products, menuSelections);
  console.log('Calculated coins:', coins);
  const [showOrderDetails, setShowOrderDetails] = useState(false);
  const [amountGiven, setAmountGiven] = useState<number | null>(null);
  const [customAmount, setCustomAmount] = useState<boolean>(false);

  // Calculate donation amount based on given amount
  const calculatedDonation = useMemo(() => {
    if (!amountGiven) return 0;
    return Math.max(0, amountGiven - total);
  }, [amountGiven, total]);

  // Calculate final total based on given amount or original total
  const calculatedTotal = useMemo(() => {
    return total + calculatedDonation;
  }, [total, calculatedDonation]);

  // Calculate the next 5 rounded amounts in steps of 5
  const roundedAmounts = useMemo(() => {
    const amounts = [];
    let currentAmount = Math.ceil(finalTotal);
    while (currentAmount % 5 !== 0) {
      currentAmount++;
    }
    for (let i = 0; i < 5; i++) {
      amounts.push(currentAmount + (i * 5));
    }
    return amounts;
  }, [finalTotal]);

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="max-w-[1200px] max-h-[90vh] overflow-hidden">
          <DialogHeader>
            <DialogTitle className="text-xl">🧾 Bestellübersicht</DialogTitle>
          </DialogHeader>
          
          <div className="flex gap-6 h-full overflow-auto">
            {/* Left side - Summary */}
            <div className="flex-1 space-y-6 min-w-0">
              {/* Show Order Details Button */}
              <Button 
                variant="outline" 
                className="w-full py-4 text-lg"
                onClick={() => setShowOrderDetails(true)}
              >
                Bestellung anzeigen
              </Button>

              {/* Payment Amount Section */}
              <div className="space-y-4 p-6 bg-gray-50 rounded-lg">
                <h3 className="text-lg font-semibold">Totalbetrag bei Spende</h3>
                
                {/* Quick Amount Buttons */}
                <div className="grid grid-cols-4 gap-2">
                  {roundedAmounts.map((amount) => (
                    <Button
                      key={amount}
                      variant={amountGiven === amount ? "default" : "outline"}
                      onClick={() => {
                        setAmountGiven(amount);
                        setCustomAmount(false);
                      }}
                      className="text-lg py-6"
                    >
                      {Math.floor(amount)}
                    </Button>
                  ))}
                  <Button
                    variant={customAmount ? "default" : "outline"}
                    onClick={() => {
                      setCustomAmount(true);
                      setAmountGiven(null);
                    }}
                    className="text-lg py-6"
                  >
                    Anderer Betrag
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setAmountGiven(null);
                      setCustomAmount(false);
                    }}
                    className="text-lg py-6 text-red-600 hover:text-red-700 hover:bg-red-50"
                  >
                    Zurücksetzen
                  </Button>
                </div>

                {/* Custom Amount Input */}
                {customAmount && (
                  <div className="space-y-2">
                    <Input
                      type="number"
                      step="0.05"
                      min={finalTotal}
                      value={amountGiven || ''}
                      onChange={(e) => setAmountGiven(parseFloat(e.target.value))}
                      className="text-lg py-6"
                      placeholder="Betrag eingeben..."
                    />
                  </div>
                )}
              </div>

              {/* Totals */}
              <div className="space-y-4 p-6 bg-gray-50 rounded-lg">
                <div className="flex justify-between text-lg">
                  <span>Zwischensumme</span>
                  <span className="font-medium">{formatPrice(total)} CHF</span>
                </div>
                {calculatedDonation > 0 && (
                  <div className="flex justify-between text-lg text-green-600">
                    <span>Spende</span>
                    <span className="font-medium">+{formatPrice(calculatedDonation)} CHF</span>
                  </div>
                )}
                <div className="flex justify-between text-xl font-bold pt-4 border-t border-gray-200">
                  <span>Total</span>
                  <span>{formatPrice(calculatedTotal)} CHF</span>
                </div>
                <div className="flex justify-between text-lg pt-4 border-t border-gray-200">
                  <span>Zahlungsmethode</span>
                  <span className="flex items-center gap-2">
                    <PaymentIcon method={paymentMethod} />
                    {getPaymentMethodName(paymentMethod)}
                  </span>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-col gap-4">
                <Button 
                  onClick={() => onConfirm(calculatedDonation)} 
                  className="w-full bg-black hover:bg-gray-800 text-white py-8 text-xl font-medium rounded-lg"
                >
                  Bestellung abschliessen
                </Button>
                <Button 
                  onClick={onClose} 
                  variant="ghost" 
                  className="text-gray-500 hover:text-gray-700 py-4 text-lg"
                >
                  Abbrechen
                </Button>
              </div>
            </div>

            {/* Right side - Coins */}
            {coins.length > 0 && (
              <div className="w-[300px] border-l border-gray-200 pl-6 overflow-auto">
                <div className="text-lg font-semibold mb-4">Jetons</div>
                <div className="grid gap-3">
                  {coins.map(({ color, count }, index) => (
                    <div 
                      key={index} 
                      className="flex items-center gap-3 bg-white px-4 py-3 rounded-lg shadow-sm"
                    >
                      <div 
                        className="w-8 h-8 rounded-full border border-gray-200"
                        style={{ backgroundColor: color }}
                      />
                      <span className="text-lg font-medium">× {count}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <OrderDetailsDialog
        isOpen={showOrderDetails}
        onClose={() => setShowOrderDetails(false)}
        order={order}
        menuItems={menuItems}
        products={products}
        menuSelections={menuSelections}
      />
    </>
  );
}
