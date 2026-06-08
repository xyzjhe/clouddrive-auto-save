import axios from 'axios'
import request from './request'

export function getVersion() {
  return request.get('/version')
}

// 独立 axios 实例，不走项目 request 拦截器（无 baseURL、无 API Key 注入）
const githubClient = axios.create({
  timeout: 8000
})

export function getLatestRelease() {
  return githubClient.get('https://api.github.com/repos/zhaocongqi/clouddrive-auto-save/releases/latest')
}
