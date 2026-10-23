<script setup>
import { onMounted, ref, watch } from 'vue'
import { User, Document, ArrowRightBold, MessageBox, Setting } from '@element-plus/icons-vue'
import { useRoute } from 'vue-router'
const activeIndex = ref('/selfMessage')
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
        <el-menu-item index="/selfMessage">
          <el-icon>
            <User />
          </el-icon>
          个人信息
        </el-menu-item>
        <el-menu-item index="/updateResume">
          <el-icon>
            <Document />
          </el-icon>
          简历修改
        </el-menu-item>
        <el-menu-item index="/studentSetting">
          <el-icon>
            <Setting />
          </el-icon>
          修改个人信息
        </el-menu-item>
        <el-menu-item index="/makeResume">
          <el-icon>
            <MessageBox />
          </el-icon>
          简历制作
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
      <router-view></router-view>
    </div>
  </div>
</template>
<style scoped lang="scss">
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