<template>
  <el-card v-show="show" shadow="hover" style="margin: 20px">
    <div slot="header" class="clearfix">
      <span><b>{{ queue.desc }}</b> - {{ queue.name }}</span>
      <el-tag type="info">{{ queue.desc }}</el-tag>
      <el-tag v-if="queue.author!=''" type="info">{{ queue.author }}</el-tag>
      <el-dropdown style="float:right" @command="qDropDownHandle">
        <span class="el-dropdown-link">
          操作<i class="el-icon-arrow-down el-icon--right"/>
        </span>
        <el-dropdown-menu slot="dropdown">
          <el-dropdown-item command="createTopic">添加topic</el-dropdown-item>
          <el-dropdown-item command="update">修改queue</el-dropdown-item>
          <el-dropdown-item command="del">删除queue</el-dropdown-item>
        </el-dropdown-menu>
      </el-dropdown>
    </div>
    <el-row :gutter="40" class="panel-group">
      <el-col :xs="12" :sm="12" :lg="12" class="card-panel-col">
        <div class="card-panel" @click="handleTopic">
          <div class="card-panel-icon-wrapper icon-topic">
            <svg-icon icon-class="topic" class-name="card-panel-icon" />
          </div>
          <div class="card-panel-description">
            <div class="card-panel-text">Topic</div>
            <count-to :start-val="0" :end-val="queue.topics" :duration="2000" class="card-panel-num"/>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :lg="12" class="card-panel-col">
        <div class="card-panel" @click="handleTopic">
          <div class="card-panel-icon-wrapper icon-qps">
            <svg-icon icon-class="qps" class-name="card-panel-icon" />
          </div>
          <div class="card-panel-description">
            <div class="card-panel-text">QPS</div>
            <count-to :start-val="0" :end-val="queue.qps" :duration="2000" class="card-panel-num"/>
          </div>
        </div>
      </el-col>
    </el-row>
  </el-card>
</template>

<script>
import CountTo from 'vue-count-to'
import { delQueue } from '../../api/queue'

export default {
  name: 'QueueCard',

  components: {
    CountTo
  },

  props: {
    queue: {
      type: Object,
      default: function() {
        return {}
      }
    }
  },

  data() {
    return {
      dialogAddTopic: false,
      formLabelWidth: '120px',
      show: true,
      form: {
        name: '',
        desc: '',
        password: ''
      }
    }
  },

  methods: {
    qDropDownHandle(cmd) {
      if (cmd === 'del') {
        if (confirm('是否删除queue:' + this.queue.name + '?')) {
          delQueue(this.queue).then(res => {
            this.show = false
          })
        }
      } else if (cmd === 'update') {
        this.$emit('queueItemAsk', { action: 'updateQueue', data: this.queue })
      } else if (cmd === 'createTopic') {
        this.$emit('queueItemAsk', { action: 'createTopic', data: this.queue })
      } else {
        console.log(cmd)
      }
    },
    handleTopic() {
      this.$router.push({
        path: '/topic/index',
        query: {
          queue: this.queue.name
        }
      })
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss" scoped>
.panel-group {
  margin-top: 18px;
  .card-panel-col {
    margin-bottom: 32px;
  }
  .card-panel {
    height: 108px;
    cursor: pointer;
    font-size: 12px;
    position: relative;
    overflow: hidden;
    color: #666;
    background: #fff;
    box-shadow: 4px 4px 40px rgba(0, 0, 0, 0.05);
    border-color: rgba(0, 0, 0, 0.05);
    &:hover {
      .card-panel-icon-wrapper {
        color: #fff;
      }
      .icon-topic {
        background: #36a3f7;
      }
      .icon-qps {
        background: #f4516c;
      }
    }
    .icon-topic {
      color: #36a3f7;
    }
    .icon-qps {
      color: #f4516c;
    }
    .card-panel-icon-wrapper {
      float: left;
      margin: 14px 0 0 14px;
      padding: 16px;
      transition: all 0.38s ease-out;
      border-radius: 6px;
    }
    .card-panel-icon {
      float: left;
      font-size: 48px;
    }
    .card-panel-description {
      float: right;
      font-weight: bold;
      margin: 26px;
      margin-left: 0px;
      .card-panel-text {
        line-height: 18px;
        color: rgba(0, 0, 0, 0.45);
        font-size: 16px;
        margin-bottom: 12px;
      }
      .card-panel-num {
        font-size: 20px;
      }
    }
  }
}
</style>
