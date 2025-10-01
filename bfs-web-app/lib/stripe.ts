import { loadStripe, Stripe } from "@stripe/stripe-js"
import { env } from "@/env.mjs"

let stripePromise: Promise<Stripe | null> | null = null

export function getStripe() {
  if (!stripePromise) {
    stripePromise = loadStripe(env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY)
  }
  return stripePromise
}

