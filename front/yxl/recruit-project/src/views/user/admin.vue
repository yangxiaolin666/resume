<script setup>
import { ref, watch } from 'vue'
const isDark = ref(window.getDarkMode ? window.getDarkMode() : true)
watch(isDark, (newValue) => {
  window.toggleDarkMode()
})
import { genFileId } from 'element-plus'
const upload = ref()
const handleExceed = (files) => {
  upload.value.clearFiles()
  const file = files[0]
  file.uid = genFileId()
  upload.value.handleStart(file)
}
const submitUpload = () => {
  console.log(upload.value);
}
</script>

<template>
  <div class="admin-container">
    <div class="header">
      <h1>管理员界面</h1>
      <div class="theme-switch">
        <el-switch v-model="isDark" inline-prompt :active-icon="'🌙'" :inactive-icon="'☀️'" active-text="暗黑"
          inactive-text="明亮" />
        <span class="theme-label">{{ isDark ? '暗黑模式' : '明亮模式' }}</span>
      </div>
    </div>
    <div class="container"> <el-upload
    ref="upload"
    class="upload-demo"
    :limit="1"
    :on-exceed="handleExceed"
    :auto-upload="false"
  >
    <template #trigger>
      <el-button type="primary">select file</el-button>
    </template>
    <el-button class="ml-3" type="success" @click="submitUpload">
      upload to server
    </el-button>
    <template #tip>
      <div class="el-upload__tip text-red">
        limit 1 file, new file will cover the old file
      </div>
    </template>
  </el-upload>
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
