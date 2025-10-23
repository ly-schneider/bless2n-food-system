import { loadStripe, Stripe } from "@stripe/stripe-js"

let stripePromise: Promise<Stripe | null> | null = null
let cachedPk: string | null = null

export function getStripe() {
  // Avoid SSR usage
  if (typeof window === "undefined") {
    try {
      console.info("[stripe-init] SSR context; returning null")
    } catch {}
    return Promise.resolve(null)
  }

  if (!stripePromise) {
    const inlinePk = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY

    const init = async (): Promise<Stripe | null> => {
      let pk = inlinePk || cachedPk
      if (!pk) {
        try {
          console.info("[stripe-init] No inline key; GET /api/public-config")
        } catch {}
        try {
          const res = await fetch("/api/public-config", { cache: "no-store" })
          if (res.ok) {
            const data = (await res.json()) as { stripePublishableKey?: string | null }
            if (data?.stripePublishableKey) {
              pk = data.stripePublishableKey
              cachedPk = pk
              try {
                console.info(`[stripe-init] Retrieved key tail=…${pk.slice(-6)}`)
              } catch {}
            } else {
              try {
                console.warn("[stripe-init] /api/public-config missing stripePublishableKey")
              } catch {}
            }
          } else {
            try {
              console.warn("[stripe-init] /api/public-config non-OK", res.status, res.statusText)
            } catch {}
          }
        } catch (e) {
          try {
            console.error("[stripe-init] Failed to fetch /api/public-config", e)
          } catch {}
        }
      }

      if (!pk) {
        try {
          console.error("[stripe-init] Publishable key unavailable; Stripe Elements disabled")
        } catch {}
        return null
      }
      try {
        console.info("[stripe-init] Loading Stripe.js…")
      } catch {}
      const s = await loadStripe(pk)
      try {
        console.info(`[stripe-init] Stripe.js loaded: ${Boolean(s)}`)
      } catch {}
      return s
    }

    stripePromise = init()
  }
  return stripePromise
}
