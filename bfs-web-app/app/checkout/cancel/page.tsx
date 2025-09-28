import Link from "next/link"

export default function CheckoutCancelPage() {
  return (
    <div className="container mx-auto px-4 py-16">
      <h1 className="text-2xl font-semibold mb-2">Zahlung abgebrochen</h1>
      <p className="mb-6">Sie können es erneut versuchen oder den Warenkorb anpassen.</p>
      <div className="flex gap-4">
        <Link className="underline" href="/checkout">Zurück zum Warenkorb</Link>
        <Link className="underline" href="/">Zur Startseite</Link>
      </div>
    </div>
  )
}

