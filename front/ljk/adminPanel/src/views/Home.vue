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

import { ElMessage, ElMessageBox } from 'element-plus'


const backendUrl = 'http://81.70.199.174:8081/admin';

const currMenu = {
  Student: 1,
  Company: 2
}

const TOKEN = ref('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjMsInVzZXJuYW1lIjoiYWRtaW4iLCJ1c2VyVHlwZSI6ImFkbWluIiwidG9rZW5UeXBlIjoiYWNjZXNzIiwiZXhwIjoxNzczMDU5NDg5fQ.u2j3CiaCIXapM4PDwpRDjM7oCvycA-nIAJfjDNcm4pw');

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

// 删除
const isDeleteMode = ref(false)
const selectedIds = ref([])

const enterDeleteMode = () => {
  isDeleteMode.value = true
  selectedIds.value = []
}
const cancelDeleteMode = () => {
  isDeleteMode.value = false
  selectedIds.value = []
}
const handleSelectionChange = (selection) => {
  if (menuSelect.value === currMenu.Student) {
    selectedIds.value = selection.map(item => item.id)
  } else {
    selectedIds.value = selection.map(item => item.id)
  }
}

// 确认批量删除
const confirmBatchDelete = () => {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请先选择要删除的项')
    return
  }
  
  const count = selectedIds.value.length
  ElMessageBox.confirm(`确认删除选中的 ${count} 项?`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    if (menuSelect.value === currMenu.Student) {
      batchDeleteStudents()
    } else {
      batchDeleteCompanies()
    }
  }).catch(() => {
    console.log('取消批量删除')
  })
}

// 批量删除学生
const batchDeleteStudents = () => {
  console.log(selectedIds.value);
  axios.delete(backendUrl + '/delStu', {
    headers: {
      'Authorization': `Bearer ${TOKEN.value}`
    },
    data: {
      stuIds: selectedIds.value
    }
  }).then(res => {
    console.log(res.data);
    if (res.data.code === 200) {
      ElMessage.success('删除成功！')
      isDeleteMode.value = false
      selectedIds.value = []
      reGetList()
    } else {
      ElMessage.error('删除失败！')
    }
  }).catch(err => {
    console.log(err)
    ElMessage.error('删除失败！')
  })
}

// 批量删除企业
const batchDeleteCompanies = () => {
  axios.delete(backendUrl + '/delCom', {
    headers: {
      'Authorization': `Bearer ${TOKEN.value}`
    },
    data: {
      comIds: selectedIds.value
    }
  }).then(res => {
    if (res.data.code === 200) {
      ElMessage.success('删除成功！')
      isDeleteMode.value = false
      selectedIds.value = []
      reGetList()
    } else {
      ElMessage.error('删除失败！')
    }
  }).catch(err => {
    console.log(err)
    ElMessage.error('删除失败！')
  })
}


//导入框显示状态
const fileList = ref([])
const importDialogVisible = ref(false)
const uploadHeaders = {
  'Authorization': `Bearer ${TOKEN.value}`
}
const uploadRef = ref(null)

