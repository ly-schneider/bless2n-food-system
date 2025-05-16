export interface Product {
  id: string;
  name: string;
  type: 'Product';
  price: number;
  stock: number;
  quantity: number;
  emoji?: string | null;
  discountPercentage?: number;
  color?: string;
}

export interface ProductGroupItem {
  id: string;
  name: string;
  emoji?: string | null;
  color?: string;
}

export interface ProductGroup {
  id: string;
  name: string;
  type: 'Product Group';
  emoji?: string | null;
  quantity: number;
  color?: string;
  products: ProductGroupItem[];
}

export interface MenuItem {
  id: string;
  name: string;
  type: 'Menu';
  price: number;
  stock: number;
  quantity: number;
  emoji?: string | null;
  products: (Product | ProductGroup)[];
  discountPercentage?: number;
  color?: string;
}

export interface OrderProduct {
  id: string;
  quantity: number;
}

export interface OrderItem {
  id: string;
  quantity: number;
  products?: OrderProduct[];
}

export type MenuOrProduct = MenuItem | Product;
