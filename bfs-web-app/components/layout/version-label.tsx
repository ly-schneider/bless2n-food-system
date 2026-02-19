export function VersionLabel({ className, version }: { className?: string; version?: string }) {
  if (!version) return null

  return <span className={`text-muted-foreground/50 text-xs ${className ?? ""}`}>v{version}</span>
}
