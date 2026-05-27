import { Metadata, Viewport } from "next"
import { VersionLabel } from "@/components/layout/version-label"
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
      <div className="flex flex-1 flex-col">{children}</div>
      <VersionLabel className="pb-6 text-center text-sm" version={process.env.APP_VERSION} />
      <ServiceWorkerRegister />
      <PWARuntime />
    </main>
  )
}
