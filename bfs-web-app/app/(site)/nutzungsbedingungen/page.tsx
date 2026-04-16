import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Nutzungsbedingungen",
  description:
    "Nutzungsbedingungen der BlessThun Food System Plattform für die Online-Essensbestellung in der Schweiz.",
}

export default function NutzungsbedingungenPage() {
  return (
    <div className="container mx-auto max-w-2xl px-4 py-8">
      <h1 className="mb-2 text-2xl font-semibold">Nutzungsbedingungen</h1>
      <p className="text-muted-foreground mb-8 text-sm">Letzte Aktualisierung: 13. April 2026</p>

      <div className="space-y-8 text-sm leading-relaxed">
        <section>
          <h2 className="mb-3 text-lg font-semibold">1. Geltungsbereich</h2>
          <p>
            Diese Nutzungsbedingungen regeln die Verwendung der BlessThun Food System (nachfolgend «Plattform») sowie
            die darüber getätigten Bestellungen von Speisen und Getränken. Mit der Nutzung der Plattform oder dem
            Aufgeben einer Bestellung akzeptieren Sie diese Nutzungsbedingungen.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">2. Betreiber</h2>
          <p>
            BlessThun
            <br />
            Industriestrasse 5
            <br />
            3600 Thun
            <br />
            Schweiz
          </p>
          <p className="mt-2">E-Mail: levyn.schneider@leys.ch</p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">3. Angebot und Verfügbarkeit</h2>
          <p>
            Über die Plattform können Speisen und Getränke an ausgewählten Standorten online bestellt werden. Das
            Angebot richtet sich nach den jeweiligen Betriebszeiten und dem aktuellen Sortiment. Wir behalten uns vor,
            das Angebot jederzeit anzupassen oder die Plattform vorübergehend ausser Betrieb zu nehmen.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">4. Registrierung und Benutzerkonto</h2>
          <p>
            Zur Nutzung der Plattform kann ein Benutzerkonto erstellt werden — entweder mit E-Mail-Adresse und Passwort
            oder über Google. Sie sind dafür verantwortlich, korrekte Angaben zu machen und Ihre Zugangsdaten
            vertraulich zu behandeln.
          </p>
          <p className="mt-2">
            Bei Verstössen gegen diese Nutzungsbedingungen oder bei missbräuchlicher Nutzung können wir Ihr Konto
            sperren oder löschen.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">5. Bestellung und Vertragsschluss</h2>
          <p>
            Die Darstellung der Produkte auf der Plattform stellt kein verbindliches Angebot dar. Mit dem Absenden einer
            Bestellung geben Sie ein verbindliches Kaufangebot ab. Der Vertrag kommt mit der Bestätigung Ihrer
            Bestellung zustande.
          </p>
          <p className="mt-2">
            Wir behalten uns vor, Bestellungen abzulehnen — insbesondere bei Nichtverfügbarkeit von Produkten oder bei
            Verdacht auf Missbrauch.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">6. Preise und Zahlung</h2>
          <p>Alle Preise sind in Schweizer Franken (CHF) angegeben.</p>
          <p className="mt-2">
            Die Zahlung erfolgt über den Zahlungsdienstleister Payrexx mittels TWINT. Mit dem Absenden der Bestellung
            autorisieren Sie die Belastung des fälligen Betrags. Kommt die Zahlung nicht zustande, wird die Bestellung
            nicht ausgeführt.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">7. Abholung</h2>
          <p>
            Bestellte Produkte sind am jeweiligen Standort abzuholen. Nach Abschluss der Bestellung erhalten Sie einen
            QR-Code, den Sie bei der Abholung vorzeigen. Wird eine Bestellung nicht innerhalb einer angemessenen Frist
            abgeholt, kann sie ohne Rückerstattung storniert werden.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">8. Stornierung</h2>
          <p>
            Eine Stornierung ist nur möglich, solange die Zubereitung noch nicht begonnen hat. Ein gesetzliches
            Widerrufsrecht für Online-Bestellungen besteht nach Schweizer Recht nicht. Bei berechtigter Stornierung wird
            der bezahlte Betrag auf den gleichen Zahlungsweg zurückerstattet.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">9. Allergene und Produktinformationen</h2>
          <p>
            Abbildungen und Beschreibungen auf der Plattform dienen der allgemeinen Orientierung und können vom
            tatsächlichen Produkt abweichen. Bei Fragen zu Inhaltsstoffen oder Allergenen wenden Sie sich bitte direkt
            an das Personal vor Ort. Die Verantwortung für die Beachtung individueller Unverträglichkeiten liegt bei
            Ihnen.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">10. Verantwortung und Haftung</h2>
          <p>
            Wir haften uneingeschränkt für Schäden aus Vorsatz oder grober Fahrlässigkeit (Art. 100 OR). Bei leichter
            Fahrlässigkeit ist die Haftung auf den vorhersehbaren, vertragstypischen Schaden beschränkt.
          </p>
          <p className="mt-2">
            Für die ununterbrochene Verfügbarkeit der Plattform übernehmen wir keine Gewähr. Für die Qualität und
            Sicherheit der Speisen gelten die Bestimmungen des Schweizer Lebensmittelgesetzes (LMG).
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">11. Höhere Gewalt</h2>
          <p>
            Bei höherer Gewalt (Naturereignisse, Epidemien, Stromausfälle, behördliche Anordnungen) sind wir von der
            Leistungspflicht befreit. Bereits geleistete Zahlungen für nicht erbrachte Leistungen werden
            zurückerstattet.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">12. Geistiges Eigentum</h2>
          <p>
            Sämtliche Inhalte der Plattform (Texte, Bilder, Logos, Software) sind urheberrechtlich geschützt. Jede
            Vervielfältigung oder Weiterverwendung ohne vorgängige schriftliche Zustimmung ist untersagt.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">13. Unzulässige Nutzung</h2>
          <p className="mb-2">Es ist untersagt, die Plattform:</p>
          <ul className="list-disc space-y-1 pl-5">
            <li>für betrügerische oder missbräuchliche Zwecke zu verwenden</li>
            <li>technische Schutzmechanismen zu umgehen oder zu manipulieren</li>
            <li>automatisierte Systeme (Bots, Scraper) ohne Genehmigung einzusetzen</li>
          </ul>
          <p className="mt-2">
            Verstösse können zur sofortigen Sperrung des Benutzerkontos und zu rechtlichen Konsequenzen führen.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">14. Datenschutz</h2>
          <p>
            Die Bearbeitung Ihrer Personendaten erfolgt gemäss unserer{" "}
            <a href="/datenschutz" className="underline">
              Datenschutzerklärung
            </a>{" "}
            und im Einklang mit dem Schweizer Datenschutzgesetz (nDSG).
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">15. Änderungen</h2>
          <p>
            Wir können diese Nutzungsbedingungen jederzeit anpassen. Änderungen werden auf der Plattform veröffentlicht.
            Bei wesentlichen Änderungen informieren wir registrierte Nutzer per E-Mail. Die weitere Nutzung der
            Plattform nach Inkrafttreten gilt als Zustimmung.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">16. Salvatorische Klausel</h2>
          <p>
            Sollte eine Bestimmung dieser Nutzungsbedingungen unwirksam sein, bleiben die übrigen Bestimmungen davon
            unberührt. Die unwirksame Bestimmung wird durch eine Regelung ersetzt, die dem wirtschaftlichen Zweck am
            nächsten kommt.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">17. Anwendbares Recht und Gerichtsstand</h2>
          <p>
            Es gilt ausschliesslich schweizerisches Recht unter Ausschluss des UN-Kaufrechts (CISG). Gerichtsstand ist
            der Wohnsitz des Konsumenten oder der Sitz des Betreibers gemäss ZPO Art. 32.
          </p>
        </section>
      </div>
    </div>
  )
}
