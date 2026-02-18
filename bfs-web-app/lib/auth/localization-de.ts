/**
 * German localization strings for authentication UI.
 *
 * These strings are used in custom auth components (login, invite accept, etc.)
 * for consistent German language support throughout the auth flow.
 */
export const authLocalizationDe = {
  // Email
  email: "E-Mail",
  emailPlaceholder: "m@beispiel.de",
  emailRequired: "E-Mail-Adresse ist erforderlich",

  // Email OTP
  emailOtp: "E-Mail-Code",
  emailOtpSendAction: "Code senden",
  emailOtpVerifyAction: "Code bestaetigen",
  emailOtpDescription: "Gib deine E-Mail-Adresse ein, um einen Code zu erhalten",
  emailOtpVerificationSent: "Bitte ueberpruefe deine E-Mails fuer den Bestaetigungscode.",

  // Sign in/out
  signIn: "Anmelden",
  signInWith: "Anmelden mit",
  signOut: "Abmelden",

  // Social/Providers
  orContinueWith: "Oder weiter mit",

  // Actions
  continue: "Weiter",
  goBack: "Zurueck",
  resendCode: "Code erneut senden",

  // Errors
  isInvalid: "ist ungültig",
  isRequired: "ist erforderlich",
  requestFailed: "Anfrage fehlgeschlagen",
} as const

export type AuthLocalizationDe = typeof authLocalizationDe
