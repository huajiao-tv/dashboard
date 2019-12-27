<template>
  <div>
    <el-table ref="topicListTable" :data="tableData" :row-class-name="tableRowClassName" style="width: 100%" stripe @row-click="expandRow" >
      <el-table-column type="expand">
        <template slot-scope="scope">
          <div style="border-left: solid 1px lightgrey;margin-top: -14px;">
            <el-alert v-if="scope.row.retry" title="" type="warning" show-icon>
              有 {{ scope.row.retry }} 记录处理失败，可以
              <el-button type="danger" plain style="padding:3px" @click="retryOperate('cleanAll', scope.row.queue, scope.row.name)">全部清除</el-button>
              <el-button type="danger" plain style="padding:3px" @click="retryOperate('recoverOne', scope.row.queue, scope.row.name)">单条重放</el-button>
              <el-button type="danger" plain style="padding:3px" @click="retryOperate('recoverAll', scope.row.queue, scope.row.name)">全部重放</el-button>
              <el-button type="success" plain style="padding:3px" @click="showError(scope.row)">查看报错</el-button>
            </el-alert>
            <el-alert :closable="false" :title="'消费节点列表, 消费文件: '+ scope.row.file" type="info" style="margin-bottom: 20px"/>
            <el-table
              :data="scope.row.machines"
              :summary-method="getSummaries"
              stripe
              show-summary
            >
              <el-table-column
                prop="Node"
                label="节点"/>
              <el-table-column
                :formatter="columnFormat"
                prop="Succ"
                label="成功总数/失败总数"/>
              <el-table-column
                prop="SuccQpm"
                label="每分钟成功数"/>
              <el-table-column
                prop="FailQpm"
                label="每分钟失败数"/>
              <el-table-column
                prop="ConsumeLe50"
                label="50%请求执行时间"/>
              <el-table-column
                prop="ConsumeLe90"
                label="90%请求执行时间"/>
              <el-table-column
                prop="ConsumeLe99"
                label="99%请求执行时间"/>
              <el-table-column
                :formatter="checkNodeScript"
                prop=""
                label="问题"
                lazy />
            </el-table>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="系统" prop="system"/>
      <el-table-column label="Queue/Topic">
        <template slot-scope="scope">
          {{ scope.row.queue }}/{{ scope.row.name }}
        </template>
      </el-table-column>
      <el-table-column label="说明" prop="desc"/>
      <el-table-column label="积压长度" prop="length" sortable/>
      <el-table-column label="失败长度" prop="retry" sortable/>
      <el-table-column label="自动重试失败长度" prop="timeout" sortable/>
      <el-table-column label="存储" prop="storage" width="160"/>
      <el-table-column :formatter="workerFormat" label="workers数量(单机/总数)" prop="workers"/>
      <el-table-column label="操作" width="280">
        <template slot-scope="scope">
          <el-tooltip class="item" effect="dark" content="重置统计数据" placement="top-start">
            <el-button icon="el-icon-warning" type="warning" size="small" circle @click.stop="CleanStatistics(scope.row.queue, scope.row.name)"/>
          </el-tooltip>
          <el-button v-show="scope.row.status == 0" icon="el-icon-success" type="success" size="small" circle @click.stop="changeStatus(scope.row, 1)"/>
          <el-button v-show="scope.row.status == 1" icon="el-icon-error" type="info" size="small" circle @click.stop="changeStatus(scope.row, 0)"/>
          <el-button icon="el-icon-refresh" size="small" circle @click.stop="refreshQueue(scope.row, 0)"/>

          <el-dropdown split-button type="primary" trigger="click" size="medium" style="margin-left:10px;" @click.stop="testTopic(scope.row)">
            测试
            <el-dropdown-menu slot="dropdown" @click.native.stop="console.log(111)">
              <el-dropdown-item @click.native.stop="modTopic(scope.row)">修改</el-dropdown-item>
              <el-dropdown-item @click.native.stop="deleteTopic(scope.row)">删除</el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="'队列测试 - ' + topicData.queue + '/' + topicData.topic" :visible.sync="dialogTest" width="900px">
      <el-alert :closable="false" title="" type="info" style="margin-bottom: 10px;">发送 json 数据: {{ topicData.file }}</el-alert>
      <el-row style="display:flex">
        <el-col :span="16"><json-editor ref="jsonEditor" :outter="jsonEditorData.outter" @inner-change="loadInnerData" /></el-col>
        <el-col :span="8" style="padding-left:10px;">
          <el-tabs v-loading="historyData.loading" id="historyWrap" v-model="historyData.tab" @tab-click="loadHistory">
            <el-tab-pane label="本地历史" name="local">
              <el-button v-for="(item, k) in historyData.local" :key="k" type="info" @click="fillHistory('local', k)">{{ item }}</el-button>
            </el-tab-pane>
            <el-tab-pane label="测试服务器历史" name="testServer">
              <el-button v-for="(item, k) in historyData.testServer" v-if="item!==''" :key="k" type="info" @click="fillHistory('testServer', k)">{{ item }}</el-button>
            </el-tab-pane>
          </el-tabs>
        </el-col>
      </el-row>
      <div v-if="noScript">
        <el-button type="danger">{{ noScript }}</el-button>
      </div>
      <div v-if="topicReturn ">
        <el-alert :closable="false" title="cgi 返回" type="info" style="margin-top: 20px; margin-bottom: 10px" />
        <json-editor ref="jsonEditorRes" :outter="topicReturn"/>
      </div>
      <div slot="footer" class="dialog-footer">
        <el-select v-model="topicData.node" placeholder="请选择节点" style="float: left">
          <el-option
            v-for="node in topicMachines"
            :key="node"
            :value="node"/>
        </el-select>
        <el-button @click="dialogTest = false">关 闭</el-button>
        <el-dropdown :loading="topicData.loading" split-button type="primary" @click.stop="sendTopicData('data')">
          发 送
          <el-dropdown-menu slot="dropdown" @click.native.stop="console.log(111)">
            <el-dropdown-item @click.native.stop="sendTopicData('ping')">Ping</el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </div>
    </el-dialog>

    <el-dialog :title="'队列修改 - ' + topicModData.queue + '/' + topicModData.name" :visible.sync="dialogMod">
      <el-form ref="modTopic" :model="topicModData" :rules="addTopicRules" label-width="140px" size="mini">
        <el-form-item label="topic名称" prop="name" >
          <el-input v-model="topicModData.name"/>
        </el-form-item>
        <el-form-item :label="topicModData.run_type ? 'url地址' : '消费文件'" prop="consume" >
          <el-input v-model="topicModData.consume" :placeholder="topicModData.run_type ? 'http请求地址' : '绝对路径，例如：/home/q/system/pepper/live/front/src/worker/liveStartPushBatchWorker.php'"/>
        </el-form-item>
        <el-form-item label="worker数量" prop="num_of_workers">
          <el-input v-model.number="topicModData.num_of_workers"/>
        </el-form-item>
        <el-form-item label="队列积压报警值" prop="alarm">
          <el-input v-model.number="topicModData.alarm">
            <template slot="prepend">为0时取默认值{{ defaultAlarm }}</template>
          </el-input>
        </el-form-item>
        <el-form-item label="retry积压报警值" prop="alarm_retry">
          <el-input v-model.number="topicModData.alarm_retry">
            <template slot="prepend">为0时取默认值{{ defaultAlarmRetry }}</template>
          </el-input>
        </el-form-item>
        <el-form-item label="运行类型" prop="run_type">
          <el-select v-model="topicModData.run_type" placeholder="请选择运行类型">
            <el-option key="0" :value="0" label="CGI类型(php-fpm:0)"/>
            <el-option key="1" :value="1" label="HTTP类型(golang,java:1)"/>
          </el-select>
        </el-form-item>
        <el-form-item label="开启自动重试" prop="is_retry">
          <el-switch
            v-model="retryConf.switch"
            active-color="#13ce66"
            inactive-color="#ff4949"
          />
        </el-form-item>
        <el-form-item v-if="retryConf.switch" label="重试有效期" prop="is_retry">
          <el-input v-model="retryConf.durTime" placeholder="单位为秒"/>
        </el-form-item>
        <el-form-item v-if="topicModData.run_type === 1" prop="http_config" label="host">
          <el-input v-model="httpConfig.host" placeholder=" '  域名 '"/>
        </el-form-item>

        <el-form-item label="存储" prop="storage">
          <el-select v-model="topicModData.storage" placeholder="请选择使用的存储">
            <el-option
              v-for="storage in storages"
              v-if="storage.system === topicModData.system"
              :key="storage.id"
              :label="storage.host + ':' + storage.port + '('+ storage.desc +')'"
              :value="storage.id"/>
          </el-select>
        </el-form-item>
        <el-form-item label="操作注释" prop="comment">
          <el-input v-model="topicModData.comment" />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitModTopic(topicModData,httpConfig)">提交修改</el-button>
        <el-button @click="dialogMod = false">关 闭</el-button>
      </div>
    </el-dialog>

    <el-dialog
      :visible.sync="dialogError"
      fullscreen
      title="错误信息-从新到旧排列"
    >
      <div v-for="(k,e) in retryErrors" :key="e+k.Time">
        <el-card class="box-card">
          <div class="text item" style="padding-bottom:5px">
            时间：{{ (new Date(k.Time*1000)).getFullYear() }}-{{ (new Date(k.Time*1000)).getMonth()+1 }}-{{ (new Date(k.Time*1000)).getDate() }} {{ (new Date(k.Time*1000)).getHours() }}:{{ (new Date(k.Time*1000)).getMinutes() }}:{{ (new Date(k.Time*1000)).getSeconds() }}
          </div>
          <div class="text item" style="padding-bottom:5px">
            队列内容：{{ k.Jobs }}
          </div>
          <div class="text item" style="padding-bottom:5px">
            <pre>报错信息：{{ k.UserMsg }}</pre>
          </div>
          <div class="text item" style="padding-bottom:5px">
            <pre>Node Hostname：{{ formatHostName(k.SysError) }}</pre>
          </div>
        </el-card>
      </div>

    </el-dialog>
  </div>
