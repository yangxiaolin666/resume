<script setup>
import { ref } from 'vue'
import { getResume } from '@/api/student';
import { ElMessage } from 'element-plus'

const pdfLoading = ref(false)
const pdfError = ref(false)

const handlegetResume = async () => {

    pdfLoading.value = true
    pdfError.value = false
    console.log(res)
    ElMessage({
      message: '简历获取成功',
      type: 'success',
      duration: 1000
    })

}

const onPdfLoad = () => {
  pdfLoading.value = false
  pdfError.value = false
}

const onPdfError = () => {
  pdfLoading.value = false
  pdfError.value = true
  ElMessage({
    message: 'PDF加载失败，请检查文件路径或网络连接',
    type: 'error',
    duration: 3000
  })
}
</script>
<template>
  <div class="pdf-container">
    <el-button 
      @click="handlegetResume" 
      type="primary"
      :loading="pdfLoading">
      {{ pdfLoading ? '加载中...' : '获取简历' }}
    </el-button>
    
    <div class="pdf-viewer">
      <!-- 加载状态 -->
      <div v-if="pdfLoading" class="loading-state">
        <el-icon class="is-loading">
          <Loading />
        </el-icon>
        <span>PDF加载中...</span>
      </div>
      
      <!-- 错误状态 -->
      <div v-else-if="pdfError" class="error-state">
        <el-icon><CircleClose /></el-icon>
        <span>PDF加载失败</span>
        <el-button @click="handlegetResume" type="text">重试</el-button>
      </div>
      
      <!-- PDF预览 -->
      <iframe 
        v-else
        src="http://81.70.199.174:8081/uploads/resume/2024081104_20260208163529_3413b8ec.pdf"
        width="100%" 
        height="800" 
        style="border: none;"
        @load="onPdfLoad"
        @error="onPdfError">
        kjl
      </iframe>
    </div>
  </div>
</template>
<style scoped></style>
