import { Product } from '@/types'
import ItemCard from './ItemCard'
import { Skeleton } from './Skeleton'

interface ProductSectionProps {
  products: Product[];
  stock: Record<string, number>;
  addToOrder: (product: Product) => void;
  isLoading: boolean;
}

export function ProductSection({ products, stock, addToOrder, isLoading }: ProductSectionProps) {
  if (isLoading) {
    return (
      <section className="mb-8">
        <h2 className="text-2xl font-bold mb-4">Products</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {[...Array(8)].map((_, index) => (
            <Skeleton key={index} className="h-40 md:h-48 w-full" />
          ))}
        </div>
      </section>
    );
  }

  return (
    <section className="mb-8">
      <h2 className="text-2xl font-bold mb-4">Products</h2>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {products.map(product => (
          <ItemCard 
            key={product.id}
            item={product}
            stock={stock[product.id] || 0}
            onClick={() => addToOrder(product)}
          />
        ))}
      </div>
    </section>
  )
}
