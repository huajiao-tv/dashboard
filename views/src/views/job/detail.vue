<template>
  <div style="padding: 20px">

    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>{{ jobConfig.name }}</span>

      </div>
      <el-table :data="tableData" style="margin-top:20px;" stripe>
        <el-table-column prop="Name" label="集群"/>
        <el-table-column prop="totalJobs" label="任务总数"/>
        <el-table-column prop="successJobs" label="成功任务数"/>
        <el-table-column prop="failJobs" label="失败任务数"/>
        <el-table-column prop="successExec" label="当天成功次数"/>
        <el-table-column prop="failExec" label="当天失败次数"/>
        <el-table-column label="操作">
          <template slot-scope="scope">
            <el-button type="success" size="mini" square @click="Dialog(scope.row.Name) " >日志</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog :visible.sync="dialogVisible " :fullscreen="true" title="日志">
      <div style="height: 100px;margin-left: 50px">
        <el-form  ref="form" label-width="100px" size="mini" style="height:100px;float: left">
          <el-input style="width: 300px;height: 100px" v-model="keyword" placeholder="模糊匹配关键字"/>
        </el-form>
        <el-button @click="searchLog()" >搜索</el-button>
      </div>
      <div v-loading="logStatus">
        <div v-for="item in logList" :key="item.Id">

          <el-card class="box-card">
            <el-row :gutter="20">
              <el-col :span="12">
                <div class="text item" style="padding-bottom:5px">
                  开始时间：{{ (new Date(item.started_at)).getFullYear()+"-"+((new Date(item.started_at)).getMonth()+1)+"-"+(new Date(item.started_at)).getDate()+" "+(new Date(item.started_at)).getHours()+":"+(new Date(item.started_at)).getMinutes() +":"+(new Date(item.started_at)).getSeconds() }}
                </div>
              </el-col>
              <el-col :span="12">
                <div class="text item" style="padding-bottom:5px">
                  结束时间：{{ (new Date(item.finished_at)).getFullYear()+"-"+((new Date(item.finished_at)).getMonth()+1)+"-"+(new Date(item.finished_at)).getDate()+" "+(new Date(item.finished_at)).getHours()+":"+(new Date(item.finished_at)).getMinutes() +":"+(new Date(item.finished_at)).getSeconds() }}
                </div>
              </el-col>
            </el-row>
            <el-row :gutter="20">
              <el-col>
                <div class="text item" style="">
                  内容：<pre>{{ item.output }}</pre>
                </div>
              </el-col>
            </el-row>

          </el-card>
        </div>
        <div style="text-align: center;margin-top: 20px;">
          <el-pagination
            :total="currentTotal"
            :page-size="10"
            background
            layout="prev, pager, next"
            @current-change="pageskip"/>

        </div>
      </div>
      <div slot="footer" class="dialog-footer">

        <el-button @click="dialogVisible=false">取 消</el-button>
      </div>
    </el-dialog>
  </div>
</template>
<script>
//  import { getSystems } from '../../api/system'
import {
  getJobNode,
  getNodeLog,
  getNodeLogTotal
}
  from '../../api/job'

export default {
  data() {
    return {
      logStatus: true,
      jobConfig: {},
      dialogVisible: false,
      tableData: [],
      logList: [],
      workingList: [],
      activeLog: [],
      currentPage: 1,
      currentNode: '',
      currentTotal: 0,
      keyword:'',
      keywordChange:false,
    }
  },

  created() {
    this.initialize()
  },
  watch: {
      keyword(val) {
          this.keywordChange = true
      }
  },

  methods: {
    initialize() {
      var lett=this;
      document.onkeydown = function (e) {
          if (e.keyCode === 13){
              lett.searchLog()
          }
      }
      this.getNodeList()
    },

    pageskip(page) {
      console.log(page)
      this.logStatus = true
      this.NodeLog(this.currentNode, page)
    },
    Dialog(node) {
      this.currentNode = node
      this.dialogVisible = true
      this.logStatus = true
      this.keyword = ''
      var thisthis = this
      /*getNodeLogTotal(this.$route.params.id, node,thisthis.keyword).then(function(back) {
        console.log(back)
        thisthis.currentTotal = back.data
        thisthis.NodeLog(node, 1,thisthis.keyword)
      })*/
      this.NodeLog(node,1,this.keyword)
    },
    getNodeList() {
      var id = this.$route.params.id

      getJobNode(id).then(res => {
        res.data.forEach(val => {
          var stats = {
            Name: val.node,
            successJobs: isNaN(val.stats.success) ? 0 : parseInt(val.stats.success),
            failJobs: isNaN(val.stats.failed) ? 0 : parseInt(val.stats.failed),
            successExec: isNaN(val.stats.day_success) ? 0 : parseInt(val.stats.day_success),
            failExec: isNaN(val.stats.day_failed) ? 0 : parseInt(val.stats.day_failed)
          }
          this.tableData.push(stats)
        })
      })
    },
    NodeLog(node, page,keyword) {
      this.currentPage = page
      getNodeLog(this.$route.params.id, node, page,keyword).then(res => {
        // console.log(res.data)
        this.logList = res.data.log
        this.logStatus = false
        if (page==1){
            this.currentTotal=res.data.count
        }
      })
    },
    searchLog(){
        if(this.keywordChange){
            this.logStatus=true
            this.NodeLog(this.currentNode,1,this.keyword)
            this.keywordChange=false
        }else {
            return
        }
    }
  }
}
</script>


<style lang="scss" scoped>
  .el-row {
    margin-bottom: 5px;
    &:last-child {
      margin-bottom: 0;
    }
  }
  .tag-dep {
    margin-left: 5px;
  }
  .button-add-dep {
    width: 130px;
  }
</style>