</template>

<style>
#historyWrap {
  height: 300px;
  overflow: auto;
}
#historyWrap .el-button {
  display: block;
  padding: 10px;
  width: 100%;
}
#historyWrap .el-button span {
  word-wrap: break-word;
  white-space: initial;
  display: block;
  text-align: left;
}
#historyWrap .el-button + .el-button {
  margin: 10px 0 0;
}
.el-table .warning-row {
  background: oldlace;
  color: red
}

</style>

<script>
import JsonEditor from '../../components/JsonEditor/index'
import { getStorages } from '../../api/storage'
import { defaultThreshold, getRetryErrors, delTopic, sendTopic, refreshQueue, recoverRetryQueue, cleanRetryQueue, updateTopic, getHistory, cleanStatistics } from '../../api/topic'

export default {
  name: 'TopicList',
  components: {
    JsonEditor
  },

  props: {
    tableData: {
      type: Array,
      default: () => { return [] }
    },
    updateCallback: { // 在数据有更新时回调，如删除数据，编辑数据等，需要上层重新赋值tableData
      type: Function,
      default: () => { }
    }
  },

  data() {
    return {
      retryConf: {
        switch: false,
        durTime: 86400
      },
      httpConfig: {},
      // send outter是用于通知jsonEditor改变的
      // inner是给jsonEditor内部改变事发通知用的
      // 用历史数据测试时，通过设置outter值来改变jsonEditor的值
      // jsonEditor编辑器内修改后会通知inner （不改变outter）
      // 最终发送数据使用inner
      jsonEditorData: { inner: null, outter: null },
      // 当前展示的pannel名
      activeName: '',
      // 调试窗口使用变量
      topicData: {
        queue: '',
        topic: '',
        node: '',
        type: 'ping',
        loading: false,
        value: '' // value 是string 是dashboard程序中接受的参数
      },
      historyData: { local: [], testServer: [], loading: false, tab: 'local' },
      topicMachines: [],
      dialogTest: false,
      topicReturn: '',
      noScript: '',
      addTopicRules: {
        comment: { required: true, message: '请填写系统操作注释', trigger: 'blur' },
        storage: { required: true, message: '请选择存储', trigger: 'blur' },
        name: { required: true, message: '请输入名称', trigger: 'blur' },
        consume: [{ required: true, message: '请输入消费文件', trigger: 'blur' },
          { validator: (rule, value, callback) => {
            if (this.topicModData.run_type === 1) {
              if (!value.match(/https?:\/\/[a-z0-9A-Z.-_:%#&]{10,}$/)) {
                callback(new Error('类型为http时，必须填写url'))
                return
              }
            }
            callback()
          } }
        ],
        num_of_workers: { required: true, message: '请输入消费worker数量', trigger: 'blur' }

      },

      // 修改窗口需要的信息
      dialogMod: false,
      topicModData: {
        storage: '',
        run_type: -1,
        consume: '',
        http_config: ''
      },
      storages: {},

      // 报错窗口
      dialogError: false,
      retryErrors: [],

      defaultAlarm: 0,
      defaultAlarmRetry: 0
    }
  },

  watch: {
    dialogTest(val) {
      val || (this.topicReturn = '')
    },

    topicReturn: function() {
      if (this.topicReturn != null) {
        if (this.topicReturn.Code === '502') {
          this.noScript = '脚本不存在！'
        } else {
          this.noScript = null
        }
      }
    }

  },

  created() {
    this.formatHostName()
    getStorages().then(res => {
      this.storages = res.data
    })
    defaultThreshold().then(res => {
      this.defaultAlarm = res.data.alarm
      this.defaultAlarmRetry = res.data.alarm_retry
    })
  },

  methods: {
    columnFormat(row, column) {
      return row.Succ + ' / ' + row.Fail
    },
    tableRowClassName({ row, rowIndex }) {
      if (row.alarm > 0) {
        if (row.length > row.alarm) {
          return 'warning-row'
        }
      } else if (this.defaultAlarm > 0 && row.length > this.defaultAlarm) {
        return 'warning-row'
      }
      if (row.alarm_retry > 0) {
        if (row.retry > row.alarm_retry) {
          return 'warning-row'
        }
      } else if (this.defaultAlarmRetry > 0 && row.retry > this.defaultAlarmRetry) {
        return 'warning-row'
      }
      return ''
    },
    showError(data) {
      getRetryErrors(data.queue, data.name, data.retry).then(res => {
        this.retryErrors = res.data
        console.log(this.$data.retryErrors)
        this.dialogError = true
      })
    },
    changeStatus(data, status) {
      this.$confirm(status === 0 ? '确认开启topic消费？' : '是否关闭消费？暂停消费可能造成积压', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        var storage = ''
        this.storages.forEach(each => {
          if (each.host + ':' + each.port === data.storage && data.system === each.system) {
            storage = each.id
          }
        })
        updateTopic(
          {
            id: data.id,
            name: data.name,
            queue: data.queue,
            consume: data.file,
            num_of_workers: data.workers,
            desc: data.desc,
            system: data.system,
            storage: storage,
            comment: 'modify status',
            status: status,
            run_type: data.run_type
          }
        ).then(res => {
          data.status = status
        })
      }).catch(() => {
      })
    },
    testTopic(data) {
      this.topicData.queue = data.queue
      this.topicData.file = data.file
      this.topicData.topic = data.name
      this.jsonEditorData.outter = []
      this.jsonEditorData.inner = []
      this.topicMachines = []
      data.machines.forEach(item => {
        this.topicMachines.push(item.Node)
      })
      this.topicData.node = (this.topicMachines.length > 0) ? this.topicMachines[0] : ''
      this.dialogTest = true

      // 获取history
      this.loadHistory('local')
      this.loadHistory('testServer')

      return false
    },
    modTopic(data) {
      console.log(data)
      this.topicModData.id = data.id
      this.topicModData.name = data.name
      this.topicModData.queue = data.queue
      this.topicModData.consume = data.file
      this.topicModData.status = data.status
      this.topicModData.run_type = data.run_type
      this.storages.forEach(each => {
        if (each.host + ':' + each.port === data.storage && data.system === each.system) {
          this.topicModData.storage = each.id
        }
      })
      if (data.http_config !== '') {
        this.httpConfig = JSON.parse(data.http_config)
      }
      if (data.topic_config !== '') {
        // data.topic_config
        this.retryConf = JSON.parse(data.topic_config)
      } else {
        this.retryConf.switch = false
        this.retryConf.durTime = 86400
      }

      this.topicModData.num_of_workers = data.workers
      this.topicModData.system = data.system
      this.topicModData.desc = data.desc
      this.topicModData.alarm = data.alarm
      this.topicModData.alarm_retry = data.alarm_retry
      this.dialogMod = true
    },
    deleteTopic(data) {
      if (!confirm('确定删除 topic: ' + data.queue + '/' + data.name + ' 吗？')) {
        return
      }
      delTopic(data).then(res => {
        this.$notify({
          title: '成功',
          message: '删除成功',
          type: 'success',
          duration: 2000
        })
        this.updateCallback()
      })
    },
    refreshQueue(data) {
      var ms = []
      data.machines.forEach(m => {
        ms.push(m.Node)
      })

      refreshQueue(data.queue, data.name, ms).then(res => {
        data.length = res.data.topicLen.Normal
        data.retry = res.data.topicLen.Retry
        data.timeout = res.data.topicLen.Timeout
        res = res.data
        data.machines.forEach(machine => {
          console.log(res[machine.Node], machine.Node, machine)
          if (res[machine.Node] !== undefined) {
            machine.ConsumeLe50 = res[machine.Node].ConsumeLe50
            machine.ConsumeLe90 = res[machine.Node].ConsumeLe90
            machine.ConsumeLe99 = res[machine.Node].ConsumeLe99
            machine.Fail = res[machine.Node].Fail
            machine.Succ = res[machine.Node].Succ
            console.log(res[machine])
          }
        })
      })
    },
    retryOperate(command, queue, topic) {
      this.$confirm('确认进行操作吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        if (command === 'cleanAll') {
          this.cleanRetryQueue(queue, topic)
        } else if (command === 'recoverOne') {
          this.recoverRetryQueue(queue, topic, 1)
        } else if (command === 'recoverAll') {
          this.recoverRetryQueue(queue, topic, 0)
        }
      }).catch(() => {
      })
    },
    recoverRetryQueue(queue, topic, count) {
      recoverRetryQueue(queue, topic, count).then(res => {
        this.$notify({
          title: '成功',
          message: '重放 ' + res.data + ' 条数据',
          type: 'success',
          duration: 2000
        })
      })
    },

    CleanStatistics(queue, topic) {
      this.$confirm('确认进行操作吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        cleanStatistics(queue, topic).then(res => {
          this.$notify({
            title: '成功',
            message: '删除 数据',
            type: 'success',
            duration: 2000
          })
        })
      }).catch(() => {
      })
    },

    cleanRetryQueue(queue, topic) {
      cleanRetryQueue(queue, topic).then(res => {
        this.$notify({
          title: '成功',
          message: '删除 ' + res.data + ' 条数据',
          type: 'success',
          duration: 2000
        })
      })
    },
    getSummaries(param) {
      const { columns, data } = param
      const sums = []
      columns.forEach((column, index) => {
        switch (index) {
          case 0:
            sums[index] = '合并'
            break
          case 1:
            sums[index] = this.getSum(column, data, 'Succ') + ' / ' + this.getSum(column, data, 'Fail')
            break
          case 2:
          case 3:
            sums[index] = this.getSum(column, data)
            break
          case 4:
          case 5:
          case 6:
            sums[index] = this.getMinMax(column, data, false)
            break
        }
      })
      return sums
    },
    sendTopicData(type) {
      this.topicReturn = null // 关闭cgi返回
      this.topicData.http
      this.topicData.value = JSON.stringify(this.jsonEditorData.inner)
      if (this.jsonEditorData.inner.length === 0) {
        this.topicData.value = '[]'
      }

      if (!this.topicData.node) {
        this.$message({ message: '请选择节点', type: 'error' })
        return
      }

      // 添加到本地记录
      if (this.topicData.value !== '[]') {
        let historyLocal = this.historyData.local
        const p = historyLocal.indexOf(this.topicData.value)

        // 如果已经有该元素就删掉
        // 如果即将溢出就删除一个
        if (p > -1) historyLocal.splice(p, 1)
        else if (historyLocal.length >= 10) historyLocal = historyLocal.slice(0, 9)

        historyLocal.unshift(this.topicData.value)
        localStorage.setItem(this.getHistoryKey(), JSON.stringify(historyLocal))
        this.historyData.local = historyLocal
      }

      if (this.topicData.loading) return
      this.topicData.loading = true
      this.topicData.type = type
      sendTopic(this.topicData).then(res => {
        this.topicData.loading = false
        this.topicReturn = res.data

        this.$notify({
          title: '成功',
          message: '测试数据发送成功',
          type: 'success',
          duration: 2000
        })
      }).catch(error => {
        this.topicData.loading = false
        console.log(['send test error', error])
      })
    },
    submitModTopic(data, httpConfig) {
      this.$refs['modTopic'].validate(valid => {
        if (valid) {
          if (!confirm('确认修改topic:' + data.queue + '/' + data.name + '吗?')) {
            return
          }
          data.http_config = JSON.stringify(httpConfig)
          this.retryConf.durTime = parseInt(this.retryConf.durTime)
          data.topic_config = JSON.stringify(this.retryConf)
          updateTopic(data).then(res => {
            this.updateCallback()
            this.dialogMod = false
          })
        }
      })
    },
    getSum(column, data, p = undefined) {
      if (p === undefined) {
        p = column.property
      }
      var val = 'N/A'
      const values = data.map(item => Number(item[p]))
      if (!values.every(value => isNaN(value))) {
        val = values.reduce((prev, curr) => {
          const value = Number(curr)
          if (!isNaN(value)) {
            return prev + curr
          } else {
            return prev
          }
        }, 0)
      }
      return val
    },
    getMinMax(column, data, min = false) {
      var val = 'N/A'
      const values = data.map(item => Number(item[column.property]))
      if (!values.every(value => isNaN(value))) {
        val = values.reduce((prev, curr) => {
          const value = Number(curr)
          if (!isNaN(value)) {
            if (min && prev < curr && prev !== 0) {
              return prev
            } else {
              return curr
            }
          } else {
            return prev
          }
        }, 0)
      }
      return val
    },
    // jsoneditor 数据修改后会调用loadInnerData
    loadInnerData(val) {
      this.jsonEditorData.inner = val
    },
    // 将历史记录填充到 发送到输入框中
    fillHistory(target, index) {
      const tmp = (target === 'local')
        ? this.historyData.local[index]
        : this.historyData.testServer[index]

      // 此处只设置 jsonEditorData.outter, jsonEditor内部会通知父组件修改 jsonEditorData.inner 的值
      this.jsonEditorData.outter = JSON.parse(tmp)
    },
    getHistoryKey() {
      const queue = this.topicData.queue
      const topic = this.topicData.topic
      return `q-${queue}-t-${topic}`
    },
    loadHistory(from) {
      if (typeof from !== 'string') {
        from = from.name
      }

      // 本地的key是queue/topic 服务器的key是queue
      if (from === 'local') {
        const key = this.getHistoryKey()
        const tmp = localStorage.getItem(key)
        this.historyData.local = tmp ? JSON.parse(tmp) : []
      } else {
        if (this.historyData.loading) return
        this.historyData.loading = true
        getHistory({ queue: this.topicData.queue }).then(res => {
          this.historyData.loading = false
          const tmpList = []
          if (res.data) {
            res.data.forEach(item => {
              tmpList.push(item.Jobs)
            })
          }
          this.historyData.testServer = tmpList
        }).catch(error => {
          console.log(error)
          this.historyData.loading = false
        })
      }
    },
    expandRow(row, event, column) {
      if (event.target.tagName !== 'DIV') return
      this.$refs.topicListTable.toggleRowExpansion(row)
    },
    formatHostName(str) {
      var st = String(str)
      var ds = st.split(':')
      return ds[1]
    },

    checkNodeScript(row, expandedRows) {
      // if (row.Question === undefined) {
      //   var item = {}
      //   item.node = row.Node
      //   item.queue = row.queue
      //   item.topic = row.topic
      //   item.value = '[]'
      //   sendTopic(item).then(res => {
      //     if (res.data.Code === '502') {
      //       row.Question = '脚本不存在！'
      //     } else {
      //       row.Question = ''
      //     }
      //   })
      //   setTimeout(function() {
      //
      //   }, 1000)
      // } else {
      if(row.Question==''){
          return ''
      }
      var res = JSON.parse(row.Question)
      console.log(res)
      if (res.code === '502') {
        return '脚本不存在！'
      } else {
        return ''
      }
    },
    workerFormat(row) {
      return row.workers + ' / ' + row.machines.length * row.workers
    }
  }
}
</script>
