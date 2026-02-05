<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Search,
  User,
  Document,
  House,
  SwitchButton
} from '@element-plus/icons-vue'

import IconSchool from '@/components/IconSchool.vue'
import { useUserStore } from '@/stores/user'
import { useStudentsStore } from '@/stores/students'
import {
  generateFilterOptions,
  getDefaultFilters,
  getFilteredAndPaginatedStudents
} from '@/utils/filterUtils'

const router = useRouter()
const userStore = useUserStore()
const studentsStore = useStudentsStore()

// 响应式数据
const activeIndex = ref('1')
const currentStudent = ref(null)
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

// 筛选条件
const filters = ref(getDefaultFilters())

// 筛选选项（从学生数据生成）
const filterOptions = computed(() => {
  return generateFilterOptions(studentsStore.allStudents)
})

// 计算属性：筛选和分页后的学生数据
const paginatedStudents = computed(() => {
  const result = getFilteredAndPaginatedStudents(
    studentsStore.allStudents,
    {
      ...filters.value,
      name: searchQuery.value
    },
    currentPage.value,
    pageSize.value
  )
  return result
})

// 处理筛选条件变化
const handleFilterChange = (key, value) => {
  filters.value[key] = value
  currentPage.value = 1 // 重置到第一页
}

// 处理多值筛选变化
const handleMultiValueChange = (key, value) => {
  filters.value[key] = value
  currentPage.value = 1
}

// 重置筛选条件
const resetFilters = () => {
  filters.value = getDefaultFilters()
  searchQuery.value = ''
  currentPage.value = 1
}

// 查看学生详情
const viewStudentDetail = (student) => {
  currentStudent.value = student
  studentsStore.setSelectedStudent(student)
  // 这里可以添加路由跳转或弹窗显示详情
  console.log('查看学生详情:', student)
}

// 跳转到用户页面
const goToUser = () => {
  router.push(`/${userStore.userType}`)
}

// 退出登录
const logOut = () => {
  userStore.setAll('', '', '')
  ElMessage({
    message: '退出登录成功',
    type: 'success',
    duration: 1000
  })
  setTimeout(() => router.push('/login'), 1100)
}

// 处理分页变化
const handlePageChange = (page) => {
  currentPage.value = page
}

// 组件挂载时初始化
onMounted(() => {
  // 如果需要从API加载数据，可以在这里调用
  console.log('Home component mounted')
})
</script>

