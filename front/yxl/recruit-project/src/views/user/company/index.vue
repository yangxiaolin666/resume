<script setup>
import { ref, onMounted, watch } from 'vue'
import { User, Document, ArrowRightBold, Plus, List } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
const router = useRouter()
const route = useRoute()
const activeIndex = ref('/like')
const updateActiveIndex = () => {
  activeIndex.value = route.path
}
onMounted(() => {
  updateActiveIndex()
})
watch(() => route.path, () => {
  updateActiveIndex()
})
</script>
<template>
  <div style="width:100%; min-width:max-content;">
    <el-menu :router="true" :default-active="activeIndex" class="fixed-menu" mode="horizontal">
      <div class="menu-left">
        <IconSchool />
        <h1>
          <p>个人主页</p>
        </h1>
      </div>

      <div class="menu-right">
        <el-menu-item index="/like">
          <el-icon>
            <Document />
          </el-icon>
          简历收藏
        </el-menu-item>
        <el-menu-item index="/myJobs">
          <el-icon>
            <List />
          </el-icon>
          发布职位
        </el-menu-item>
        <el-menu-item index="/change">
          <el-icon>
            <User />
          </el-icon>
          修改个人信息
        </el-menu-item>
        <el-menu-item index="/showData">
          返回
          <el-icon class="icon">
            <ArrowRightBold />
          </el-icon>
        </el-menu-item>
      </div>
    </el-menu>
    <div class="content-area">
      <router-view />
    </div>
  </div>
</template>
<style scoped>
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

.content-area {
  margin-top: 80px;
  min-height: calc(100vh - 80px);
}
</style>
