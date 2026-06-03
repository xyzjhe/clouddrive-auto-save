/**
 * 统一的工具函数模块
 * 集中管理格式化、状态映射等共享逻辑
 */

/**
 * 格式化字节数为可读字符串
 * @param {number} bytes 字节数
 * @param {number} digits 小数位数，默认 2
 * @returns {string} 格式化后的字符串，如 "1.50 GB"
 */
export function formatSize(bytes, digits = 2) {
  const b = Number(bytes)
  if (isNaN(b) || b <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const k = 1024
  const i = Math.floor(Math.log(b) / Math.log(k))
  // 使用 parseFloat 去除尾零：2.00 → "2"，1.50 → "1.5"
  return `${parseFloat((b / Math.pow(k, i)).toFixed(digits))} ${units[i]}`
}

/**
 * 格式化时间为相对时间或绝对时间
 * @param {string} timeStr ISO 时间字符串
 * @param {string} fallback 默认文本
 * @returns {string} 格式化后的时间文本
 */
export function formatTime(timeStr, fallback = '从未执行') {
  if (!timeStr || timeStr.startsWith('0001')) return fallback
  const date = new Date(timeStr)
  if (isNaN(date.getTime())) return fallback

  const now = new Date()
  const diff = now - date
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (seconds < 60) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  if (days < 30) return `${days} 天前`

  return date.toLocaleDateString('zh-CN')
}

/**
 * 任务状态 → el-tag type 映射
 */
export const statusTagTypeMap = {
  pending: 'info',
  running: 'primary',
  success: 'success',
  failed: 'danger',
}

/**
 * 获取任务状态对应的 tag type
 * @param {string} status 任务状态
 * @returns {string} el-tag type
 */
export function getStatusTagType(status) {
  return statusTagTypeMap[status] || 'info'
}

/**
 * 任务状态 → 中文标签
 */
export const statusLabelMap = {
  pending: '等待中',
  running: '运行中',
  success: '成功',
  failed: '失败',
}

/**
 * 获取任务状态的中文标签
 * @param {string} status 任务状态
 * @returns {string} 中文标签
 */
export function getStatusLabel(status) {
  return statusLabelMap[status] || status
}
