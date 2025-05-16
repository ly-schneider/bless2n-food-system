import { MenuItem, Product } from '@/types'

export const menuItems: MenuItem[] = [
  { 
    id: 'menu1', 
    name: 'Smash Burger Menu', 
    type: 'Menu' as const,
    price: 14.00, 
    discountPercentage: 20, 
    stock: 20,
    products: [
      { id: 'p1', name: 'Burger', type: 'Product' as const, emoji: '🍔', quantity: 1 },
      { id: 'p2', name: 'Pommes', type: 'Product' as const, emoji: '🍟', quantity: 1 },
      { id: 'p3', name: 'Getränk', type: 'Product' as const, emoji: '🥤', quantity: 1 }
    ]
  } as MenuItem,
  { 
    id: 'menu2', 
    name: 'Smash Burger Menu', 
    type: 'Menu' as const,
    price: 8.50, 
    discountPercentage: 0, 
    stock: 15,
    products: [
      { id: 'p1', name: 'Burger', type: 'Product' as const, emoji: '🍔', quantity: 1 },
      { id: 'p2', name: 'Pommes', type: 'Product' as const, emoji: '🍟', quantity: 1 },
      { id: 'p3', name: 'Getränk', type: 'Product' as const, emoji: '🥤', quantity: 1 }
    ]
  } as MenuItem,
  { 
    id: 'menu3', 
    name: 'Smash Burger Menu', 
    type: 'Menu' as const,
    price: 8.50, 
    discountPercentage: 10, 
    stock: 10,
    products: [
      { id: 'p1', name: 'Burger', type: 'Product' as const, emoji: '🍔', quantity: 1 },
      { id: 'p2', name: 'Pommes', type: 'Product' as const, emoji: '🍟', quantity: 1 },
      { id: 'p3', name: 'Getränk', type: 'Product' as const, emoji: '🥤', quantity: 1 }
    ]
  } as MenuItem,
]

export const products: Product[] = [
  { id: 'product1', name: 'Smash Burger', type: 'Product', price: 8.50, discountPercentage: 15, stock: 50, emoji: '🍔' } as Product,
  { id: 'product2', name: 'Cheeseburger', type: 'Product', price: 9.50, stock: 40, emoji: '🧀' } as Product,
  { id: 'product3', name: 'Chicken Burger', type: 'Product', price: 9.00, discountPercentage: 10, stock: 30, emoji: '🍗' } as Product,
  { id: 'product4', name: 'Veggie Burger', type: 'Product', price: 8.50, stock: 25, emoji: '🥗' } as Product,
  { id: 'product5', name: 'Fries', type: 'Product', price: 4.50, stock: 100, emoji: '🍟' } as Product,
  { id: 'product6', name: 'Onion Rings', type: 'Product', price: 5.00, discountPercentage: 20, stock: 75, emoji: '🧅' } as Product,
  { id: 'product7', name: 'Soda', type: 'Product', price: 3.50, stock: 200, emoji: '🥤' } as Product,
  { id: 'product8', name: 'Milkshake', type: 'Product', price: 6.00, stock: 60, emoji: '🥛' } as Product,
]
