<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../components/Header.vue'
import SelectorFixed from '../components/SelectorFixed.vue'
import SelectItem from '../components/SelectItem.vue'
import Student_item from '../components/Student_item.vue'
import { Student } from '../types/student.js'

import { stu_options } from '../staticData/student_options.js'

const menuSelect = ref(1); // 1: 学生信息, 2: 企业信息
const handleMenuSelect = (index) => {
  menuSelect.value = index;
  console.log("改变");
};
const handleLogout = () => {
  console.log("退出登录");
}
const selectedStatus = ref({});

const handleChange = (value) => {
  console.log('筛选：', value);
  console.log('筛选条件：', selectedStatus.value);
};

const handleReset = () => {
  selectedStatus.value = {};
  console.log('已重置筛选条件');
};


const tempStudent = new Student({
  id: 1,
  stuId: '20230001',
  name: '张三',
  age: 22,
  gender: '男',
  nation: '汉族',
  major: '计算机科学与技术',
  degree: '本科',
  graduationYear: 2023,
  internship: 2,
  languages: ['英语', '日语'],
  languageLevel: '六级',
  cities: ['北京', '上海'],
  provinces: ['广东', '江苏'],
  certNames: ['AWS', 'PMP'],
  certLevel: '高级',
  internshipCompany: '腾讯',
  specialities: ['Java', 'Python'],
  education: '本科',
  internshipCnt: 1,
  language: '中文',
})

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
          <SelectorFixed title="筛选" @reset="handleReset">
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
                v-model="selectedStatus.internshipCnt"
                @change="handleChange"
            ></SelectItem>
            <SelectItem
              title="语言"
              filter-key="status"
              :label="stu_options.language"
                :options="[]"
                v-model="selectedStatus.language"
                @change="handleChange"
            ></SelectItem>
          </SelectorFixed>
        </div> <!-- selector-container -->
        <div class="student-list-container">
          <Student_item class="student-header" :isHeader="true"></Student_item>
          <div class="student-container">
            <!-- <Student_item v-for="item in stuList"
              :studentInfo="item"></Student_item> -->
          </div>
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
