import request from '@/utils/request'

export function getMachines(forCron) {
  return request({
    url: '/api/machine/list',
    method: 'post',
    data: {
      cron: forCron
    }
  })
}

export function addMachines(items, forCron) {
  items.cron = forCron
  return request({
    url: '/api/machine/add',
    method: 'post',
    data: items
  })
}

export function delMachine(items, forCron) {
  items.cron = forCron
  return request({
    url: '/api/machine/delete',
    method: 'post',
    data: items
  })
}
