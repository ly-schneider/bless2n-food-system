"use client"

// Minimal bridge shim for native WebView environments
// Exposes globalThis.PosBridge expected by /pos UI by delegating to window.BFS if available

declare global {
  var PosBridge:
    | {
        print?: (s: string) => void
        payWithCard?: (p: { amountCents: number; currency: string; reference: string }) => void
      }
    | undefined
  // window.BFS shape (optional)
  interface Window {
    BFS?: {
      sumup?: {
        pay?: (p: {
          amount: number
          currency: string
          metadata?: Record<string, unknown>
          correlationId?: string
        }) => void
      }
      printer?: { print?: (p: { mode: "system" | "escpos"; content: string; correlationId?: string }) => void }
    }
  }
}

// Install shim once per page load
;(() => {
  if (typeof window === "undefined") return
  try {
    if (!globalThis.PosBridge && window.BFS) {
      globalThis.PosBridge = {
        print: (s: string) => {
          const correlationId = `print_${Date.now()}`
          window.BFS?.printer?.print?.({ mode: "escpos", content: s, correlationId })
        },
        payWithCard: (p) => {
          const correlationId = p.reference || `pos_${Date.now()}`
          window.BFS?.sumup?.pay?.({
            amount: p.amountCents / 100,
            currency: p.currency,
            metadata: { reference: p.reference },
            correlationId,
          })
        },
      }
    }
  } catch {
    // ignore
  }
})()

export {}
