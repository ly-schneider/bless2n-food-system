/**
 * Returns the CSRF token from the browser cookie following the double-submit pattern.
 * Uses `__Host-csrf` on HTTPS and `csrf` on HTTP.
 */
export function getCSRFToken(): string | null {
  if (typeof document === "undefined") return null
  const name = (document.location.protocol === "https:" ? "__Host-" : "") + "csrf"
  const m = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/([.$?*|{}()\[\]\\/+^])/g, "\\$1") + "=([^;]*)"))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}
