import { Suspense } from "react"
import LoginClient from "./login-client"

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="container mx-auto px-4 pt-24 pb-10">Lade Anmeldeseite…</div>}>
      <LoginClient />
    </Suspense>
  )
}
