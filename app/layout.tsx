import "./globals.css";
import Image from "next/image";
import { fontVariables } from './fonts';

const defaultUrl = process.env.VERCEL_URL
  ? `https://${process.env.VERCEL_URL}`
  : "http://localhost:3000";

export const metadata = {
  metadataBase: new URL(defaultUrl),
  title: "Bestellungs System",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="de" className={fontVariables}>
      <body className="font-custom bg-background text-foreground overflow-hidden select-none">
        <main className="min-h-screen flex flex-col items-center">
          <div className="flex-1 w-full flex flex-col gap-6 items-center">
            <nav className="w-full flex justify-center border-b border-b-foreground/10 h-16">
              <div className="w-full flex justify-between items-center p-3 px-5 text-sm">
                <div className="flex gap-5 items-center font-semibold">
                  <Image src={"https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/Icf-logo.svg/250px-Icf-logo.svg.png"} alt="ICF Bern Image" width={80} height={0}></Image>
                </div>
                {/* <HeaderAuth /> */}
              </div>
            </nav>
            <div className="flex flex-col gap-20 w-full px-5 pb-5">{children}</div>
          </div>
        </main>
      </body>
    </html>
  );
}
