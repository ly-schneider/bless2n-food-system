const CACHE = "bfs-station-v3"
const SHELL = [
  "/station",
  "/station.webmanifest",
  "/icons/station-192.png",
  "/icons/station-512.png",
  "/apple-touch-icon.png",
]

const NETWORK_FIRST = new Set(["/station"])

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(CACHE)
      .then((c) => c.addAll(SHELL))
      .then(() => self.skipWaiting())
  )
})

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim())
  )
})

self.addEventListener("fetch", (event) => {
  if (event.request.method !== "GET") return
  const url = new URL(event.request.url)
  if (url.origin !== self.location.origin) return
  if (!SHELL.includes(url.pathname)) return

  if (NETWORK_FIRST.has(url.pathname)) {
    event.respondWith(
      fetch(event.request)
        .then((res) => {
          if (res.ok) {
            const copy = res.clone()
            caches.open(CACHE).then((c) => c.put(event.request, copy))
          }
          return res
        })
        .catch(() => caches.match(event.request))
    )
    return
  }

  event.respondWith(
    caches.match(event.request).then(
      (cached) =>
        cached ||
        fetch(event.request).then((res) => {
          if (res.ok) {
            const copy = res.clone()
            caches.open(CACHE).then((c) => c.put(event.request, copy))
          }
          return res
        })
    )
  )
})

self.addEventListener("message", (event) => {
  if (event.data === "hard-refresh") {
    caches.keys().then((keys) => Promise.all(keys.map((k) => caches.delete(k))))
    self.registration.unregister()
  }
})
