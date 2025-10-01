"use client"
import { useRouter, useSearchParams } from 'next/navigation'
import { useState } from 'react'
import { useAuth } from '@/contexts/auth-context'
import type { User } from '@/types'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [step, setStep] = useState<'start' | 'code'>('start')
  const [otp, setOtp] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()
  const sp = useSearchParams()
  const next = sp.get('next') || '/'
  const { setAuth } = useAuth()

  const requestCode = async () => {
    setLoading(true); setError(null)
    try {
      await fetch('/api/auth/otp/request', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email }) })
      setStep('code')
    } catch {
      setError('Something went wrong. Please try again.')
    } finally { setLoading(false) }
  }

  const verifyCode = async () => {
    setLoading(true); setError(null)
    try {
      const res = await fetch('/api/auth/otp/verify', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, otp }) })
      if (!res.ok) throw new Error('Invalid code')
      const data = await res.json() as { access_token: string; expires_in: number; user: User; is_new?: boolean }
      setAuth(data.access_token, data.expires_in, data.user)
      if (data.is_new) {
        router.replace('/profile')
      } else {
        router.replace(next)
      }
    } catch {
      setError('Invalid code. Please try again.')
    } finally { setLoading(false) }
  }

  return (
    <div className="max-w-sm mx-auto py-16">
      <h1 className="text-2xl font-semibold mb-6">Sign in</h1>
      {step === 'start' && (
        <div className="space-y-4">
          <label className="block text-sm font-medium">Email</label>
          <input className="w-full border rounded px-3 py-2" type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="you@example.com" />
          <button className="w-full bg-black text-white rounded px-3 py-2 disabled:opacity-50" onClick={requestCode} disabled={loading || !email}>
            {loading ? 'Sending…' : 'Send code'}
          </button>
          <p className="text-xs text-gray-500">We’ll email you a one-time 6-digit code. Avoid shared devices.</p>
          {error && <p className="text-sm text-red-600">{error}</p>}
        </div>
      )}
      {step === 'code' && (
        <div className="space-y-4">
          <label className="block text-sm font-medium">Enter code</label>
          <input className="w-full border rounded px-3 py-2 tracking-widest text-center" inputMode="numeric" pattern="[0-9]*" maxLength={6} value={otp} onChange={e => setOtp(e.target.value.replace(/\D+/g, ''))} />
          <button className="w-full bg-black text-white rounded px-3 py-2 disabled:opacity-50" onClick={verifyCode} disabled={loading || otp.length < 6}>
            {loading ? 'Verifying…' : 'Verify'}
          </button>
          <p className="text-xs text-gray-500">Didn’t get it? Check spam or request again. We’ll never ask for your code.</p>
          {error && <p className="text-sm text-red-600">{error}</p>}
        </div>
      )}
    </div>
  )
}
