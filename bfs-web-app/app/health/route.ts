import { NextResponse } from "next/server"

export async function GET() {
  const payload = {
    status: "healthy",
    timestamp: new Date().toISOString(),
  }
  return NextResponse.json(payload, { status: 200 })
}
