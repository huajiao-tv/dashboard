<template>
  <div style="padding: 20px">
    <el-button type="primary" @click="dialogAdd = true">添加系统</el-button>
    <el-button type="primary" @click="topicDialogHandle('add')">添加 topic</el-button>
    <el-row v-loading="loadingSystem">
      <el-card v-for="system in systems" :key="system.name" shadow="nvever" style="margin-top: 20px">
        <div slot="header" class="clearfix">
          <span><b>{{ system.desc }}</b> - {{ system.name }}</span>
        </div>
        <el-row :gutter="40">
          <el-col :span="12">
            <bus-card
              :name="system.name"
              :machine="system.machines"
              :storage="system.storages"
              :topic="system.topics"/>
          </el-col>
          <el-col :span="12">
            <cron-card
              :name="system.name"
              :machine="system.cron_machines"
              :jobs="system.jobs"
              :success="system.job_success_count"
              :failure="system.job_fail_count"
            />
          </el-col>
        </el-row>
      </el-card>
    </el-row>

    <el-dialog :visible.sync="dialogAdd" title="添加系统">
      <el-form ref="systemForm" :rules="addSystemRules" :model="systemItem">
        <el-form-item :label-width="formLabelWidth" label="系统名称" prop="name">
          <el-input v-model="systemItem.name" placeholder="系统名称，不要使用中文，例如：live"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="系统说明" prop="desc">
          <el-input v-model="systemItem.desc" placeholder="系统中文说明，尽量简洁，例如：直播"/>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogAdd = false">取 消</el-button>
        <el-button type="primary" @click="addSystem">确 定</el-button>
      </div>
    </el-dialog>

    <topicEditor ref="topicDialog" :storages="storages" :systems="systems" :queue-options="queueOptions" @topicDialogAsk="topicDialogHandle"/>
  </div>
</template>
                    $this->handler = explode(",", $value);

<script>
import topicEditor from '@/views/topic/topic-editor.vue'
import BusCard from './bus-card'
import CronCard from './cron-card'
import { getStorages } from '../../api/storage'
import { getQueues } from '../../api/queue'
import { getSystems, addSystem } from '../../api/system'
// import { getMachines } from '../../api/cluster'
// import {
//   getJobs
// }
//   from '../../api/job'
export default {
  components: {
    BusCard, CronCard, topicEditor
  },

  data() {
    return {
      loadingSystem: false,
      systems: [],
      storages: {},
      queueOptions: {},
      dialogAdd: false,
      formLabelWidth: '120px',
      systemItem: {
        name: '',
        desc: ''
      },
      addSystemRules: {
        name: [{ required: true, message: '请填写系统名称', trigger: 'blur' }],
        desc: [{ required: true, message: '请填写系统描述', trigger: 'blur' }]
      },
      TaskMashions: {}
    }
  },

  watch: {
    dialogAdd(val) {
      val || this.resetForm()
    }
  },

  created() {
    this.initialize()
  },

  methods: {
    initialize() {
      this.loadingSystem = true
      getSystems().then(res => {
        this.loadingSystem = false
        this.systems = res.data
      }).catch(error => {
        this.loadingSystem = false
        console.log(['system initialize error', error])
      })
      getQueues().then(res => {
        res.data.forEach(queue => {
          this.queueOptions[queue.name] = queue.desc
        })
      })
      getStorages().then(res => {
        res.data.forEach(st => {
          if (this.storages[st.system] === undefined) {
            this.storages[st.system] = []
          }
          this.storages[st.system].push({
            id: st.id,
            desc: st.host + ':' + st.port + '(' + st.desc + ')'
          })
        })
      })
    },

    resetForm() {
      this.systemItem.name = ''
      this.systemItem.desc = ''
      this.$refs['systemForm'].clearValidate()
    },

    addSystem() {
      this.$refs['systemForm'].validate(valid => {
        if (valid) {
          addSystem(this.systemItem).then(res => {
            if (res.code === 0) {
              this.$notify({
                title: '成功',
                message: '保存成功',
                type: 'success',
                duration: 2000
              })
              this.dialogAdd = false
              this.initialize()
            }
          })
        }
      })
    },
    topicDialogHandle(ask) {
      if (ask === 'add') {
        this.$refs.topicDialog.resetTopicForm()
        this.$refs.topicDialog.dialogAddTopic = true
      } else if (ask === 'saved') {
        this.initialize()
      }
    }
  }
}
</script>
