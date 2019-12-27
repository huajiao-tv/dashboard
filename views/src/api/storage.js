import request from '@/utils/request'

export function getStorages() {
  return request({
    url: '/api/storage/list',
    method: 'get'
  })
}

export function addStorage(item) {
  return request({
    url: '/api/storage/add',
    method: 'post',
    data: item
  })
}

export function updateStorage(item) {
  return request({
    url: '/api/storage/update',
    method: 'post',
    data: item
  })
}
