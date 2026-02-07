export function VersionLabel({ className }: { className?: string }) {
  const version = process.env.NEXT_PUBLIC_APP_VERSION
  if (!version) return null

  return <span className={`text-muted-foreground/50 text-xs ${className ?? ""}`}>v{version}</span>
}
