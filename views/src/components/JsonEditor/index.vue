<template>
  <div ref="jsoneditor" style="height:100%;"/>
</template>

<script>
// outter 数据变动通知 innerData
// innerData 变动不能再通知outter

import ace from 'ace-builds'
import 'ace-builds/webpack-resolver' // 在 webpack 环境中使用必须要导入
import 'ace-builds/src-noconflict/theme-monokai' // 默认设置的主题
import 'ace-builds/src-noconflict/mode-javascript' // 默认设置的语言模式

import JSONEditor from 'jsoneditor/dist/jsoneditor-minimalist.js'
import 'jsoneditor/dist/jsoneditor.min.css'
import _ from 'lodash'

export default {
  name: 'JsonEditor',
  // outter用于外部通知内部 innerData用于内部通知外部
  props: {
    outter: {
      required: true,
      type: [Object, String, Array]
    },
    options: {
      type: Object,
      default: () => {
        return {}
      }
    },
    onChange: {
      type: Function,
      default: () => {
        return () => { }
      }
    }
  },
  data() {
    return {
      innerData: null,
      editor: null
    }
  },
  watch: {
    outter(val) {
      console.log(['parent change outter', val])
      this.editor.set(val)
      this.innerData = this.editor.get()
      this.$emit('inner-change', this.innerData)
    }
  },
  mounted() {
    const container = this.$refs.jsoneditor
    const options = _.extend({
      ace: ace,
      mode: 'code',
      modes: ['code', 'text'],
      onChange: this._onChange
    }, this.options)

    this.editor = new JSONEditor(container, options)
    this.editor.set(this.outter)
  },
  beforeDestroy() {
    if (this.editor) {
      this.editor.destroy()
      this.editor = null
    }
  },
  methods: {
    _onChange(e) {
      this.$emit('inner-change', this.editor.get())
    }
  }
}
</script>

<style>
</style>
