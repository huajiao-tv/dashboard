<template>
  <el-dialog :close-on-click-modal="false" :visible.sync="dialogAddTopic" title="添加 Topic">
    <el-form ref="topicForm" :model="topicItem" :rules="addTopicRules">
      <el-form-item :label-width="formLabelWidth" label="系统" prop="system">
        <el-select v-model="topicItem.system" filterable placeholder="请选择所属系统" @change="topicItem.storage=null">
          <el-option
            v-for="system in systems"
            :key="system.name"
            :label="system.desc + ' - ' + system.name"
            :value="system.name"/>
        </el-select>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="队列" prop="queue">
        <el-select v-model="topicItem.queue" filterable placeholder="请选择所属队列">
          <el-option
            v-for="(desc, name) in queueOptions"
            :key="name"
            :label="desc + ' - ' + name "
            :value="name"/>
        </el-select>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="存储" prop="storage">
        <el-select v-model="topicItem.storage" placeholder="请选择使用的存储">
          <el-option
            v-for="storage in storages[topicItem.system]"
            :key="storage.id"
            :label="storage.desc"
            :value="storage.id"/>
        </el-select>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="topic 名称" prop="name">
        <el-input v-model="topicItem.name" placeholder="topic 名称，不要使用中文，例如：default"/>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="topic 说明" prop="desc">
        <el-input v-model="topicItem.desc" placeholder="topic 中文说明，尽量简洁，例如：开播提醒"/>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="密码" prop="password">
        <el-input v-model="topicItem.password" placeholder="topic 消费密码，可不填写，默认与queue密码一致"/>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="开启自动重试" prop="is_retry">
        <el-switch
          v-model="retryConf.switch"
          active-color="#13ce66"
          inactive-color="#ff4949"
          @change="retryChange"/>
      </el-form-item>
      <el-form-item v-if="retryConf.switch" :label-width="formLabelWidth" label="重试有效期" prop="is_retry">
        <el-input v-model="retryConf.durTime" placeholder="单位为秒"/>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="运行类型" prop="run_type">
        <el-select v-model="topicItem.run_type" placeholder="请选择运行类型">
          <el-option key="0" :value="0" label="CGI类型(php-fpm)"/>
          <el-option key="1" :value="1" label="HTTP类型(golang,java)"/>
        </el-select>
      </el-form-item>
      <el-form-item v-if="topicItem.run_type === 1" :label-width="formLabelWidth" prop="http_config" label="host">
        <el-input v-model.trim="httpConfig.host" placeholder=" '  域名 '"/>
      </el-form-item>

      <el-form-item :label-width="formLabelWidth" :label="topicItem.run_type ? 'url地址' : '消费文件'" prop="consume">
        <el-input v-model.trim="topicItem.consume" :placeholder="topicItem.run_type ? 'http请求地址' : '绝对路径，例如：/home/q/system/pepper/live/front/src/worker/liveStartPushBatchWorker.php'"/>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="操作注释" prop="comment">
        <el-input v-model="topicItem.comment" placeholder="操作注释"/>
      </el-form-item>
      <el-form-item :label-width="formLabelWidth" label="消费状态" prop="status">
        <el-radio-group v-model="topicItem.status">
          <el-radio :label="0">开启</el-radio>
          <el-radio :label="1">关闭</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    <div slot="footer" class="dialog-footer">
      <el-button @click="dialogAddTopic = false">取 消</el-button>
      <el-button type="primary" @click="saveTopic">确 定</el-button>
    </div>
  </el-dialog>
</template>

<script>
import { addTopic } from '@/api/topic'
import { deepClone } from '@/utils/index.js'
export default {
  props: {
    storages: { type: Object, default: () => { return {} } },
    systems: { type: Array, default: () => { return [] } },
    queueOptions: { type: Object, default: () => { return {} } },
    toModifyItem: { type: Object, default: () => {
      return {
        system: '',
        queue: '',
        storage: '',
        name: '',
        desc: '',
        password: '',
        consume: '',
        comment: '',
        status: 0
      }
    } }
  },
  data() {
    return {
      retryConf: {
        switch: false,
        durTime: 86400
      },
      dialogAddTopic: false,
      formLabelWidth: '120px',
      httpConfig: {},
      topicItem: {
        system: '',
        queue: '',
        storage: '',
        name: '',
        desc: '',
        password: '',
        consume: '',
        comment: '',
        status: 0,
        run_type: -1
      },
      addTopicRules: {
        system: [{ required: true, message: '请选择系统', trigger: 'blur' }],
        queue: [{ required: true, message: '请选择队列', trigger: 'blur' }],
        storage: [{ required: true, message: '请选择存储', trigger: 'blur' }],
        name: [
          { required: true, message: '请填写 topic 名称', trigger: 'blur' }
        ],
        desc: [
          { required: true, message: '请填写 topic 描述', trigger: 'blur' }
        ],
        consume: [
          { required: true, message: '请填写 topic 消费文件', trigger: 'blur' },
          { validator: (rule, value, callback) => {
            if (this.topicItem.run_type === 1) {
              if (!value.match(/https?:\/\/[a-z0-9A-Z.-_:%#&]{10,}$/)) {
                callback(new Error('类型为http时，必须填写url'))
                return
              }
            }
            callback()
          } }
        ],
        comment: [{ required: true, message: '请填写注释', trigger: 'blur' }],
        status: [{ required: true, message: '请选择消费状态', trigger: 'blur' }],
        run_type: [{ required: true, message: '请选择运行类型', trigger: 'blur' }]
      }
    }
  },
  methods: {

    retryChange() {
      console.log(this.retryConf)
    },
    mount() {
      this.topicItem = deepClone(this.toModifyItem)
      this.retryConf = JSON.parse(this.topicItem.topic_config)
    },
    resetTopicForm() {
      this.topicItem = {
        system: '',
        queue: '',
        storage: '',
        name: '',
        desc: '',
        password: '',
        consume: '',
        comment: ''
      }
      this.$nextTick(() => {
        this.$refs['topicForm'].clearValidate()
      })
    },
    saveTopic() {
      this.$refs['topicForm'].validate(valid => {
        if (!valid) {
          console.log('不行', valid)
          return
        }
        this.topicItem.http_config = JSON.stringify(this.httpConfig)
        this.retryConf.durTime = parseInt(this.retryConf.durTime)
        this.topicItem.topic_config = JSON.stringify(this.retryConf)
        addTopic(this.topicItem).then(res => {
          if (res.code === 0) {
            this.$notify({
              title: '成功',
              message: '保存成功',
              type: 'success',
              duration: 2000
            })
            this.dialogAddTopic = false
            this.$emit('topicDialogAsk', 'saved')
          }
        })
      })
    }
  }
}
</script>

