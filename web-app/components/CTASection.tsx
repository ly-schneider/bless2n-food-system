import Link from "next/link";
import { Button } from "./ui/button";

export default function CTASection() {
  return (
    <section className="py-20" id="cta">
      <div className="container mx-auto px-4 md:px-6">
        <div className="bg-foreground rounded-2xl overflow-hidden shadow-xl">
          <div className="grid grid-cols-1 md:grid-cols-2">
            <div className="p-8 md:p-12 flex items-center">
              <div>
                <h2 className="text-3xl md:text-4xl font-bold mb-4 text-white">Bereit, Ihr Event auf das nächste Level zu heben?</h2>
                <p className="text-white/80 text-lg mb-6">
                  Kontaktieren Sie uns noch heute für ein unverbindliches Angebot und machen Sie Ihr nächstes Event zum bargeldlosen Erfolg.
                </p>
                <div className="flex flex-wrap gap-4">
                  <Link href="/angebot">
                    <Button className="bg-background text-foreground hover:bg-gray-100">Angebot anfordern</Button>
                  </Link>
                </div>
              </div>
            </div>
            <div className="hidden md:block bg-[url('/rentro-product-image.png')] bg-cover bg-center"></div>
          </div>
        </div>
      </div>
    </section>
  );
}
