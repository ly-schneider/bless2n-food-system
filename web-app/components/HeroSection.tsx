import Link from "next/link";
import { Button } from "./ui/button";

export default function HeroSection() {
  return (
    <section className="pt-32 pb-20 overflow-hidden relative" id="hero">
      <div className="absolute top-0 right-0 w-1/2 h-full bg-gradient-to-bl from-primary/50 to-accent/50"></div>
      <div className="container mx-auto px-4 md:px-6">
        <div className="grid grid-cols-1 md:grid-cols-12 gap-8">
          <div className="md:col-span-7 z-10">
            <h1 className="text-4xl md:text-6xl font-bold mb-6 leading-tight">Cashless Checkout & Profi-Kasse in 5 Minuten.</h1>
            <p className="text-lg text-gray-700 mb-6 max-w-2xl">
              Miete das rentro POS Kit: Tablet-Kasse mit TWINT-Integration in einer Box. Plug-&-Play und Live-Analytics für Festivals, Messen &
              Pop-ups jeder Grösse.
            </p>

            <ul className="space-y-1 mb-8">
              <li className="flex items-start">
                <span className="mr-2 text-xl">•</span>
                <span>Setup in unter 5 Minuten - auspacken, einschalten, kassieren</span>
              </li>
              <li className="flex items-start">
                <span className="mr-2 text-xl">•</span>
                <span>Schneller Checkout mit TWINT. Besucher zahlen in Sekunden.</span>
              </li>
              <li className="flex items-start">
                <span className="mr-2 text-xl">•</span>
                <span>Self-Checkout oder Dual-Display mit intuitivem UX</span>
              </li>
              <li className="flex items-start">
                <span className="mr-2 text-xl">•</span>
                <span>Live-Umsatz-Analytics zeigen Peaks & Topseller unmittelbar</span>
              </li>
              <li className="flex items-start">
                <span className="mr-2 text-xl">•</span>
                <span>Schweizer Qualität</span>
              </li>
            </ul>

            <div className="flex flex-wrap gap-4">
              <Link href="/angebot">
                <Button className="bg-foreground text-background px-8 py-6 rounded-md">Unverbindliches Angebot holen</Button>
              </Link>
              <Link href="/#funktionen">
                <Button variant="outline" className="border-black px-8 py-6 rounded-md">
                  Funktionen entdecken
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
