'use client';

import { useState, useEffect, useMemo } from 'react'
import { Button } from "@/components/ui/button"
import { Menu, User as LucideUser, Minus, Plus } from 'lucide-react'
import { useSupabase } from '@/hooks/useSupabase'
import type { User as SupabaseUser } from '@supabase/supabase-js'
import { MenuItem, Product, ProductGroup, ProductGroupItem, OrderItem, OrderProduct } from '@/types'
import { calculateDiscountedPrice } from '@/utils/priceUtils'
import { OrderSidebar } from './OrderSidebar'
import { ProductSection } from './ProductSection'
import { MenuSection } from './MenuSection'
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { TipModal } from '@/components/ui/TipModal'
import { RoundUpModal } from '@/components/ui/RoundUpModal'
import { useToast } from "@/hooks/use-toast"
import { PaymentMethodDialog } from './PaymentMethodDialog'
import { ProductGroupDialog } from './ProductGroupDialog'
import { OrderSummaryDialog } from "@/components/OrderSummaryDialog";

type OrderEntry = {
  baseId: string;
  selections?: Record<string, string>;
  quantity: number;
}

export default function HomeClient() {
  const { supabase, user } = useSupabase()
  const [order, setOrder] = useState<Record<string, OrderEntry>>({})
  const [currentProductGroups, setCurrentProductGroups] = useState<ProductGroup[]>([])
  const [currentMenuId, setCurrentMenuId] = useState<string | null>(null)
  const [stock, setStock] = useState<Record<string, number>>({})
  const [menuItems, setMenuItems] = useState<MenuItem[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isAuthLoading, setIsAuthLoading] = useState(false)
  const [isPaymentMethodDialogOpen, setIsPaymentMethodDialogOpen] = useState(false)
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState<"CSH" | "CRE" | "TWI" | "EMP" | "VIP" | "KUL" | "GUT" | "DIV">("CSH")
  const [isJsonPreviewOpen, setIsJsonPreviewOpen] = useState(false)
  const [orderJsonPreview, setOrderJsonPreview] = useState<any>(null)
  const { toast } = useToast()
  const [menuSelections, setMenuSelections] = useState<Record<string, Record<string, string>>>({});
  const [currentDonationAmount, setCurrentDonationAmount] = useState(0);

  useEffect(() => {
    console.log("Current user state:", user)
  }, [user])

  async function fetchData() {
    setIsLoading(true)
    setError(null)
    try {
      console.log('Fetching menu card data...')
      const { data: sessionData, error: sessionError } = await supabase.auth.getSession()
      if (sessionError) {
        toast({
          title: "Fehler beim Laden der Daten",
          description: sessionError.message,
          variant: "destructive",
        })
        throw sessionError
      }

      const { data, error } = await supabase.functions.invoke<(MenuItem | Product)[]>('menucard_GET', {
        method: 'GET',
      })

      if (error) {
        toast({
          title: "Fehler beim Laden der Menükarte",
          description: error.message,
          variant: "destructive",
        })
        throw error
      }

      if (!data) {
        throw new Error('Keine Daten von der API erhalten')
      }

      console.log('Received data from menucard_GET function:', data)

      const sanitizedData = data
        .filter((item: any) => item.price !== null)
        .map((item: any) => ({
          ...item,
          price: typeof item.price === 'number' ? item.price : 0,
          stock: typeof item.stock === 'number' ? item.stock : 0,
        }))

      const newMenuItems: MenuItem[] = sanitizedData
        .filter((item: any) => item.type === 'Menu') as MenuItem[]

      const newProducts: Product[] = sanitizedData
        .filter((item: any) => item.type === 'Product') as Product[]

      const initialStock: Record<string, number> = {}
      sanitizedData.forEach((item: MenuItem | Product) => {
        initialStock[item.id] = item.stock
      })

      setMenuItems(newMenuItems)
      setProducts(newProducts)
      setStock(initialStock)

    } catch (error: unknown) {
      console.error('Error fetching data:', error)
      setError(error instanceof Error ? error.message : 'Ein unbekannter Fehler ist aufgetreten')
      toast({
        title: "Fehler",
        description: error instanceof Error ? error.message : 'Ein unbekannter Fehler ist aufgetreten',
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const updateLocalStock = (baseId: string, change: number, selections?: Record<string, string>) => {
    setStock(prev => {
      const newStock = { ...prev };
      
      // Find if this is a menu item
      const menuItem = menuItems.find(m => m.id === baseId);
      if (menuItem) {
        // Update stock for the menu itself
        newStock[baseId] = Math.max(0, (prev[baseId] || 0) - change);
        
        // Update stock for fixed products in the menu
        menuItem.products
          .filter(p => p.type === 'Product')
          .forEach(product => {
            const productQuantity = (product as Product).quantity || 1;
            newStock[product.id] = Math.max(0, (prev[product.id] || 0) - (change * productQuantity));
          });
        
        // Update stock for selected products from product groups
        if (selections) {
          Object.values(selections).forEach(productId => {
            newStock[productId] = Math.max(0, (prev[productId] || 0) - change);
          });
        }
      } else {
        // Regular product
        newStock[baseId] = Math.max(0, (prev[baseId] || 0) - change);
      }
      
      return newStock;
    });
  };

  const handleQuantityChange = (
    baseId: string, 
    change: number, 
    selections?: Record<string, string>
  ) => {
    // Check stock availability for menu items and their products
    if (change > 0) {
      const menuItem = menuItems.find(m => m.id === baseId);
      if (menuItem) {
        // Check menu stock
        if ((stock[baseId] || 0) < change) {
          toast({
            title: "Nicht genügend Lagerbestand",
            description: `${menuItem.name} ist nicht mehr verfügbar.`,
            variant: "destructive",
          });
          return;
        }
        
        // Check stock for fixed products
        const fixedProducts = menuItem.products.filter(p => p.type === 'Product');
        for (const product of fixedProducts) {
          const productQuantity = (product as Product).quantity || 1;
          if ((stock[product.id] || 0) < (change * productQuantity)) {
            toast({
              title: "Nicht genügend Lagerbestand",
              description: `${product.name} ist nicht mehr verfügbar.`,
              variant: "destructive",
            });
            return;
          }
        }
        
        // Check stock for selected products
        if (selections) {
          for (const productId of Object.values(selections)) {
            if ((stock[productId] || 0) < change) {
              const product = products.find(p => p.id === productId);
              toast({
                title: "Nicht genügend Lagerbestand",
                description: `${product?.name || 'Ein Produkt'} ist nicht mehr verfügbar.`,
                variant: "destructive",
              });
              return;
            }
          }
        }
      } else {
        // Regular product stock check
        if ((stock[baseId] || 0) < change) {
          const product = products.find(p => p.id === baseId);
          toast({
            title: "Nicht genügend Lagerbestand",
            description: `${product?.name || 'Dieser Artikel'} ist nicht mehr verfügbar.`,
            variant: "destructive",
          });
          return;
        }
      }
    }

    let stockUpdateNeeded = false;
    let stockUpdateAmount = 0;
    let stockUpdateSelections = selections;

    setOrder(prev => {
      // Find existing identical item
      const existingKey = Object.entries(prev).find(([_, item]) => 
        item.baseId === baseId && 
        areSelectionsIdentical(item.selections, selections)
      )?.[0];

      if (existingKey) {
        // Update existing item
        const newQuantity = prev[existingKey].quantity + change;
        if (newQuantity <= 0) {
          stockUpdateNeeded = true;
          stockUpdateAmount = -prev[existingKey].quantity;
          stockUpdateSelections = prev[existingKey].selections;
          const { [existingKey]: _, ...rest } = prev;
          return rest;
        }
        stockUpdateNeeded = true;
        stockUpdateAmount = change;
        return {
          ...prev,
          [existingKey]: {
            ...prev[existingKey],
            quantity: newQuantity
          }
        };
      } else if (change > 0) {
        // Add new item with unique key
        const newKey = `${baseId}_${Date.now()}`;
        stockUpdateNeeded = true;
        stockUpdateAmount = change;
        return {
          ...prev,
          [newKey]: {
            baseId,
            selections,
            quantity: change
          }
        };
      }
      return prev;
    });

    // Update stock after order state is updated
    if (stockUpdateNeeded) {
      updateLocalStock(baseId, stockUpdateAmount, stockUpdateSelections);
    }
  };

  const handleMenuClick = (menuItem: MenuItem) => {
    const productGroups = menuItem.products.filter((p): p is ProductGroup => p.type === 'Product Group');
    
    if (productGroups.length === 0) {
      // Simple menu without selections
      handleQuantityChange(menuItem.id, 1);
      return;
    }

    // Initialize selections object for this menu
    const initialSelections: Record<string, string> = {};
    
    // Pre-select single-product groups
    const remainingGroups = productGroups.filter(group => {
      // Check if this group has only one available product
      const availableProducts = group.products.filter(product => 
        stock[product.id] > 0
      );

      if (availableProducts.length === 1) {
        // Auto-select the only available product
        initialSelections[group.id] = availableProducts[0].id;
        return false; // Remove this group from remaining groups
      }
      return true; // Keep this group in remaining groups
    });

    setMenuSelections(prev => ({
      ...prev,
      [menuItem.id]: initialSelections
    }));

    if (remainingGroups.length === 0) {
      // If all groups were auto-selected, add to order immediately
      handleQuantityChange(menuItem.id, 1, initialSelections);
    } else {
      // Show selection dialog for remaining groups
      setCurrentMenuId(menuItem.id);
      setCurrentProductGroups(remainingGroups);
    }
  };

  const handleProductClick = (product: Product) => {
    handleQuantityChange(product.id, 1);
  };

  const handleProductGroupSelection = (selectedProduct: ProductGroupItem) => {
    if (!currentMenuId || !currentProductGroups[0]) return;

    const currentGroup = currentProductGroups[0];
    const groupId = currentGroup.id;
    const productId = selectedProduct.id;
    
    // Get current selections for this menu
    const currentSelections = {
      ...(menuSelections[currentMenuId] || {}),
      [groupId]: productId
    };

    // Remove the first product group from the queue
    const remainingGroups = currentProductGroups.slice(1);
    setCurrentProductGroups(remainingGroups);

    if (remainingGroups.length === 0) {
      // All selections are complete, add to order
      handleQuantityChange(currentMenuId, 1, currentSelections);
      
      // Clear the selections for this menu
      setMenuSelections(prev => {
        const { [currentMenuId]: _, ...rest } = prev;
        return rest;
      });
      
      setCurrentMenuId(null);
    } else {
      // Store partial selections for next group
      setMenuSelections(prev => ({
        ...prev,
        [currentMenuId]: currentSelections
      }));
    }
  };

  // Helper function to compare selections
  const areSelectionsIdentical = (
    selections1?: Record<string, string>,
    selections2?: Record<string, string>
  ) => {
    if (!selections1 && !selections2) return true;
    if (!selections1 || !selections2) return false;
    
    const keys1 = Object.keys(selections1);
    const keys2 = Object.keys(selections2);
    
    if (keys1.length !== keys2.length) return false;
    
    return keys1.every(key => selections1[key] === selections2[key]);
  };

  const handleSignIn = async () => {
    setIsAuthLoading(true)
    try {
      console.log("Attempting to sign in...")
      const { error } = await supabase.auth.signInWithOAuth({
        provider: 'google',
        options: {
          redirectTo: `${window.location.origin}/auth/callback`
        }
      })
      if (error) throw error
    } catch (error) {
      console.error('Error signing in:', error)
    } finally {
      setIsAuthLoading(false)
    }
  }

  const handleSignOut = async () => {
    try {
      console.log("Attempting to sign out...")
      await supabase.auth.signOut()
    } catch (error) {
      console.error('Error signing out:', error)
    }
  }

  const handlePlaceOrder = async () => {
    if (Object.keys(order).length === 0) {
      toast({
        title: "Leere Bestellung",
        description: "Bitte fügen Sie Ihrer Bestellung Artikel hinzu",
        variant: "destructive",
      })
      return
    }

    if (finalTotal > 10000) {
      toast({
        title: "Ungültiger Betrag",
        description: "Bestellsumme darf CHF 10'000 nicht überschreiten",
        variant: "destructive",
      })
      return
    }

    setIsPaymentMethodDialogOpen(true)
  }

  const handlePaymentMethodSelect = async (method: "CSH" | "CRE" | "TWI" | "EMP" | "VIP" | "KUL" | "GUT" | "DIV") => {
    type OrderPayload = {
      payment_method: string;
      total_amount: number;
      donation_amount: number;
      items: {
        id: string;
        quantity: number;
        name: string;
        price: number;
        discount_percentage?: number;
        type: string;
        selections?: {
          group_id: string;
          group_name?: string;
          product_id: string;
          product_name?: string;
        }[];
      }[];
    };

    setSelectedPaymentMethod(method);
    setIsPaymentMethodDialogOpen(false);

    // Prepare the order data
    const orderData: OrderPayload = {
      payment_method: method,
      total_amount: total + currentDonationAmount,
      donation_amount: currentDonationAmount,
      items: Object.entries(order).map(([itemId, item]) => {
        const menuItem = [...menuItems, ...products].find(i => i.id === item.baseId);
        if (!menuItem) return null;

        const orderItem = {
          id: item.baseId,
          quantity: item.quantity,
          name: menuItem.name,
          price: menuItem.price,
          discount_percentage: menuItem.discountPercentage,
          type: menuItem.type
        };

        if (menuItem.type === 'Menu' && item.selections) {
          return {
            ...orderItem,
            selections: Object.entries(item.selections).map(([groupId, productId]) => {
              const group = (menuItem as MenuItem).products.find(p => p.id === groupId && p.type === 'Product Group') as ProductGroup;
              const selectedProduct = group?.products.find(p => p.id === productId);
              return {
                group_id: groupId,
                group_name: group?.name,
                product_id: productId,
                product_name: selectedProduct?.name
              };
            })
          };
        }

        return orderItem;
      }).filter((item): item is NonNullable<typeof item> => item !== null)
    };

    setOrderJsonPreview(orderData);
    setIsJsonPreviewOpen(true);
  };

  const handleConfirmOrder = async () => {
    setIsJsonPreviewOpen(false);
    
    try {
      setIsLoading(true);

      const orderItems: OrderItem[] = Object.entries(order).map(([_, item]) => {
        const menuItem = menuItems.find(m => m.id === item.baseId);
        if (menuItem) {
          // Get all fixed products from the menu
          const fixedProducts = menuItem.products
            .filter(p => p.type === 'Product')
            .map(p => ({
              id: p.id,
              quantity: (p as Product).quantity || 1
            }));

          // Get all selected products from product groups
          const selectedProducts = item.selections 
            ? Object.entries(item.selections).map(([groupId, productId]) => ({
                id: productId,
                quantity: 1
              }))
            : [];

          // Combine fixed products and selected products
          return {
            id: item.baseId,
            quantity: item.quantity,
            products: [...fixedProducts, ...selectedProducts]
          };
        } else {
          // This is a regular product
          return {
            id: item.baseId,
            quantity: item.quantity
          };
        }
      });

      const orderPayload = {
        user_id: user?.id,
        donation_amount: currentDonationAmount,
        total_amount: total + currentDonationAmount,
        discount_code: '',
        payment_method: selectedPaymentMethod,
        items: orderItems,
      };

      // Copy order payload to clipboard
      await navigator.clipboard.writeText(JSON.stringify(orderPayload, null, 2));
      console.log('Order JSON copied to clipboard:', orderPayload);

      const response = await fetch('https://vykrkekwtijhkuwgvtbj.supabase.co/functions/v1/make_order', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${(await supabase.auth.getSession()).data.session?.access_token}`,
        },
        body: JSON.stringify(orderPayload),
      });

      const result = await response.json();

      if (!response.ok) {
        const errorMessage = result.error || '';
        
        if (response.status === 400) {
          let title = "Ungültige Bestellung";
          let description = "Bitte überprüfen Sie Ihre Eingaben.";

          if (errorMessage.includes("stock")) {
            title = "Nicht genügend Lagerbestand";
            description = "Ein oder mehrere Artikel sind nicht mehr verfügbar.";
          } else if (errorMessage.includes("total_amount")) {
            title = "Ungültiger Betrag";
            description = "Bitte überprüfen Sie den Gesamtbetrag.";
          } else if (errorMessage.includes("items")) {
            title = "Ungültige Artikel";
            description = "Bitte überprüfen Sie Ihre Bestellung.";
          }

          toast({
            title,
            description,
            variant: "destructive",
          });
        } else if (response.status === 500) {
          toast({
            title: "Serverfehler",
            description: "Bitte versuchen Sie es später erneut.",
            variant: "destructive",
          });
        }

        throw new Error(errorMessage || `HTTP error! status: ${response.status}`);
      }

      // Clear the order after successful submission
      setOrder({});
      setMenuSelections({});
      setIsPaymentMethodDialogOpen(false);

      toast({
        title: "Bestellung erfolgreich",
        description: (
          <div className="mt-2 flex flex-col space-y-2">
            <p>Bestellnummer: {result.order_id}</p>
            <p>Gesamtbetrag: CHF {finalTotal.toFixed(2)}</p>
            {currentDonationAmount > 0 && (
              <p className="text-sm text-gray-600">
                Inkl. Spende: CHF {currentDonationAmount.toFixed(2)}
              </p>
            )}
          </div>
        ),
        duration: 5000,
        variant: "success",
      });

      // Refresh the data
      await fetchData();

    } catch (error) {
      console.error('Error placing order:', error);
      
      // If order fails, refresh stock to ensure accuracy
      await fetchData();
      
      if (error instanceof Error && !error.message.includes('HTTP error')) {
        toast({
          title: "Bestellung fehlgeschlagen",
          description: "Ein unerwarteter Fehler ist aufgetreten. Bitte versuchen Sie es erneut.",
          variant: "destructive",
        });
      }
    } finally {
      setIsLoading(false);
    }
  }

  // Calculate the total
  const total = useMemo(() => {
    return Object.entries(order).reduce((sum, [_, item]) => {
      const menuItem = [...menuItems, ...products].find(i => i.id === item.baseId);
      if (menuItem) {
        const price = menuItem.price ?? 0;
        const discountPercentage = menuItem.discountPercentage ?? 0;
        const discountedPrice = calculateDiscountedPrice(price, discountPercentage) ?? 0;
        return sum + (discountedPrice * item.quantity);
      }
      return sum;
    }, 0);
  }, [order, menuItems, products]);

  // Calculate the suggested round-up amount (to next full franc)
  const roundToNext1 = (num: number) => Math.ceil(num);
  const suggestedRoundUp = Math.ceil(total) === Math.floor(total) 
    ? 1 // If total is a whole number, suggest CHF 1
    : roundToNext1(total) - total; // Otherwise, round up to next franc
  const roundedTotal = total + suggestedRoundUp;
  
  // Calculate the final total including any donation
  const finalTotal = total + currentDonationAmount;
  
  // Calculate the actual donation amount based on any additional amount over the base total
  const donationAmount = currentDonationAmount;

  // Helper function to get sorted order entries
  const getSortedOrderEntries = () => {
    return Object.entries(order)
      .sort(([_, a], [__, b]) => {
        // First sort by baseId
        const baseCompare = a.baseId.localeCompare(b.baseId);
        if (baseCompare !== 0) return baseCompare;
        
        // Then sort by selections if they exist
        if (!a.selections || !b.selections) return 0;
        return JSON.stringify(a.selections).localeCompare(JSON.stringify(b.selections));
      });
  };

  const handleTipToggle = (enabled: boolean) => {
  }

  const handleCloseModal = () => {
  }

  const handleSaveTotal = (newTotal: number) => {
  }

  // Memoize the sorted order entries
  const sortedOrderEntries = useMemo(() => {
    return Object.entries(order)
      .sort(([_, a], [__, b]) => {
        const baseCompare = a.baseId.localeCompare(b.baseId);
        if (baseCompare !== 0) return baseCompare;
        
        if (!a.selections || !b.selections) return 0;
        return JSON.stringify(a.selections).localeCompare(JSON.stringify(b.selections));
      });
  }, [order]);

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <header className="flex justify-between items-center p-2 md:p-4 border-b bg-white">
        <Menu className="w-5 h-5 md:w-7 md:h-7" aria-label="Menu" />
        <h1 className="text-lg md:text-2xl font-bold">BlessThun</h1>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-8 w-8 md:h-11 md:w-11 rounded-full">
              {user ? (
                <Avatar className="h-8 w-8 md:h-11 md:w-11">
                  <AvatarImage src={user.user_metadata?.avatar_url} alt={user.user_metadata?.full_name || ''} />
                  <AvatarFallback>{user.user_metadata?.full_name?.charAt(0) || user.email?.charAt(0) || '?'}</AvatarFallback>
                </Avatar>
              ) : (
                <LucideUser className="h-5 w-5 md:h-7 md:w-7" />
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56 md:w-72" align="end" forceMount>
            {user ? (
              <>
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user.user_metadata?.full_name || user.email}</p>
                    <p className="text-xs leading-none text-muted-foreground">{user.email}</p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleSignOut}>
                  Abmelden
                </DropdownMenuItem>
              </>
            ) : (
              <DropdownMenuItem onClick={handleSignIn} disabled={isAuthLoading}>
                {isAuthLoading ? 'Anmeldung...' : 'Mit Google anmelden'}
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </header>

      <div className="flex flex-col md:flex-row flex-1 overflow-hidden">
        <main className="flex-1 overflow-auto p-2 md:p-6">
          <MenuSection 
            menuItems={menuItems} 
            stock={stock} 
            onClick={handleMenuClick}
            isLoading={isLoading}
          />
          <ProductSection
            products={products}
            stock={stock}
            addToOrder={(product) => handleQuantityChange(product.id, 1)}
            isLoading={isLoading}
          />
        </main>

        <aside className="w-full md:w-96 border-t md:border-l md:border-t-0 flex flex-col bg-gray-50">
          <div className="flex-1 overflow-auto p-2 md:p-6">
            <h2 className="text-xl md:text-2xl font-bold mb-2 md:mb-4 text-center">Bestellung</h2>
            <div className="space-y-2 md:space-y-3 font-mono text-sm md:text-base">
              {/* Regular order items */}
              {sortedOrderEntries.map(([key, item]) => {
                const menuItem = [...menuItems, ...products].find(i => i.id === item.baseId);
                if (!menuItem) return null;
                
                const discountedPrice = calculateDiscountedPrice(menuItem.price, menuItem.discountPercentage) ?? 0;
                
                return (
                  <div key={key} className="flex flex-col space-y-1">
                    <div className="flex justify-between items-center">
                      <span className="flex-1 font-bold">{menuItem.name}</span>
                      <div className="flex items-center">
                        <Button 
                          size="sm" 
                          variant="outline" 
                          className="h-6 w-6 md:h-8 md:w-8 p-0" 
                          onClick={() => handleQuantityChange(item.baseId, -1, item.selections)}
                        >
                          <Minus className="w-3 h-3 md:w-4 md:h-4" />
                        </Button>
                        <span className="mx-2 md:mx-3 md:text-lg">{item.quantity}</span>
                        <Button 
                          size="sm" 
                          variant="ghost" 
                          className="h-6 w-6 md:h-8 md:w-8 p-0" 
                          onClick={() => handleQuantityChange(item.baseId, 1, item.selections)} 
                          disabled={stock[item.baseId] === 0}
                        >
                          <Plus className="w-3 h-3 md:w-4 md:h-4" />
                        </Button>
                      </div>
                      <span className="text-right ml-2 md:ml-3 md:text-lg">
                        CHF {(discountedPrice * item.quantity).toFixed(2)}
                      </span>
                    </div>
                    {/* Show selections if they exist */}
                    {item.selections && menuItem.type === 'Menu' && (
                      <div className="ml-4 text-xs text-gray-500">
                        {(menuItem as MenuItem).products.map((product, index) => {
                          if (product.type === 'Product') {
                            return <div key={index}>• {product.name}</div>;
                          } else if (product.type === 'Product Group') {
                            const selectedProductId = item.selections?.[product.id];
                            const selectedProduct = selectedProductId
                              ? product.products.find(p => p.id === selectedProductId)
                              : null;
                            
                            return (
                              <div key={index}>
                                • {product.name}: {selectedProduct ? selectedProduct.name : 'Keine Auswahl'}
                              </div>
                            );
                          }
                          return null;
                        })}
                      </div>
                    )}
                  </div>
                );
              })}

              {/* Total amount */}
              <div className="border-t border-dashed pt-2 mt-2 md:pt-4 md:mt-4">
                <div className="flex justify-between font-bold md:text-xl">
                  <span>Gesamt</span>
                  <span>CHF {finalTotal.toFixed(2)}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="border-t p-2 md:p-6 bg-white">
            <Button 
              className="w-full bg-black hover:bg-gray-800 text-white text-lg md:text-xl py-6 md:py-8" 
              onClick={handlePlaceOrder}
              disabled={isLoading}
            >
              {isLoading ? 'Wird verarbeitet...' : 'Bezahlen'}
            </Button>
          </div>
        </aside>
      </div>

      {isPaymentMethodDialogOpen && (
        <PaymentMethodDialog
          isOpen={isPaymentMethodDialogOpen}
          onClose={() => setIsPaymentMethodDialogOpen(false)}
          onSelect={handlePaymentMethodSelect}
        />
      )}

      {isJsonPreviewOpen && (
        <OrderSummaryDialog
          isOpen={isJsonPreviewOpen}
          onClose={() => setIsJsonPreviewOpen(false)}
          onConfirm={(newDonationAmount) => {
            setCurrentDonationAmount(newDonationAmount);
            handleConfirmOrder();
          }}
          order={order}
          menuItems={menuItems}
          products={products}
          menuSelections={Object.fromEntries(
            Object.entries(order)
              .filter(([_, entry]) => entry.selections)
              .map(([_, entry]) => [entry.baseId, entry.selections!])
          )}
          total={total}
          donationAmount={currentDonationAmount}
          finalTotal={finalTotal}
          paymentMethod={selectedPaymentMethod}
        />
      )}

      {currentProductGroups.length > 0 && (
        <ProductGroupDialog
          productGroup={currentProductGroups[0]}
          onSelect={handleProductGroupSelection}
          onClose={() => {
            setCurrentProductGroups([])
            setCurrentMenuId(null)
          }}
          stock={stock}
        />
      )}
    </div>
  )
}
