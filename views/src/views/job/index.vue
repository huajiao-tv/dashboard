<template>
  <div style="padding: 20px">
    <!--<el-button type="primary" @click="dialogAdd = true">回收站</el-button>-->
    <el-button type="primary" @click="jobDialogHandle('add')">添加</el-button>

    <el-tabs v-model="activeName" @tab-click="setQuey">
      <template v-for="item in tableData">
        <el-tab-pane :label="systemOptions[item.system] + ' - ' + item.system" :name="item.system" :key="item.system">
          <el-card class="box-card">
            <div slot="header" class="clearfix">
              <span>{{ item.system }}</span>
            </div>
            <div class="text item">
              <el-row>
                <el-col :span="4">任务总数 : {{ item.totalJobs }}</el-col>
                <el-col :span="4">成功总数 : {{ item.successJobs }}</el-col>
                <el-col :span="4">失败总数 : {{ item.failJobs }}</el-col>
                <el-col :span="4">当日成功总数 : {{ item.successExec }}</el-col>
                <el-col :span="4">当日失败总数 : {{ item.failExec }}</el-col>
              </el-row>

            </div>
          </el-card>
          <el-card class="box-card">
            <div slot="header" class="clearfix">
              <span>任务列表</span>

            </div>
            <div class="text item">
              <el-table :data="item.jobs" stripe>
                <el-table-column prop="id" label="ID"/>
                <el-table-column prop="name" label="任务"/>
                <el-table-column prop="desc" label="任务说明"/>
                <el-table-column prop="concurrency" label="任务并发"/>
                <el-table-column prop="successJobs" label="成功次数"/>
                <el-table-column prop="failExec" label="失败次数"/>
                <el-table-column :formatter="statusFormat" label="状态"/>
                <el-table-column label="操作" min-width="100">
                  <template slot-scope="scope">
                    <el-button-group>
                      <el-tooltip class="item" effect="dark" content="开启任务" placement="top-start">
                        <el-button
                          v-if="scope.row.status == 1"
                          type="success"
                          size="mini"
                          icon="el-icon-check"
                          square
                          @click="jobStart(scope.row.id)"/>
                      </el-tooltip>
                      <el-tooltip class="item" effect="dark" content="暂停任务" placement="top-start">
                        <el-button
                          v-if="scope.row.status == 0"
                          type="warning"
                          size="mini"
                          icon="el-icon-close"
                          square
                          @click="jobPause(scope.row.id)"/>
                      </el-tooltip>

                      <router-link :to="{path: '/cron/detail/'+scope.row.id }">
                        <el-tooltip class="item" effect="dark" content="查看任务运行状态" placement="top-end">
                          <el-button square type="primary" size="mini" icon="el-icon-menu"/>
                        </el-tooltip>
                      </router-link>

                      <el-dropdown @command="handleActionCommand(scope.row, $event)">
                        <el-button
                          type="success"
                          size="mini"
                          square>
                          更多<i class="el-icon-arrow-down el-icon--right"/>
                        </el-button>
                        <el-dropdown-menu slot="dropdown">
                          <el-dropdown-item command="updateJob">修改配置</el-dropdown-item>
                          <el-dropdown-item command="deleteJob">删除任务</el-dropdown-item>
                          <el-dropdown-item command="copyJob" divided>复制任务</el-dropdown-item>
                          <el-dropdown-item command="testJob">测试任务</el-dropdown-item>
                        </el-dropdown-menu>
                      </el-dropdown>

                    </el-button-group>
                  </template>
                </el-table-column>
              </el-table>

            </div>
          </el-card>

        </el-tab-pane>
      </template>
    </el-tabs>

    <jobEditor ref="jobDialog" :systems="systemOptions" :jobs="jobs" @jobDialogAsk="jobDialogHandle"/>

    <el-dialog :visible.sync="dialogDetail" title="执行结果">
      <div v-loading="loadStatus" style="height: 500px">
        <el-card class="box-card">

          <div class="text item" style="padding-bottom:5px">
            开始时间：{{ (new Date(testLog.started_at)).getFullYear()+"-"+((new Date(testLog.started_at)).getMonth()+1)+"-"+(new Date(testLog.started_at)).getDate()+" "+(new Date(testLog.started_at)).getHours()+":"+(new Date(testLog.started_at)).getMinutes() +":"+(new Date(testLog.started_at)).getSeconds() }}
          </div>
          <div class="text item" style="padding-bottom:5px">
            结束时间：{{ (new Date(testLog.finished_at)).getFullYear()+"-"+((new Date(testLog.finished_at)).getMonth()+1)+"-"+(new Date(testLog.finished_at)).getDate()+" "+(new Date(testLog.finished_at)).getHours()+":"+(new Date(testLog.finished_at)).getMinutes() +":"+(new Date(testLog.finished_at)).getSeconds() }}
          </div>
          <div class="text item" style="padding-bottom:5px">
            内容：<pre>{{ testLog.output }}</pre>
        </div></el-card>
      </div>
    </el-dialog>

    <el-dialog :visible.sync="dialogRemove" title="删除任务">
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogRemove=false">取 消</el-button>
        <el-button type="primary" @click="remvoeJob(tmpData.id)">确定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getSystems } from '../../api/system'