const handleImport = () => {
  importDialogVisible.value = true
  console.log(importDialogVisible.value);
}
const beforeRemove = (file, fileList) => {
  ElMessageBox.confirm(`确定移除 ${ file.name }？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    fileList.splice(fileList.indexOf(file), 1);
  }).catch(() => {
    ElMessage({
      type: 'info',
      message: '已取消移除'
    });
  });
}
const handleRemove = (file, fileList) => {
  console.log(file, fileList);
}
const handleExceed = (files, fileList) => {
  fileList.splice(0, fileList.length);
  fileList.push(...files.slice(0, 1));
}
const handleUploadSuccess = (res) => {
  console.log(res);
  if (res.code === 200) {
    ElMessage.success('上传成功！');
    importDialogVisible.value = false;
    reGetList();
    fileList.value = [];
  } else {
    ElMessage.error('上传失败！');
  }
}
const handleUploadError = (res, file) => {
  ElMessage.error('上传失败！');
}
const beforeUpload = (file) => {
  const isExcel = /\.(xls|xlsx)$/.test(file.name);
  const isLt2M = file.size / 1024 / 1024 < 2;
  
  if (!isExcel) {
    ElMessage.error('只能上传xls/xlsx格式的Excel文件！');
    return false;
  }
  if (!isLt2M) {
    ElMessage.error('文件大小不能超过2MB！');
    return false;
  }
  return true;
};

// 确认上传
const handleConfirmUpload = () => {
  uploadRef.value.submit();
};



const handleExport = () => {
  console.log('导出操作，还没有');
}


const handleDelete = (id, name) => {
  if (menuSelect.value === currMenu.Student) {
    deleteStudent(id, name)
  } else if (menuSelect.value === currMenu.Company) {
    deleteCompany(id, name)
  }
}
const deleteStudent = (stuIds, stuName) => {
  ElMessageBox.confirm('确认删除该学生 ' + stuName + " ?", '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    console.log('删除学生 ' + stuName);
    axios.delete(backendUrl + '/delStu', {
      headers: {
        'Authorization': `Bearer ${TOKEN.value}`
      },
      body: {
        stuIds: stuIds
      }
    }).then(res => {
      console.log(res.data);
      if (res.data.code === 200) {
        ElMessage.success('删除成功！');
        reGetList();
      } else {
        ElMessage.error('删除失败！');
      }
    }).catch(err => {
      console.log(err);
    })
  }).catch(() => {
    console.log('取消删除学生 ' + stuName);
  });
}
const deleteCompany = (companyIds, companyName) => {
  ElMessageBox.confirm('确认删除该公司 ' + companyName + " ?", '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    console.log('删除公司 ' + companyName);
    axios.delete(backendUrl + '/delCom', {
      headers: {
        'Authorization': `Bearer ${TOKEN.value}`
      },
      body: {
        companyIds: companyIds
      }
    }).then(res => {
      if (res.data.code === 200) {
        ElMessage.success('删除成功！');
        reGetList();
      } else {
        ElMessage.error('删除失败！');
      }
    }).catch(err => {
      console.log(err);
    })
  }).catch(() => {
    console.log('取消删除公司 ' + companyName);
  });
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
            @selection-change="handleSelectionChange"
            >
            <el-table-column v-if="isDeleteMode" type="selection" width="35"></el-table-column>
            <!-- <el-table-column v-else label="" width="35"></el-table-column> -->
            <el-table-column prop="stuId" label="学号" width="120"></el-table-column>
            <el-table-column prop="name" label="姓名" width="120"></el-table-column>
            <el-table-column prop="gender" label="性别" width="120"></el-table-column>
            <el-table-column prop="age" label="年龄" width="80"></el-table-column>
            <el-table-column prop="nation" label="民族" width="120"></el-table-column>
            <el-table-column prop="degree" label="学位" width="120"></el-table-column>
            <el-table-column prop="major" label="专业" width="160"></el-table-column>
            <el-table-column prop="graduationYear" label="毕业年份" width="80"></el-table-column>
            <el-table-column label="..." width="50"></el-table-column>
            <el-table-column v-if="!isDeleteMode" label="操作" width="200" class-name="operation-column">
              <template #default="scope">
                <el-button
                  size="medium"
                  @click="handleEdit(scope.row)">查看/编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-container">
            <div class="pagination-left">
              <el-dropdown>
                <el-button class="action-btn">
                  导入/导出 <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleImport()">&nbsp;&nbsp;导入&nbsp;&nbsp;</el-dropdown-item>
                    <el-dropdown-item @click="handleExport()">&nbsp;&nbsp;导出&nbsp;&nbsp;</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <div class="delete-btn-container" v-if="!isDeleteMode">
                <el-button class="action-btn" type="danger" @click="enterDeleteMode">删除</el-button>
              </div>
              <div class="delete-btn-container" v-else>
                <el-button class="action-btn" type="danger" @click="confirmBatchDelete">确认删除</el-button>
                <el-button class="action-btn" @click="cancelDeleteMode">取消删除</el-button>
              </div>
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
            @selection-change="handleSelectionChange"
            >
            <el-table-column v-if="isDeleteMode" type="selection" width="35"></el-table-column>
            <el-table-column prop="id" label="" width="70" class-name="id-column"></el-table-column>
            <el-table-column prop="comName" label="企业名称" width="300"></el-table-column>
            <el-table-column prop="comRegion" label="所在城市" width="300"></el-table-column>
            <el-table-column v-if="!isDeleteMode" label="操作" width="200" class-name="operation-column">
              <template #default="scope">
                <el-button
                  size="medium"
                  @click="handleEdit(scope.$index, scope.row)">查看/编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-container">
            <div class="pagination-left">
              <el-dropdown>
                <el-button class="action-btn">
                  导入导出 <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleImport()">&nbsp;&nbsp;导入&nbsp;&nbsp;</el-dropdown-item>
                    <el-dropdown-item @click="handleExport()">&nbsp;&nbsp;导出&nbsp;&nbsp;</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <div class="delete-btn-container" v-if="!isDeleteMode">
                <el-button class="action-btn" type="danger" @click="enterDeleteMode">删除</el-button>
              </div>
              <div class="delete-btn-container" v-else>
                <el-button class="action-btn" type="danger" @click="confirmBatchDelete">确认删除</el-button>
                <el-button class="action-btn" @click="cancelDeleteMode">取消删除</el-button>
              </div>
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
          
          width="60%"
          :show-close="true"
          :close-on-click-modal="false"
        >
          <el-upload
            ref="uploadRef"
            class="upload-container"
            drag
            :limit="1"
            :on-exceed="handleExceed"
            :on-remove="handleRemove"
            :file-list="fileList"
            :before-remove="beforeRemove"
            :headers="uploadHeaders"
            :action="menuSelect === 1 ? backendUrl + '/addStu' : backendUrl + '/addCom'"
            :on-success="handleUploadSuccess"
            :on-error="handleUploadError"
            :before-upload="beforeUpload"
            :auto-upload="false"
            accept=".xls,.xlsx"
            :name="menuSelect === 1 ? 'stuFile' : 'comFile'"
            :multiple="false">
              <img style="width: 100px; height: 100px;" src="/upload.svg"></img>
              <div style="font-size: 14px; color: #666666; margin-top: 12px;">将文件拖到此处，或<em style="color: #409eff;">点击上传</em></div>
              <div style="font-size: 12px; color: #999999; margin-top: 12px;" slot="tip">只能上传xls/xlsx文件，且不超过2MB</div>
          </el-upload>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="importDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="handleConfirmUpload">确认导入</el-button>
            </span>
          </template>
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

.pagination-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pagination-left .action-btn {
  width: 120px;
}

.delete-btn-container {
  display: flex;
  gap: 8px;
}

</style>
