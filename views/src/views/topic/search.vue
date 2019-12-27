<template>
  <div>
    <el-row style="padding-top:5px">
      <el-col :span="6">
        <el-form ref="form" label-width="80px" size="mini">
          <el-form-item label="搜索" size="small">
            <el-input v-model="keyword" placeholder="同时搜索queue和topic可以写成 queue/topic"/>
          </el-form-item>
        </el-form>
      </el-col>
      <el-col :span="6" style="margin-left:4px;">
        <el-tag v-if="system" size="smal" type="success" closable @close="system='';wordChange()" >system:{{ system }}</el-tag>
        <el-tag v-if="queue" type="success" size="smal" closable @close="queue='';wordChange()" >queue:{{ queue }}</el-tag>
        <el-tag v-if="topic" type="success" size="smal" closable @close="topic='';wordChange()" >topic:{{ topic }}</el-tag>
      </el-col>
    </el-row>
    <div>
      <el-alert
        v-if="tooManyWarning"
        :title="'本页加载只' + (tableData.length?tableData.length : '30+') + '个topic，加载有些缓慢，可以输入完整词'"
        type="warning"
        center
        show-icon
      />
      <el-alert
        v-if="nothing"
        :title="'无结果'"
        type="warning"
        center
        show-icon
      />
      <topic-list v-loading="loadingTopic" :table-data="tableData" :update-callback="wordChange" />
    </div>
  </div>
</template>

<script>
import TopicList from './topicList'
import JsonEditor from '../../components/JsonEditor/index'
import { searchTopic } from '../../api/topic'

export default {
  components: {
    JsonEditor,
    TopicList
  },

  data() {
    return {
      loadingTopic: false,
      tooManyWarning: true,
      nothing: false,
      system: '',
      queue: '',
      topic: '',
      keyword: '',
      tableData: []
    }
  },

  watch: {
    keyword(val) {
      if (this.keyword !== undefined && this.keyword !== '' && this.keyword !== false) {
        this.wordChange()
      } else {
        this.tooManyWarning = false
      }
    }
  },

  created() {
    this.system = this.$route.query.system
    this.queue = this.$route.query.queue
    this.keyword = this.$route.query.keyword
    if (this.keyword !== undefined && this.keyword !== '' && this.keyword !== false) {
      this.wordChange()
    } else {
      this.tooManyWarning = false
    }
  },

  methods: {
    wordChange() {
      this.loadingTopic = true
      this.nothing = false
      searchTopic(this.system, this.queue, this.topic, this.keyword).then(res => {
        this.loadingTopic = false
        this.tableData = res.data
        if (this.tableData.length > 0) {
          // this.activeName = this.tableData[0].queue + '/' + this.tableData[0].name
          this.current = this.keyword
        } else {
          this.nothing = true
        }
        if (res.data.length < 30) {
          this.tooManyWarning = false
        } else {
          this.tooManyWarning = true
        }
      }).catch(res => {
        this.loadingTopic = false
        this.tooManyWarning = false
      })
    }
  }
}
</script>
