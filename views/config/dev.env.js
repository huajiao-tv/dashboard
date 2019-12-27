'use strict'
const merge = require('webpack-merge')
const prodEnv = require('./prod.env')

module.exports = merge(prodEnv, {
  NODE_ENV: '"development"',
  BASE_API: '""',
  WEBSOCKET_URL:'"ws://127.0.0.1:12335"'
  // WEBSOCKET_URL:'""'
})
