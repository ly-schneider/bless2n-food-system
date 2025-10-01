export default function DatenschutzPage() {
  return (
    <div className="container mx-auto max-w-3xl px-4 py-8">
      <h1 className="mb-6 text-3xl font-semibold">Datenschutzerklärung</h1>

      <div className="max-w-none text-sm leading-relaxed [&_div]:mb-6 [&_div:last-child]:mb-0">
        <div>
          <p>
            <strong>Verantwortliche Stelle.</strong>
            <br />
            [Unternehmensname], [Adresse], [E-Mail]. Wir entscheiden über Zwecke und Mittel der Verarbeitung deiner
            Personendaten.
          </p>
        </div>
        <div>
          <p>
            <strong>Welche Daten wir verarbeiten (datenminimiert).</strong>
          </p>
          <ul>
            <li>
              <strong>Bestelldaten:</strong> Artikel, Beträge, Abhol-/Lieferdetails, Bestellstatus.
            </li>
            <li>
              <strong>Kontaktdaten (optional):</strong> E-Mail für Quittungen/Status; Telefon für Liefer­rückfragen.
            </li>
            <li>
              <strong>Zahlung:</strong> über unsere Zahlungsdienstleister; wir erhalten Status/Transaktions-ID, nicht
              deine vollständigen Karten-/TWINT-Details. Wir erheben nur, was für die Leistungserbringung und Sicherheit
              nötig ist.
            </li>
          </ul>
        </div>
        <div>
          <p>
            <strong>Warum wir verarbeiten (Rechtsgrundlagen).</strong>
          </p>
          <ul>
            <li>
              <strong>Vertragserfüllung:</strong> Zahlung abwickeln, Bestellung zubereiten und übergeben, Support
              leisten.
            </li>
            <li>
              <strong>Berechtigtes Interesse:</strong> Betrugsprävention, Systemsicherheit, einfache Betriebsanalysen.
            </li>
          </ul>
        </div>
        <div>
          <p>
            <strong>Wer Daten erhält (Auftragsverarbeiter).</strong>
          </p>
          <ul>
            <li>
              <strong>Zahlungen:</strong> z. B. Stripe/TWINT; sie verarbeiten Zahlungen sicher und teilen nur notwendige
              Infos (z. B. Erfolg/Fehlschlag, Transaktions-ID).
            </li>
            <li>
              <strong>Hosting/IT:</strong> vertraglich gebundene Dienstleister mit Auftragsverarbeitungsverträgen. Wir
              bleiben für den Schutz deiner Daten verantwortlich.
            </li>
          </ul>
        </div>
        <div>
          <p>
            <strong>Wo Daten verarbeitet werden.</strong>
            <br />
            Vorrangig in der Schweiz/EWR. Bei Übermittlungen ins Ausland setzen wir die gesetzlich erforderlichen
            Garantien ein.
          </p>
        </div>
        <div>
          <p>
            <strong>Wer darf bestellen.</strong>
            <br />
            Unser Angebot richtet sich an urteilsfähige Personen. Wenn wir feststellen, dass eine Bestellung ohne
            notwendige Zustimmung erfolgt ist, löschen wir die Daten oder holen die Zustimmung der
            erziehungsberechtigten Person ein.
          </p>
        </div>
        <div>
          <p>
            <strong>Aufbewahrung.</strong>
            <br />
            Nur so lange wie nötig für die Bestellung, gesetzliche Pflichten (z. B. Buchhaltung) oder zur Klärung von
            Anliegen — danach löschen oder anonymisieren wir.
          </p>
        </div>
        <div>
          <p>
            <strong>Deine Rechte.</strong>
            <br />
            Auskunft, Berichtigung, Löschung, Datenübertragung; Widerspruch gegen bestimmte Verarbeitungen. Kontakt:
            [privacy@…].
          </p>
        </div>
        <div>
          <p>
            <strong>Cookies &amp; Tracking.</strong>
            <br />
            Standardmässig nur technisch notwendige Cookies für Bestellung und Sicherheit. Analyse/Marketing-Cookies
            sind <strong>aus</strong>, ausser du stimmst zu.
          </p>
        </div>
        <div>
          <p>
            <strong>Sicherheit.</strong>
            <br />
            Wir setzen angemessene technische und organisatorische Massnahmen zum Schutz deiner Daten ein.
          </p>
        </div>
        <div>
          <p>
            <strong>Kontakt &amp; Beschwerden.</strong>
            <br />
            Fragen oder Anliegen: [privacy@…]. Du kannst dich auch an den EDÖB wenden.
          </p>
        </div>
        <div>
          <p>
            <strong>Updates.</strong>
            <br />
            Änderungen veröffentlichen wir hier und kennzeichnen das Gültigkeitsdatum.
          </p>
        </div>
      </div>
    </div>
  )
}
