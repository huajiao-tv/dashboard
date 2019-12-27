import request from '@/utils/request'

export function getQueues() {
  return request({
    url: '/api/queue/list',
    method: 'get'
  })
}

export function getQueueSystem() {
  return request({
    url: '/api/queue/listSystem',
    method: 'get'
  })
}

export function updateQueue(item) {
  return request({
    url: '/api/queue/update',
    method: 'post',
    data: item
  })
}

export function addQueue(item) {
  return request({
    url: '/api/queue/add',
    method: 'post',
    data: item
  })
}

export function delQueue(item) {
  return request({
    url: '/api/queue/delete',
    method: 'post',
    data: item
  })
}

export function exportQueue(ids) {
  return request({
    url: '/api/queue/export',
    method: 'post',
    data: ids
  })
}

export function importQueue(data) {
  return request({
    url: '/api/queue/import',
    method: 'post',
    data: data
  })
}