<template>
  <div class="home-container">
    <!-- 顶部导航栏 -->
    <el-menu :default-active="activeIndex" class="home-nav" mode="horizontal" :ellipsis="false">
      <div class="nav-left">
        <IconSchool />
        <h1>学生招聘平台</h1>
      </div>

      <div class="nav-right">
        <el-menu-item index="3">
          <el-icon class="nav-icon">
            <Document />
          </el-icon>
        </el-menu-item>
        <el-menu-item index="4">
          <el-popover placement="bottom-end" width="200">
            <template #reference>
              <el-icon class="nav-icon">
                <User />
              </el-icon>
            </template>
            <div class="user-menu">
              <div class="menu-item" @click="goToUser">
                <el-icon>
                  <House />
                </el-icon>
                <span>个人主页</span>
              </div>
              <div class="menu-item" @click="logOut">
                <el-icon>
                  <SwitchButton />
                </el-icon>
                <span>退出登录</span>
              </div>
            </div>
          </el-popover>
        </el-menu-item>
      </div>
    </el-menu>

    <!-- 主要内容区域 -->
    <main class="main-content">
      <!-- 筛选面板 -->
      <aside class="filter-sidebar">
        <el-card class="filter-card">
          <template #header>
            <div class="filter-header">
              <span>筛选条件</span>
              <el-button type="text" @click="resetFilters"> 重置 </el-button>
            </div>
          </template>

          <!-- 搜索框 -->
          <div class="filter-section">
            <el-input v-model="searchQuery" placeholder="搜索学生姓名..." clearable style="margin-bottom: 16px"
              @input="currentPage = 1">
              <template #prefix>
                <el-icon>
                  <Search />
                </el-icon>
              </template>
            </el-input>
          </div>

          <!-- 学位筛选 -->
          <div class="filter-section">
            <h4>学位</h4>
            <el-radio-group v-model="filters.degree" size="small"
              @change="(value) => handleFilterChange('degree', value)">
              <el-radio-button label=""> 全部 </el-radio-button>
              <el-radio-button v-for="degree in filterOptions.degrees" :key="degree" :label="degree">
                {{ degree }}
              </el-radio-button>
            </el-radio-group>
          </div>

          <!-- 专业筛选 -->
          <div class="filter-section">
            <h4>专业</h4>
            <el-select v-model="filters.major" placeholder="选择专业" clearable style="width: 100%" size="small"
              @change="(value) => handleFilterChange('major', value)">
              <el-option label="全部专业" value="" />
              <el-option v-for="major in filterOptions.majors" :key="major" :label="major" :value="major" />
            </el-select>
          </div>

          <!-- 专业技能筛选 -->
          <div class="filter-section">
            <h4>专业技能</h4>
            <el-select v-model="filters.specialities" placeholder="选择技能" multiple clearable style="width: 100%"
              size="small" @change="(value) => handleMultiValueChange('specialities', value)">
              <el-option v-for="skill in filterOptions.specialities" :key="skill" :label="skill" :value="skill" />
            </el-select>
          </div>

          <!-- 语言筛选 -->
          <div class="filter-section">
            <h4>语言能力</h4>
            <el-select v-model="filters.languages" placeholder="选择语言" multiple clearable style="width: 100%"
              size="small" @change="(value) => handleMultiValueChange('languages', value)">
              <el-option v-for="lang in filterOptions.languages" :key="lang" :label="lang" :value="lang" />
            </el-select>
          </div>

          <!-- 实习次数筛选 -->
          <div class="filter-section">
            <h4>实习经验</h4>
            <el-select v-model="filters.internship" placeholder="选择实习次数" clearable style="width: 100%" size="small"
              @change="(value) => handleFilterChange('internship', value)">
              <el-option label="不限" :value="null" />
              <el-option label="无实习经验" :value="0" />
              <el-option label="1次实习" :value="1" />
              <el-option label="2次及以上" :value="2" />
            </el-select>
          </div>

          <!-- 毕业年份筛选 -->
          <div class="filter-section">
            <h4>毕业年份</h4>
            <el-select v-model="filters.graduationYear" placeholder="选择毕业年份" clearable style="width: 100%" size="small"
              @change="(value) => handleFilterChange('graduationYear', value)">
              <el-option label="不限" :value="null" />
              <el-option label="2023年及以后" :value="2023" />
              <el-option label="2024年及以后" :value="2024" />
              <el-option label="2025年及以后" :value="2025" />
            </el-select>
          </div>

          <!-- 城市筛选 -->
          <div class="filter-section">
            <h4>城市</h4>
            <el-select v-model="filters.cities" placeholder="选择城市" multiple clearable style="width: 100%" size="small"
              @change="(value) => handleMultiValueChange('cities', value)">
              <el-option v-for="city in filterOptions.cities" :key="city" :label="city" :value="city" />
            </el-select>
          </div>

          <!-- 证书筛选 -->
          <div class="filter-section">
            <h4>证书</h4>
            <el-select v-model="filters.certNames" placeholder="选择证书" multiple clearable style="width: 100%"
              size="small" @change="(value) => handleMultiValueChange('certNames', value)">
              <el-option v-for="cert in filterOptions.certNames" :key="cert" :label="cert" :value="cert" />
            </el-select>
          </div>
        </el-card>
      </aside>

      <!-- 学生列表 -->
      <div class="student-content">
        <el-card class="student-card">
          <template #header>
            <div class="student-header">
              <span>学生列表</span>
              <span class="student-count">
                共 {{ paginatedStudents.total }} 名学生
              </span>
            </div>
          </template>

          <div v-if="paginatedStudents.data.length > 0" class="student-grid">
            <el-row :gutter="16">
              <el-col v-for="student in paginatedStudents.data" :key="student.id" :xs="24" :sm="12" :md="8">
                <el-card class="student-item" shadow="hover">
                  <template #header>
                    <div class="student-item-header">
                      <span class="student-name">{{ student.name }}</span>
                      <el-tag :type="student.internship === 0 ? 'info' : 'success'" size="small">
                        {{
                          student.internship === 0
                            ? '应届生'
                            : student.internship + '次实习'
                        }}
                      </el-tag>
                    </div>
                  </template>

                  <div class="student-info">
                    <div class="info-row">
                      <span class="label">学号：</span>
                      <span>{{ student.stuId }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">专业：</span>
                      <span>{{ student.major }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">学位：</span>
                      <span>{{ student.degree }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">年龄：</span>
                      <span>{{ student.age }}岁</span>
                    </div>
                    <div class="info-row">
                      <span class="label">毕业年份：</span>
                      <span>{{ student.graduationYear }}</span>
                    </div>

                    <!-- 技能标签 -->
                    <div class="skills-section">
                      <h5>专业技能</h5>
                      <div class="skills-tags">
                        <el-tag v-for="skill in student.specialities.slice(0, 3)" :key="skill" type="primary"
                          size="small" class="skill-tag">
                          {{ skill }}
                        </el-tag>
                        <el-tag v-if="student.specialities.length > 3" size="small">
                          +{{ student.specialities.length - 3 }}
                        </el-tag>
                      </div>
                    </div>

                    <!-- 语言能力 -->
                    <div v-if="student.languages.length > 0" class="languages-section">
                      <h5>语言能力</h5>
                      <div class="languages-tags">
                        <el-tag v-for="lang in student.languages" :key="lang" type="warning" size="small">
                          {{ lang }} {{ student.languageLevel }}
                        </el-tag>
                      </div>
                    </div>

                    <!-- 证书 -->
                    <div v-if="student.certNames.length > 0" class="certificates-section">
                      <h5>证书</h5>
                      <div class="certificates-tags">
                        <el-tag v-for="cert in student.certNames.slice(0, 2)" :key="cert" type="success" size="small">
                          {{ cert }}
                        </el-tag>
                        <el-tag v-if="student.certNames.length > 2" size="small">
                          +{{ student.certNames.length - 2 }}
                        </el-tag>
                      </div>
                    </div>

                    <!-- 实习公司 -->
                    <div v-if="student.internshipCompany" class="internship-section">
                      <h5>实习公司</h5>
                      <p class="internship-company">
                        {{ student.internshipCompany }}
                      </p>
                    </div>

                    <!-- 意向城市 -->
                    <div v-if="student.cities.length > 0" class="cities-section">
                      <h5>意向城市</h5>
                      <div class="cities-tags">
                        <el-tag v-for="city in student.cities.slice(0, 2)" :key="city" type="info" size="small">
                          {{ city }}
                        </el-tag>
                        <el-tag v-if="student.cities.length > 2" size="small">
                          +{{ student.cities.length - 2 }}
                        </el-tag>
                      </div>
                    </div>
                  </div>

                  <template #footer>
                    <div class="student-actions">
                      <el-button type="primary" size="small" @click="viewStudentDetail(student)">
                        查看详情
                      </el-button>
                      <el-button type="success" size="small">
                        收藏简历
                      </el-button>
                    </div>
                  </template>
                </el-card>
              </el-col>
            </el-row>

            <!-- 分页 -->
            <div class="pagination-container">
              <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 50]"
                layout="total, sizes, prev, pager, next, jumper" :total="paginatedStudents.total"
                @size-change="(size) => (pageSize = size)" @current-change="handlePageChange" />
            </div>
          </div>

          <div v-else class="empty-state">
            <el-empty description="没有找到符合条件的学生">
              <el-button type="primary" @click="resetFilters">
                重置筛选条件
              </el-button>
            </el-empty>
          </div>
        </el-card>
      </div>
    </main>
  </div>
</template>

<style scoped>
.home-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
}

/* 导航栏样式 */
.home-nav {
  background-color: #fff;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  height: 80px;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.nav-left h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.nav-right {
  display: flex;
  align-items: center;
}

.nav-icon {
  font-size: 24px;
  cursor: pointer;
  transition: all 0.3s;
}

.nav-icon:hover {
  color: #409eff;
  transform: scale(1.1);
}

.user-menu {
  padding: 8px 0;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 8px 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.menu-item:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

.menu-item span {
  margin-left: 12px;
  font-size: 14px;
}

/* 主要内容样式 */
.main-content {
  flex: 1;
  display: flex;
  gap: 20px;
  padding: 20px;
  margin-top: 80px;
}

.filter-sidebar {
  width: 280px;
  flex-shrink: 0;
}

.filter-card {
  border-radius: 12px;
  overflow: hidden;
  background-color: #fff;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-section {
  margin-bottom: 20px;
}

.filter-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #606266;
}

/* 学生列表样式 */
.student-content {
  flex: 1;
}

.student-card {
  border-radius: 12px;
  overflow: hidden;
  background-color: #fff;
}

.student-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.student-count {
  font-size: 14px;
  color: #909399;
}

.student-grid {
  margin-top: 20px;
}

.student-item {
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s;
  background-color: #fff;
  margin-bottom: 20px;
}

.student-item:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 20px rgba(0, 0, 0, 0.15);
}

.student-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.student-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.student-info {
  margin-top: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
}

.label {
  font-weight: 500;
  color: #606266;
}

.skills-section,
.languages-section,
.certificates-section,
.internship-section,
.cities-section {
  margin-top: 16px;
}

.skills-section h5,
.languages-section h5,
.certificates-section h5,
.internship-section h5,
.cities-section h5 {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: #606266;
}

.skills-tags,
.languages-tags,
.certificates-tags,
.cities-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.skill-tag {
  margin-bottom: 4px;
}

.internship-company {
  margin: 0;
  font-size: 13px;
  color: #606266;
}

.student-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 16px;
}

.pagination-container {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}

/* 空状态样式 */
.empty-state {
  padding: 60px 0;
  text-align: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .main-content {
    flex-direction: column;
    padding: 10px;
  }

  .filter-sidebar {
    width: 100%;
  }

  .nav-left h1 {
    font-size: 16px;
  }
}
</style>
