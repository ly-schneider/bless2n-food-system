"use client";

import Image from "next/image";
import { Button } from "./ui/button";
import { useState, useEffect } from "react";
import Link from "next/link";

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 10);
    };

    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <nav className={`w-full fixed top-0 left-0 z-50 transition-all duration-300 ${scrolled ? "bg-background shadow-sm py-3" : "py-5"}`}>
      <div className="container mx-auto px-4 md:px-6 flex items-center justify-between">
        <div className="flex items-center">
          <a href="/">
            <Image src={"/rentro-text-black.png"} width={100} height={50} alt="rentro Logo"></Image>
          </a>
        </div>

        <div className="hidden md:flex items-center space-x-8 font-medium">
          <a href="/#funktionen" className="hover:text-foreground/70">
            Funktionen
          </a>
          <a href="/#preise" className="hover:text-foreground/70">
            Preise
          </a>
          <a href="/#einsaetze" className="hover:text-foreground/70">
            Einsätze
          </a>
          <a href="/#kontakt" className="hover:text-foreground/70">
            Kontakt
          </a>
        </div>

        <div className="hidden md:flex items-center space-x-6">
          <Link href="/angebot">
            <Button className="bg-foreground text-background px-6 py-3 rounded-md">Jetzt Angebot holen</Button>
          </Link>
          <Link href={"/dashboard"}>Dashboard</Link>
        </div>
      </div>
    </nav>
  );
}
