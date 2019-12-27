<template>
  <div style="padding: 20px">
    <el-button type="primary" @click="dialogAdd = true">添加</el-button>
    <el-tabs v-model="activeName">
      <template v-for="(data, system) in tableData">
        <el-tab-pane :label="systemOptions[system] + ' - ' + system" :key="system" :name="system">
          <el-table
            :data="tableData[system]"
            stripe
          >
            <el-table-column
              prop="host"
              label="IP"/>
            <el-table-column
              prop="port"
              label="端口"/>
            <el-table-column
              prop="status"
              label="状态">
              <template slot-scope="scope">

                <el-tooltip v-if="scope.row.RunStatus === 'error'" :content="scope.row.Msg " class="item" effect="dark">

                  <el-tag size="mini" type="danger">异常  </el-tag>
                </el-tooltip>

                <el-tag v-if="scope.row.RunStatus !== 'error' " size="mini" type="success">正常</el-tag>
              </template>
            </el-table-column>
            <el-table-column

              prop="desc"
              label="说明"/>

            <el-table-column
              :formatter="qpsShow"
              label="ops_per_sec"/>

            <el-table-column
              :formatter="useMemoryShow"
              label="使用内存"/>
            <el-table-column
              :formatter="connShow"
              label="连接数"/>

            <el-table-column
              label="操作">
              <template slot-scope="scope">
                <el-button type="warning" size="mini" icon="el-icon-edit" circle @click="editStorage(scope.row)"/>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </template>
    </el-tabs>

    <el-dialog :visible.sync="dialogAdd" :title="editDialog ? '修改存储' : '新增存储'">
      <el-form ref="storageForm" :model="storageItem" :rules="addStorageRules">
        <el-form-item :label-width="formLabelWidth" label="系统" prop="system">
          <el-select v-model="storageItem.system" :disabled="editDialog ? true : false" placeholder="请选择系统">
            <el-option
              v-for="(desc, name) in systemOptions"
              :key="name"
              :label="desc + ' - ' + name"
              :value="name"/>
          </el-select>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="存储类型" prop="type">
          <el-select v-model="storageItem.type" placeholder="请选择存储类型">
            <el-option
              v-for="(desc, name) in stroageType"
              :key="name"
              :label="desc + ' - ' + name"
              :value="name"/>
          </el-select>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="IP" prop="host">
          <el-input v-model="storageItem.host" placeholder="存储对应的 IP"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="端口" prop="port">
          <el-input v-model.number="storageItem.port" placeholder="存储对应的端口"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="密码" prop="password">
          <el-input v-model="storageItem.password" placeholder="存储对应的密码"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="redis备注" prop="desc">
          <el-input v-model="storageItem.desc" placeholder="redis备注，简短说明"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="最大连接数" prop="max_conn_num">
          <el-input v-model.number="storageItem.max_conn_num" placeholder="最大连接数，默认100"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="最大空闲连接数" prop="max_idle_num">
          <el-input v-model.number="storageItem.max_idle_num" placeholder="最大空闲连接数，默认100"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="最大空闲时间" prop="max_idle_seconds">
          <el-input v-model.number="storageItem.max_idle_seconds" placeholder="最大空闲时间，默认30秒，3000000000"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="操作注释" prop="comment">
          <el-input v-model="storageItem.comment" placeholder="操作注释"/>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogAdd = false">取 消</el-button>
        <el-button type="primary" @click="addStorage">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getSystems } from '../../api/system'
import { getStorages, addStorage, updateStorage } from '../../api/storage'

export default {
  data() {
    return {
      activeName: '',
      tableData: {},
      systemOptions: {},
      stroageType: {
        redis: 'Redis'
      },
      addSystem: '',
      dialogAdd: false,
      editDialog: false,
      formLabelWidth: '120px',
      storageItem: {
        id: 0,
        system: '',
        host: '',
        port: '',
        password: '',
        comment: ''
      },
      addStorageRules: {
        system: [{ required: true, message: '请选择系统', trigger: 'blur' }],
        host: [{ required: true, message: '请填写域名或IP', trigger: 'blur' }],
        port: [
          {
            type: 'number',
            required: true,
            message: '请填写存储端口',
            trigger: 'blur'
          }
        ],
        password: [
          { required: false, message: '请填写存储密码', trigger: 'blur' }
        ],
        comment: [
          { required: true, message: '请填写操作注释', trigger: 'blur' }
        ],
        max_conn_num: [{ type: 'number' }],
        max_idle_num: [{ type: 'number' }],
        max_idle_seconds: [{ type: 'number' }]
      }
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
      getSystems().then(res => {
        res.data.forEach(system => {
          this.systemOptions[system.name] = system.desc
        })

        getStorages().then(res => {
          this.tableData = {}
          res.data.forEach(st => {
            if (this.tableData[st.system] === undefined) {
              this.tableData[st.system] = []
            }
            this.tableData[st.system].push(st)
          })
          if (this.$route.query.system !== undefined) {
            this.activeName = this.$route.query.system
          } else if (res.data.length > 0) {
            this.activeName = res.data[0].system
          }
        })
      })
    },
    showdata(row, column, cv, index) {
      if (row.RunStatus === 'error') {
        return row.desc + ':' + row.Msg
      } else {
        return row.desc
      }
    },
    qpsShow(row, column, cv, index) {
      if (row.RunStatus === 'error') {
        return ''
      }
      var reg2 = new RegExp('instantaneous_ops_per_sec\:([0-9]+)', 'gi')
      var lt2 = reg2.exec(row.RunStatus)
      return lt2.length >= 2 ? lt2[1] : ''
    },
    useMemoryShow(row, column, cv, index) {
      if (row.RunStatus === 'error') {
        return ''
      }
      var reg2 = new RegExp('used_memory\:([0-9]+)', 'gi')
      var lt2 = reg2.exec(row.RunStatus)

      return lt2.length >= 2 ? parseInt(lt2[1] / 1024 / 1024) + 'MB' : ''
    },
    connShow(row, column, cv, index) {
      if (row.RunStatus === 'error') {
        return ''
      }
      var reg = new RegExp('connected_clients\:([0-9]+)', 'gi')
      var lt2 = reg.exec(row.RunStatus)
      return lt2.length >= 2 ? lt2[1] : ''
    },
    resetForm() {
      this.storageItem = {
        id: 0,
        system: '',
        host: '',
        port: '',
        password: '',
        comment: ''
      }
      this.editDialog = false
      this.$refs['storageForm'].clearValidate()
    },
    addStorage() {
      this.$refs['storageForm'].validate(valid => {
        if (valid) {
          this.storageItem.port = parseInt(this.storageItem.port)
          if (this.editDialog) {
            updateStorage(this.storageItem).then(res => {
              if (res.code === 0) {
                this.$notify({
                  title: '成功',
                  message: '修改成功',
                  type: 'success',
                  duration: 2000
                })
                this.dialogAdd = false
                this.initialize()
              }
            })
          } else {
            addStorage(this.storageItem).then(res => {
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
        }
      })
    },
    editStorage(item) {
      this.editDialog = true
      this.storageItem = item
      this.dialogAdd = true
    }
  }
}
</script>
