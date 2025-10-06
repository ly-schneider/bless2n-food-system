import { NextResponse } from 'next/server'
import { headers as nextHeaders, cookies } from 'next/headers'
import { API_BASE_URL } from '@/lib/api'

type ForwardPayload = {
  url: string
  method: string
  headers?: Record<string, string>
  body?: string | null
}

export async function POST(req: Request) {
  try {
    const hdrs = await nextHeaders()
    const cookieStore = await cookies()

    // CSRF double-submit validation: header must match cookie value
    const csrfHeader = hdrs.get('X-CSRF') || hdrs.get('x-csrf') || ''
    const csrfCookie = cookieStore.get('__Host-csrf')?.value || cookieStore.get('csrf')?.value || ''
    if (!csrfHeader || !csrfCookie || csrfHeader !== csrfCookie) {
      return NextResponse.json({ error: true, message: 'Forbidden' }, { status: 403 })
    }

    const data = (await req.json()) as ForwardPayload
    const { url, method, headers: fwdHeaders = {}, body } = data

    // Allow only forwarding to our backend base URL
    if (!url.startsWith(API_BASE_URL)) {
      return NextResponse.json({ error: true, message: 'Invalid forward target' }, { status: 400 })
    }

    // Prepare headers for backend request
    const outHeaders: Record<string, string> = { ...fwdHeaders }
    // Remove any existing CSRF headers to prevent duplication, then add the validated one
    Object.keys(outHeaders).forEach(key => {
      if (key.toLowerCase() === 'x-csrf') {
        delete outHeaders[key]
      }
    })
    outHeaders['X-CSRF'] = csrfHeader

    // Prepare cookies: only forward authentication-related cookies
    const rtCookie = cookieStore.get('__Host-rt')?.value || cookieStore.get('rt')?.value
    const cookiePairs: string[] = []
    // Include both variants to be robust across dev/https
    cookiePairs.push(`csrf=${encodeURIComponent(csrfCookie)}`)
    cookiePairs.push(`__Host-csrf=${encodeURIComponent(csrfCookie)}`)
    if (rtCookie) {
      cookiePairs.push(`rt=${encodeURIComponent(rtCookie)}`)
      cookiePairs.push(`__Host-rt=${encodeURIComponent(rtCookie)}`)
    }
    const cookieHeader = cookiePairs.join('; ')

    const res = await fetch(url, {
      method: method || 'POST',
      headers: {
        ...outHeaders,
        Cookie: cookieHeader,
      },
      body: body ?? undefined,
    })

    // For 204 No Content, return an empty body to avoid NextResponse errors
    if (res.status === 204) {
      return new NextResponse(null, { status: 204 })
    }

    const contentType = res.headers.get('content-type') || ''
    const payload = contentType.includes('application/json') ? await res.json().catch(() => ({})) : await res.text()
    return new NextResponse(
      typeof payload === 'string' ? payload : JSON.stringify(payload),
      { status: res.status, headers: { 'content-type': contentType || 'application/json' } }
    )
  } catch (e) {
    return NextResponse.json({ error: true, message: 'Proxy error' }, { status: 500 })
  }
}
