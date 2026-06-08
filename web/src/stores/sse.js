/**
 * SSE Pinia Store
 * 统一管理 EventSource 连接，支持引用计数、自动重连、事件回调
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useSSEStore = defineStore('sse', () => {
  // 日志列表：[{ text, time }]
  const logs = ref([])
  // 连接状态
  const connected = ref(false)

  // 内部状态（非响应式，不需要暴露）
  let eventSource = null
  let refCount = 0
  let retryTimer = null
  let retryDelay = 5000
  const maxRetryDelay = 60000

  // 事件回调注册表：{ eventType: Set<callback> }
  const listeners = {
    task_update: new Set(),
    task_delete: new Set(),
    stats_update: new Set(),
    search_validate: new Set(),
    progress: new Set()
  }

  /**
   * 获取 SSE URL（携带 token 参数）
   */
  function getSSEUrl() {
    const token = localStorage.getItem('ucas_api_key') || ''
    const base = '/api/dashboard/logs'
    return token ? `${base}?token=${encodeURIComponent(token)}` : base
  }

  /**
   * 建立或复用 EventSource 连接
   */
  function connect() {
    if (eventSource) return

    const url = getSSEUrl()
    eventSource = new EventSource(url)

    eventSource.onopen = () => {
      connected.value = true
      // 连接成功后重置退避延迟
      retryDelay = 5000
    }

    eventSource.onmessage = (event) => {
      const msg = event.data
      if (!msg) return

      // 解析 [PROGRESS:...] 消息
      if (msg.includes('[PROGRESS:')) {
        const match = msg.match(/\[PROGRESS:(.+)\]/)
        if (match) {
          listeners.progress.forEach(cb => {
            try { cb(match[1]) } catch (e) { console.error('进度回调异常:', e) }
          })
        }
        return
      }

      // 解析 [EVENT:...] 消息
      // 支持两种后端格式：
      //   管道符格式：[EVENT:search_validate|{...}]（BroadcastSearchValidate）
      //   纯 JSON 格式：[EVENT:{"type":"task_update","task":{...}}]（BroadcastTaskUpdate 等）
      if (msg.includes('[EVENT:')) {
        // 优先尝试管道符格式
        const pipeMatch = msg.match(/\[EVENT:(.+?)\|(.*)\]/)
        if (pipeMatch) {
          const eventType = pipeMatch[1]
          try {
            const payload = JSON.parse(pipeMatch[2])
            if (listeners[eventType]) {
              listeners[eventType].forEach(cb => {
                try { cb(payload) } catch (e) { console.error(`事件回调异常(${eventType}):`, e) }
              })
            }
          } catch (e) {
            console.error('解析事件失败:', e)
          }
        } else {
          // 回退到纯 JSON 格式
          const jsonMatch = msg.match(/\[EVENT:(.+)\]/)
          if (jsonMatch) {
            try {
              const ev = JSON.parse(jsonMatch[1])
              if (ev.type && listeners[ev.type]) {
                listeners[ev.type].forEach(cb => {
                  try { cb(ev) } catch (e) { console.error(`事件回调异常(${ev.type}):`, e) }
                })
              }
            } catch (e) {
              console.error('解析事件失败:', e)
            }
          }
        }
        return
      }

      // 普通日志
      logs.value.push({ text: msg, time: new Date().toLocaleTimeString() })
      if (logs.value.length > 200) logs.value.shift()
    }

    eventSource.onerror = () => {
      connected.value = false
      eventSource.close()
      eventSource = null
      // 指数退避重连
      scheduleReconnect()
    }
  }

  /**
   * 关闭 EventSource 连接
   */
  function disconnect() {
    if (retryTimer) {
      clearTimeout(retryTimer)
      retryTimer = null
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    connected.value = false
    retryDelay = 5000
  }

  /**
   * 指数退避重连调度
   */
  function scheduleReconnect() {
    if (refCount <= 0) return

    retryTimer = setTimeout(() => {
      retryTimer = null
      // 重连前再次检查引用计数，防止 unsubscribe 后仍触发
      if (refCount > 0) {
        connect()
      }
    }, retryDelay)

    // 指数退避：5s → 10s → 20s → 40s → 60s（上限）
    retryDelay = Math.min(retryDelay * 2, maxRetryDelay)
  }

  /**
   * 订阅 SSE 连接（引用计数 +1）
   * @returns {Function} 取消订阅函数（引用计数 -1）
   */
  function subscribe() {
    refCount++
    if (refCount === 1) {
      connect()
    }
    // 返回取消订阅函数
    return () => {
      refCount = Math.max(0, refCount - 1)
      if (refCount === 0) {
        disconnect()
      }
    }
  }

  /**
   * 注册事件回调
   * @param {string} eventType 事件类型：task_update | task_delete | stats_update | search_validate | progress
   * @param {Function} callback 回调函数
   * @returns {Function} 取消注册函数
   */
  function on(eventType, callback) {
    if (!listeners[eventType]) {
      console.warn(`未知的事件类型: ${eventType}`)
      return () => {}
    }
    listeners[eventType].add(callback)
    return () => {
      listeners[eventType].delete(callback)
    }
  }

  /**
   * 清空日志
   */
  function clearLogs() {
    logs.value = []
  }

  /**
   * 添加历史日志（time 为空串），追加到现有日志之后
   * @param {string[]} historyLogs 历史日志文本数组
   */
  function addHistoryLogs(historyLogs) {
    const items = historyLogs.map(text => ({ text, time: '' }))
    logs.value = [...items, ...logs.value]
  }

  return {
    logs,
    connected,
    subscribe,
    on,
    clearLogs,
    addHistoryLogs
  }
})
