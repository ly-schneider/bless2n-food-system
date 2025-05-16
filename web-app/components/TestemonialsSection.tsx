export default function TestimonialsSection() {
  const testimonials = [
    {
      quote: "Mit rentro konnten wir die Warteschlangen an unserem Festival drastisch reduzieren. Das System war in Minuten eingerichtet und lief den ganzen Tag ohne Probleme.",
      author: "Markus Schmidt",
      position: "Eventmanager, Zürich Open Air"
    },
    {
      quote: "Besonders beeindruckend ist die Offline-Funktionalität. Bei unserem letzten Event hatten wir Netzwerkprobleme, aber das rentro-System lief einfach weiter.",
      author: "Lisa Müller",
      position: "Organisatorin, Food Festival Basel"
    },
    {
      quote: "Die Echtzeit-Analytics haben uns geholfen, Engpässe zu identifizieren und sofort zu reagieren. Das hat unseren Umsatz deutlich gesteigert.",
      author: "Thomas Weber",
      position: "Messe-Koordinator, Handwerksmesse München"
    }
  ];

  return (
    <section className="py-20 bg-background" id="testimonials">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Was unsere Kunden sagen</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">
            Erfahrungen von Veranstaltern, die rentro bereits erfolgreich eingesetzt haben
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {testimonials.map((testimonial, index) => (
            <div 
              key={index} 
              className="bg-muted p-8 rounded-lg"
            >
              <div className="mb-6 text-2xl font-serif text-gray-400">"</div>
              <p className="text-gray-700 mb-6 italic">"{testimonial.quote}"</p>
              <div>
                <p className="font-semibold">{testimonial.author}</p>
                <p className="text-sm text-gray-500">{testimonial.position}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}