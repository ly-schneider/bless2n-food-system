import "./globals.css";
import { fontVariables } from "./fonts";
import { Metadata } from "next";
import ContactSection from "@/components/ContactSection";
import Footer from "@/components/Footer";
import Navbar from "@/components/Navbar";

const defaultUrl = process.env.VERCEL_URL
  ? `https://${process.env.VERCEL_URL}`
  : "http://localhost:3000";

export const metadata: Metadata = {
  metadataBase: new URL(defaultUrl),
  title: "rentro - POS Systeme zum mieten",
  description: "Mieten Sie POS Systeme für Ihren Live-Event",
  openGraph: {
    title: "rentro - POS Systeme zum mieten",
    description: "Mieten Sie POS Systeme für Ihren Live-Event",
    url: defaultUrl,
    siteName: "rentro",
    images: [
      {
        url: `${defaultUrl}/og-image.png`,
        width: 1200,
        height: 630,
        alt: "rentro - POS Systeme zum mieten",
      },
    ],
    locale: "de_DE",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "rentro - POS Systeme zum mieten",
    description: "Mieten Sie POS Systeme für Ihren Live-Event",
    images: [`${defaultUrl}/og-image.png`],
    creator: "@rentro",
  },
  icons: {
    icon: "/favicon.ico",
    shortcut: "/favicon.ico",
    apple: "/favicon.ico",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="de" className={fontVariables}>
      <body className="font-custom bg-background text-foreground overflow-x-hidden flex flex-col min-h-screen">
        <Navbar />
        {children}
        <ContactSection />
        <Footer />
      </body>
    </html>
  );
}
