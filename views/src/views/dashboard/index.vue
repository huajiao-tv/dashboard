<template>
  <div class="dashboard-container">
    <div v-if="isAdmin" class="dashboard-text"><el-button type="primary" @click="jump">用户管理</el-button></div>
    <hr>
    <div>name:{{ name }}</div>
    <div >roles:<span v-for="role in roles" :key="role">{{ role }}</span></div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'

export default {
  name: 'Dashboard',
  computed: {
    ...mapGetters([
      'name',
      'roles'
    ]),
    isAdmin() {
      return this.roles.indexOf('administration') >= 0
    }
  },
  methods: {
    jump(params) {
      this.$router.push({ path: '/auth/list' })
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss" scoped>
.dashboard {
  &-container {
    margin: 30px;
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }
}
</style>
