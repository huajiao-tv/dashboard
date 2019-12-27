import Cookies from 'js-cookie'

const TokenKey = 'Admin-Token'

export default {
  TokenKey
}

export function getToken() {
  var token = Cookies.get(TokenKey)
  if (isNaN(token)) {
    token = window.localStorage.getItem(TokenKey)
  }
  return token
}

export function setToken(token) {
  window.localStorage.setItem(TokenKey, token)
  return Cookies.set(TokenKey, token)
}

export function removeToken() {
  window.localStorage.removeItem(TokenKey)
  return Cookies.remove(TokenKey)
}
