import Link from "next/link";
import { Button } from "./ui/button";
import { CheckCircle2 } from "lucide-react";

export default function PricingSection() {
  const plans = [
    {
      name: "Event Lite",
      price: "390",
      duration: "für 3 Tage • CHF 120 je Zusatztag",
      description: "TWINT-Flatrate für kleine & mittlere Anlässe – inkl. persönlichem Vor-Ort Setup",
      features: [
        "Unlimitierte POS-Lizenzen",
        "Self-Checkout oder Dual-Display",
        "TWINT-Integration",
        "Basis-Dashboard in Echtzeit",
        "Vor-Ort Setup bis 3 h",
        "E-Mail Support",
      ],
      highlighted: true,
    },
    {
      name: "Event Pro",
      price: "780",
      duration: "für 3 Tage • CHF 200 je Zusatztag",
      description: "Rundum-Service für Festivals & Grossanlässe – inkl. Vor-Ort Setup und Premium-Features",
      features: [
        "Alle Lite-Leistungen",
        "Brandbares POS-Interface",
        "Erweiterte Echtzeit-Analytics Dashboard",
        "Kundenspezifische Anpassungen",
        "Technischer Support vor Ort",
        "E-Mail und Telefon Support",
      ],
    },
  ];

  return (
    <section className="py-20" id="preise">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Transparente Preise</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">Wählen Sie den Plan, der zu Ihren Bedürfnissen passt. Keine versteckten Gebühren.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
          {plans.map((plan, index) => (
            <div
              key={index}
              className={`rounded-lg overflow-hidden ${plan.highlighted ? "shadow-xl border-2 border-foreground" : "shadow-md border border-gray-200"}`}
            >
              <div className={`p-8 ${plan.highlighted ? "bg-foreground text-background" : "bg-background"}`}>
                <h3 className="text-2xl font-bold mb-2">{plan.name}</h3>
                <div className="flex items-end mb-4">
                  <span className="text-4xl font-bold">CHF {plan.price}</span>
                  <span className="ml-1 text-sm">{plan.duration}</span>
                </div>
                <p className={`${plan.highlighted ? "text-white/80" : "text-gray-600"} mb-6`}>{plan.description}</p>
              </div>
              <div className="bg-background p-8">
                <ul className="space-y-4 mb-8">
                  {plan.features.map((feature, featureIndex) => (
                    <li key={featureIndex} className="flex items-start">
                      <CheckCircle2 className="h-6 w-6 mr-2 pt-0.5 flex-shrink-0 text-foreground" />
                      <span>{feature}</span>
                    </li>
                  ))}
                </ul>
                <Link href={"/angebot?angebot=" + (plan.name === "Event Lite" ? "lite" : "pro")}>
                  <Button
                    className={`w-full ${plan.highlighted ? "bg-foreground text-white" : "bg-background text-foreground border border-foreground hover:bg-foreground hover:text-background transition-all duration-300"}`}
                  >
                    Angebot anfordern
                  </Button>
                </Link>
              </div>
            </div>
          ))}
        </div>

        <div className="text-center mt-12">
          <p className="text-gray-600 mb-4">Benötigen Sie eine massgeschneiderte Lösung?</p>
          <Link href="/#kontakt">
            <Button variant="outline" className="border-foreground text-foreground hover:bg-foreground hover:text-white">
              Kontaktieren Sie uns
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
}
