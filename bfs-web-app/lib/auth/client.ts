"use client"

/**
 * Better Auth client re-exports.
 *
 * Re-exports from the root lib/auth-client.ts for backwards compatibility
 * with existing imports from "@/lib/auth/client".
 */
export { authClient, signIn, signOut, useSession, emailOtp } from "../auth-client"
