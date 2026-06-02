import request from './request'

export function searchResources(params) {
  return request({
    url: '/search',
    method: 'get',
    params
  })
}

export function listSearchSources() {
  return request({
    url: '/search/sources',
    method: 'get'
  })
}

export function getSearchConfig() {
  return request({
    url: '/search/config',
    method: 'get'
  })
}

export function updateSearchConfig(data) {
  return request({
    url: '/search/config',
    method: 'put',
    data
  })
}

export function validateLink(url, timeoutMs = 5000) {
  return request({
    url: '/search/validate',
    method: 'get',
    params: { url },
    timeout: timeoutMs,
    // 链接校验的错误是预期内的（无效链接），不弹全局 toast
    skipErrorHandler: true
  })
}
