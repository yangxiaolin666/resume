<script setup>
import { useStudentsStore } from '@/stores/students'
import studentCard from '@/components/studentCard.vue'
const StudentsStore = useStudentsStore()
const students = StudentsStore.students
</script>
<template>
  <div class="container">
    <el-row>
      <el-col v-for="item in students" :key="item.id" :xs="24" :sm="12" :md="8" :lg="6">
      <studentCard :item="item" />
      </el-col>
    </el-row>
  </div>
</template>
<style scoped>

.container {
  padding: 20px;
  max-height: 100vh;
}
</style>
