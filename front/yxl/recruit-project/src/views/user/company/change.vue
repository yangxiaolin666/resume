<script setup>
import { ref } from 'vue'
import IconSchool from '@/components/IconSchool.vue'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'
import { changePwd } from '@/api/login'
const username = ref('')
const password = ref('')
const newPassword = ref('')
const rePassword = ref('')
const userStore = useUserStore()
const router = useRouter()
const change = async () => {
  const res = await changePwd(password.value, newPassword.value, rePassword.value)
  console.log(res)
  ElMessage({
    type: 'success',
    message: '修改成功,请重新登录',
    duration: 1000
  })
  userStore.setAll('', '', '')
  setTimeout(() => router.push('/login'), 1100)
}
</script>

<template>
  <div class="login-container">
    <div class="login-box">
      <!-- 左侧导航栏 -->
      <div class="login-sidebar">
        <IconSchool style="margin-left: 40px;
        width:120px;height: 120px;stroke: white;" />
        <div class="logo">
          修改密码
        </div>
      </div>
      <div class="login-form-container">
        <div class="login-form">
          <h2>
            原密码修改密码
          </h2>
          <el-form :model="{ username, password }" label-width="0">
            <el-form-item>
              <el-input v-model="username" placeholder="请输入账号" :prefix-icon="User" style="width: 300px; height: 40px" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="password" type="password" placeholder="请输入原密码" :prefix-icon="Lock" show-password
                style="width: 300px; height: 40px" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="newPassword" type="password" placeholder="请输入新密码" :prefix-icon="Lock" show-password
                style="width: 300px; height: 40px" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="rePassword" type="password" placeholder="请确认新密码" :prefix-icon="Lock" show-password
                style="width: 300px; height: 40px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" style="width: 300px; height: 45px" @click="change">
                修改
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  width: 100vw;
  height: 100vh;
  background-size: cover;
  background-position: center;
  display: flex;
  justify-content: center;
  align-items: center;
  transition: background-color 0.3s ease;
}

.login-box {
  width: 750px;
  height: 500px;
  background-color: var(--el-bg-color);
  border-radius: 15px;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
  display: flex;
  overflow: hidden;
  transition: all 0.3s ease;
}

.login-sidebar {
  width: 200px;
  background-color: var(--el-color-primary);
  padding: 20px 0;
}

.logo {
  font-size: 30px;
  font-weight: bold;
  color: white;
  text-align: center;
  margin-bottom: 40px;
  padding: 0 20px;
}

.login-menu {
  border-right: none;
  background-color: transparent;
}

.login-menu .el-menu-item {
  height: 60px;
  line-height: 60px;
  font-size: 16px;
  color: var(--el-text-color-primary);
  background-color: var(--el-bg-color);
  border: none;
}

.login-menu .el-menu-item:hover {
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

.login-menu .el-menu-item.is-active {
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-8);
}

.login-form-container {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: var(--el-bg-color);
}

.login-form {
  text-align: center;
}

.login-form h2 {
  font-size: 40px;
  margin-bottom: 95px;
  color: var(--el-text-color-primary);
}

.el-form-item {
  margin-bottom: 16px;
}

/* 暗黑模式下的特殊样式 */
:global(.dark) .login-box {
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.3);
}

:global(.dark) .login-sidebar {
  background-color: var(--el-color-primary-dark-2);
}

:global(.dark) .login-menu .el-menu-item {
  color: rgba(255, 255, 255, 0.9);
}

:global(.dark) .login-menu .el-menu-item:hover {
  background-color: rgba(255, 255, 255, 0.15);
}

:global(.dark) .login-menu .el-menu-item.is-active {
  background-color: rgba(255, 255, 255, 0.25);
}
</style>
