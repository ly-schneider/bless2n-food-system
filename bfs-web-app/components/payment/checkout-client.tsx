"use client"

import { Elements, PaymentElement, useElements, useStripe } from "@stripe/react-stripe-js"
import { useEffect, useMemo, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useAuth } from "@/contexts/auth-context"
import { useCart } from "@/contexts/cart-context"
import { attachReceiptEmail, createPaymentIntent } from "@/lib/api/payments"
import { getStripe } from "@/lib/stripe"
import { formatChf } from "@/lib/utils"
import { ArrowLeft } from "lucide-react"
import { useRouter } from "next/navigation"

type Props = {
  returnPath?: string
}

export function CheckoutClient({ returnPath = "/food/checkout/payment/next" }: Props) {
  const { user, accessToken } = useAuth()
  const { cart } = useCart()
  const [clientSecret, setClientSecret] = useState<string | null>(null)
  const clientSecretRef = useRef<string | null>(null)
  const [paymentIntentId, setPaymentIntentId] = useState<string | null>(null)
  const [initialEmail, setInitialEmail] = useState<string>("")
  const [email, setEmail] = useState<string>("")
  const [error, setError] = useState<string | null>(null)
  const initSeq = useRef(0)

  useEffect(() => {
    clientSecretRef.current = clientSecret
  }, [clientSecret])

  // Build a stable fingerprint of the cart + user to reuse existing PI in dev/strict mode or re-mounts
  const cartFingerprint = useMemo(() => {
    const items = cart.items.map((i) => ({
      productId: i.product.id,
      quantity: i.quantity,
      configuration: i.configuration,
    }))
    return JSON.stringify({ items, user: user?.id || null })
  }, [cart.items, user?.id])

  function stableAttemptId(fingerprint: string) {
    // Small stable hash of the fingerprint for idempotency
    let h = 0
    for (let i = 0; i < fingerprint.length; i++) {
      h = (Math.imul(31, h) + fingerprint.charCodeAt(i)) | 0
    }
    return `attempt-${(h >>> 0).toString(36)}`
  }

  // Initialize PI when cart is present
  useEffect(() => {
    const init = async () => {
      const seq = ++initSeq.current
      setError(null)
      if (cart.items.length === 0) return
      // Reuse an existing PI for this fingerprint if present (prevents duplicates in Strict Mode/dev)
      try {
        const raw = sessionStorage.getItem("bfs.pi.current")
        if (raw) {
          const parsed = JSON.parse(raw) as {
            pi: string
            clientSecret: string
            fingerprint: string
            orderId?: string
            email?: string
          }
          if (parsed && parsed.fingerprint === cartFingerprint && parsed.pi && parsed.clientSecret) {
            setPaymentIntentId(parsed.pi)
            setClientSecret(parsed.clientSecret)
            const e = parsed.email || user?.email || ""
            setInitialEmail(e)
            setEmail(e)
            return
          }
        }
      } catch {}

      const items = cart.items.map((i) => ({
        productId: i.product.id,
        quantity: i.quantity,
        configuration: i.configuration,
      }))
      const prefill = user?.email ?? undefined
      try {
        const attemptId = stableAttemptId(cartFingerprint)
        const res = await createPaymentIntent({ items, customerEmail: prefill, attemptId }, accessToken || undefined)
        setClientSecret(res.clientSecret)
        setPaymentIntentId(res.paymentIntentId)
        const e = prefill || ""
        setInitialEmail(e)
        setEmail(e)
        // analytics (optional)
        try {
          const win = window as unknown as {
            analytics?: { track?: (name: string, props?: Record<string, unknown>) => void }
          }
          win.analytics?.track?.("payment_initiated", { pi: res.paymentIntentId })
        } catch {}
        // Persist PI in session to avoid duplicate creation on dev strict mode remounts
        try {
          sessionStorage.setItem(
            "bfs.pi.current",
            JSON.stringify({
              pi: res.paymentIntentId,
              clientSecret: res.clientSecret,
              orderId: res.orderId,
              fingerprint: cartFingerprint,
              email: e,
            })
          )
        } catch {}
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Fehler beim Initialisieren der Zahlung"
        // Suppress transient error if another init pass succeeded
        if (initSeq.current === seq && !clientSecretRef.current) {
          setError(msg)
        }
      }
    }
    void init()
  }, [cartFingerprint, user?.email, accessToken])

  const stripePromise = useMemo(() => getStripe(), [])
  const options = useMemo(() => {
    if (!clientSecret) return undefined
    return {
      clientSecret,
      appearance: { theme: "stripe" as const },
      // Disable all billing detail collection (we render our own optional email input)
      loader: "auto" as const,
    }
  }, [clientSecret])

  if (!clientSecret) {
    return (
      <div className="rounded-md border p-4">
        <p className="text-muted-foreground">Zahlung wird vorbereitet…</p>
        {error && <p className="mt-2 text-red-600">{error}</p>}
      </div>
    )
  }

  return (
    <Elements stripe={stripePromise} options={options}>
      <CheckoutForm
        email={email}
        setEmail={setEmail}
        initialEmail={initialEmail}
        paymentIntentId={paymentIntentId!}
        returnPath={returnPath}
        amountCents={cart.totalCents}
      />
    </Elements>
  )
}

