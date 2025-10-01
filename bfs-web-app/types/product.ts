import type { CategoryDTO } from './category';
import type { Cents } from './common';
import type { MenuDTO } from './menu';

export type ProductType = "simple" | "menu";

export interface Product {
  id: string;
  categoryId: string;
  type: ProductType;
  name: string;
  image: string | null;
  priceCents: Cents;
  isActive: boolean;
  createdAt: string; // ISO date
  updatedAt: string; // ISO date
}

export interface ProductSummaryDTO {
  id: string;
  category: CategoryDTO;
  type: ProductType;
  name: string;
  image: string | null;
  priceCents: Cents;
  isActive: boolean;
  availableQuantity?: number | null;
  isAvailable?: boolean;
  isLowStock?: boolean;
}

export interface ProductDTO extends ProductSummaryDTO {
  menu: MenuDTO | null;
}
