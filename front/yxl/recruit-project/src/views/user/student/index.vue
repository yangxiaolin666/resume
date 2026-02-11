<script setup>
import { onMounted, ref , watch} from 'vue'
import { User, Document, ArrowRightBold } from '@element-plus/icons-vue'
import { useRoute } from 'vue-router'
const activeIndex = ref('/selfMessage')
const route=useRoute()
const updateActiveIndex = () => {
  activeIndex.value = route.path 
}
onMounted(()=>{
  updateActiveIndex()
})
watch(()=>route.path,()=>{
  updateActiveIndex()
})
</script>
<template>
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
      <el-menu-item index="/selfMessage">
        <el-icon>
          <User />
        </el-icon>
        个人信息
      </el-menu-item>
      <el-menu-item index="/updateResume">
        <el-icon>
          <Document />
        </el-icon>
        简历修改
      </el-menu-item>
      <el-menu-item index="/showData">
        返回
        <el-icon class="icon">
          <ArrowRightBold />
        </el-icon>
      </el-menu-item>
    </div>
  </el-menu>
 <router-view></router-view>
</template>
<style scoped lang="scss"></style>