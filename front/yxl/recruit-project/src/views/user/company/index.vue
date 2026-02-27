<script setup>
import { ref, onMounted, watch } from 'vue'
import { User, Document, ArrowRightBold } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
const router = useRouter()
const route = useRoute()
const activeIndex = ref('/like')
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
    <el-menu :router="true" :default-active="activeIndex" class="el-menu-demo" mode="horizontal" style="
      display: flex;
      justify-content: space-between;
      height: 80px;
      user-select: none;
    ">
      <div style="display: flex; align-items: center">
        <IconSchool />
        <h1>
          <p>个人主页</p>
        </h1>
      </div>

      <div style="display: flex; align-items: center">
        <el-menu-item index="/like">
          <el-icon>
            <Document />
          </el-icon>
          简历收藏
        </el-menu-item>
        <el-menu-item index="/change">
          <el-icon>
            <User />
          </el-icon>
          修改密码
        </el-menu-item>
        <el-menu-item index="/showData">
          返回
          <el-icon class="icon">
            <ArrowRightBold />
          </el-icon>
        </el-menu-item>
      </div>
    </el-menu>
    <router-view />
  </div>
</template>
<style scoped></style>
