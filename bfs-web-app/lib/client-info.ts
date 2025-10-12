// Client information detection using UA-CH with fallbacks
// Safe to import on server; only touches navigator inside functions.

export type ClientInfo = {
  os: string
  model: string
  browser: string
  browserVersion?: string
  arch?: string
  deviceModel?: string
}

type UADataBrand = { brand: string; version: string }

const hasNavigator = () => typeof navigator !== "undefined"

type NavigatorLike = Navigator & {
  brave?: { isBrave?: () => boolean | Promise<boolean> }
  // Not always present in older lib.dom typings
  userAgentData?: unknown
}

type UAData = {
  brands?: UADataBrand[]
  platform?: string
  getHighEntropyValues?: (
    hints: Array<"architecture" | "bitness" | "model" | "platformVersion" | "uaFullVersion">
  ) => Promise<{
    architecture?: string
    bitness?: string
    model?: string
    platformVersion?: string
    uaFullVersion?: string
  }>
}

function parseBrowserFromUA(ua: string, nav: NavigatorLike | undefined): { name: string; version?: string } {
  // Brave detection via API if present
  if (typeof nav?.brave?.isBrave === "function") {
    return { name: "Brave" }
  }
  if (/Edg\//.test(ua)) {
    const m = ua.match(/Edg\/(\S+)/)
    return { name: "Edge", version: m?.[1] }
  }
  if (/OPR\//.test(ua)) {
    const m = ua.match(/OPR\/(\S+)/)
    return { name: "Opera", version: m?.[1] }
  }
  if (/Firefox\//.test(ua)) {
    const m = ua.match(/Firefox\/(\S+)/)
    return { name: "Firefox", version: m?.[1] }
  }
  if (/Chrome\//.test(ua)) {
    const m = ua.match(/Chrome\/(\S+)/)
    return { name: "Chrome", version: m?.[1] }
  }
  if (/Safari\//.test(ua)) {
    const m = ua.match(/Version\/(\S+)/)
    return { name: "Safari", version: m?.[1] }
  }
  return { name: "Unknown" }
}

export async function getClientInfo(): Promise<ClientInfo> {
  const nav: NavigatorLike | undefined = hasNavigator() ? (navigator as NavigatorLike) : undefined
  const ua: string = nav?.userAgent || ""
  const uaData: UAData | undefined = nav?.userAgentData as UAData | undefined

  // Browser from brands or UA
  let browserName = ""
  let browserVersion: string | undefined
  if (uaData?.brands?.length) {
    const preferred = uaData.brands.find((b) =>
      /Brave|Google Chrome|Microsoft Edge|Opera|Firefox|Safari/i.test(b.brand)
    )
    const nonGeneric =
      preferred || uaData.brands.find((b) => !/Chromium|Not.?A.?Brand/i.test(b.brand)) || uaData.brands[0]
    browserName = (nonGeneric?.brand || "").replace(/\s+\d+$/g, "")
    browserVersion = nonGeneric?.version
    if (/Chrome/i.test(browserName) && typeof nav?.brave?.isBrave === "function") {
      browserName = "Brave"
    }
  } else {
    const parsed = parseBrowserFromUA(ua, nav)
    browserName = parsed.name
    browserVersion = parsed.version
    if (browserName === "Chrome" && typeof nav?.brave?.isBrave === "function") {
      browserName = "Brave"
    }
  }

  // OS/platform
  let os = uaData?.platform || ""
  let arch = ""
  let deviceModel = ""
  try {
    if (uaData?.getHighEntropyValues) {
      const hints = await uaData.getHighEntropyValues([
        "architecture",
        "bitness",
        "model",
        "platformVersion",
        "uaFullVersion",
      ])
      arch =
        hints?.architecture && hints?.bitness
          ? `${hints.architecture}${hints.bitness === "64" ? "64" : ""}`
          : hints?.architecture || ""
      deviceModel = hints?.model || ""
      if (!browserVersion && hints?.uaFullVersion) browserVersion = hints.uaFullVersion
      if (os === "macOS" && hints?.platformVersion) {
        os = `macOS ${hints.platformVersion}`
      } else if ((os === "Windows" || os === "Windows NT") && hints?.platformVersion) {
        os = `Windows ${hints.platformVersion}`
      }
    }
  } catch {
    // ignore
  }

  if (!os) {
    if (/Windows NT/.test(ua)) {
      const m = ua.match(/Windows NT ([0-9.]+)/)
      os = m ? `Windows ${m[1]}` : "Windows"
    } else if (/Mac OS X/.test(ua)) {
      const m = ua.match(/Mac OS X ([0-9_]+)/)
      const ver = m?.[1] || ""
      os = m ? (ver ? `macOS ${ver.replaceAll("_", ".")}` : "macOS") : "macOS"
    } else if (/iPhone|iPad|iPod/.test(ua)) {
      const m = ua.match(/OS ([0-9_]+) like Mac OS X/)
      const ver = m?.[1] || ""
      os = m ? (ver ? `iOS ${ver.replaceAll("_", ".")}` : "iOS") : "iOS"
    } else if (/Android/.test(ua)) {
      const m = ua.match(/Android ([0-9.]+)/)
      os = m ? `Android ${m[1]}` : "Android"
    } else if (/CrOS/.test(ua)) {
      os = "ChromeOS"
    } else if (/Linux/.test(ua)) {
      os = "Linux"
    } else {
      os = nav?.platform || "Unknown"
    }
  }

  // Build model string from browser + version and optionally arch/device model
  let model = browserName || "Browser"
  if (browserVersion) model += ` ${browserVersion}`
  const archSuffix = arch ? ` (${arch})` : ""
  model += deviceModel ? ` – ${deviceModel}${archSuffix}` : archSuffix

  return { os, model, browser: browserName, browserVersion, arch, deviceModel }
}
