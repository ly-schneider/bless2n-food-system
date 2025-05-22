import React from "react";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";

export default function FAQSection() {
  const faqs = [
    {
      question: "Wie lange dauert die Einrichtung des rentro Systems?",
      answer: "Die Einrichtung dauert in der Regel weniger als 5 Minuten. Das System ist nach dem Auspacken sofort einsatzbereit - Sie müssen nur das Tablet einschalten, mit dem Event linken und können direkt loslegen."
    },
    {
      question: "Funktioniert das System auch ohne Internetverbindung?",
      answer: "Nein, rentro benötigt eine Internetverbindung, um TWINT-Transaktionen zu verarbeiten und Echtzeitdaten zu liefern. Wir empfehlen, eine stabile WLAN-Verbindung oder mobile Daten zu nutzen."
    },
    {
      question: "Was bedeutet 'unlimitierte POS-Lizenzen'?",
      answer: "Das bedeutet, dass Sie so viele Kassensysteme (POS) einrichten können, wie Sie benötigen. Es gibt keine Begrenzung für die Anzahl der Geräte, die Sie gleichzeitig verwenden können."
    },
    {
      question: "Was bedeutet 'Dual-Display'?",
      answer: "Das bedeutet, dass Sie zwei Bildschirme für das Kassensystem verwenden können - einen für den Kassierer und einen für den Kunden. Dies verbessert die Benutzererfahrung und ermöglicht eine einfachere Interaktion."
    },
    {
      question: "Welche Zahlungsmethoden werden unterstützt?",
      answer: "rentro ist eine TWINT-Lösung und unterstützt daher nur TWINT-Zahlungen."
    },
    {
      question: "Muss ich die Hardware kaufen oder kann ich sie mieten?",
      answer: "rentro bietet ein flexibles Mietmodell an. Sie zahlen nur für die Zeit, in der Sie das System tatsächlich nutzen, was es ideal für temporäre Events macht."
    },
    {
      question: "Wie funktioniert das mit den Verkaufsberichten?",
      answer: "Alle Verkaufsdaten werden in Echtzeit erfasst und auf dem Dashboard angezeigt. Sie können Umsätze, Spitzenzeiten und Bestseller sofort erkennen und entsprechend reagieren."
    },
    {
      question: "Bietet rentro auch Support während der Veranstaltung?",
      answer: "Ja, je nach gewähltem Paket haben Sie Zugang zu unserem Support-Team per E-Mail, Telefon oder sogar mit einer dedizierten Person vor Ort."
    }
  ];

  return (
    <section className="py-20 bg-muted" id="faq">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Häufig gestellte Fragen</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">
            Antworten auf die wichtigsten Fragen zu unserem Kassensystem
          </p>
        </div>
        
        <div className="max-w-3xl mx-auto">
          <Accordion type="single" collapsible className="space-y-4">
            {faqs.map((faq, index) => (
              <AccordionItem key={index} value={`item-${index}`} className="bg-background rounded-lg shadow-sm">
                <AccordionTrigger className="px-6 py-4 text-left font-medium">{faq.question}</AccordionTrigger>
                <AccordionContent className="px-6 pb-4 text-gray-600">
                  {faq.answer}
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </div>
        
        <div className="text-center mt-12">
          <p className="text-gray-600 mb-4">Haben Sie weitere Fragen?</p>
          <a href="#kontakt" className="text-rentro font-medium hover:underline">
            Kontaktieren Sie uns
          </a>
        </div>
      </div>
    </section>
  );
}