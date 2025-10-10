import { Suspense } from "react"
import AcceptInviteClient from "./accept-invite-client"

export default function AcceptInvitePage() {
  return (
    <Suspense fallback={<div className="container mx-auto px-4 pt-24 pb-10">Prüfe Einladung…</div>}>
      <AcceptInviteClient />
    </Suspense>
  )
}
