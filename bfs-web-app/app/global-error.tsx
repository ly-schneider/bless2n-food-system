"use client"

import { useEffect } from "react"

export default function GlobalError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    console.error("Global application error:", error)
  }, [error])

  return (
    <html lang="de">
      <body
        style={{
          margin: 0,
          minHeight: "100vh",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#E9E7E6",
          fontFamily: "system-ui, -apple-system, sans-serif",
          padding: "1rem",
        }}
      >
        <div style={{ maxWidth: "28rem", textAlign: "center" }}>
          <div
            style={{
              width: "5rem",
              height: "5rem",
              margin: "0 auto 1.5rem",
              borderRadius: "50%",
              backgroundColor: "rgba(252, 102, 102, 0.1)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="40"
              height="40"
              viewBox="0 0 24 24"
              fill="none"
              stroke="#FC6666"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" />
              <path d="M12 9v4" />
              <path d="M12 17h.01" />
            </svg>
          </div>

          <h1 style={{ fontSize: "1.5rem", fontWeight: 600, marginBottom: "1rem", color: "#000" }}>
            Ein kritischer Fehler ist aufgetreten
          </h1>

          <p style={{ color: "#7B7B7B", marginBottom: "2rem", lineHeight: 1.5 }}>
            Die Anwendung konnte nicht geladen werden. Bitte versuche es erneut oder lade die Seite neu.
          </p>

          <div style={{ display: "flex", gap: "0.75rem", justifyContent: "center", flexWrap: "wrap" }}>
            <button
              onClick={reset}
              style={{
                display: "inline-flex",
                alignItems: "center",
                justifyContent: "center",
                gap: "0.5rem",
                padding: "0.625rem 1.5rem",
                backgroundColor: "#F87778",
                color: "#FDFDFD",
                border: "none",
                borderRadius: "9999px",
                fontSize: "0.875rem",
                fontWeight: 500,
                cursor: "pointer",
                transition: "background-color 0.15s",
              }}
              onMouseOver={(e) => (e.currentTarget.style.backgroundColor = "#F66969")}
              onMouseOut={(e) => (e.currentTarget.style.backgroundColor = "#F87778")}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
                <path d="M21 3v5h-5" />
              </svg>
              Erneut versuchen
            </button>

            {/* eslint-disable-next-line @next/next/no-html-link-for-pages -- global-error cannot use Next.js components */}
            <a
              href="/"
              style={{
                display: "inline-flex",
                alignItems: "center",
                justifyContent: "center",
                gap: "0.5rem",
                padding: "0.625rem 1.5rem",
                backgroundColor: "#FDFDFD",
                color: "#000",
                border: "1px solid #D7D7D7",
                borderRadius: "7px",
                fontSize: "0.875rem",
                fontWeight: 500,
                cursor: "pointer",
                textDecoration: "none",
                transition: "background-color 0.15s",
              }}
              onMouseOver={(e) => (e.currentTarget.style.backgroundColor = "#F3F3F3")}
              onMouseOut={(e) => (e.currentTarget.style.backgroundColor = "#FDFDFD")}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                <polyline points="9 22 9 12 15 12 15 22" />
              </svg>
              Zur Startseite
            </a>
          </div>

          {error.digest && (
            <div
              style={{
                marginTop: "2rem",
                padding: "1rem",
                backgroundColor: "#FDFDFD",
                border: "1px solid #D7D7D7",
                borderRadius: "0.5rem",
              }}
            >
              <p style={{ fontSize: "0.75rem", color: "#7B7B7B", margin: 0 }}>
                Fehler-ID:{" "}
                <code
                  style={{
                    backgroundColor: "#FDFDFD",
                    padding: "0.125rem 0.375rem",
                    borderRadius: "0.25rem",
                    fontFamily: "monospace",
                  }}
                >
                  {error.digest}
                </code>
              </p>
            </div>
          )}
        </div>
      </body>
    </html>
  )
}
