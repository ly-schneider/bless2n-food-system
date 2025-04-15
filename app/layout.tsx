import "./globals.css";
import Image from "next/image";
import { fontVariables } from './fonts';
import { Metadata } from 'next'
import { Toaster } from "sonner";
import AuthNavigation from "@/components/auth-navigation";
import { LockProvider, LockButton } from "@/components/lock-provider";
import Navigation from "@/components/navigation";

const defaultUrl = process.env.VERCEL_URL
  ? `https://${process.env.VERCEL_URL}`
  : "http://localhost:3000";

export const metadata: Metadata = {
  metadataBase: new URL(defaultUrl),
  title: "Bestellungs System",
  manifest: '/manifest.webmanifest', // Add manifest reference here
  appleWebApp: {
    capable: true,
    statusBarStyle: 'default',
    title: 'Bestellungs System'
  },
  applicationName: 'Bestellungs System',
  formatDetection: {
    telephone: false
  }
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="de" className={fontVariables}>
      <head>
        <link rel="apple-touch-icon" href="/icons/icon-192x192.png" />
      </head>
      <body className="font-custom bg-background text-foreground overflow-hidden select-none">
        <LockProvider>
          <main className="min-h-screen flex flex-col items-center">
            <div className="flex-1 w-full flex flex-col gap-6 items-center">
              <nav className="w-full flex justify-center border-b border-b-foreground/10 h-16">
                <div className="w-full flex justify-between items-center p-3 px-5 text-sm">
                  <div className="flex gap-5 items-center font-medium">
                    <Image src={"https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/Icf-logo.svg/250px-Icf-logo.svg.png"} alt="ICF Bern Image" width={80} height={0}></Image>
                  </div>
                  <div className="flex items-center gap-2">
                    <Navigation />
                  </div>
                </div>
              </nav>
              <div className="flex flex-col gap-20 w-full">{children}</div>
            </div>
          </main>
          <Toaster position="top-center" richColors />
        </LockProvider>
      </body>
    </html>
  );
}
