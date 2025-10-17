import { loadStripe, Stripe } from "@stripe/stripe-js"

let stripePromise: Promise<Stripe | null> | null = null

export function getStripe() {
  // Avoid crashing during SSR/prerender. Only attempt to load Stripe in the browser.
  if (typeof window === "undefined") {
    return Promise.resolve(null)
  }

  if (!stripePromise) {
    const pk = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
    if (!pk) {
      // Defer failure to runtime UX instead of crashing build.
      if (process.env.NODE_ENV !== "production") {
        console.warn("NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY is not set; Stripe elements will be disabled.")
      }
      return Promise.resolve(null)
    }
    stripePromise = loadStripe(pk)
  }
  return stripePromise
}
