import request from '@/utils/request'

export function login(username, password) {
  return request({
    url: '/api/auth/login',
    method: 'post',
    data: {
      username,
      password
    }
  })
}

export function getInfo(token) {
  return request({
    url: '/api/auth/info'
  })
}

export function userlist(search) {
  return request({
    url: '/api/auth/list?keyword=' + search
  })
}

export function deluser(id) {
  return request({
    url: '/api/auth/del',
    method: 'post',
    data: {
      id,
      username: 'xx'
    }
  })
}

export function editroles(id, roles) {
  return request({
    url: '/api/auth/editroles',
    method: 'post',
    data: {
      id,
      roles,
      username: 'xx'
    }
  })
}
