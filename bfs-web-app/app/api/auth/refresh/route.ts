import { cookies, headers } from 'next/headers'
import { NextResponse } from 'next/server'
import { API_BASE_URL } from '@/lib/api'

export async function POST(_req: Request) {
  const cookieStore = await cookies()
  const hdrs = await headers()
  const proto = (hdrs.get('x-forwarded-proto') || '').toLowerCase()
  const secure = proto === 'https'
  const rtName = secure ? '__Host-rt' : 'rt'
  const csrfName = secure ? '__Host-csrf' : 'csrf'
  const rt = cookieStore.get(rtName)?.value

  if (!rt) {
    return NextResponse.json({ error: true, message: 'Unauthorized' }, { status: 401 })
  }

  // Call backend refresh - no CSRF required since backend generates new CSRF tokens
  const res = await fetch(`${API_BASE_URL}/v1/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Call': '1',
      // Forward refresh cookie for validation
      'Cookie': `${rtName}=${encodeURIComponent(rt)}`,
    },
  })

  if (!res.ok) {
    // Clear cookies on failure (reuse/invalid)
    cookieStore.set({ name: rtName, value: '', path: '/', httpOnly: true, secure, sameSite: 'lax', maxAge: -1 })
    cookieStore.set({ name: csrfName, value: '', path: '/', httpOnly: false, secure, sameSite: 'lax', maxAge: -1 })
    return NextResponse.json({ error: true, message: 'Unauthorized' }, { status: 401 })
  }

  const data = await res.json() as { access_token: string; expires_in: number; token_type: string; user: unknown; refresh_token?: string; csrf_token?: string }

  // Set cookies directly from response body (reliable for internal calls)
  if (data.refresh_token) {
    cookieStore.set({ name: rtName, value: data.refresh_token, httpOnly: true, secure, sameSite: 'lax', path: '/', maxAge: 7 * 24 * 60 * 60 })
  }
  const csrf = data.csrf_token || generateRandom(16)
  cookieStore.set({ name: csrfName, value: csrf, httpOnly: false, secure, sameSite: 'lax', path: '/', maxAge: 7 * 24 * 60 * 60 })

  return NextResponse.json({
    access_token: data.access_token,
    expires_in: data.expires_in,
    token_type: data.token_type,
    user: data.user,
  })
}

function generateRandom(n: number) {
  const bytes = crypto.getRandomValues(new Uint8Array(n))
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_' // URL-safe
  let out = ''
  for (let i = 0; i < bytes.length; i++) {
    const b = bytes[i] ?? 0
    out += chars[b % chars.length]
  }
  return out
}
