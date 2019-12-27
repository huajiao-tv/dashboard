<template>
  <el-card shadow="hover">
    <div slot="header" class="clearfix">
      <span>pepper_bus</span>
    </div>
    <div class="panel-group">
      <div class="card-panel-col">
        <div class="card-panel" @click="handleTopic">
          <div class="card-panel-icon-wrapper icon-topic">
            <svg-icon icon-class="topic" class-name="card-panel-icon" />
          </div>
          <div class="card-panel-description">
            <div class="card-panel-text">Topic</div>
            <count-to :start-val="0" :end-val="topic" :duration="2000" class="card-panel-num"/>
          </div>
        </div>
      </div>
      <div class="card-panel-col">
        <div class="card-panel" @click="handleMachine">
          <div class="card-panel-icon-wrapper icon-machine">
            <svg-icon icon-class="servers" class-name="card-panel-icon" />
          </div>
          <div class="card-panel-description">
            <div class="card-panel-text">消费机器</div>
            <count-to :start-val="0" :end-val="machine" :duration="2000" class="card-panel-num"/>
          </div>
        </div>
      </div>
      <div class="card-panel-col">
        <div class="card-panel" @click="handleStorage">
          <div class="card-panel-icon-wrapper icon-storage">
            <svg-icon icon-class="storage" class-name="card-panel-icon" />
          </div>
          <div class="card-panel-description">
            <div class="card-panel-text">存储</div>
            <count-to :start-val="0" :end-val="storage" :duration="2000" class="card-panel-num"/>
          </div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script>
import CountTo from 'vue-count-to'

export default {
  name: 'SystemCard',

  components: {
    CountTo
  },

  props: {
    name: {
      type: String,
      default: ''
    },
    machine: {
      type: Number,
      default: 0
    },
    storage: {
      type: Number,
      default: 0
    },
    topic: {
      type: Number,
      default: 0
    }
  },

  methods: {
    handleMachine() {
      this.$router.push({
        name: 'Cluster_Bus',
        query: {
          system: this.name
        }
      })
    },
    handleStorage() {
      this.$router.push({
        name: 'Bus_Storage',
        query: {
          system: this.name
        }
      })
    },
    handleTopic() {
      this.$router.push({
        name: 'Topic',
        query: {
          system: this.name
        }
      })
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss" scoped>
.panel-group {
  .card-panel-col {
    margin-bottom: 10px;
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
      .icon-machine {
        background: #40c9c6;
      }
      .icon-topic {
        background: #36a3f7;
      }
      .icon-storage {
        background: #f4516c;
      }
    }
    .icon-machine {
      color: #40c9c6;
    }
    .icon-topic {
      color: #36a3f7;
    }
    .icon-storage {
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
