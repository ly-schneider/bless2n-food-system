"use client"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { readErrorMessage } from "@/lib/http"
import { useEffect, useState } from "react"

export default function AcceptInvitePage() {
  const fetchAuth = useAuthorizedFetch()
  const [status, setStatus] = useState<"idle"|"ok"|"error">("idle")
  const [message, setMessage] = useState<string>("")

  useEffect(() => {
    const url = new URL(window.location.href)
    const token = url.searchParams.get('token')
    if (!token) { setStatus('error'); setMessage('Missing token'); return }
    ;(async () => {
      try {
        const res = await fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/invites/accept`, {
          method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ token })
        })
        if (!res.ok) throw new Error(await readErrorMessage(res))
        setStatus('ok'); setMessage('Invitation accepted. You can now log in as admin.')
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : 'Failed to accept invite'
        setStatus('error'); setMessage(msg)
      }
    })()
  }, [fetchAuth])

  return (
    <div className="max-w-md mx-auto px-4 py-10">
      <h1 className="text-2xl font-semibold mb-3">Admin Invite</h1>
      {status === 'idle' && <p className="text-gray-600">Processing…</p>}
      {status !== 'idle' && (
        <div className={status === 'ok' ? 'text-green-700' : 'text-red-700'}>{message}</div>
      )}
      <div className="mt-6">
        <a href="/login" className="underline">Go to login</a>
      </div>
    </div>
  )
}
