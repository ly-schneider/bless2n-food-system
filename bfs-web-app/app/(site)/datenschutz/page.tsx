import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Datenschutzerklärung",
  description:
    "Datenschutzerklärung der BlessThun Food System Plattform gemäss dem Schweizer Datenschutzgesetz (nDSG).",
}

export default function PrivacyPolicyPage() {
  return (
    <div className="container mx-auto max-w-2xl px-4 py-8">
      <h1 className="mb-2 text-2xl font-semibold">Datenschutzerklärung</h1>
      <p className="text-muted-foreground mb-8 text-sm">Letzte Aktualisierung: 13. April 2026</p>

      <div className="space-y-8 text-sm leading-relaxed">
        <section>
          <p>
            Diese Datenschutzerklärung informiert Sie gemäss dem Schweizer Bundesgesetz über den Datenschutz (nDSG)
            darüber, wie wir Ihre Personendaten im Zusammenhang mit der Nutzung der BlessThun Food System (nachfolgend
            «Plattform») bearbeiten.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">1. Verantwortliche Stelle</h2>
          <p>Verantwortlich für die Datenbearbeitung ist:</p>
          <p className="mt-2">
            BlessThun Food
            <br />
            Industriestrasse 5
            <br />
            3600 Thun
            <br />
            Schweiz
          </p>
          <p className="mt-2">E-Mail: levy.schneider@leys.ch</p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">2. Erhobene Personendaten</h2>
          <p className="mb-2">Wir bearbeiten folgende Kategorien von Personendaten:</p>
          <ul className="list-disc space-y-1 pl-5">
            <li>
              <strong>Kontodaten:</strong> Name, E-Mail-Adresse, Passwort (verschlüsselt), Google-Konto-Informationen
              bei Anmeldung via Google
            </li>
            <li>
              <strong>Bestelldaten:</strong> Bestellte Artikel, Bestellverlauf, Zeitstempel, Bestellstatus
            </li>
            <li>
              <strong>Zahlungsdaten:</strong> Zahlungsmethode, Transaktions-IDs und Zahlungsstatus (die eigentliche
              Zahlungsabwicklung erfolgt durch Payrexx; wir speichern keine Kreditkarten- oder TWINT-Kontodaten)
            </li>
            <li>
              <strong>Technische Daten:</strong> IP-Adresse, Browsertyp, Betriebssystem, Gerätetyp,
              Sitzungsinformationen
            </li>
            <li>
              <strong>Kommunikationsdaten:</strong> Inhalte von E-Mails, die im Zusammenhang mit Bestellungen versendet
              werden (z.B. Bestellbestätigungen, Quittungen)
            </li>
          </ul>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">3. Zwecke der Datenbearbeitung</h2>
          <p className="mb-2">Wir bearbeiten Ihre Personendaten zu folgenden Zwecken:</p>
          <ul className="list-disc space-y-1 pl-5">
            <li>Erstellung und Verwaltung Ihres Benutzerkontos</li>
            <li>Abwicklung und Erfüllung Ihrer Bestellungen</li>
            <li>Zahlungsabwicklung über Payrexx (TWINT)</li>
            <li>Versand von transaktionalen E-Mails (Bestellbestätigungen, Quittungen)</li>
            <li>Sitzungsverwaltung und Authentifizierung</li>
            <li>Gewährleistung der Sicherheit und Funktionalität der Plattform</li>
            <li>Erfüllung gesetzlicher Aufbewahrungspflichten</li>
          </ul>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">4. Empfänger und Auftragsbearbeiter</h2>
          <p className="mb-2">
            Zur Erbringung unserer Dienstleistungen geben wir Personendaten an folgende Dritte weiter:
          </p>
          <ul className="list-disc space-y-1 pl-5">
            <li>
              <strong>Payrexx AG</strong> (Thun, Schweiz) — Zahlungsabwicklung (TWINT)
            </li>
            <li>
              <strong>Plunk</strong> — Versand transaktionaler E-Mails
            </li>
            <li>
              <strong>Google LLC</strong> — Authentifizierung via Google OAuth (sofern Sie diese Anmeldemethode wählen)
            </li>
            <li>
              <strong>Microsoft Azure</strong> — Cloud-Hosting und Infrastruktur
            </li>
          </ul>
          <p className="mt-2">
            Mit allen Auftragsbearbeitern bestehen Vereinbarungen zur Auftragsbearbeitung gemäss Art. 9 nDSG.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">5. Datenübermittlung ins Ausland</h2>
          <p>
            Bestimmte Auftragsbearbeiter haben ihren Sitz ausserhalb der Schweiz. Daten können in folgende Länder
            übermittelt werden:
          </p>
          <ul className="mt-2 list-disc space-y-1 pl-5">
            <li>
              <strong>USA</strong> (Google, ggf. Plunk) — Die Übermittlung erfolgt auf Grundlage von
              Standarddatenschutzklauseln (Standard Contractual Clauses, SCC) oder anderer geeigneter Garantien gemäss
              Art. 16–17 nDSG.
            </li>
            <li>
              <strong>EU/EWR</strong> (ggf. Azure-Rechenzentren) — Die EU/der EWR verfügt über ein vom Bundesrat
              anerkanntes angemessenes Datenschutzniveau.
            </li>
          </ul>
          <p className="mt-2">
            Soweit möglich, werden Daten in Schweizer Rechenzentren (Azure Switzerland North) verarbeitet.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">6. Cookies und Sitzungsverwaltung</h2>
          <p>
            Wir verwenden technisch notwendige Cookies für die Sitzungsverwaltung und Authentifizierung. Diese Cookies
            sind für den Betrieb der Plattform erforderlich und ermöglichen es Ihnen, angemeldet zu bleiben und
            Bestellungen aufzugeben. Es werden keine Tracking- oder Werbe-Cookies eingesetzt.
          </p>
          <p className="mt-2">
            Sie können Cookies in Ihren Browsereinstellungen deaktivieren. Beachten Sie jedoch, dass die Plattform ohne
            Cookies nicht vollständig funktionsfähig ist.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">7. Aufbewahrungsdauer</h2>
          <p className="mb-2">
            Wir bewahren Ihre Personendaten nur so lange auf, wie es für die genannten Zwecke erforderlich ist:
          </p>
          <ul className="list-disc space-y-1 pl-5">
            <li>
              <strong>Kontodaten:</strong> Bis zur Löschung Ihres Kontos
            </li>
            <li>
              <strong>Bestell- und Zahlungsdaten:</strong> 10 Jahre gemäss gesetzlicher Aufbewahrungspflicht (Art. 958f
              OR)
            </li>
            <li>
              <strong>Technische Daten / Logs:</strong> Maximal 90 Tage
            </li>
          </ul>
          <p className="mt-2">Nach Ablauf der Aufbewahrungsfrist werden Ihre Daten gelöscht oder anonymisiert.</p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">8. Datensicherheit</h2>
          <p>
            Wir treffen angemessene technische und organisatorische Massnahmen zum Schutz Ihrer Personendaten gemäss
            Art. 8 nDSG. Dazu gehören insbesondere verschlüsselte Datenübertragung (TLS), Zugangskontrollen,
            verschlüsselte Passwortspeicherung sowie regelmässige Sicherheitsüberprüfungen.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">9. Ihre Rechte</h2>
          <p className="mb-2">Gemäss nDSG stehen Ihnen folgende Rechte zu:</p>
          <ul className="list-disc space-y-1 pl-5">
            <li>
              <strong>Auskunftsrecht (Art. 25 nDSG):</strong> Sie haben das Recht, Auskunft darüber zu erhalten, ob und
              welche Personendaten wir über Sie bearbeiten.
            </li>
            <li>
              <strong>Recht auf Datenherausgabe (Art. 28 nDSG):</strong> Sie können verlangen, dass Ihnen Ihre Daten in
              einem gängigen elektronischen Format herausgegeben werden.
            </li>
            <li>
              <strong>Recht auf Berichtigung:</strong> Sie können die Korrektur unrichtiger Personendaten verlangen.
            </li>
            <li>
              <strong>Recht auf Löschung:</strong> Sie können die Löschung Ihrer Personendaten verlangen, sofern keine
              gesetzlichen Aufbewahrungspflichten entgegenstehen. Sie können Ihr Konto jederzeit in den
              Profileinstellungen löschen.
            </li>
          </ul>
          <p className="mt-2">
            Zur Ausübung Ihrer Rechte kontaktieren Sie uns bitte per E-Mail an [kontakt@example.ch]. Wir werden Ihr
            Anliegen innert 30 Tagen bearbeiten.
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">10. Beschwerderecht</h2>
          <p>
            Sie haben das Recht, eine Beschwerde beim Eidgenössischen Datenschutz- und Öffentlichkeitsbeauftragten
            (EDÖB) einzureichen:
          </p>
          <p className="mt-2">
            EDÖB
            <br />
            Feldeggweg 1
            <br />
            3003 Bern
            <br />
            Schweiz
            <br />
            www.edoeb.admin.ch
          </p>
        </section>

        <section>
          <h2 className="mb-3 text-lg font-semibold">11. Änderungen</h2>
          <p>
            Wir behalten uns vor, diese Datenschutzerklärung jederzeit anzupassen. Die aktuelle Fassung ist auf der
            Plattform abrufbar. Bei wesentlichen Änderungen informieren wir Sie über die auf der Plattform angegebene
            E-Mail-Adresse.
          </p>
        </section>
      </div>
    </div>
  )
}
