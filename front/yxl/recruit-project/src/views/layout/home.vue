<script setup>
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import {
    Search,
    User,
    House,
    SwitchButton,
    Document
  } from '@element-plus/icons-vue'
  import IconSchool from '@/components/IconSchool.vue'
  import { useUserStore } from '@/stores/user'
  import { useStudentsStore } from '@/stores/students'
  import studentCard from '@/components/studentCard.vue'
  const router = useRouter()
  const userStore = useUserStore()
  const StudentsStore = useStudentsStore()
  const students = StudentsStore.students
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
</script>

<template>
  <div class="home-container">
    <el-menu :default-active="activeIndex" class="el-menu-demo" mode="horizontal" style="
        display: flex;
        justify-content: space-between;
        height: 80px;
        user-select: none;
      ">
      <div style="display: flex; align-items: center">
        <IconSchool />
        <h1>
          <p>学生招聘平台</p>
        </h1>
      </div>
      <div style="display: flex; align-items: center">
        <el-menu-item @click="router.push('/updateResume')">
          <el-icon>
            <Document />
          </el-icon>
          简历修改
        </el-menu-item>
        <el-menu-item index="4">
          <el-popover placement="bottom-end">
            <template #reference>
              <el-icon class="nav-icon">
                <User />
              </el-icon>
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
    <main class="main-content">
      <aside class="filter-sidebar">
        <el-card class="filter-card">
          <template #header>
            <div class="filter-header">
              <span>筛选条件</span>
              <el-button type="text">
                重置
              </el-button>
            </div>
          </template>

          <!-- 搜索框 -->
          <div class="filter-section">
            <el-input>
              <template #prefix>
                <el-icon>
                  <Search />
                </el-icon>
              </template>
            </el-input>
          </div>

          <!-- 学位筛选 -->
          <div class="filter-section">
            <h4>学位</h4>
            <el-radio-group size="small">
              <el-radio-button label="">
                全部
              </el-radio-button>
              <el-radio-button />
            </el-radio-group>
          </div>
        </el-card>
      </aside>
      <div>
        <el-row>
          <el-col v-for="item in students" :key="item.id" :xs="24" :sm="12" :md="8" :lg="8">
            <studentCard :item="item" />
          </el-col>
        </el-row>
      </div>
    </main>
  </div>
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
