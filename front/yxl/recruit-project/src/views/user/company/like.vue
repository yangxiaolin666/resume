<script setup>
import { useStudentsStore } from '@/stores/students'
import studentCard from '@/components/studentCard.vue'
const StudentsStore = useStudentsStore()
const students = StudentsStore.students
</script>
<template>
  <div style="max-height: 100vh; padding-right: 10px">
    <el-row>
      <el-col v-for="item in students" :key="item.id" :xs="24" :sm="12" :md="8" :lg="6">
        <studentCard :item="item" />
      </el-col>
    </el-row>
  </div>
</template>
<style scoped></style>
