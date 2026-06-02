/**
 * SSE 连接共享模块
 * 统一管理 EventSource 连接的创建、消息解析和自动重连
 */

/**
 * 创建 SSE 连接
 * @param {string} url SSE 端点 URL
 * @param {Object} handlers 回调函数集合
 * @param {function(string)} handlers.onLog 接收到日志文本
 * @param {function(Object)} handlers.onTaskUpdate 接收到任务更新事件
 * @param {function(Object)} handlers.onStatsUpdate 接收到统计更新事件
 * @param {function(Object)} handlers.onTaskDelete 接收到任务删除事件
 * @param {function(Object)} handlers.onSearchValidate 接收到搜索验证结果事件
 * @returns {{ close: function() }} 包含 close 方法的控制对象
 */
export function createSSEConnection(url, handlers = {}) {
  let eventSource = null
  let retryTimer = null

  function connect() {
    if (eventSource) {
      eventSource.close()
    }

    eventSource = new EventSource(url)

    eventSource.onmessage = (event) => {
      const text = event.data
      if (!text) return

      // 检查是否为结构化事件
      const eventMatch = text.match(/\[EVENT:(\w+)\|(.*)\]/)
      if (eventMatch) {
        const eventType = eventMatch[1]
        try {
          const payload = JSON.parse(eventMatch[2])
          switch (eventType) {
            case 'task_update':
              handlers.onTaskUpdate?.(payload)
              break
            case 'task_delete':
              handlers.onTaskDelete?.(payload)
              break
            case 'stats_update':
              handlers.onStatsUpdate?.(payload)
              break
            case 'search_validate':
              handlers.onSearchValidate?.(payload)
              break
          }
        } catch {
          // 非 JSON 载荷，忽略
        }
        return
      }

      // 普通日志文本
      handlers.onLog?.(text)
    }

    eventSource.onerror = () => {
      eventSource.close()
      eventSource = null
      // 5 秒后自动重连
      retryTimer = setTimeout(connect, 5000)
    }
  }

  connect()

  return {
    close() {
      if (retryTimer) {
        clearTimeout(retryTimer)
        retryTimer = null
      }
      if (eventSource) {
        eventSource.close()
        eventSource = null
      }
    },
  }
}
