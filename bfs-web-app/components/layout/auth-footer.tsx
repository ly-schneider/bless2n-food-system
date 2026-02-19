import Link from "next/link"
import { VersionLabel } from "@/components/layout/version-label"

const version = process.env.APP_VERSION

export default function AuthFooter() {
  return (
    <footer id="auth-footer" className={`text-muted-foreground mb-4 w-full border-t border-gray-200/70 py-4 text-sm`}>
      <div className="container mx-auto px-4">
        <nav className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
          <Link href="/agb" className="hover:underline">
            AGB
          </Link>
          <span className="text-gray-300">•</span>
          <Link href="/datenschutz" className="hover:underline">
            Datenschutzerklärung
          </Link>
          <span className="text-gray-300">•</span>
          <Link
            href="https://github.com/ly-schneider/bless2n-food-system"
            className="hover:underline"
            target="_blank"
            rel="noopener noreferrer"
          >
            GitHub
          </Link>
        </nav>
        <VersionLabel className="mt-2 block text-center" version={version} />
      </div>
    </footer>
  )
}
