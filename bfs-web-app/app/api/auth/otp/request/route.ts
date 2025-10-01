import { headers } from 'next/headers'
import { NextResponse } from 'next/server'
import { API_BASE_URL } from '@/lib/api'

export async function POST(req: Request) {
  try {
    const { email } = (await req.json()) as { email: string }
    const hdrs = await headers()
    const ua = hdrs.get('user-agent') || ''
    const xff = hdrs.get('x-forwarded-for') || ''
    // Call backend (always return 202/generic)
    await fetch(`${API_BASE_URL}/v1/auth/otp/request`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // forward browser context so backend can label device + ip nicely
        'X-Forwarded-User-Agent': ua,
        ...(xff ? { 'X-Forwarded-For': xff } : {}),
      },
      body: JSON.stringify({ email }),
    })
  } catch {}
  return NextResponse.json({ message: "If the email exists, you'll receive a code." }, { status: 202 })
}
