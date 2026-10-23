<script setup>
import { useUserStore } from '@/stores/user';
import {
  User,
  House,
  SwitchButton,
  PieChart,
  Discount,
  Moon,
  Sunny,
  ChatDotRound,
} from '@element-plus/icons-vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getConversations } from '@/api/chat'
import { getMyProfile } from '@/api/login'
import { unreadTotal, setUnreadTotal } from '@/stores/chatUnread'
const router = useRouter()
const goToUser = () => {
  router.push(`/${userStore.userType}`)
}
const logOut = () => {
  userStore.clear()
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
  refreshUnread()
  syncProfile()
  // 轮询刷新未读数，保证停留在其他页面时也能实时看到红点
  unreadTimer = setInterval(refreshUnread, 15000)
})
watch(() => route.path, () => {
  updateActiveIndex()
  refreshUnread()
})
onBeforeUnmount(() => {
  if (unreadTimer) clearInterval(unreadTimer)
})
const userStore = useUserStore()

// ---------- 昵称（右上角显示） ----------
const displayName = computed(() => userStore.displayName || userStore.username)

// 老版本登录状态里没有昵称，进入页面时补拉一次
async function syncProfile() {
  if (userStore.displayName || !userStore.token) return
  try {
    const { data } = await getMyProfile()
    userStore.setDisplayName(data.data.displayName)
  } catch (e) {
    // 拉取失败时先显示账号名，不影响使用
  }
}

// ---------- 顶部“消息”未读红点 ----------
let unreadTimer = null

async function refreshUnread() {
  // 管理员没有聊天入口，直接清零
  if (userStore.userType !== 'student' && userStore.userType !== 'company') {
    setUnreadTotal(0)
    return
  }
  try {
    const res = await getConversations()
    const list = res.data?.data?.list || []
    // 每个会话的 unread 相加 = 全部未读总数
    setUnreadTotal(list.reduce((sum, item) => sum + (Number(item.unread) || 0), 0))
  } catch (e) {
    // 网络异常时静默忽略，不打扰用户
  }
}

const isDark = ref(window.getDarkMode ? window.getDarkMode() : true)

const toggleTheme = () => {
  isDark.value = !isDark.value
  if (window.toggleDarkMode) {
    window.toggleDarkMode()
  }
}
</script>

<template>
  <div class="layout-root">
    <el-menu :default-active="activeIndex" :router="true" class="fixed-menu" mode="horizontal">
      <div class="menu-left">
        <IconSchool />
        <h1>
          <p>招聘平台</p>
        </h1>
      </div>

      <div class="menu-right">
        <el-menu-item index="/home" v-if="(userStore.userType === 'company')">
          <el-icon>
            <Discount />
          </el-icon>
          应届生列表
        </el-menu-item>
        <el-menu-item index="/Job" v-if="(userStore.userType === 'company')">
          <el-icon>
            <Discount />
          </el-icon>
          招聘信息
        </el-menu-item>
        <el-menu-item index="/showData">
          <el-icon>
            <PieChart />
          </el-icon>
          应届生数据
        </el-menu-item>
        <el-menu-item index="/chat" v-if="userStore.userType === 'student' || userStore.userType === 'company'">
          <el-icon>
            <ChatDotRound />
          </el-icon>
          消息
          <span v-if="unreadTotal > 0" class="unread-badge">
            {{ unreadTotal > 99 ? '99+' : unreadTotal }}
          </span>
        </el-menu-item>
        <el-menu-item :index="activeIndex" style="color:var(--el-text-color)">
          <el-popover placement="bottom-end">
            <template #reference>
              <div style="display: flex; align-items: center;">
                <el-icon class="nav-icon">
                  <User />
                </el-icon>
                <span>{{ displayName }}</span>
              </div>
            </template>
            <div class="menu-item" @click="toggleTheme">
              <el-icon>
                <Moon v-if="isDark" />
                <Sunny v-else />
              </el-icon>
              <span>{{ isDark ? '暗黑模式' : '明亮模式' }}</span>
            </div>
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
    <div class="content-area">
      <router-view></router-view>
    </div>
  </div>

</template>

<style scoped>
/* 宽度必须由容器决定，不能用 max-content，
   否则聊天页会按最长消息的单行宽度把整页撑宽 */
.layout-root {
  width: 100%;
  min-width: 0;
}

.fixed-menu {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 80px;
  background: var(--el-bg-color);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  user-select: none;
}

.menu-left {
  display: flex;
  align-items: center;
}

.menu-right {
  display: flex;
  align-items: center;
}

/* 顶部“消息”未读红点：放在“消息”文字右侧 */
.unread-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  width: 18px;
  height: 18px;
  margin-left: 6px;
  padding: 0;
  border-radius: 50%;
  background: var(--el-color-danger);
  color: #fff;
  font-size: 10px;
  line-height: 1;
  flex-shrink: 0;
}

.content-area {
  margin-top: 80px;
  min-height: calc(100vh - 80px);
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 8px 8px;
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