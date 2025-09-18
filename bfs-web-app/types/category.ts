export interface Category {
  id: string;
  name: string;
  isActive: boolean;
  createdAt: string; // ISO date
  updatedAt: string; // ISO date
}

export interface CategoryDTO {
  id: string;
  name: string;
  isActive: boolean;
}