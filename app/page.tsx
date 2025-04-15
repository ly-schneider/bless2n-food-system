import { Cart } from "@/components/cart";
import ProductList from "@/components/product-list";
import { CartProvider } from "@/contexts/cart-context";
import { createClient } from "@/utils/supabase/server";

export default async function HomePage() {
  const supabase = await createClient();

  const {
    data: { user },
  } = await supabase.auth.getUser();

  return (
    <div className="flex-1 w-full h-screen overflow-hidden">
      <CartProvider>
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 h-full">
          {/* Product list section - scrollable area */}
          <div className="col-span-1 md:col-span-2 lg:col-span-3 overflow-y-auto h-[calc(100vh-100px)] pr-4">
            <ProductList />
          </div>
          
          {/* Cart section - fixed position */}
          <div className="col-span-1 border-l border-border bg-background h-[calc(100vh-100px)]">
            <Cart />
          </div>
        </div>
      </CartProvider>
    </div>
  );
}
