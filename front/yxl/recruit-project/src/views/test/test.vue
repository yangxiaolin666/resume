<template>
  <div class="pdf-container">
    <h2>简历预览</h2>
   <div class="pdf-viewer">
      <embed 
        src="http://81.70.199.174:8081/resume/uploads/resume/2024081104_20260208163529_3413b8ec.pdf"
        type="application/pdf" 
        width="100%" 
        height="800" />
    </div>
    
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Download, Refresh } from '@element-plus/icons-vue'

const pdfUrl = ref('http://81.70.199.174:8081/resume/uploads/resume/2024081104_20260208163529_3413b8ec.pdf')

const downloadPdf = () => {
  const link = document.createElement('a')
  link.href = pdfUrl.value
  link.download = '简历.pdf'
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

</script>

<style scoped>
.pdf-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.pdf-container h2 {
  text-align: center;
  margin-bottom: 20px;
  color: var(--el-text-color-primary);
}

.pdf-viewer {
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}



</style>