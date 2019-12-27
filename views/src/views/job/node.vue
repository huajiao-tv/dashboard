<template>
  <div style="padding: 20px">
    <!--<el-button type="primary" @click="dialogAdd = true">回收站</el-button>-->
    <el-button type="primary" @click="jobDialogHandle('add')">添加</el-button>
    <!--{{ tableData[1].list[0].Nodes["10.142.106.179"].success }}-->

    <el-table ref="tableShow" :data="tableData" :row-key="rowKey" :expand-row-keys="expendKey" style="margin-top:20px;" stripe @row-click="rowClick" @expand-change="currentRow">
      <el-table-column label="" type="expand" >

        <template slot-scope="scope">
          <div style="border-left: solid 1px lightgrey;margin-top: -14px;">
            <el-alert :closable="false" title="任务列表" type="info" style="margin-bottom: 20px"/>
            <el-table :data="scope.row.list" stripe>
              <el-table-column prop="name" label="任务"/>
              <el-table-column prop="desc" label="任务说明"/>
              <el-table-column prop="concurrency" label="任务并发"/>
              <el-table-column :formatter="formatSucNum" prop="sucNum" label="成功次数"/>
              <el-table-column :formatter="formatFailExec" prop="failExec" label="失败次数"/>
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

                    <el-tooltip class="item" effect="dark" content="查看任务配置" placement="top-start">
                      <el-button square type="info" size="mini" icon="el-icon-view" @click="viewJob(scope.row.id)"/>
                    </el-tooltip>
                    <router-link :to="{path: '/cron/detail/'+scope.row.id }">
                      <el-tooltip class="item" effect="dark" content="查看任务运行状态" placement="top-start">

                        <el-button square type="primary" size="mini" icon="el-icon-menu"/>

                      </el-tooltip>
                    </router-link>
                    <el-tooltip class="item" effect="dark" content="删除任务" placement="top-start">
                      <el-button
                        type="danger"
                        size="mini"
                        icon="el-icon-delete"
                        square
                        @click="showRemoveDialog(scope.row)"/>
                    </el-tooltip>
                  </el-button-group>
                </template>
              </el-table-column>
            </el-table>

          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="节点"/>
      <el-table-column prop="jobNum" label="任务总数"/>
      <el-table-column prop="sucNum" label="总成功数"/>
      <el-table-column prop="failJobs" label="失败任务数"/>
      <el-table-column prop="successExec" label="当天成功次数"/>
      <el-table-column prop="failExec" label="当天失败次数"/>
      <el-table-column label="操作">
        <template slot-scope="scope">
          <el-button type="success" size="mini" icon="el-icon-plus" square @click.stop="jobDialogHandle('add')"/>
        </template>
      </el-table-column>
    </el-table>

    <jobEditor ref="jobDialog" :systems="systemOptions" :jobs="jobs" @jobDialogAsk="jobDialogHandle"/>

    <el-dialog :visible.sync="dialogDetail" title="依赖关系">

      <div v-if="depsData != null ">
        <el-tree
          :data="depsData"
          :props="defaultProps"
          default-expand-all/>
      </div>
      <div style="height: 500px">
        <json-editor ref="jsonEditor" :outter="jobDetail" :options="jsOption"/>
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
import {
  remvoeJob,
  setJobState,
  jobPause,
  depsTree,
  nodeList
}
  from '../../api/job'

