<script setup>
import { ref } from 'vue'
import IconSchool from '@/components/IconSchool.vue'
import { User, Lock, OfficeBuilding, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { loginToR } from '@/api/login'
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'
const username = ref('')
const password = ref('')
const userType = ref('student')
const userStore = useUserStore()
const router = useRouter()
const loginTo = async () => {
  const {
    data: { data }
  } = await loginToR(username.value, password.value, userType.value)
  console.log(data)
  userStore.setAll(data.userId,data.username, data.userType, data.token, data.displayName)
  ElMessage({
    message: '登录成功',
    type: 'success',
    duration: 1000
  })
  setTimeout(() => {
    if (userType.value === 'admin') {
      loginToAdmin()
    } else {
      router.push('/chat')
    }
  }, 1100)
}
const loginToAdmin = () => {
  const adminToken = userStore.token
  if (!adminToken) {
    ElMessage({
      message: '管理员登录失败，未获取到token',
      type: 'error',
      duration: 1000
    })
    return
  }

  const encodeToken = encodeURIComponent(adminToken)
  const adminBaseUrl = import.meta.env.VITE_ADMIN_URL || 'http://47.93.133.36'
  const adminUrl = `${adminBaseUrl}/admin/?token=${encodeToken}`
  window.location.replace(adminUrl)
}


</script>

<template>
  <div class="login-container">
    <div class="login-box">
      <!-- 左侧导航栏 -->
      <div class="login-sidebar">
        <IconSchool style="margin-left: 40px;width:120px;
        stroke: white;height: 120px;" />
        <div class="logo">
          招聘系统
        </div>
        <el-menu :default-active="userType" class="login-menu"
         @select="(key) => {userType = key;console.log(userType);
        }">
          <el-menu-item index="student">
            <el-icon>
              <User />
            </el-icon>
            <span>学生</span>
          </el-menu-item>
          <el-menu-item index="company">
            <el-icon>
              <OfficeBuilding />
            </el-icon>
            <span>企业</span>
          </el-menu-item>
          <el-menu-item index="admin">
            <el-icon>
              <Setting />
            </el-icon>
            <span>管理员</span>
          </el-menu-item>
        </el-menu>
      </div>
      <div class="login-form-container">
        <div class="login-form">
          <h2>
            {{
              userType === 'student'
                ? '学生登录'
                : userType === 'company'
                  ? '企业登录'
                  : '管理员登录'
            }}
          </h2>
          <el-form :model="{ username, password }" label-width="0">
            <el-form-item>
              <el-input v-model="username" placeholder="请输入账号" :prefix-icon="User" style="width: 300px; height: 40px" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="password" type="password" placeholder="请输入密码" :prefix-icon="Lock" show-password
                style="width: 300px; height: 40px" />
            </el-form-item>
            <el-form-item @click="loginTo">
              <el-button type="primary" style="width: 300px; height: 45px">
                登录
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
