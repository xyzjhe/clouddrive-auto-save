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
