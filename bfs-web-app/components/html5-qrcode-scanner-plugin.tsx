"use client"

import {
  Html5Qrcode,
  type Html5QrcodeCameraScanConfig,
  Html5QrcodeScannerState,
  type QrcodeErrorCallback,
  type QrcodeSuccessCallback,
} from "html5-qrcode"
import { useEffect, useId, useMemo } from "react"

type Html5QrcodeScannerPluginProps = {
  cameraId?: string
  fps?: number
  qrbox?: Html5QrcodeCameraScanConfig["qrbox"]
  aspectRatio?: number
  disableFlip?: boolean
  verbose?: boolean
  qrCodeSuccessCallback: QrcodeSuccessCallback
  qrCodeErrorCallback?: QrcodeErrorCallback
  onReady?: (scanner: Html5Qrcode) => void
  onStartError?: (error: unknown) => void
  onStop?: () => void
}

const createConfig = (props: Html5QrcodeScannerPluginProps): Html5QrcodeCameraScanConfig => {
  const config: Html5QrcodeCameraScanConfig = {
    fps: props.fps ?? 10,
    qrbox: props.qrbox ?? 250,
  }
  if (props.qrbox) config.qrbox = props.qrbox
  if (props.aspectRatio) config.aspectRatio = props.aspectRatio
  if (props.disableFlip !== undefined) config.disableFlip = props.disableFlip
  return config
}

const Html5QrcodeScannerPlugin = (props: Html5QrcodeScannerPluginProps) => {
  const reactId = useId()
  const regionId = useMemo(() => `html5qr-code-full-region-${reactId.replace(/[:]/g, "")}`, [reactId])

  useEffect(() => {
    const region = document.getElementById(regionId)
    if (region) {
      region.classList.add("relative", "w-full", "aspect-square", "bg-black")
    }
    const styleEl = document.createElement("style")
    styleEl.setAttribute("data-html5qr-style", regionId)
    styleEl.innerHTML = `
#${regionId} { position: relative; display: block; width: 100% !important; max-width: 28rem; aspect-ratio: 1 / 1; margin: 0 auto; }
@media (max-width: 640px) { #${regionId} { max-width: 100vw !important; } }
#${regionId} video { width: 100% !important; height: 100% !important; object-fit: cover; display: block !important; }
#${regionId} video:nth-of-type(n+2) { display: none !important; }
`
    document.head.appendChild(styleEl)
    return () => {
      styleEl.remove()
    }
  }, [regionId])

  useEffect(() => {
    const config = createConfig(props)
    const startConfig = props.cameraId ? { deviceId: { exact: props.cameraId } } : { facingMode: "environment" }
    const html5Qrcode = new Html5Qrcode(regionId, props.verbose === true)
    let cancelled = false

    html5Qrcode
      .start(startConfig, config, props.qrCodeSuccessCallback, props.qrCodeErrorCallback)
      .then(() => {
        if (cancelled) return
        props.onReady?.(html5Qrcode)
      })
      .catch((error) => {
        if (cancelled) return
        console.error("Failed to start html5-qrcode", error)
        props.onStartError?.(error)
      })

    return () => {
      cancelled = true
      props.onStop?.()
      const state = (() => {
        try {
          return html5Qrcode.getState()
        } catch {
          return undefined
        }
      })()
      const isActive =
        html5Qrcode.isScanning || state === Html5QrcodeScannerState.SCANNING || state === Html5QrcodeScannerState.PAUSED
      const stopPromise = isActive ? html5Qrcode.stop().catch(() => {}) : Promise.resolve()
      stopPromise.finally(() => {
        try {
          html5Qrcode.clear()
        } catch {}
      })
    }
  }, [
    props.aspectRatio,
    props.cameraId,
    props.disableFlip,
    props.fps,
    props.onReady,
    props.onStartError,
    props.onStop,
    props.qrCodeErrorCallback,
    props.qrCodeSuccessCallback,
    props.qrbox,
    props.verbose,
    regionId,
  ])

  return <div id={regionId} className="w-full overflow-hidden rounded-2xl" />
}

export default Html5QrcodeScannerPlugin
