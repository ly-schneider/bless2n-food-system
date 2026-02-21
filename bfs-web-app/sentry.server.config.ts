import * as Sentry from "@sentry/nextjs"

const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN

if (dsn) {
  const appUrl = process.env.NEXT_PUBLIC_APP_URL ?? ""

  Sentry.init({
    dsn,
    environment: appUrl.includes("staging") ? "staging" : "production",
    release: process.env.APP_VERSION,
    tracesSampleRate: 0.2,
    enableLogs: true,
    integrations: [
      Sentry.consoleLoggingIntegration({ levels: ["warn", "error"] }),
    ],
  })
}
