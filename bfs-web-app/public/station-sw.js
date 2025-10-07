// Minimal SW to cache the /station shell for offline viewing (no queuing)
const CACHE = 'bfs-station-v1'
const SHELL = ['/station']
self.addEventListener('install', (event) => {
  event.waitUntil(caches.open(CACHE).then((c) => c.addAll(SHELL)).then(() => self.skipWaiting()))
})
self.addEventListener('activate', (event) => {
  event.waitUntil(self.clients.claim())
})
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)
  if (url.pathname === '/station') {
    event.respondWith(
      caches.match(event.request).then((cached) => cached || fetch(event.request).then((res) => {
        const copy = res.clone()
        caches.open(CACHE).then((c) => c.put(event.request, copy))
        return res
      }))
    )
  }
})

