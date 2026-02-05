<script setup>
import { ref } from 'vue'
import { User, Document, ArrowRightBold } from '@element-plus/icons-vue'
import {  useRouter } from 'vue-router'
const router = useRouter()
const activeIndex = ref('1')
</script>
<template>
  <el-menu :router="true" :default-active="activeIndex" class="el-menu-demo" mode="horizontal"
    style="display: flex; justify-content: space-between; height: 80px;user-select: none; " active-text-color="#000">
    <div style="display: flex; align-items: center">
      <IconSchool />
      <h1>
        <p>学生招聘平台</p>
      </h1>
    </div>

    <div style="display: flex; align-items: center">
      <el-menu-item index="/like">
        <el-icon class="icon">
          <Document />
        </el-icon>简历收藏
      </el-menu-item>
      <el-menu-item index="/change">
        <el-icon class="icon">
          <User />
        </el-icon>修改密码
      </el-menu-item>
      <el-menu-item @click="router.push('/home')">
        返回
        <el-icon class="icon">
          <ArrowRightBold />
        </el-icon>
      </el-menu-item>
    </div>
  </el-menu>
  <router-view></router-view>
</template>
<style scoped></style>
