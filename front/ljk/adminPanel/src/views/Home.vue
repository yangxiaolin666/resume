<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../components/Header.vue'
import SelectorFixed from '../components/SelectorFixed.vue'
import SelectItem from '../components/SelectItem.vue'

import { stu_options } from '../staticData/student_options.js'

const menuSelect = ref(1); // 1: 学生信息, 2: 企业信息
const handleMenuSelect = (index) => {
  menuSelect.value = index;
  console.log("改变");
};
const handleLogout = () => {
  console.log("退出登录");
}
const selectedStatus = ref([]);

const handleChange = (value) => {
  console.log('筛选：', value);
  console.log('筛选条件：', selectedStatus.value);
};


</script>

<template>
  <Header
    @menu-change="handleMenuSelect"
    @logout="handleLogout"
  ></Header>
  <div class="main-container">
    
    <div class="content-container">
      <!-- 学生信息 -->
      <div class="page student-container" v-if="menuSelect === 1">
        <div class="selector-container">
          <SelectorFixed title="筛选">
            <SelectItem
              title="学历"
              filter-key="status"
              :label="stu_options.education"
                :options="[]"
                v-model="selectedStatus.education"
                @change="handleChange"
            ></SelectItem>
            <SelectItem
              title="性别"
              filter-key="status"
              :label="stu_options.gender"
                :options="[]"
                v-model="selectedStatus.gender"
                @change="handleChange"
            ></SelectItem>
            <SelectItem
              title="专业"
              filter-key="status"
              :label="stu_options.major"
                :options="[]"
                v-model="selectedStatus.major"
                @change="handleChange"
            ></SelectItem>
            <SelectItem
              title="实习次数"
              filter-key="status"
              :label="stu_options.internshipCnt"
                :options="[]"
                v-model="selectedStatus"
                @change="handleChange"
            ></SelectItem>
            <SelectItem
              title="语言"
              filter-key="status"
              :label="stu_options.language"
                :options="[]"
                v-model="selectedStatus"
                @change="handleChange"
            ></SelectItem>
          </SelectorFixed>
        </div>
      </div>
      <!-- 企业信息 -->
      <div class="page company-container" v-if="menuSelect === 2">
        <div class="selector-container">
          <SelectorFixed title="筛选" :show="false"></SelectorFixed>
        </div>
        <h1>企业信息</h1>
      </div>
      <!-- 统计数据 -->
      <div class="page" v-if="menuSelect === 3">
        <h1>统计数据</h1>
      </div>
      <!-- 设置 -->
      <div class="page" v-if="menuSelect === 4">
        <h1>设置</h1>
      </div>
    </div>
  </div>
  

</template>

<style scoped>



.main-container {
  display: flex;
  height: calc(100vh - var(--height-header));
}
.menu-container {
  height: 100vh;
  width: 180px;
  background-color: #f2f2f2;
  margin: 2px 0px;
  margin-right: 5px;
  
}
.content-container {
  width: 100vw;
  height: 100%;
  margin: 2px 2px;
  background-color: #f2f2f2;
}



.page {
  display: flex;
}



</style>
