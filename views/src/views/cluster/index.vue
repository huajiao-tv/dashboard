<template>
  <div style="padding: 20px">
    <el-button type="primary" @click="dialogAdd = true">添加</el-button>
    <el-tabs v-model="activeName">
      <el-tab-pane v-for="(data, name) in tableData" :key="name" :label="systemOptions[name] + ' - ' + name" :name="name">
        <el-table
          :data="data"
          stripe
        >
          <el-table-column
            prop="ip"
            label="IP"/>
          <el-table-column
            prop="status"
            label="状态">
            <template slot-scope="scope">
              <el-tag size="mini" type="success">正常</el-tag>
            </template>
          </el-table-column>
          <el-table-column
            label="操作">
            <template slot-scope="scope">
              <el-button type="danger" size="mini" icon="el-icon-delete" circle @click="deleteMachine(scope.row)"/>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog :visible.sync="dialogAdd" title="新增服务器">
      <el-form ref="clusterForm" :model="clusterItem" :rules="addMachineRules">
        <el-form-item :label-width="formLabelWidth" label="系统" prop="system">
          <el-select v-model="clusterItem.system" placeholder="请选择系统">
            <el-option
              v-for="(desc, name) in systemOptions"
              :key="name"
              :label="desc + ' - ' + name"
              :value="name"/>
          </el-select>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="节点列表" prop="ips">
          <el-input v-model="clusterItem.ips" type="textarea" rows="10" placeholder="节点列表，第行一个，例如：10.138.114.222:19840"/>
        </el-form-item>
        <el-form-item :label-width="formLabelWidth" label="操作注释" prop="comment">
          <el-input v-model="clusterItem.comment" placeholder="操作注释"/>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogAdd = false">取 消</el-button>
        <el-button type="primary" @click="addCluster">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getMachines, addMachines, delMachine } from '../../api/cluster'
import { getSystems } from '../../api/system'

export default {
  data() {
    return {
      forCron: false,
      activeName: '',
      tableData: {},
      systemOptions: {},
      addSystem: '',
      dialogAdd: false,
      formLabelWidth: '120px',
      clusterItem: {
        system: '',
        ips: '',
        comment: ''
      },
      addMachineRules: {
        system: [{ required: true, message: '请选择系统', trigger: 'blur' }],
        ips: [{ required: true, message: '请填写节点列表，每行一个', trigger: 'blur' }],
        comment: [{ required: true, message: '请填写操作注释', trigger: 'blur' }]
      }
    }
  },

  watch: {
    dialogAdd(val) {
      val || this.resetForm()
    },
    '$route'(to, from) {
      this.initialize(to)
    }
  },

  created() {
    this.initialize(this.$route)
  },

  methods: {
    initialize(router) {
      if (router.name === 'Cluster_Bus') {
        this.forCron = false
      } else {
        this.forCron = true
      }
      getSystems().then(res => {
        res.data.forEach(system => {
          this.systemOptions[system.name] = system.desc
        })
        getMachines(this.forCron).then(res => {
          this.tableData = {}
          res.data.forEach(st => {
            if (this.tableData[st.system] === undefined) {
              this.tableData[st.system] = []
            }
            this.tableData[st.system].push({
              id: st.id,
              ip: st.ip,
              status: st.status
            })
          })
          if (this.$route.query.system !== undefined) {
            if (this.tableData[this.$route.query.system] === undefined) {
              this.$notify({
                title: this.$route.query.system + ' 集群不存在',
                message: '未找到对应的集群',
                type: 'success',
                duration: 2000
              })
            }
            this.activeName = this.$route.query.system
          } else if (res.data.length > 0) {
            this.activeName = res.data[0].system
          }
        })
      })
    },
    resetForm() {
      this.clusterItem = {
        system: '',
        ips: '',
        comment: ''
      }
      this.$refs['clusterForm'].clearValidate()
    },
    addCluster() {
      this.$refs['clusterForm'].validate((valid) => {
        if (valid) {
          addMachines(this.clusterItem, this.forCron).then(res => {
            if (res.code === 0) {
              this.$notify({
                title: '成功',
                message: '保存成功',
                type: 'success',
                duration: 2000
              })
              this.dialogAdd = false
              this.initialize(this.$route)
            }
          })
        }
      })
    },
    deleteMachine(item) {
      if (!confirm('确定删除 ' + item.ip + ' 吗？')) {
        return
      }
      delMachine(item, this.forCron).then(res => {
        this.$notify({
          title: '成功',
          message: '删除成功',
          type: 'success',
          duration: 2000
        })
        this.initialize(this.$route)
      })
    }
  }
}
</script>
