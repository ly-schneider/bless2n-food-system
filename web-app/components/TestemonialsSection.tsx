import Link from "next/link";
import { Button } from "./ui/button";

export default function TestimonialsSection() {
  const testimonials = [
    {
      quote: "Die App war für unser Musical eine echte Unterstützung – intuitiv, zuverlässig und sie hat unserem Team bei der Arbeit sehr geholfen.",
      author: "ICF Bern",
      position: "Leitung Musical «Joyride»",
      path: "/kunden/icf-bern",
    },
  ];

  return (
    <section className="py-20 bg-background" id="einsaetze">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Was unsere Kunden sagen</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">Erfahrungen von Veranstaltern, die rentro bereits erfolgreich eingesetzt haben</p>
        </div>

        <div className="flex flex-col md:flex-row gap-6">
          {testimonials.map((testimonial, index) => (
            <div key={index} className="bg-muted p-8 rounded-lg w-[50%] mx-auto">
              <div className="mb-6 text-2xl font-serif text-gray-400">"</div>
              <p className="text-gray-700 mb-6 italic">"{testimonial.quote}"</p>
              <div className="flex items-center gap-4 justify-between">
                <div>
                  <p className="font-semibold">{testimonial.author}</p>
                  <p className="text-sm text-gray-500">{testimonial.position}</p>
                </div>
                {/* <Link href={testimonial.path}>
                  <Button variant="outline" className="bg-transparent border-black px-4 py-2 rounded-md text-sm">
                    Mehr erfahren
                  </Button>
                </Link> */}
              </div>
            </div>
          ))}
          <div className="bg-muted p-8 rounded-lg w-[50%] mx-auto">
            <div className="mb-6 text-2xl font-serif text-gray-400">"</div>
            <p className="font-semibold mb-6 text-xl">Sei du die nächste Erfolgsgeschichte...</p>
          </div>
        </div>
      </div>
    </section>
  );
}
