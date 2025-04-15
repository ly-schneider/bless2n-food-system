import { Cart } from "@/components/cart";
import ProductList from "@/components/product-list";
import { CartProvider } from "@/contexts/cart-context";

export default async function HomePage() {
  return (
    <div className="flex-1 w-full h-screen overflow-hidden px-5 pb-5">
      <CartProvider>
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 h-full">
          <div className="col-span-1 md:col-span-2 lg:col-span-3 overflow-y-auto h-[calc(100vh-110px)] pr-4">
            <ProductList />
          </div>
          
          <div className="col-span-1 border-l border-border bg-background h-[calc(100vh-110px)]">
            <Cart />
          </div>
        </div>
      </CartProvider>
    </div>
  );
}
