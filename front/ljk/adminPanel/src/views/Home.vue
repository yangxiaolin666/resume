<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../components/Header.vue'
import SelectorFixed from '../components/SelectorFixed.vue'
import SelectItem from '../components/SelectItem.vue'
import Student_item from '../components/Student_item.vue'
import { Student } from '../types/student.js'
import { ArrowDown } from '@element-plus/icons-vue'

import { stu_options } from '../staticData/student_options.js'
import axios from 'axios'




const backendUrl = 'http://81.70.199.174:8081/admin';

const currMenu = {
  Student: 1,
  Company: 2
}

const TOKEN = ref('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjMsInVzZXJuYW1lIjoiYWRtaW4iLCJ1c2VyVHlwZSI6ImFkbWluIiwidG9rZW5UeXBlIjoiYWNjZXNzIiwiZXhwIjoxNzcyNzE4OTAyfQ.c1es0MRjTukRk1ZSYn82wdWf501Pbx8qZgWgOuRNcng');

const menuSelect = ref(1); // 1: 学生信息, 2: 企业信息

const handleMenuSelect = (index) => {
  menuSelect.value = index;
  console.log("to" + (menuSelect.value === currMenu.Student ? "学生信息" : "企业信息"));
  
  if (menuSelect.value === currMenu.Student) {
    if (currentStudentList.value.length === 0) {
      reGetList();
    }
  } else if (menuSelect.value === currMenu.Company) {
    if (currentCompanyList.value.length === 0) {
      reGetList();
    }
  }
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


onMounted(() => {
  // 1. 初始化列表
  reGetList();
});

const reGetList = () => {
  if (menuSelect.value === currMenu.Student) {
    getStudentList()
    console.log("getStudentList");
  } else if (menuSelect.value === currMenu.Company) {
    getCompanyList()
    console.log("getCompanyList");
  }
}



const getStudentList = (name) => {
  var token = TOKEN.value
  axios.get(backendUrl + '/searchStu', {
    headers: {
      'Authorization': `Bearer ${token}`
    },
    params: {
      page: currentPage.value,
      pageSize: pageSize.value,
      username: name ? name : null
    }
  }).then(res => {
    if (res.data.code === 200) {
      console.log(res.data.data);
      currentStudentList.value = res.data.data.list
      totalStudent.value = res.data.data.pagination.total
    }
  }).catch(err => {
    console.log(err);
  })
}

const getCompanyList = (name) => {
  var token = TOKEN.value
  axios.get(backendUrl + '/searchCom', {
    headers: {
      'Authorization': `Bearer ${token}`
    },
    params: {
      page: currentPageCompany.value,
      pageSize: pageSizeCompany.value,
      companyName: name ? name : null
    }
  }).then(res => {
    if (res.data.code === 200) {
      console.log(res.data.data);
      currentCompanyList.value = res.data.data.list
      totalCompany.value = res.data.data.pagination.total
    }
  }).catch(err => {
    console.log(err);
  })
}

const handleSearch = (name) => {
  if (menuSelect.value === currMenu.Student) {
    getStudentList(name)
  } else if (menuSelect.value === currMenu.Company) {
    getCompanyList(name)
  }
}

// 列表分页
const currentPage = ref(1)
const pageSize = ref(20)
const currentPageCompany = ref(1)
const pageSizeCompany = ref(20)

const totalStudent = ref(0);
const totalCompany = ref(0);
const currentStudentList = ref([]);
const currentCompanyList = ref([]);

const handleSizeChange = (val) => {
  pageSize.value = val
  currentPage.value = 1
  reGetList()
}
const handleCurrentChange = (val) => {
  currentPage.value = val
  reGetList()
}
const handleSizeChangeCompany = (val) => {
  pageSizeCompany.value = val
  currentPageCompany.value = 1
  reGetList()
}
const handleCurrentChangeCompany = (val) => {
  currentPageCompany.value = val
  reGetList()
}
const getPaginatedStudentList = () => {
  return currentStudentList.value
}
const getPaginatedCompanyList = () => {
  return currentCompanyList.value
}


//导入框显示状态
const importDialogVisible = ref(false)

const handleImport = () => {
  importDialogVisible.value = true
  console.log(importDialogVisible.value);
}



const handleExport = () => {
  console.log('导出操作');
}

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
            :data="getPaginatedStudentList()" 
            style="width: 100%"
            height="80%"
            stripe
            >
            <el-table-column prop="stuId" label="学号" width="120"></el-table-column>
            <el-table-column prop="name" label="姓名" width="120"></el-table-column>
            <el-table-column prop="gender" label="性别" width="120"></el-table-column>
            <el-table-column prop="age" label="年龄" width="80"></el-table-column>
            <el-table-column prop="nation" label="民族" width="120"></el-table-column>
            <el-table-column prop="degree" label="学位" width="120"></el-table-column>
            <el-table-column prop="major" label="专业" width="160"></el-table-column>
            <el-table-column prop="graduationYear" label="毕业年份" width="80"></el-table-column>
            <el-table-column label="..." width="50"></el-table-column>
            <el-table-column label="操作" width="200" class-name="operation-column">
              <template #default="scope">
                <el-button
                  size="small"
                  @click="handleEdit(scope.$index, scope.row)">查看/编辑</el-button>
                <el-button
                  size="small"
                  type="danger"
                  @click="handleDelete(scope.$index, scope.row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-container">
            <div class="pagination-left">
              <el-dropdown>
                <el-button>
                  导入/导出 <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleImport()">导入</el-dropdown-item>
                    <el-dropdown-item @click="handleExport()">导出</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <div class="pagination-center">
              <el-pagination
                v-model:current-page="currentPage"
                v-model:page-size="pageSize"
                :page-sizes="[10, 20, 50]"
                :small="true"
                :background="true"
                layout="total, sizes, prev, pager, next, jumper"
                :total="totalStudent"
                @size-change="handleSizeChange"
                @current-change="handleCurrentChange"
              />
            </div>
          </div>
        </div>
      </div>
      <!-- 企业信息 -->
      <div class="page company-container" v-if="menuSelect === 2">
        <div class="company-list-container">
          <el-table 
            :data="getPaginatedCompanyList()" 
            style="width: 100%"
            height="80%"
            stripe
            >
            <el-table-column prop="id" label="" width="70" class-name="id-column"></el-table-column>
            <el-table-column prop="comName" label="企业名称" width="300"></el-table-column>
            <el-table-column prop="comRegion" label="所在城市" width="300"></el-table-column>
            <el-table-column label="操作" width="200" class-name="operation-column">
              <template #default="scope">
                <el-button
                  size="small"
                  @click="handleEdit(scope.$index, scope.row)">查看/编辑</el-button>
                <el-button
                  size="small"
                  type="danger"
                  @click="handleDelete(scope.$index, scope.row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-container">
            <div class="pagination-left">
              <el-dropdown>
                <el-button>
                  导入导出 <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleImport()">导入</el-dropdown-item>
                    <el-dropdown-item @click="handleExport">导出</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <div class="pagination-center">
              <el-pagination
                v-model:current-page="currentPageCompany"
                v-model:page-size="pageSizeCompany"
                :page-sizes="[10, 20, 50]"
                :small="true"
                :background="true"
                layout="total, sizes, prev, pager, next, jumper"
                :total="totalCompany"
                @size-change="handleSizeChangeCompany"
                @current-change="handleCurrentChangeCompany"
              />
            </div>
          </div>
        </div>
      </div>
      <!-- 导入信息页 -->
      <div class="import-container">
        <el-dialog
          :title="menuSelect === 1 ? '导入学生信息' : '导入企业信息'"
          v-model="importDialogVisible"
          action="https://jsonplaceholder.typicode.com/posts/"
          width="60%"
        >
          <el-upload
            class="upload-container"
            drag
            multiple>
              <img style="width: 100px; height: 100px;" src="/upload.svg"></img>
              <div style="font-size: 14px; color: #666666; margin-top: 12px;">将文件拖到此处，或<em style="color: #409eff;">点击上传</em></div>
              <div style="font-size: 12px; color: #999999; margin-top: 12px;" slot="tip">只能上传xls/xlsx文件，且不超过2MB</div>
          </el-upload>
        </el-dialog>
      </div>
      <!-- 详情页 -->

      <!-- 统计数据 -->
      <!-- <div class="page" v-if="menuSelect === 3">
        <h1>统计数据</h1>
      </div> -->
      <!-- 设置 -->
      <!-- <div class="page" v-if="menuSelect === 4">
        <h1>设置</h1>
      </div> -->
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

.pagination-container {
  display: flex;
  align-items: center;
  margin-top: 24px;
  padding: 16px 0 8px;
  border-top: 1px solid #ebeef5;
}

.pagination-left {
  margin-right: auto;
}

.pagination-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.pagination-container :deep(.el-pagination) {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pagination-container :deep(.el-pagination.is-background .el-pager li) {
  background-color: #f4f4f5;
  border-radius: 4px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.pagination-container :deep(.el-pagination.is-background .el-pager li:not(.disabled).is-active) {
  background-color: #409eff;
  color: #ffffff;
}

.pagination-container :deep(.el-pagination.is-background .el-pager li:hover:not(.disabled)) {
  background-color: #e6f7ff;
  color: #409eff;
}

.pagination-container :deep(.el-pagination.is-background .btn-prev),
.pagination-container :deep(.el-pagination.is-background .btn-next) {
  background-color: #f4f4f5;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.pagination-container :deep(.el-pagination.is-background .btn-prev:hover),
.pagination-container :deep(.el-pagination.is-background .btn-next:hover) {
  background-color: #e6f7ff;
  color: #409eff;
}

.pagination-container :deep(.el-pagination__sizes .el-select .el-input__wrapper) {
  border-radius: 4px;
}

.pagination-container :deep(.el-pagination__jump) {
  margin-left: 8px;
}

.pagination-container :deep(.el-pagination__total) {
  color: #606266;
  font-weight: 500;
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

.company-container {
  padding: 0px 20px;
}


</style>
