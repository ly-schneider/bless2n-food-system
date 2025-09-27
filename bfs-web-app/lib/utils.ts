import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Formats a price in cents to Swiss Francs.
// Examples:
//  - 500  -> "CHF 5.-"
//  - 550  -> "CHF 5.50"
export function formatChf(cents: number): string {
  if (isNaN(cents)) return "CHF 0.-"
  const isWholeFranc = cents % 100 === 0
  if (isWholeFranc) {
    return `CHF ${Math.round(cents / 100)}.-`
  }
  return `CHF ${(cents / 100).toFixed(2)}`
}
