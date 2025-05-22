import { CheckCircle2 } from "lucide-react";

export default function FeaturesSection() {
  const features = [
    {
      title: "Schnelle Einrichtung",
      description: "Setup in weniger als 5 Minuten. Alles was Sie brauchen ist in einem Paket.",
      icon: CheckCircle2
    },
    {
      title: "TWINT-Integration",
      description: "Akzeptieren Sie Zahlungen mit TWINT und bieten Sie Ihren Kunden mehr Flexibilität.",
      icon: CheckCircle2
    },
    {
      title: "Echtzeit-Analytics",
      description: "Verfolgen Sie Verkäufe und Trends live während Ihrer Veranstaltung.",
      icon: CheckCircle2
    },
    {
      title: "Benutzerfreundlichkeit",
      description: "Intuitive Benutzeroberfläche, die minimale Schulung erfordert.",
      icon: CheckCircle2
    },
    {
      title: "Schweizer Qualität und Support",
      description: "Profitieren Sie von unserem erstklassigen Support und der Zuverlässigkeit eines Schweizer Unternehmens.",
      icon: CheckCircle2
    }
  ];

  return (
    <section className="py-20 bg-rentro-light" id="funktionen">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Unsere Funktionen</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">
            rentro bietet alles, was Sie für ein erfolgreiches Event-Kassensystem benötigen
          </p>
        </div>

        <div className="flex flex-wrap justify-center gap-8">
          {features.map((feature, index) => (
            <div 
              key={index} 
              className="w-[30%] p-6 bg-background rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300"
            >
              {/* <div className="mb-4 text-rentro">
                <feature.icon className="w-10 h-10" />
              </div> */}
              <h3 className="text-xl font-semibold mb-3">{feature.title}</h3>
              <p className="text-gray-600">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}