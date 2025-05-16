import { Button } from "./ui/button";
import { CheckCircle2 } from "lucide-react";

export default function PricingSection() {
  const plans = [
    {
      name: "Basic",
      price: "99",
      duration: "pro Tag",
      description: "Ideal für kleine Pop-up Stores und kurze Events",
      features: [
        "1x iPad mit Kassensoftware",
        "1x Kartenleser",
        "TWINT QR-Integration",
        "24h Support via Email",
        "Grundlegende Verkaufsberichte"
      ]
    },
    {
      name: "Standard",
      price: "199",
      duration: "pro Tag",
      description: "Perfekt für mittlere Veranstaltungen und Festivals",
      features: [
        "2x iPad mit Kassensoftware",
        "2x Kartenleser",
        "TWINT QR-Integration",
        "Apple Pay & Google Pay",
        "Echtzeit-Analytics Dashboard",
        "24/7 Priority Support",
        "Offline-Modus"
      ],
      highlighted: true
    },
    {
      name: "Premium",
      price: "399",
      duration: "pro Tag",
      description: "Für grosse Events mit hohem Durchsatz",
      features: [
        "5x iPad mit Kassensoftware",
        "5x Kartenleser",
        "Alle Zahlungsmethoden",
        "Erweiterte Analytics",
        "Dedizierter Support-Mitarbeiter",
        "Vor-Ort Setup (optional)",
        "Kundenspezifische Anpassungen"
      ]
    }
  ];

  return (
    <section className="py-20" id="preise">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Transparente Preise</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">
            Wählen Sie den Plan, der zu Ihren Bedürfnissen passt. Keine versteckten Gebühren.
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {plans.map((plan, index) => (
            <div 
              key={index} 
              className={`rounded-lg overflow-hidden ${plan.highlighted ? 'shadow-xl border-2 border-foreground' : 'shadow-md border border-gray-200'}`}
            >
              <div className={`p-8 ${plan.highlighted ? 'bg-foreground text-background' : 'bg-background'}`}>
                <h3 className="text-2xl font-bold mb-2">{plan.name}</h3>
                <div className="flex items-end mb-4">
                  <span className="text-4xl font-bold">CHF {plan.price}</span>
                  <span className="ml-1 text-sm">{plan.duration}</span>
                </div>
                <p className={`${plan.highlighted ? 'text-white/80' : 'text-gray-600'} mb-6`}>
                  {plan.description}
                </p>
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
                <Button 
                  className={`w-full ${plan.highlighted ? 'bg-foreground text-white' : 'bg-background text-foreground border border-foreground hover:bg-gray-50'}`}
                >
                  Angebot anfordern
                </Button>
              </div>
            </div>
          ))}
        </div>
        
        <div className="text-center mt-12">
          <p className="text-gray-600 mb-4">Benötigen Sie eine massgeschneiderte Lösung?</p>
          <Button variant="outline" className="border-foreground text-foreground hover:bg-foreground hover:text-white">
            Kontaktieren Sie uns
          </Button>
        </div>
      </div>
    </section>
  );
}