import { NextResponse } from "next/server"

export const dynamic = "force-dynamic" // ensure runtime evaluation

export function GET() {
  const pk = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || process.env.STRIPE_PUBLISHABLE_KEY
  if (!pk) {
    return NextResponse.json({ error: "missing publishable key" }, { status: 500 })
  }
  return NextResponse.json({ stripePk: pk })
}
