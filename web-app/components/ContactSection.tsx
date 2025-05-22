import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Textarea } from "./ui/textarea";

export default function ContactSection() {
  return (
    <section className="py-20 bg-background" id="kontakt">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Kontaktieren Sie uns</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">
            Haben Sie Fragen oder möchten Sie ein individuelles Angebot erhalten? 
            Unser Team steht Ihnen gerne zur Verfügung.
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-12 max-w-5xl mx-auto">
          <div>
            <form>
              <div className="space-y-4">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium mb-1">
                    Name
                  </label>
                  <Input id="name" placeholder="Ihr Name" required />
                </div>
                
                <div>
                  <label htmlFor="email" className="block text-sm font-medium mb-1">
                    E-Mail
                  </label>
                  <Input id="email" type="email" placeholder="ihre.email@beispiel.com" required />
                </div>
                
                <div>
                  <label htmlFor="phone" className="block text-sm font-medium mb-1">
                    Telefonnummer
                  </label>
                  <Input id="phone" placeholder="+41 XX XXX XX XX" />
                </div>
                
                <div>
                  <label htmlFor="message" className="block text-sm font-medium mb-1">
                    Nachricht
                  </label>
                  <Textarea 
                    id="message" 
                    placeholder="Beschreiben Sie Ihre Anfrage oder Ihr Event..." 
                    rows={5} 
                    required
                  />
                </div>
                
                <Button type="submit" variant={"default"} className="bg-foreground w-full">
                  Nachricht senden
                </Button>
              </div>
            </form>
          </div>
          
          <div className="flex flex-col justify-between">
            <div>
              <h3 className="text-xl font-semibold mb-4">Kontaktinformationen</h3>
              <div className="space-y-3">
                <p className="flex items-center">
                  <span className="mr-3 text-rentro">📞</span>
                  <span>+41 079 757 16 08</span>
                </p>
                <p className="flex items-center">
                  <span className="mr-3 text-rentro">✉️</span>
                  <span>info@rentro.ch</span>
                </p>
              </div>
              
              <div className="mt-8">
                <h3 className="text-xl font-semibold mb-4">Geschäftszeiten</h3>
                <div className="space-y-2">
                  <p>Montag - Freitag: 9:00 - 18:00 Uhr</p>
                  <p>Samstag & Sonntag: Geschlossen</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
