import request from '@/utils/request'
import {keywords} from "d3/dist/package";

export function getJobs() {
  return request({
    url: '/api/task/list',
    method: 'get'
  })
}

export function getJob(id) {
  return request({
    url: '/api/job/detail',
    method: 'get',
    params: {
      id: id
    }
  })
}

export function addJob(item) {
  return request({
    url: '/api/task/add',
    method: 'post',
    data: item
  })
}

export function getJobNode(id) {
  return request({
    url: '/api/task/nodes',
    method: 'get',
    params: {
      id: id
    }
  })
}
export function updateJob(item) {
  return request({
    url: '/api/task/update',
    method: 'post',
    data: item
  })
}

export function jobStart(id) {
  return request({
    url: '/api/task/start',
    method: 'post',
    data: id
  })
}

export function jobPause(id) {
  return request({
    url: '/api/task/pause',
    method: 'post',
    data: id
  })
}

export function setJobState(item) {
  return request({
    url: '/api/job/update',
    method: 'post',
    data: item
  })
}

export function remvoeJob(id) {
  return request({
    url: '/api/task/remove',
    method: 'post',
    data: id
  })
}

export function depsTree(id) {
  return request({
    url: '/api/task/tree',
    method: 'get',
    params: {
      id: id
    }
  })
}

export function nodeList() {
  return request({
    url: '/api/task/nodeList',
    method: 'get'
  })
}

export function getNodeLog(id, node, page,keyword) {
  return request({
    url: '/api/task/nodeLog',
    method: 'get',
    params: {
      id: id,
      node: node,
      page: page,
      search:keyword,
    }
  })
}

export function getNodeLogTotal(id, node,keyword) {
  return request({
    url: '/api/task/nodeLogTotal',
    method: 'get',
    params: {
      id: id,
      node: node,
      searchWord: keyword,
    }
  })
}

export function getTask(id) {
  return request({
    url: '/api/task/task',
    method: 'get',
    params: {
      id: id
    }
  })
}
export function taskTest(id) {
  return request({
    url: '/api/task/test',
    method: 'get',
    params: {
      id: id
    }
  })
}
export function taskTestLog(id) {
  return request({
    url: '/api/task/testLog',
    method: 'get',
    params: {
      id: id
    }
  })
}

