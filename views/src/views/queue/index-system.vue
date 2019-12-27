<template>
  <div style="padding: 20px">
    <el-button type="primary" @click="wakeDialog('queue', 'create')">添加 Queue</el-button>
    <el-button type="primary" @click="topicDialogHandle('add')">添加 Topic</el-button>
    <el-button :loading="loading.export" type="success" @click="exportQueue()">批量导出</el-button>
    <el-button type="success" @click="importData.visible=true">批量导入</el-button>
    <el-tabs v-model="activeName" @tab-click="setQuey">
      <template v-for="item in queuesSystem" >
        <el-tab-pane :label="item.system.name + ' - ' + item.system.desc">
          <el-card class="box-card">
              <span>包含该系统的topic的queue列表</span>
          <el-table v-loading="queueLoading" :data="item.queue" :default-sort="{prop:'name', order:'ascending'}" prop="name" style="margin-top:20px;" @selection-change="handleSelectionChange">
            <el-table-column type="selection" width="45" />
            <el-table-column label="队列名称" prop="name" sortable/>
            <el-table-column label="说明">
              <template slot-scope="scope">
                <el-tooltip :content="scope.row.comment" class="item" effect="dark" placement="top-start">
                  <i class="el-icon-info"/>
                </el-tooltip>
                {{ scope.row.desc }}
              </template>
            </el-table-column>
            <el-table-column label="Topics" prop="topics" width="80">
              <template slot-scope="scope">
                <el-button type="primary" size="mini" @click="redirectTopic(scope.row.name)">{{ scope.row.topics }}</el-button>
              </template>
            </el-table-column>
            <el-table-column label="QPS" prop="qps" width="80" sortable/>
            <el-table-column label="Author" prop="author" width="200" sortable/>

            <el-table-column label="操作" width="250">
              <template slot-scope="scope">
                <el-button size="mini" @click="wakeDialog('topic', 'create', scope.row)">添加topic</el-button>
                <el-button size="mini" @click="wakeDialog('queue', 'update', scope.row)">编辑</el-button>
                <el-button size="mini" type="danger" @click="removeQueue(scope.row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          </el-card>
        </el-tab-pane>
      </template>
    </el-tabs>

    <topicEditor ref="topicDialog" :storages="storages" :systems="systems" :queue-options="queueOptions" @topicDialogAsk="topicDialogHandle"/>

    <el-dialog :visible.sync="dialogQueue.visible" :title="dialogQueue.title">
      <el-form ref="queueForm" :model="queueItem" :rules="saveQueueRules">
        <el-form-item :label-width="formLabelWidth" label="队列名称" prop="name">
          <el-input v-model="queueItem.name" placeholder="队列名称，不要使用中文，例如：live_start"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="队列说明" prop="desc">
          <el-input v-model="queueItem.desc" placeholder="队列中文说明，尽量简洁，例如：开播"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="队列认证" prop="password">
          <el-input v-model="queueItem.password" placeholder="队列写入密码"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="操作注释" prop="comment">
          <el-input v-model="queueItem.comment" placeholder="操作注释"/>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogQueue.visible = false">取 消</el-button>
        <el-button type="primary" @click="saveQueue">确 定</el-button>
      </div>
    </el-dialog>

    <el-dialog :visible.sync="importData.visible" title="导入队列" width="900px">
      <el-row style="display:flex">
        <el-col :span="16" style="height:300px">
          <json-editor ref="jsonEditor" :outter="importData.data" @inner-change="loadInnerData"/>
        </el-col>
        <el-col :span="8" style="padding-left:10px;">
          <el-upload id="importJsonFile" :auto-upload="false" :show-file-list="false" :on-change="readImportJson" class="upload-demo" drag action="#" >
            <i class="el-icon-upload"/>
            <div class="el-upload__text">将文件拖到此处，或<em>点击选择文件</em></div>
          </el-upload>
          <el-button @click="importData.visible = false">关闭</el-button>
          <el-button :loading="loading.import" type="primary" @click="importQueue">导入</el-button>

          <el-alert title="" type="info" style="margin-top:10px;">
            如果依赖的system和storage不存在则无法导入<br>
            已存在的queue和topic不会被覆盖；
          </el-alert>
        </el-col>
      </el-row>
      <div v-if="importData.resList.length">
        <el-table :data="importData.resList" style="width: 100%" height="250">
          <el-table-column fixed prop="queue" width="200" label="Queue"/>
          <el-table-column prop="topic" width="200" label="Topic"/>
          <el-table-column prop="res" label="导入结果">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.res === 'success'" type="success" >success</el-tag>
              <el-tag v-else type="danger">{{ scope.row.res }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<style>
  #importJsonFile .el-upload-dragger {
    width: 270px;
  }
