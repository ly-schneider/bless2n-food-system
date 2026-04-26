import { Metadata, Viewport } from "next"
import Image from "next/image"
import { VersionLabel } from "@/components/layout/version-label"
import { FullscreenButton } from "@/components/station/fullscreen-button"
import { PWARuntime } from "@/components/station/pwa-runtime"
import { ServiceWorkerRegister } from "@/components/station/service-worker-register"

export const metadata: Metadata = {
  title: "Station Portal",
  description: "Stationsportal für das BlessThun Food Bestellsystem.",
  manifest: "/station.webmanifest",
  applicationName: "BlessThun Station",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
    title: "Station",
  },
  icons: {
    icon: [
      { url: "/icons/station-192.png", sizes: "192x192", type: "image/png" },
      { url: "/icons/station-512.png", sizes: "512x512", type: "image/png" },
    ],
    apple: [{ url: "/apple-touch-icon.png", sizes: "180x180", type: "image/png" }],
  },
}

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
  viewportFit: "cover",
  themeColor: "#000000",
}

export default function StationLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="flex min-h-screen flex-col px-2">
      <div className="relative flex items-center justify-center pt-6">
        <div className="inline-flex items-center gap-3">
          <Image src="/assets/images/blessthun.png" alt="BlessThun Logo" width={40} height={40} className="h-10 w-10" />
          <span className="text-lg font-semibold tracking-tight">BlessThun Food</span>
        </div>
        <div className="absolute top-6 right-2">
          <FullscreenButton />
        </div>
      </div>
      <div className="flex flex-1 flex-col">{children}</div>
      <VersionLabel className="pb-6 text-center text-sm" version={process.env.APP_VERSION} />
      <ServiceWorkerRegister />
      <PWARuntime />
    </main>
  )
}
