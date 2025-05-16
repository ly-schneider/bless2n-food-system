export function calculateDiscountedPrice(price: number | null, discountPercentage?: number): number | null {
  if (price === null) return null;
  if (!discountPercentage) return price;
  return price * (1 - discountPercentage / 100);
}
