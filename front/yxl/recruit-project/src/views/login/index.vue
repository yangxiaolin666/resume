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
const userType = ref('admin')
const userStore = useUserStore()
const router = useRouter()
const loginTo = async () => {
  const {
    data: { data }
  } = await loginToR(username.value, password.value, userType.value)
  userStore.setAll(data.username, data.userType, data.token)
  ElMessage({
    message: '登录成功',
    type: 'success',
    duration: 1000
  })
  setTimeout(() => router.push('/home'), 1100)
}
</script>

<template>
  <div class="login-container">
    <div class="login-box">
      <!-- 左侧导航栏 -->
      <div class="login-sidebar">
        <IconSchool style="margin-left: 40px;width:120px;height: 120px;" />
        <div class="logo">
          招聘系统
        </div>
        <el-menu :default-active="userType" class="login-menu" @select="(key) => {userType = key;console.log(userType);
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

      <!-- 右侧登录表单 -->
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
}

.login-box {
  width: 750px;
  height: 500px;
  background-color: white;
  border-radius: 15px;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
  display: flex;
  overflow: hidden;
}

.login-sidebar {
  width: 200px;
  background-color: var(--main-blue);
  padding: 20px 0;
  border-right: 1px solid #e4e7ed;
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
}

.login-form-container {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: white;
}

.login-form {
  text-align: center;
}

.login-form h2 {
  font-size: 40px;
  margin-bottom: 95px;
  color: #303133;
}

.el-menu-item {
  height: 60px;
  line-height: 60px;
  font-size: 16px;
}

.el-form-item {
  margin-bottom: 16px;
}

.el-menu-item.is-active {
  color: #409eff;
  background-color: #ecf5ff;
}
</style>
