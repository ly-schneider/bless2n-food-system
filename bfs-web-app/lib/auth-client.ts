"use client"

import { adminClient, emailOTPClient } from "better-auth/client/plugins"
import { createAuthClient } from "better-auth/react"
import { ac, adminRole, customerRole } from "./auth/permissions"

export const authClient = createAuthClient({
  baseURL: process.env.NEXT_PUBLIC_APP_URL,
  plugins: [
    emailOTPClient(),
    adminClient({
      ac,
      roles: {
        customer: customerRole,
        admin: adminRole,
      },
    }),
  ],
})

// Export commonly used methods for convenience
export const { signIn, signOut, useSession } = authClient

// Plugin-specific exports
export const emailOtp = authClient.emailOtp
export const adminApi = authClient.admin
