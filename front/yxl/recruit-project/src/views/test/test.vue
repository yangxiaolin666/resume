<template>
  <div class="job-list">
    <jobCard 
      v-for="job in jobList" 
      :key="job.id" 
      :item="job" 
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import jobCard from '@/components/jobCard.vue'

const jobList = ref([
  {
    "id": 1,
    "com_id": 100,
    "comName": "某某科技有限公司",
    "title": "Java开发工程师",
    "description": "负责后端系统开发，参与需求分析和系统设计，编写高质量的代码。",
    "external_link": "https://example.com/jobs/123",
    "status": "active",
    "created_at": "2026-03-17 10:00:00",
    "updated_at": "2026-03-17 14:30:00"
  }
])
</script>

<style scoped>
.job-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
</style>