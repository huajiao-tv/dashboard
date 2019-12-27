<template>
  <div>
    <el-alert
      v-if="tooManyWarning"
      :title="'本页加载' + (tableData.length?tableData.length : '30+') + '个topic，加载有些缓慢，试试使用搜索功能'"
      type="warning"
      center
      show-icon/>

    <topic-list v-loading="loadingTopic" :table-data="tableData" :update-callback="initialize" />
  </div>
</template>

<script>
import TopicList from './topicList'
import { getTopic } from '../../api/topic'

export default {
  components: {
    TopicList
  },

  data() {
    return {
      loadingTopic: false,
      system: '',
      queue: '',
      tableData: [],
      tooManyWarning: false
    }
  },

  watch: {
    dialogTest(val) {
      val || (this.topicReturn = '')
    }
  },

  created() {
    this.system = this.$route.query.system
    this.queue = this.$route.query.queue
    this.initialize()
  },
  methods: {
    initialize() {
      this.loadingTopic = true
      getTopic(this.system, this.queue).then(res => {
        this.loadingTopic = false

        this.tableData = res.data
        if (this.tableData.length > 0) {
          this.tableData.forEach(one => {
            one.machines.forEach(m => {
              m.queue = one.queue
              m.topic = one.name
            })
          })
          // this.activeName = this.tableData[0].queue + '/' + this.tableData[0].name
        } else {
          this.$notify({
            title: 'topic 不存在',
            message: '未找到对应的 topic，请确认是否建立了topic并查看权限问题',
            type: 'success',
            duration: 2000
          })
        }
        if (res.data.length < 30) {
          this.tooManyWarning = false
        }
      }).catch(error => {
        this.loadingTopic = false
        console.log(error)
      })
    }
  }
}
</script>
