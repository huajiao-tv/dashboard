import request from '@/utils/request'

export function getTopic(system, queue) {
  var uri = '/api/topic/system/summary'
  if (system === undefined) {
    uri = '/api/topic/queue/summary'
  }
  return request({
    url: uri,
    method: 'get',
    params: {
      system,
      queue
    }
  })
}

export function getRetryErrors(queue, topic, count) {
  return request({
    url: '/api/topic/get/retry/errors',
    method: 'get',
    params: {
      queue,
      topic,
      count
    }
  })
}

export function searchTopic(system, queue, topic, keyword) {
  return request({
    url: '/api/topic/search',
    method: 'post',
    data: {
      system,
      queue,
      topic,
      keyword
    }
  })
}

export function addTopic(item) {
  return request({
    url: '/api/topic/add',
    method: 'post',
    data: item
  })
}

export function delTopic(item) {
  return request({
    url: '/api/topic/delete',
    method: 'post',
    data: item
  })
}

export function sendTopic(item) {
  return request({
    url: '/api/topic/send',
    method: 'post',
    data: item
  })
}

export function refreshQueue(queue, topic, machines) {
  return request({
    url: '/api/topic/get/length',
    method: 'post',
    data: {
      queue,
      topic,
      machines
    }
  })
}

export function recoverRetryQueue(queue, topic, count) {
  return request({
    url: '/api/topic/recover/retry',
    method: 'post',
    data: {
      queue,
      topic,
      count
    }
  })
}

export function cleanRetryQueue(queue, topic) {
  return request({
    url: '/api/topic/clean/retry',
    method: 'post',
    data: {
      queue,
      topic
    }
  })
}
export function cleanStatistics(queue, topic) {
  return request({
    url: '/api/topic/clean/statistics',
    method: 'post',
    data: {
      queue,
      topic
    }
  })
}

export function updateTopic(item) {
  return request({
    url: '/api/topic/update',
    method: 'post',
    data: item
  })
}

export function getHistory(params) {
  return request({
    url: '/api/topic/history',
    method: 'post',
    data: params
  })
}

export function defaultThreshold() {
  return request({
    url: '/api/topic/default/threshold',
    method: 'post'
  })
}
