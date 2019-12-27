<template>
  <div>
    <el-table
      :data="tableData"
      stripe
      style="width: 100%">
      <el-table-column
        prop="id"
        label="ID"
        width="50"/>
      <el-table-column
        prop="username"
        label="用户名"/>
      <el-table-column
        prop="email"
        label="邮箱"/>
      <el-table-column
        prop="roles"
        label="权限"/>
      <el-table-column
        prop="create_at"
        label="创建时间"/>
      <el-table-column
        label="操作">
        <template slot-scope="scope">
          <el-button type="danger" size="mini" icon="el-icon-error" circle @click="del(scope.row)"/>
          <el-button type="success" size="mini" icon="el-icon-edit" circle @click="edit(scope.row)"/>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :visible.sync="roleEdit.visible" :title="roleEdit.title">

      <el-checkbox v-model="roleEdit.admin">管理员</el-checkbox>
      <hr>
      <div style="margin: 15px 0;"/>
      <el-checkbox-group v-model="roleEdit.roles">
        <el-checkbox v-for="system in roleEdit.systems" :label="system" :key="system" :disabled="roleEdit.admin">{{ system }}</el-checkbox>
      </el-checkbox-group>
      <div slot="footer" class="dialog-footer">
        <el-button @click="roleEdit.visible = false">取 消</el-button>
        <el-button type="primary" @click="editRoles">确 定</el-button>
      </div>
    </el-dialog>

  </div>
</template>
<script>
import { userlist, deluser, editroles } from '../../api/login'
import { getSystems } from '../../api/system'

export default {
  data() {
    return {
      search: '',
      tableData: [],
      roleEdit: {
        admin: false,
        systems: [],
        roles: [''],
        visible: false,
        title: '',
        id: ''
      }
    }
  },
  created() {
    this.initialize()
  },

  methods: {
    initialize() {
      getSystems().then(res => {
        this.roleEdit.systems = []
        res.data.forEach((element, v) => {
          this.roleEdit.systems.push(element.name)
        })
      })
      userlist(this.search).then(res => {
        this.tableData = res.data
        if (this.tableData.length === 0) {
          this.$notify({
            title: '用户列表为空',
            message: '找不到一个用户，检查您的搜索词或者权限',
            type: 'success',
            duration: 2000
          })
        }
      })
    },
    del(data) {
      if (!confirm('确定删除用户: ' + data.username + '/' + data.email + ' 吗？')) {
        return
      }
      deluser(data.id).then(res => {
        this.$notify({
          title: '成功',
          message: '删除成功',
          type: 'success',
          duration: 2000
        })
        this.initialize()
      })
    },
    edit(data) {
      this.roleEdit.id = data.id
      this.roleEdit.visible = true
      this.roleEdit.admin = false
      data.roles.split(',').forEach(element => {
        if (element === 'administration') {
          this.roleEdit.admin = true
        } else {
          this.roleEdit.roles.push(element)
        }
      })
      this.roleEdit.title = data.username + '/' + data.email + '权限编辑'
    },
    editRoles() {
      var role = []
      this.roleEdit.roles.forEach(element => {
        if (element !== '' && element !== 'administration' && role.indexOf(element) === -1) {
          role.push(element)
        }
      })
      if (!this.roleEdit.admin && role.length === 0) {
        if (!confirm('确定将此用户设置成无任何权限吗？')) {
          return
        }
      }
      var rolestring = role.join(',')
      if (this.roleEdit.admin) {
        rolestring = 'administration'
      }
      editroles(this.roleEdit.id, rolestring).then(res => {
        this.roleEdit.visible = false
        this.$notify({
          title: '成功',
          message: '修改成功',
          type: 'success',
          duration: 2000
        })
        this.initialize()
      })
    }
  }
}
</script>
