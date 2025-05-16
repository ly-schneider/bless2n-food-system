import { CheckCircle2 } from "lucide-react";

export default function FeaturesSection() {
  const features = [
    {
      title: "Schnelle Einrichtung",
      description: "Setup in weniger als 5 Minuten. Alles was Sie brauchen ist in einem Paket.",
      icon: CheckCircle2
    },
    {
      title: "Bargeldlose Zahlungen",
      description: "Akzeptieren Sie alle gängigen bargeldlosen Zahlungsmethoden mit einem Gerät.",
      icon: CheckCircle2
    },
    {
      title: "Offline-Modus",
      description: "Auch bei schwacher Internetverbindung bleibt Ihr Kassensystem voll funktionsfähig.",
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
      title: "Premium Support",
      description: "Schweizer Support-Team, das Ihnen bei allen Fragen zur Seite steht.",
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
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <div 
              key={index} 
              className="bg-background p-6 rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300"
            >
              <div className="mb-4 text-rentro">
                <feature.icon className="w-10 h-10" />
              </div>
              <h3 className="text-xl font-semibold mb-3">{feature.title}</h3>
              <p className="text-gray-600">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}