import axios from 'axios'
import { ElMessage } from 'element-plus'

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10000,
  paramsSerializer: params => {
    const parts = []
    for (const key in params) {
      const value = params[key]
      if (Array.isArray(value)) {
        value.forEach(v => {
          parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(v)}`)
        })
      } else if (value !== undefined && value !== null) {
        parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
      }
    }
    return parts.join('&')
  }
})

// 请求拦截器
service.interceptors.request.use(
  config => {
    const apiKey = localStorage.getItem('ucas_api_key')
    if (apiKey) {
      config.headers['X-API-Key'] = apiKey
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  response => {
    const res = response.data
    return res
  },
  error => {
    // 调用方可通过 config.skipErrorHandler 自行处理错误（如链接校验场景）
    if (error.config?.skipErrorHandler) {
      return Promise.reject(error)
    }
    let msg
    if (error.response?.status === 401) {
      msg = '认证失败，请检查 API Key 配置'
    } else if (error.code === 'ECONNABORTED' || error.message?.includes('timeout')) {
      msg = '请求超时，请稍后重试'
    } else {
      msg = error.response?.data?.error || error.response?.data?.message || error.message || '请求失败'
    }
    ElMessage({
      message: msg,
      type: 'error',
      duration: 5 * 1000
    })
    return Promise.reject(error)
  }
)

export default service
