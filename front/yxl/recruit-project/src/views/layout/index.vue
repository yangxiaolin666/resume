<script setup>
import { useUserStore } from '@/stores/user';
import {
  User,
  House,
  SwitchButton,
  PieChart,
  Discount,
} from '@element-plus/icons-vue'
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
const router = useRouter()
const goToUser = () => {
  router.push(`/${userStore.userType}`)
}
const logOut = () => {
  userStore.setAll('', '', '')
  ElMessage({
    message: '退出登录成功',
    type: 'success',
    duration: 1000
  })
  setTimeout(() => router.push('/login'), 1100)
}
const activeIndex = ref('/showData')
const route = useRoute()
const updateActiveIndex = () => {
  activeIndex.value = route.path
}
onMounted(() => {
  updateActiveIndex()
})
watch(() => route.path, () => {
  updateActiveIndex()
})
const userStore = useUserStore()
</script>

<template>
  <el-menu :default-active="activeIndex" :router="true" class="el-menu-demo" mode="horizontal" style="
        display: flex;
        justify-content: space-between;
        height: 80px;
      ">
    <div style="display: flex; align-items: center">
      <IconSchool />
      <h1>
        <p>学生招聘平台</p>
      </h1>
    </div>
    <div style="display: flex; align-items: center">
      <el-menu-item index="/showData">
        <el-icon>
          <PieChart />
        </el-icon>
        应届生数据
      </el-menu-item>
      <el-menu-item index="/home" v-if="true||(userStore.userType === 'company')">
        <el-icon>
          <Discount />
        </el-icon>
        应届生列表
      </el-menu-item>
      <el-menu-item>
        <el-popover placement="bottom-end">
          <template #reference>
            <div style="display: flex; align-items: center;">
              <el-icon class="nav-icon">
                <User />
              </el-icon>
              <span>{{ userStore.userType }}</span>
            </div>
          </template>
          <div class="menu-item" @click="goToUser">
            <el-icon>
              <House />
            </el-icon>
            <span>个人主页</span>
          </div>
          <div class="menu-item" @click="logOut">
            <el-icon>
              <SwitchButton />
            </el-icon>
            <span>退出登录</span>
          </div>
        </el-popover>
      </el-menu-item>
    </div>
  </el-menu>
  <router-view></router-view>
</template>

<style scoped>
.menu-item {
  display: flex;
  align-items: center;
  padding: 8px 0px;
  cursor: pointer;
}

.menu-item:hover {
  color: #409eff;
}

.menu-item span {
  padding-left: 30px;
  font-size: 14px;
}

/* 主要内容样式 */
.main-content {
  display: flex;
  gap: 20px;
  padding: 20px;
}

.filter-sidebar {
  margin-top: 10px;
  width: 280px;
  flex-shrink: 0;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
