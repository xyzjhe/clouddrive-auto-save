// web/public/sw.js
const CACHE_NAME = 'ucas-v1'
const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/manifest.json'
]

// 安装事件 - 缓存静态资源
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS)
    })
  )
  self.skipWaiting()
})

// 激活事件 - 清理旧缓存
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      )
    })
  )
  self.clients.claim()
})

// 请求事件 - 网络优先策略
self.addEventListener('fetch', (event) => {
  // 仅拦截和缓存 GET 请求，避免非 GET 请求（如 POST、PUT、DELETE）写入缓存引发 TypeError
  if (event.request.method !== 'GET') {
    return
  }

  // 仅处理 http 和 https 协议，过滤 chrome-extension 和 websocket
  const url = new URL(event.request.url)
  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    return
  }

  // 跳过 API 请求
  if (url.pathname.includes('/api/')) {
    return
  }

  event.respondWith(
    fetch(event.request)
      .then((response) => {
        // 成功获取资源后，更新缓存
        const responseClone = response.clone()
        caches.open(CACHE_NAME).then((cache) => {
          cache.put(event.request, responseClone)
        })
        return response
      })
      .catch(() => {
        // 网络失败时，尝试从缓存获取
        return caches.match(event.request)
      })
  )
})
