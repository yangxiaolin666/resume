<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../components/Header.vue'
import SelectorFixed from '../components/SelectorFixed.vue'
import SelectItem from '../components/SelectItem.vue'
import Student_item from '../components/Student_item.vue'
import { Student } from '../types/student.js'

import { stu_options } from '../staticData/student_options.js'
import { stuList } from '../staticData/student_data.js'
import { companyList } from '../staticData/company_data.js'

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

const handleSearch = (query) => {
  console.log('搜索关键词：', query);
  // 后续搜索逻辑将在此实现
};


</script>

<template>
  <Header
    @menu-change="handleMenuSelect"
    @logout="handleLogout"
    @search="handleSearch"
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
              title="第二语言"
              filter-key="status"
              :label="stu_options.language"
                :options="[]"
                v-model="selectedStatus.language"
                @change="handleChange"
            ></SelectItem>
            <SelectItem
              title="技能水平"
              filter-key="status"
              :label="stu_options.skillLevel"
                :options="[]"
                v-model="selectedStatus.skillLevel"
                @change="handleChange"
            ></SelectItem>
          </SelectorFixed>
        </div> <!-- selector-container -->

        <div class="student-list-container">
          <el-table 
            :data="stuList" 
            style="width: 100%"
            height="80%"
            stripe
            >
            <el-table-column prop="id" label="" width="70" class-name="id-column"></el-table-column>
            <el-table-column prop="name" label="姓名" width="120"></el-table-column>
            <el-table-column prop="gender" label="性别" width="120"></el-table-column>
            <el-table-column prop="age" label="年龄" width="120"></el-table-column>
            <el-table-column prop="nation" label="民族" width="120"></el-table-column>
            <el-table-column prop="education" label="学历" width="120"></el-table-column>
            <el-table-column prop="major" label="专业" width="180"></el-table-column>
            <el-table-column prop="language" label="第二语言" width="90"></el-table-column>
            <el-table-column label="..." width="50"></el-table-column>
            <el-table-column label="操作" width="200" class-name="operation-column">
              <template #default="scope">
                <el-button
                  size="mini"
                  @click="handleEdit(scope.$index, scope.row)">查看/编辑</el-button>
                <el-button
                  size="mini"
                  type="danger"
                  @click="handleDelete(scope.$index, scope.row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="student-operate-container">
            <div class="operate-left">
              <el-tooltip content="导入后会覆盖原列表" placement="bottom">
                <el-button>导入</el-button>
              </el-tooltip>
              <el-tooltip content="导出为Excel文件" placement="bottom">
                <el-button>导出</el-button>
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>
      <!-- 企业信息 -->
      <div class="page company-container" v-if="menuSelect === 2">
        <div class="company-list-container">
          <el-table 
            :data="companyList" 
            style="width: 100%"
            height="80%"
            stripe
            >
            <el-table-column prop="id" label="" width="70" class-name="id-column"></el-table-column>
            <el-table-column prop="name" label="企业名称" width="300"></el-table-column>
            <el-table-column prop="city" label="所在城市" width="300"></el-table-column>
            <el-table-column label="操作" width="200" class-name="operation-column">
              <template #default="scope">
                <el-button
                  size="mini"
                  @click="handleEdit(scope.$index, scope.row)">查看/编辑</el-button>
                <el-button
                  size="mini"
                  type="danger"
                  @click="handleDelete(scope.$index, scope.row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="company-operate-container">
            <div class="operate-left">
              <el-tooltip content="导入后会覆盖原列表" placement="bottom">
                <el-button>导入</el-button>
              </el-tooltip>
              <el-tooltip content="导出为Excel文件" placement="bottom">
                <el-button>导出</el-button>
              </el-tooltip>
            </div>
          </div>
        </div>
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
  width: 100vw;
  height: 100%;
}

.student-list-container {
  flex: 1;
  padding: 24px;
  background-color: #ffffff;
  overflow-y: auto;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
}

.student-list-container :deep(.el-table__row) {
  height: 54px;
}

.student-list-container :deep(.el-table td) {
  padding: 6px 0;
}

.student-list-container :deep(.el-table th) {
  padding: 6px 0;
}

.student-list-container :deep(.el-table .operation-column) {
  background-color: #ffffff !important;
}

.student-list-container :deep(.el-table__row .operation-column) {
  background-color: #ffffff !important;
}

.student-list-container :deep(.el-table__row--striped .operation-column) {
  background-color: #ffffff !important;
}

.student-list-container :deep(.el-table__row:hover .operation-column) {
  background-color: #ffffff !important;
}

.student-list-container :deep(.el-table .id-column) {
  color: #999999 !important;
}

.student-list-container :deep(.el-table__row .id-column) {
  color: #999999 !important;
}

.student-list-container :deep(.el-table__row--striped .id-column) {
  color: #999999 !important;
}

.student-list-container :deep(.el-table__row:hover .id-column) {
  color: #999999 !important;
}

.student-operate-container {
  margin-top: 24px;
  padding: 16px 0 8px;
  border-top: 1px solid #ebeef5;
}

.operate-left {
  display: flex;
  align-items: stretch;
  gap: 12px;
  width: fit-content;
}

.operate-left .el-button {
  width: 90px;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.operate-left .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.company-list-container {
  flex: 1;
  padding: 24px;
  background-color: #ffffff;
  overflow-y: auto;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
}

.company-list-container :deep(.el-table__row) {
  height: 54px;
}

.company-list-container :deep(.el-table td) {
  padding: 6px 0;
}

.company-list-container :deep(.el-table th) {
  padding: 6px 0;
}

.company-list-container :deep(.el-table .operation-column) {
  background-color: #ffffff !important;
}

.company-list-container :deep(.el-table__row .operation-column) {
  background-color: #ffffff !important;
}

.company-list-container :deep(.el-table__row--striped .operation-column) {
  background-color: #ffffff !important;
}

.company-list-container :deep(.el-table__row:hover .operation-column) {
  background-color: #ffffff !important;
}

.company-list-container :deep(.el-table .id-column) {
  color: #999999 !important;
}

.company-list-container :deep(.el-table__row .id-column) {
  color: #999999 !important;
}

.company-list-container :deep(.el-table__row--striped .id-column) {
  color: #999999 !important;
}

.company-list-container :deep(.el-table__row:hover .id-column) {
  color: #999999 !important;
}

.company-operate-container {
  margin-top: 24px;
  padding: 16px 0 8px;
  border-top: 1px solid #ebeef5;
}

.company-container {
  padding: 0px 20px;
}

</style>
