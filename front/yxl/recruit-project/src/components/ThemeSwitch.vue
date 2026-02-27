<template>
  <div class="theme-switch">
    <el-switch 
      v-model="isDark" 
      inline-prompt
      size="small"
    />
    <span class="theme-label">{{ isDark ? '暗黑模式' : '明亮模式' }}</span>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const isDark = ref(window.getDarkMode ? window.getDarkMode() : true)

watch(isDark, (newValue) => {
  window.toggleDarkMode()
})
</script>

<style scoped>
.theme-switch {
  display: flex;
  align-items: center;
  gap: 15px;
}

.theme-label {
  font-size: 14px;
  color: var(--el-text-color-regular);
}
</style>