</style>

<script>
    import QueueCard from './card'
    import { getSystems } from '../../api/system'
    import { getQueues,getQueueSystem, addQueue, updateQueue, delQueue, exportQueue, importQueue } from '../../api/queue'
    import { getStorages } from '../../api/storage'
    import topicEditor from '@/views/topic/topic-editor.vue'
    import { deepClone } from '@/utils/index'
    import jsonEditor from '@/components/JsonEditor'

    export default {
        components: {
            QueueCard, topicEditor, jsonEditor
        },

        data() {
            return {
                exportSelection: [], // 要导出的数据
                importData: { visible: false, data: [], realData: [], resList: [] },
                loading: { export: false, import: false },
                queueLoading: false,
                systems: [],
                storages: {},
                queueOptions: {},
                topicItem: { queue: '' }, // 必须已存在的属性才能双向绑定 否则赋值后 select中其他选项会失效
                queues: [],
                queuesSystem: [],
                dialogQueue: {
                    title: '添加Queue',
                    visible: false
                },
                formLabelWidth: '120px',
                queueItem: {
                    id: '',
                    name: '',
                    desc: '',
                    password: '',
                    comment: ''
                },
                systemOptions: {},
                saveQueueRules: {
                    // name password 校验比较特殊 所以在wakeDialog中处理
                    desc: [{ required: true, message: '请填写队列说明', trigger: 'blur' }],
                    comment: [{ required: true, message: '请填写操作注释', trigger: 'blur' }]
                }
            }
        },

        created() {
            this.initialize()
        },

        methods: {
            initialize() {
                getSystems().then(res => {
                    this.systems = res.data
                })
                if (!this.queueLoading) {
                    this.queueLoading = true
                    getQueueSystem().then(res => {
                        this.queueLoading = false
                        this.queuesSystem=res.data
                        res.data.forEach(item => {
                            item.queue.forEach(queue => {
                                this.queueOptions[queue.name] = queue.desc
                                }
                            )
                        })
                    }).catch(error => {
                        this.queueLoading = false
                        console.log(error)
                    })
                }

                getStorages().then(res => {
                    this.storages = {}
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
            redirectTopic(queue) {
                this.$router.push({
                    path: '/topic/index',
                    query: {
                        queue: queue
                    }
                })
            },
            removeQueue(data) {
                if (confirm('是否删除queue:' + data.name + '?')) {
                    delQueue(data).then(res => {
                        this.initialize()
                    })
                }
            },
            resetQueueForm(data) {
                const { id, name, desc, comment } = data || {}
                this.queueItem = {
                    id: id,
                    name: name,
                    desc: desc,
                    password: '',
                    comment: comment
                }
                // 如果直接获取refs的queueForm会失败
                this.$nextTick(() => {
                    this.$refs.queueForm.clearValidate()
                })
            },
            // 呼起form对话框 用于添加修改queue 添加topic
            wakeDialog(source, action, data) {
                if (source === 'queue') {
                    this.dialogQueue.visible = true
                    if (action === 'create') {
                        this.dialogQueue.title = '添加 Queue'
                        this.resetQueueForm()

                        this.saveQueueRules.password = [{ required: true, message: '请填写队列密码', trigger: 'blur' }]
                        this.saveQueueRules.name = [{ required: true, message: '请填写队列名', trigger: 'blur' }]
                    } else {
                        this.dialogQueue.title = '修改 Queue'
                        this.resetQueueForm(deepClone(data))

                        delete this.saveQueueRules.password
                        delete this.saveQueueRules.name
                    }
                } else if (source === 'topic' && action === 'create') {
                    this.topicDialogHandle('add')
                    if (data.name) {
                        this.$refs.topicDialog.topicItem.queue = data.name
                    }
                }
            },
            queueItemAskHandle(params) {
                if (params.action === 'updateQueue') {
                    this.wakeDialog('queue', 'update', params.data)
                } else if (params.action === 'createTopic') {
                    this.wakeDialog('topic', 'create', params.data)
                }
            },
            saveQueue() {
                this.$refs['queueForm'].validate(valid => {
                    if (!valid) return
                    this.queueItem.id > 0
                        ? updateQueue(this.queueItem).then(res => {
                            if (res.code === 0) {
                                this.$notify({
                                    title: '成功',
                                    message: '保存成功',
                                    type: 'success',
                                    duration: 2000
                                })
                                this.dialogQueue.visible = false
                                this.initialize()
                            }
                        }) : addQueue(this.queueItem).then(res => {
                            if (res.code === 0) {
                                this.$notify({
                                    title: '成功',
                                    message: '保存成功',
                                    type: 'success',
                                    duration: 2000
                                })
                                this.dialogQueue.visible = false
                                this.initialize()
                            }
                        })
                })
            },
            topicDialogHandle(ask) {
                if (ask === 'add') {
                    this.$refs.topicDialog.resetTopicForm()
                    this.$refs.topicDialog.dialogAddTopic = true
                } else if (ask === 'saved') {
                    this.initialize()
                }
            },
            // jsoneditor 数据修改后会调用loadInnerData
            loadInnerData(val) {
                this.importData.realData = val
            },
            handleSelectionChange(val) {
                this.exportSelection = val
            },
            exportQueue() {
                if (this.exportSelection.length === 0) {
                    this.$message.error('请选择要导出的数据')
                    return
                }
                var ids = []
                this.exportSelection.forEach(val => {
                    ids.push(val.id)
                })

                if (this.loading.export) return
                this.loading.export = true

                exportQueue({ ids: ids }).then(res => {
                    if (res.code === 0) {
                        const content = JSON.stringify(res.data)
                        const blob = new Blob([content])
                        const fileName = `export-queue.json`
                        if ('download' in document.createElement('a')) {
                            console.log('非IE下载')
                            const elink = document.createElement('a')
                            elink.download = fileName
                            elink.style.display = 'none'
                            elink.href = URL.createObjectURL(blob)
                            document.body.appendChild(elink)
                            elink.click()
                            URL.revokeObjectURL(elink.href) // 释放URL 对象
                            document.body.removeChild(elink)
                        } else {
                            console.log('IE10+下载')
                            navigator.msSaveBlob(blob, fileName)
                        }
                    } else {
                        this.$message.error(res.message)
                    }

                    this.loading.export = false
                }).catch(err => {
                    this.loading.export = false
                    console.log(err)
                })
            },
            readImportJson(file, fileList) {
                console.log(file, fileList)
                fileList = [file]

                var reader = new FileReader()
                reader.onloadend = (evt) => {
                    if (evt.target.readyState === FileReader.DONE) {
                        var base64Str = evt.target.result.split(',')[1]
                        this.importData.data = JSON.parse(decodeURIComponent(escape(window.atob(base64Str))))
                    }
                }
                reader.readAsDataURL(file.raw)
            },
            importQueue() {
                if (this.importData.realData.length <= 0) {
                    this.$message.error('请选择要导入的数据')
                    return
                }
                if (this.loading.import) return
                this.loading.import = true
                this.importData.resList = []
                importQueue({ data: this.importData.realData }).then(res => {
                    this.loading.import = false
                    if (res.code === 0) {
                        this.importData.resList = res.data
                    } else {
                        this.$message.error(res.message)
                    }
                }).catch(error => {
                    console.log(error)
                    this.loading.import = false
                })
            }
        }
    }
</script>
