import Mock from 'mockjs'

Mock.mock('/api/auth/login', (req, res) => {
  return {
    code: 0,
    message: '',
    data: {
      token: 'test'
    }
  }
})

Mock.mock('/api/auth/info', (req, res) => {
  return {
    code: 0,
    message: '',
    data: {
      name: 'test',
      email: 'test@xxx.com',
      avatar: 'http://image.huajiao.com/b3e1be4fe93e001ff9c33fc7ebf6a759.jpg',
      roles: [
        'admin'
      ]
    }
  }
})

Mock.mock('/api/system/list', (req, res) => {
  return {
    code: 0,
    message: '',
    data: [
      {
        name: 'live',
        desc: '直播',
        machines: 10,
        storages: 2,
        topics: 5,
        cron_machines: 2,
        jobs: 10
      }, {
        name: 'test',
        desc: '测试',
        machines: 9,
        storages: 2,
        topics: 1,
        cron_machines: 2,
        jobs: 10
      }
    ]
  }
})

Mock.mock('/api/queue/list', (req, res) => {
  return {
    code: 0,
    message: '',
    data: [
      {
        name: 'LiveStart',
        desc: '开播',
        topics: 5
      }, {
        name: 'test',
        desc: '测试',
        topics: 1
      }
    ]
  }
})

Mock.mock('/api/machine/list', (req, res) => {
  var body = JSON.parse(req.body)
  if (body.cron) {
    return {
      code: 0,
      message: '',
      data: [
        {
          system: 'live',
          ip: '127.0.0.1:12306',
          status: 0
        }, {
          system: 'live',
          ip: '127.0.0.2:12306',
          status: 0
        }
      ]
    }
  }
  return {
    code: 0,
    message: '',
    data: [
      {
        system: 'live',
        ip: '127.0.0.1:12017',
        status: 0
      }, {
        system: 'live',
        ip: '127.0.0.2:12017',
        status: 0
      }, {
        system: 'live',
        ip: '127.0.0.3:12017',
        status: 0
      }
    ]
  }
})

Mock.mock('/api/storage/list', (req, res) => {
  return {
    code: 0,
    message: '',
    data: [
      {
        id: 1,
        system: 'live',
        host: '127.0.0.1',
        port: 6379,
        status: 0
      }, {
        id: 2,
        system: 'live',
        host: '127.0.0.2',
        port: 6379,
        status: 0
      }, {
        id: 3,
        system: 'live',
        host: '127.0.0.3',
        port: 6379,
        status: 0
      }
    ]
  }
})

Mock.mock('/api/topic/system/summary', (req, res) => {
  return {
    code: 0,
    message: '',
    data: [
      {
        queue: 'LiveStart',
        name: 'Push',
        desc: '开播提醒',
        file: 'test.php',
        workers: 10,
        storage: '127.0.0.1:6379',
        length: 100,
        retry: 0,
        machines: [
          {
            node: '127.0.0.1:12017',
            success_count: 100,
            fail_count: 10,
            cgi_qps: 12.1,
            min_consume_time: 120,
            max_consume_time: 300,
            avg_consume_time: 200
          }, {
            node: '127.0.0.2:12017',
            success_count: 150,
            fail_count: 5,
            cgi_qps: 10.5,
            min_consume_time: 100,
            max_consume_time: 500,
            avg_consume_time: 400
          }
        ]
      }, {
        queue: 'LiveStart',
        name: 'Push2',
        desc: '开播提醒222',
        file: 'test.php',
        workers: 10,
        storage: '127.0.0.1:6379',
        length: 100,
        retry: 0,
        machines: [
          {
            node: '127.0.0.1:12017',
            success_count: 100,
            fail_count: 10,
            cgi_qps: 12.1,
            min_consume_time: 120,
            max_consume_time: 300,
            avg_consume_time: 200
          }, {
            node: '127.0.0.2:12017',
            success_count: 150,
            fail_count: 5,
            cgi_qps: 10.5,
            min_consume_time: 100,
            max_consume_time: 500,
            avg_consume_time: 400
          }
        ]
      }
    ]
  }
})

Mock.mock('/job/list', (req, res) => {
  return {
    code: 0,
    message: '',
    data: [
      {
        id: 1,
        name: 'live',
        nodeCount: 10,
        sucExecsTotal: 100,
        errCount: 6379,
        status: 0
      }, {
        id: 2,
        name: 'live1',
        nodeCount: 10,
        sucExecsTotal: 100,
        errCount: 6379,
        status: 0
      }, {
        id: 3,
        name: 'live2',
        nodeCount: 10,
        sucExecsTotal: 100,
        errCount: 6379,
        status: 0
      }
    ]
  }
})

Mock.mock(RegExp('/api/job/detail.*'), (req, res) => {
  return {
    code: 0,
    message: '',
    data: [
      {
        name: 'live',
        nodeCount: 10,
        sucExecsTotal: 100,
        errCount: 6379,
        status: 0
      }, {
        name: 'live1',
        nodeCount: 10,
        sucExecsTotal: 100,
        errCount: 6379,
        status: 0
      }, {
        name: 'live2',
        nodeCount: 10,
        sucExecsTotal: 100,
        errCount: 6379,
        status: 0
      }
    ]
  }
})