function CheckoutForm({
  email,
  setEmail,
  initialEmail,
  paymentIntentId,
  returnPath,
  amountCents,
}: {
  email: string
  setEmail: (v: string) => void
  initialEmail: string
  paymentIntentId: string
  returnPath: string
  amountCents: number
}) {
  const stripe = useStripe()
  const elements = useElements()
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const btnRef = useRef<HTMLButtonElement>(null)
  const { accessToken } = useAuth()
  const router = useRouter()

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      if (!stripe || !elements) return
      // Update receipt email if changed (including clearing)
      if (email !== initialEmail) {
        await attachReceiptEmail(paymentIntentId, email || "", accessToken || undefined)
      }
      // Provide billing_details.name because we disabled collection via Payment Element
      // Fallback to a generic label if no user/admin name available
      const billingName = (() => {
        // We don't have customer names for normal users; admin may have first/last names
        // Use email localpart as a nicer fallback if present, else a generic value
        if (email) return email.split("@")[0] || "Customer"
        return "Customer"
      })()
      const returnUrl = `${window.location.origin}${returnPath}?pi=${encodeURIComponent(paymentIntentId)}`
      const { error } = await stripe.confirmPayment({
        elements,
        confirmParams: {
          return_url: returnUrl,
          payment_method_data: {
            billing_details: {
              name: billingName,
              email: email || "",
              phone: "",
              address: {
                country: "CH",
                postal_code: "",
                line1: "",
                city: "",
                state: "",
              },
            },
          },
        },
      })
      if (error) {
        setError(error.message || "Zahlung fehlgeschlagen. Bitte erneut versuchen.")
        btnRef.current?.focus()
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Zahlung fehlgeschlagen. Bitte erneut versuchen."
      setError(msg)
      btnRef.current?.focus()
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4" aria-busy={submitting}>
      <div className="flex items-center justify-between rounded-[11px] border p-4">
        <span>Gesamtsumme</span>
        <strong>{formatChf(amountCents)}</strong>
      </div>

      <div className="flex flex-col gap-1">
        <label htmlFor="receipt-email" className="text-sm">
          E-Mail für Quittung (optional)
        </label>
        <Input
          id="receipt-email"
          type="email"
          placeholder="deine@email.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          inputMode="email"
          autoComplete="email"
          className="w-full rounded-[11px] border px-3 py-5"
        />
      </div>

      <PaymentElement
        options={
          {
            // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
            fields: { billingDetails: { name: "never", email: "never", phone: "never", address: "never" } },
            layout: "tabs",
          } as unknown as Parameters<typeof PaymentElement>[0]["options"]
        }
      />

      {error && (
        <p className="text-red-600" role="alert">
          {error}
        </p>
      )}

      <div className="bg-background fixed right-0 bottom-0 left-0 z-50 p-4">
        <div className="max-w-xl mx-auto flex items-center justify-between gap-3">
          <Button
            onClick={() => {
              router.back()
            }}
            size="icon"
            variant="outline"
            className="size-12 shrink-0 rounded-full bg-white text-black"
          >
            <ArrowLeft className="size-5" />
          </Button>

          <Button
            ref={btnRef}
            type="submit"
            disabled={!stripe || !elements || submitting}
            aria-busy={submitting}
            className="rounded-pill h-12 flex-1 text-base font-medium md:min-w-64 md:flex-none"
          >
            {submitting ? "Verarbeiten…" : "Jetzt mit TWINT zahlen"}
          </Button>
        </div>
      </div>
    </form>
  )
}
