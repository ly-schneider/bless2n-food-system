"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import { API_BASE_URL } from "@/lib/api"
import Image from "next/image"

type StationStatus = { exists: boolean; approved: boolean; name?: string }

function randKey(): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
  const bytes = new Uint8Array(24)
  if (typeof crypto !== "undefined" && crypto.getRandomValues) crypto.getRandomValues(bytes)
  return Array.from(bytes, (b) => chars[b % chars.length]).join("")
}

export default function StationPage() {
  const videoRef = useRef<HTMLVideoElement | null>(null)
  const [devices, setDevices] = useState<MediaDeviceInfo[]>([])
  const [deviceId, setDeviceId] = useState<string>("")
  const [streaming, setStreaming] = useState(false)
  const [status, setStatus] = useState<StationStatus | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  type PublicOrderItem = { id: string; orderId: string; productId: string; title: string; quantity: number; pricePerUnitCents?: number; productImage?: string | null; isRedeemed?: boolean; parentItemId?: string | null; menuSlotId?: string | null; menuSlotName?: string | null }
  type VerifyResult = { orderId: string; items: PublicOrderItem[] }
  type RedeemResult = { orderId: string; stationId: string; matched: number; redeemed: number; items: PublicOrderItem[]; redeemedAt: string }
  const [result, setResult] = useState<VerifyResult | null>(null)
  const [scanned, setScanned] = useState<string | null>(null)
  const [redeemResp, setRedeemResp] = useState<RedeemResult | null>(null)
  const [manual, setManual] = useState("")
  const [offline, setOffline] = useState(false)

  const stationKey = useMemo(() => {
    if (typeof window === "undefined") return ""
    let k = localStorage.getItem("bfs.stationKey")
    if (!k) { k = `st_${randKey()}`; localStorage.setItem("bfs.stationKey", k) }
    return k
  }, [])

  useEffect(() => {
    setOffline(!navigator.onLine)
    const on = () => setOffline(false)
    const off = () => setOffline(true)
    window.addEventListener("online", on)
    window.addEventListener("offline", off)
    return () => { window.removeEventListener("online", on); window.removeEventListener("offline", off) }
  }, [])

  useEffect(() => {
    // Register a simple SW to cache this shell
    if (typeof window !== "undefined" && "serviceWorker" in navigator) {
      navigator.serviceWorker.register("/station-sw.js").catch(() => {})
    }
  }, [])

  useEffect(() => {
    // fetch station status
    ;(async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/v1/stations/me`, { headers: { "X-Station-Key": stationKey } })
        const json = await res.json()
        setStatus(json as StationStatus)
      } catch {
        setStatus({ exists: false, approved: false })
      }
    })()
  }, [stationKey])

  useEffect(() => {
    // list cameras
    (async () => {
      try {
        const ds = await navigator.mediaDevices.enumerateDevices()
        const vids = ds.filter((d) => d.kind === "videoinput")
        setDevices(vids)
        if (vids.length && !deviceId) { const first = vids[0]; if (first) setDeviceId(first.deviceId) }
      } catch {
        // no permission yet or insecure context
      }
    })()
  }, [deviceId])

  async function startCamera() {
    if (!deviceId) return
    setError(null)
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { deviceId: { exact: deviceId } } })
      const v = videoRef.current
      if (v) { v.srcObject = stream; await v.play(); setStreaming(true) }
    } catch {
      setError("Camera access denied or unavailable. You can use manual entry.")
    }
  }

  async function stopCamera() {
    const v = videoRef.current
    if (v?.srcObject) {
      const tracks = (v.srcObject as MediaStream).getTracks()
      tracks.forEach((t) => t.stop())
      v.srcObject = null
    }
    setStreaming(false)
  }

  async function scanOnce() {
    setError(null)
    setBusy(true)
    try {
      const code = manual.trim() || await decodeOnce(videoRef.current)
      if (!code) { setError("No code detected"); return }
      const verify = await fetch(`${API_BASE_URL}/v1/stations/verify-qr`, { method: "POST", headers: { "Content-Type": "application/json", "X-Station-Key": stationKey }, body: JSON.stringify({ code }) })
      type Problem = { detail?: string }
      if (!verify.ok) { const j = (await verify.json().catch(()=>({}))) as Problem; throw new Error(j.detail || `Error ${verify.status}`) }
      const data = (await verify.json()) as VerifyResult
      setResult(data)
      setScanned(code)
      setRedeemResp(null)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Scan failed")
    } finally {
      setBusy(false)
    }
  }

  async function redeem() {
    if (!result) return
    if (offline) return
    setBusy(true)
    setError(null)
    try {
      const idem = `idem_${Date.now()}_${Math.random().toString(36).slice(2)}`
      const res = await fetch(`${API_BASE_URL}/v1/stations/redeem`, { method: "POST", headers: { "Content-Type": "application/json", "X-Station-Key": stationKey, "Idempotency-Key": idem }, body: JSON.stringify({ code: scanned }) })
      const json = (await res.json().catch(()=>({}))) as { detail?: string; message?: string } | RedeemResult
      if (!res.ok) throw new Error((json as { detail?: string; message?: string }).detail || (json as { detail?: string; message?: string }).message || `Error ${res.status}`)
      setRedeemResp(json as RedeemResult)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Redeem failed")
    } finally { setBusy(false) }
  }

  async function requestVerification() {
    const deviceLabel = (devices[0]?.label || navigator.userAgent).slice(0, 80)
    const os = navigator.platform || "web"
    const name = prompt("Station label (e.g., Grill #1)") || deviceLabel || "Station"
    if (!name) return
    setBusy(true)
    try {
      await fetch(`${API_BASE_URL}/v1/stations/requests`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ name, model: deviceLabel, os, deviceKey: stationKey }) })
      // refresh status
      const r = await fetch(`${API_BASE_URL}/v1/stations/me`, { headers: { "X-Station-Key": stationKey } })
      const js = await r.json(); setStatus(js as StationStatus)
    } catch { setError("Request failed") } finally { setBusy(false) }
  }

  return (
    <div className="min-h-screen bg-background text-foreground p-6 flex flex-col items-center gap-4">
      <h1 className="text-3xl font-primary">Station</h1>
      {offline && (
        <div role="status" aria-live="polite" className="text-destructive bg-destructive/10 px-3 py-2 rounded">
          Offline. Scanning okay; redemption disabled.
        </div>
      )}
      {status && !status.approved && (
        <div className="w-full max-w-xl border border-border bg-card rounded p-4 flex flex-col gap-3">
          <div className="font-semibold">This device is not an approved station.</div>
          <button onClick={requestVerification} className="bg-primary text-primary-foreground rounded px-4 py-3 disabled:opacity-50" disabled={busy}>Request Station Verification</button>
          <div className="text-sm text-muted-foreground">After approval, this screen will unlock redemption.</div>
        </div>
      )}

      <div className="w-full max-w-xl border border-border bg-card rounded p-4 flex flex-col gap-4">
        <label className="text-sm">Camera</label>
        <select className="border border-border rounded px-3 py-2" value={deviceId} onChange={(e) => setDeviceId(e.target.value)}>
          {devices.map((d) => (<option key={d.deviceId} value={d.deviceId}>{d.label || `Camera ${d.deviceId.slice(0,6)}`}</option>))}
        </select>
        <video ref={videoRef} className="w-full bg-muted rounded" muted playsInline />
        <div className="flex gap-2">
          {!streaming ? (
            <button onClick={startCamera} className="bg-primary text-primary-foreground rounded px-4 py-3">Start Camera</button>
          ) : (
            <button onClick={stopCamera} className="bg-secondary text-secondary-foreground rounded px-4 py-3">Stop Camera</button>
          )}
          <button onClick={scanOnce} className="bg-selected text-sidebar-primary-foreground rounded px-4 py-3 disabled:opacity-50" disabled={busy}>Scan</button>
        </div>
        <div className="flex gap-2 items-center">
          <input className="flex-1 border border-border rounded px-3 py-2" placeholder="Manual QR payload" value={manual} onChange={(e) => setManual(e.target.value)} />
          <button onClick={scanOnce} className="bg-secondary text-secondary-foreground rounded px-4 py-2 disabled:opacity-50" disabled={busy}>Submit</button>
        </div>
        {error && <div role="alert" aria-live="assertive" className="text-destructive">{error}</div>}
      </div>

      {result && (
        <div className="w-full max-w-xl border border-border bg-card rounded p-4">
          <div className="font-semibold mb-3">Zu verteilende Artikel</div>
          <div className="flex flex-col gap-3">
            {result.items?.map((it) => (
              <div key={it.id} className="rounded-xl border p-3">
                <div className="flex items-center gap-3">
                  {it.productImage ? (
                    <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
                      <Image src={it.productImage} alt={it.title} fill sizes="64px" className="h-full w-full object-cover" />
                    </div>
                  ) : null}
                  <div className="min-w-0 flex-1">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="truncate font-medium">{it.title}</p>
                        {it.menuSlotName && (
                          <div className="mt-1 text-xs text-muted-foreground">{it.menuSlotName}: {it.title}</div>
                        )}
                      </div>
                      <div className="shrink-0 text-right">
                        <p className="text-sm text-muted-foreground">x{it.quantity}</p>
                        <span className={`inline-block mt-1 text-xs rounded px-2 py-1 ${it.isRedeemed ? 'bg-muted text-muted-foreground' : 'bg-primary text-primary-foreground'}`}>{it.isRedeemed ? 'Redeemed' : 'Unredeemed'}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
          <div className="mt-4 flex gap-2">
            <button className="bg-primary text-primary-foreground rounded px-4 py-3 disabled:opacity-50" disabled={busy || offline || !(status?.approved)} onClick={redeem}>Redeem</button>
            <button className="bg-secondary text-secondary-foreground rounded px-4 py-3" onClick={() => { setResult(null); setRedeemResp(null); setManual(""); setScanned(null) }}>Scan next</button>
          </div>
        </div>
      )}

      {redeemResp && (
        <div className="w-full max-w-xl border border-border bg-card rounded p-4">
          <div className="font-semibold mb-2">Receipt</div>
          <div className="text-sm">Redeemed: {String(redeemResp.redeemed)} / Matched: {String(redeemResp.matched)}</div>
          <div className="mt-2">
            <ul className="space-y-1">
              {redeemResp.items?.map((it: PublicOrderItem) => (
                <li key={it.id} className="flex justify-between text-sm">
                  <span>{it.title}</span>
                  <span className={it.isRedeemed ? "text-foreground" : "text-muted-foreground"}>{it.isRedeemed ? "Redeemed" : "Remaining"}</span>
                </li>
              ))}
            </ul>
          </div>
          <div className="mt-3">
            <button className="bg-secondary text-secondary-foreground rounded px-4 py-3" onClick={() => { setResult(null); setRedeemResp(null); setScanned(null) }}>Scan next</button>
          </div>
        </div>
      )}
    </div>
  )
}

async function decodeOnce(video: HTMLVideoElement | null): Promise<string | null> {
  try {
    const { BrowserQRCodeReader } = await import("@zxing/browser")
    const reader = new BrowserQRCodeReader()
    const res = await reader.decodeOnceFromVideoDevice(undefined, video!)
    return res.getText()
  } catch {
    return null
  }
}
