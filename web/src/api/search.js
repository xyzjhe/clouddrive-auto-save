import request from './request'

export function searchResources(params) {
  return request({
    url: '/search',
    method: 'get',
    params,
    timeout: 30000 // 搜索需要调用多个源，超时设为 30s
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

export function validateBatch(data) {
  return request({
    url: '/search/validate_batch',
    method: 'post',
    data,
    timeout: 35000
  })
}
