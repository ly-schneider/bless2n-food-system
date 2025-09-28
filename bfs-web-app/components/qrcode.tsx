"use client"

import { toDataURL } from "qrcode"
import { useEffect, useMemo, useState } from "react"

type Props = { value: string; size?: number; className?: string }

export default function QRCode({ value, size = 220, className }: Props) {
  const [src, setSrc] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const url = await toDataURL(value, {
          width: size,
          margin: 1,
          color: {
            dark: "#000000",
            light: "#E9E7E6",
          },
        })
        if (!cancelled) setSrc(url)
      } catch (err) {
        console.error("Failed generating QR code", err)
        if (!cancelled) setSrc(null)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [value, size])

  if (!src) return null
  return <img src={src} alt="QR Code" width={size} height={size} className={className} />
}
