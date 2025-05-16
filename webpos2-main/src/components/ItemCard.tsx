import React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { MenuItem, Product } from '@/types'

interface ItemCardProps {
  item: MenuItem | Product;
  stock: number;
  onClick: () => void;
}

export default function ItemCard({ item, stock, onClick }: ItemCardProps) {
  const isOutOfStock = stock <= 0;
  const isProduct = (item: MenuItem | Product): item is Product => {
    return item.type === 'Product';
  };

  const isMenuItemWithProducts = (item: MenuItem | Product): item is MenuItem => {
    return item.type === 'Menu' && 'products' in item;
  };

  // Function to convert hex color to RGBA with opacity
  const getBackgroundColor = (color: string | undefined) => {
    if (!color) return undefined;
    
    // If color is already in rgba format, just modify its opacity
    if (color.startsWith('rgba')) {
      return color.replace(/[\d.]+\)$/g, '0.4)');
    }
    
    // Convert hex to rgba
    let hex = color.replace('#', '');
    if (hex.length === 3) {
      hex = hex.split('').map(char => char + char).join('');
    }
    const r = parseInt(hex.substring(0, 2), 16);
    const g = parseInt(hex.substring(2, 4), 16);
    const b = parseInt(hex.substring(4, 6), 16);
    return `rgba(${r}, ${g}, ${b}, 0.4)`;
  };

  return (
    <Card 
      className={`relative cursor-pointer hover:shadow-lg transition-all min-h-[160px] flex flex-col ${isOutOfStock ? 'opacity-50 cursor-not-allowed' : ''}`}
      onClick={!isOutOfStock ? onClick : undefined}
      style={{
        backgroundColor: ((item as any).color) ? getBackgroundColor((item as any).color) : undefined,
      }}
    >
      {item.emoji && (
        <div className="absolute top-2 right-2 text-2xl">
          {item.emoji}
        </div>
      )}
      
      <CardHeader className="p-2 md:p-3">
        <CardTitle className="text-xl md:text-2xl flex items-center font-roboto">
          {item.name}
          <span className="hidden md:inline-flex ml-2 px-2 py-0.5 bg-white/50 text-gray-700 rounded-full text-xs md:text-sm">
            {stock}
          </span>
        </CardTitle>
      </CardHeader>

      <CardContent className="p-2 md:p-3 flex-1 flex flex-col">
        <div className="flex-1">
          {isMenuItemWithProducts(item) && item.products.map((product) => (
            <p key={product.id} className="text-xs md:text-sm text-gray-600">
              {product.quantity}x {product.name} {product.emoji}
            </p>
          ))}
        </div>
        
        <div className="font-roboto text-sm md:text-base text-gray-700 mt-auto pt-2">
          {isProduct(item) ? (
            <>CHF {item.price.toFixed(2)}</>
          ) : (
            <>CHF {(item as MenuItem).price.toFixed(2)}</>
          )}
          {(item as any).discountPercentage && (
            <span className="ml-2 text-xs text-green-600">
              -{(item as any).discountPercentage}%
            </span>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