import jobEditor from '@/views/job/job-editor.vue'
import JsonEditor from '../../components/JsonEditor/index'
import { getMachines } from '../../api/cluster'
import { getSystems } from '../../api/system'
import contt from '../../utils/auth.js'
var stateMap = [
  '运行中',
  '停止中',
  '已删除'
]
export default {

  components: { JsonEditor, jobEditor },
  data() {
    return {
      expendKey: [],
      dialogRemove: false,
      dialogDetail: false,
      jobDetail: '',
      nodelist: [],
      tableData: [],
      SystemJob: [],
      OriginalData: [],
      systemOptions: {},
      jobs: [],
      tmpData: {},
      jsOption: {
        // mode: 'view'
      },
      depsData: [],
      defaultProps: {
        children: 'Childrens',
        label: function(data, node) {
          return data.name + '-' + stateMap[data.status]
        }
      }

    }
  },

  created() {
    this.initialize()
    var wsUrl = 'ws://' + window.location.host
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
        var token = window.localStorage.getItem(contt.TokenKey)
        console.log(token)
        var it = {
          token: token
        }
        ws.send(JSON.stringify(it))
        // setInterval(function() {
        //   var da = new Date()
        //   var ii = { 'time': da.toString() }
        //   ws.send(JSON.stringify(ii))
        // }, 10000)
      }
    }
  },

  methods: {
    initialize() {
      getSystems().then(res => {
        res.data.forEach(system => {
          this.systemOptions[system.name] = system.desc
        })
        getMachines(true).then(res => {
          this.nodelist = res.data
          nodeList().then(res => {
            if (res.data != null) {
              this.OriginalData = res.data
              this.getJobSystem()
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
        // getJobs().then(res => {
        //   this.jobs = res.data
        //   this.getJobSystem()
        // })
      })
    },
    showRemoveDialog(id) {
      this.tmpData.id = id
      this.dialogRemove = true
    },
    /**
       * 获取系统数据
       */
    getJobSystem() {
      var system = {}
      this.nodelist.forEach(val => {
        system[val.ip] = []
      })
      this.OriginalData.forEach(val => {
        console.log(val)
        for (var k in val.Nodes) {
          if (system[k] === undefined) {
            system[k] = []
          }
          system[k].push(val)
        }
      })

      this.tableData = []
      for (var i in system) {
        var tmp = {
          name: i,
          sucNum: 0,
          failJobs: 0,
          successExec: 0,
          failExec: 0,
          jobNum: system[i].length,
          list: system[i]
        }
        system[i].forEach(val => {
          for (var kk in val.Nodes) {
            tmp.sucNum += val.Nodes[kk].success
            tmp.failJobs += val.Nodes[kk].failed
            tmp.successExec += val.Nodes[kk].day_success
            tmp.failExec += val.Nodes[kk].day_failed
          }
        })

        this.tableData.push(tmp)
      }
      // console.log(this.tableData)
    },
    viewJob(id) {
      this.getTree(id)
      this.dialogDetail = true
      this.OriginalData.forEach(val => {
        if (val.id === id) {
          this.jobDetail = val
        }
      })
    },

    jobPause(id) {
      jobPause(id).then(res => {
        this.tableData.forEach(val => {
          for (var i in val.list) {
            if (val.list[i].id === id) {
              val.list.splice(i, 1)
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
      // console.log(row.status)
      if (row.status === 1) {
        return '停止中'
      }
      return '运行中'
    },

    getTree(id) {
      depsTree(id).then(res => {
        // console.log(res)
        this.depsData = res.data.Childrens
      })
    },
    rowKey(row) {
      // console.log(row)
      return row.name
    },
    rowClick(row, event, column) {
      // console.log(this)
      // console.log(row)
      if (this.expendKey.length > 0 && this.expendKey.indexOf(row.name) >= 0) {
        this.expendKey = []
      } else {
        this.expendKey = []
        this.expendKey.push(row.name)
      }
    },

    jobDialogHandle(ask) {
      if (ask === 'add') {
        this.$refs.jobDialog.resetJobForm()
        this.$refs.jobDialog.dialogVisible = true
      } else if (ask === 'saved') {
        this.initialize()
      }
    },
    formatSucNum(row, column, cellValue, index) {
      // console.log(this.expendKey)
      if (row.Nodes[this.expendKey[0]] === undefined) {
        return 0
      }
      return row.Nodes[this.expendKey[0]].success
    },
    formatFailExec(row, column, cellValue, index) {
      // console.log(this.expendKey)
      if (row.Nodes[this.expendKey[0]] === undefined) {
        return 0
      }
      return row.Nodes[this.expendKey[0]].failed
    },
    currentRow(row, expandedRows) {
      this.rowClick(row)
    }
  }
}
</script>