import {
  remvoeJob,
  setJobState,
  getJobs,
  jobStart,
  jobPause,
  depsTree,
  taskTest,
  taskTestLog
}
  from '../../api/job'

import { getToken } from '../../utils/auth.js'
import jobEditor from '@/views/job/job-editor.vue'
import JsonEditor from '../../components/JsonEditor/index'
var stateMap = [
  '运行中',
  '停止中',
  '已删除'
]
export default {

  components: { JsonEditor, jobEditor },
  data() {
    return {
      testBeginTime: {},
      loadStatus: true,
      timer: {},
      testLog: {},
      expendKey: [],
      dialogRemove: false,
      dialogDetail: false,
      jobDetail: '',
      tableData: [],
      systemOptions: {},
      jobSystems: {},
      jobs: [],
      tmpData: {},
      activeName: '',
      jsOption: {
        // mode: 'view'
      },
      depsData: [],
      defaultProps: {
        children: 'Childrens',
        label: function(data, node) {
          return data.name + '-' + stateMap[data.status]
        }
      },
      ws: {}
    }
  },

  created() {
    this.initialize()
    // console.log(process.env.WEBSOCKET_URL)
    var wsUrl = 'ws://' + window.location.hostname
    if (process.env.WEBSOCKET_URL !== '') {
      wsUrl = process.env.WEBSOCKET_URL
    }
    var ws = new WebSocket(wsUrl + '/push')
    if (ws === undefined) {
      var timer = setInterval(
        this.initialize, 30000
      )
      console.log(timer)
    } else {
      // ws.onopen(ws.send('tt'))
      ws.onmessage = ee => {
        console.log(ee)
        this.initialize()
      }
      ws.onopen = function() {
        var it = {
          token: getToken()
        }
        // console.log(getToken())
        ws.send(JSON.stringify(it))
      }
    }
  },

  mounted() {
  },

  methods: {
    initialize() {
      getSystems().then(res => {
        res.data.forEach(system => {
          this.systemOptions[system.name] = system.desc
        })

        getJobs().then(res => {
          if (res.data != null) {
            this.jobs = res.data
            this.getJobSystem()
            this.setActiveName()
          } else {
            this.$notify({
              title: 'job不存在',
              message: '请添加job',
              type: 'success',
              duration: 2000
            })
          }
        })
      })
    },
    showRemoveDialog(id) {
      this.tmpData.id = id
      this.dialogRemove = true
    },

    getStat(val) {
      // is
      if (isNaN(val.Statics.success)) {
        return 0
      } else {
        return parseInt(val.Statics.success)
      }
    },

    setQuey(system) {
      console.log(this.tableData[system.index].system)
      var name = this.tableData[system.index].system
      var path = this.$router.history.current.path
      this.$router.push({ path, query: { system: name }})
      this.$route.query.system = name
    },

    setActiveName() {
      if (this.$route.query.system === undefined && this.activeName === '0') {
        this.activeName = this.tableData[0].system
      } else if (this.activeName === '0') {
        // if (this.jobSystems.)
        this.activeName = this.$route.query.system
        if (this.jobSystems[this.activeName ] === undefined) {
          this.$notify({
            title: '系统不存在',
            message: '请添加系统',
            type: 'success',
            duration: 2000
          })
        }
      }
    },
    /**
       * 获取系统数据
       */
    getJobSystem() {
      var system = {}
      this.jobs.forEach(val => {
        system[val.system] = {
          system: val.system,
          totalJobs: 0,
          successJobs: 0,
          failJobs: 0,
          successExec: 0,
          failExec: 0,
          jobs: []
        }
      })
      this.jobSystems = system
      this.jobs.forEach(val => {
        var job = {
          id: val.id,
          name: val.name,
          desc: val.desc,
          concurrency: val.concurrency,
          status: val.status,
          nodeStatus :val.nodeStatus,
          sid: val.system,
          successJobs: isNaN(val.Statics.success) ? 0 : parseInt(val.Statics.success),
          failJobs: isNaN(val.Statics.failed) ? 0 : parseInt(val.Statics.failed),
          successExec: isNaN(val.Statics.day_success) ? 0 : parseInt(val.Statics.day_success),
          failExec: isNaN(val.Statics.day_failed) ? 0 : parseInt(val.Statics.day_failed)
        }

        system[val.system].successJobs += job.successJobs
        system[val.system].failJobs += job.failJobs
        system[val.system].successExec += job.successExec
        system[val.system].failExec += job.failExec
        system[val.system].jobs.push(job)
        system[val.system].totalJobs = system[val.system].jobs.length
      })
      this.tableData = []
      for (var i in system) {
        // console.log(system[i])
        this.tableData.push(system[i])
      }
    },
    // viewJob(id) {
    //   this.getTree(id)
    //   this.dialogDetail = true
    //   this.jobs.forEach(val => {
    //     if (val.id === id) {
    //       this.jobDetail = val
    //     }
    //   })
    // },
    jobStart(id) {
      jobStart(id).then(res => {
        this.tableData.forEach(val => {
          for (var i in val.jobs) {
            if (val.jobs[i].id === id) {
              val.jobs[i].status = 0
            }
          }
          // val.totalJobs = val.jobs.length
        })
      })
    },
    jobPause(id) {
      jobPause(id).then(res => {
        this.tableData.forEach(val => {
          for (var i in val.jobs) {
            if (val.jobs[i].id === id) {
              val.jobs[i].status = 1
            }
          }
          // val.totalJobs = val.jobs.length
        })
      })
    },
    setJobState(id) {
      setJobState(id)
    },
    remvoeJob(row) {
      remvoeJob(row.id).then(res => {
        this.tableData.forEach(val => {
          if (val.system === row.sid) {
            for (var i in val.jobs) {
              if (val.jobs[i].id === row.id) {
                val.jobs.splice(i, 1)
              }
            }
            val.totalJobs = val.jobs.length
          }
        })
      })

      this.dialogRemove = false
    },

    statusFormat(row, column, cellValue, index) {
      if (row.status === 1) {
        return '停止中'
      };
      if (row.nodeStatus === 1){
          return '运行异常'
      }
      return '运行中'
    },

    getTree(id) {
      depsTree(id).then(res => {
        // console.log(res)

        this.depsData = res.data.Childrens
      })
    },

    handleActionCommand(row, cmd) {
      if (cmd === 'updateJob') {
        this.jobUpdateDialogHandle(row)
      } else if (cmd === 'deleteJob') {
        this.showRemoveDialog(row)
      } else if (cmd === 'copyJob') {
        this.jobCopyHandle(row)
      } else if (cmd === 'testJob') {
        this.jobTest(row.id)
      }
    },

    rowKey(row) {
      // console.log(row)
      return row.system
    },
    rowClick(row, event, column) {
      // console.log(this.$refs)
      // console.log(row)
      if (this.expendKey.length > 0 && this.expendKey.indexOf(row.system) >= 0) {
        this.expendKey = []
      } else {
        this.expendKey = []
        this.expendKey.push(row.system)
      }

      // return row.system
    },

    jobDialogHandle(ask) {
      if (ask === 'add') {
        this.$refs.jobDialog.resetJobForm()
        this.$refs.jobDialog.dialogVisible = true
      } else if (ask === 'saved') {
        this.initialize()
      }
    },

    jobUpdateDialogHandle(job) {
      this.$refs.jobDialog.resetJobForm()
      this.jobs.forEach(val => {
        if (val.id === job.id) {
          // var jobDetail = val
          this.$refs.jobDialog.setjobForm(val)
          this.$refs.jobDialog.dialogVisible = true
          return
        }
      })
    },

    jobCopyHandle(job) {
      this.$refs.jobDialog.resetJobForm()
      this.jobs.forEach(val => {
        if (val.id === job.id) {
          var jobDetail = val
          jobDetail.id = null
          this.$refs.jobDialog.setjobForm(jobDetail)
          this.$refs.jobDialog.dialogVisible = true
          return
        }
      })
    },
    // todo 和服务器时间差的有点大
    jobTest(id) {
      this.testBeginTime = (new Date()).getTime() * 1000000
      taskTest(id).then(res => {
        this.loadStatus = true
        this.dialogDetail = true

        this.timer = setInterval(() => {
          taskTestLog(id).then(res => {
            if (res.data != null || this.dialogDetail === false) {
              if (this.dialogDetail === false) {
                this.loadStatus = false
                clearInterval(this.timer)
              } else if (res.data != null && res.data.length > 0) {
                this.testLog = res.data[0]
                // var datasplit = this.testLog.dispatch_id.split(':')
                // console.log(this.testBeginTime)
                // console.log(datasplit[2])
                // if (datasplit[2] >= this.testBeginTime) {
                this.loadStatus = false
                clearInterval(this.timer)
                // }
              }
            }
          })
        }, 10000)
      })
    }
  }
}
</script>
