const CACHE_NAME = 'teachhub-v2';
const SHELL_ASSETS = [
  '/static/css/style.css',
  '/static/js/htmx.min.js',
  '/static/js/livekit-client.umd.js',
  '/static/js/pdf.min.js',
  '/static/js/pdf.worker.min.js',
  '/static/manifest.json'
];

// Install: cache shell assets
self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(SHELL_ASSETS))
      .then(() => self.skipWaiting())
  );
});

// Activate: clean old caches
self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.filter(k => k !== CACHE_NAME).map(k => caches.delete(k)))
    ).then(() => self.clients.claim())
  );
});

// Fetch: network-first for HTML, cache-first for static assets
self.addEventListener('fetch', event => {
  const url = new URL(event.request.url);

  // Static assets: cache-first
  if (url.pathname.startsWith('/static/')) {
    event.respondWith(
      caches.match(event.request).then(cached => {
        return cached || fetch(event.request).then(resp => {
          const clone = resp.clone();
          caches.open(CACHE_NAME).then(cache => cache.put(event.request, clone));
          return resp;
        });
      })
    );
    return;
  }

  // Everything else: network-first (server-rendered pages)
  event.respondWith(
    fetch(event.request).catch(() => caches.match(event.request))
  );
});
