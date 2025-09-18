import type { Cents } from './common';

export type OrderStatus = "pending" | "paid" | "cancelled" | "refunded";

export interface Order {
  id: string;
  customerId: string | null;
  contactEmail: string | null;
  totalCents: Cents;
  status: OrderStatus;
  createdAt: string; // ISO date
  updatedAt: string; // ISO date
}

export type OrderItemType = "simple" | "bundle" | "component";

export interface OrderItem {
  id: string;
  orderId: string;
  productId: string;
  title: string;
  quantity: number;
  pricePerUnitCents: Cents;
  parentItemId: string | null;
  menuSlotId: string | null;
  menuSlotName: string | null;
  isRedeemed: boolean;
  redeemedAt: string | null; // ISO date
}