import request from '@/utils/request'

export function getSystems() {
  return request({
    url: '/api/system/list',
    method: 'get'
  })
}

export function addSystem(item) {
  return request({
    url: '/api/system/add',
    method: 'post',
    data: item
  })
}
