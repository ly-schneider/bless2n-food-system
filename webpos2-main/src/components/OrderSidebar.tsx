import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import { MenuItem, Product } from '@/types'
import { useToast } from '@/hooks/use-toast'

interface OrderSidebarProps {
  order: Record<string, number>;
  menuItems: MenuItem[];
  products: Product[];
  handleQuantityChange: (itemId: string, change: number) => void;
  stock: Record<string, number>;
}

export function OrderSidebar({ order, menuItems, products, handleQuantityChange, stock }: OrderSidebarProps) {
  // Implement your OrderSidebar logic here

  return (
    <aside className="w-full md:w-72 border-t md:border-l md:border-t-0 flex flex-col bg-gray-50">
      {/* Your OrderSidebar JSX */}
    </aside>
  )
}