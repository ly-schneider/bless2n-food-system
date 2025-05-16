import { createRouteHandlerClient } from '@supabase/auth-helpers-nextjs'
import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'

export const runtime = 'edge'

export async function GET(request: Request) {
  const requestUrl = new URL(request.url)
  const code = requestUrl.searchParams.get('code')

  if (code) {
    const supabase = createRouteHandlerClient({ cookies })
    try {
      await supabase.auth.exchangeCodeForSession(code)
      console.log("Successfully exchanged code for session")
    } catch (error) {
      console.error("Error exchanging code for session:", error)
    }
  }

  // Redirect to the home page after authentication attempt
  return NextResponse.redirect(new URL('/', requestUrl.origin))
}
