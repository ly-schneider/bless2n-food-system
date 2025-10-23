import { loadStripe, Stripe } from "@stripe/stripe-js"

let stripePromise: Promise<Stripe | null> | null = null

export function getStripe() {
  if (typeof window === "undefined") {
    return Promise.resolve(null)
  }

  if (!stripePromise) {
    const inlinePk = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY

    const init = async (): Promise<Stripe | null> => {
      if (!inlinePk) {
        if (process.env.NODE_ENV !== "production") {
          console.warn(
            "Stripe publishable key not available; disabling Elements."
          )
        }
        return null
      }
      return loadStripe(inlinePk)
    }

    stripePromise = init()
  }
  return stripePromise
}
