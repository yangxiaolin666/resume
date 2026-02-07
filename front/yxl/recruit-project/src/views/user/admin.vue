<script setup>
import { ref, onMounted, watch } from 'vue'
const isDark = ref(window.getDarkMode ? window.getDarkMode() : true)
watch(isDark, (newValue) => {
    window.toggleDarkMode()
})
</script>

<template>
    <div class="admin-container">
    <div class="header">
      <h1>管理员界面</h1>
      <div class="theme-switch">
        <el-switch
          v-model="isDark"
          inline-prompt
          :active-icon="'🌙'"
          :inactive-icon="'☀️'"
          active-text="暗黑"
          inactive-text="明亮"
        />
        <span class="theme-label">{{ isDark ? '暗黑模式' : '明亮模式' }}</span>
      </div>
    </div>
    
    <div class="content">
      <p>这是管理员界面的主要内容区域。</p>
      <p>当前主题：{{ isDark ? '暗黑模式' : '明亮模式' }}</p>
      
      <!-- 测试Element Plus组件在暗黑模式下的表现 -->
      <div class="component-test">
        <el-button type="primary">主要按钮</el-button>
        <el-button type="success">成功按钮</el-button>
        <el-button type="warning">警告按钮</el-button>
        <el-button type="danger">危险按钮</el-button>
        
        <el-card class="test-card" header="测试卡片">
          <p>这是一个测试卡片，用于展示暗黑模式下的组件样式</p>
          <el-tag type="primary">标签</el-tag>
          <el-tag type="success">成功</el-tag>
          <el-tag type="warning">警告</el-tag>
          <el-tag type="danger">危险</el-tag>
        </el-card>
        
        <!-- 更多组件测试 -->
        <el-alert title="成功提示" type="success" show-icon />
        <el-alert title="警告提示" type="warning" show-icon />
        <el-alert title="错误提示" type="error" show-icon />
      </div>
    </div>
  </div>
</template>

<style scoped>
.admin-container {
  padding: 20px;
  min-height: 100vh;
  transition: all 0.3s ease;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--el-border-color);
}

.theme-switch {
  display: flex;
  align-items: center;
  gap: 10px;
}

.theme-label {
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.content {
  padding: 20px;
  background: var(--el-bg-color-page);
  border-radius: 8px;
  border: 1px solid var(--el-border-color-light);
}

.component-test {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.test-card {
  max-width: 400px;
}
</style>
