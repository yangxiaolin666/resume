<script setup>
import {
  Delete,
  Search,
} from '@element-plus/icons-vue'
import { ref, computed, onMounted } from 'vue'
import jobCard from '@/components/jobCard.vue'
import { getJobPostings } from '@/api/recruitment'

const jobList = ref([])

const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const keyword = ref('')
const status = ref('')

const jobCategories = ref([
  'Java开发工程师',
  'Python开发工程师',
  '前端开发工程师',
  '后端开发工程师',
  '全栈工程师',
  '算法工程师',
  '测试工程师',
  '运维工程师',
  '产品经理',
  'UI设计师'
])

const selectedCategories = ref([])
const dateRange = ref([])

const mockJobData = [
  {
    "id": 1,
    "com_id": 100,
    "comName": "百度科技有限公司",
    "title": "Java开发工程师",
    "description": "负责后端系统开发，参与需求分析和系统设计，编写高质量的代码。",
    "external_link": "www.baidu.com",
    "status": "active",
    "created_at": "2026-03-17 10:00:00",
    "updated_at": "2026-03-17 14:30:00"
  },
  {
    "id": 2,
    "com_id": 100,
    "comName": "百度科技有限公司",
    "title": "Python开发工程师",
    "description": "负责Python后端开发，参与需求分析和系统设计，编写高质量的代码。",
    "external_link": "www.baidu.com",
    "status": "active",
    "created_at": "2026-03-16 10:00:00",
    "updated_at": "2026-03-16 14:30:00"
  },
  {
    "id": 3,
    "com_id": 100,
    "comName": "百度科技有限公司",
    "title": "前端开发工程师",
    "description": "负责前端页面开发，参与需求分析和系统设计，编写高质量的代码。",
    "external_link": "www.baidu.com",
    "status": "active",
    "created_at": "2026-03-15 10:00:00",
    "updated_at": "2026-03-15 14:30:00"
  },
  {
    "id": 4,
    "com_id": 100,
    "comName": "百度科技有限公司",
    "title": "算法工程师",
    "description": "负责算法模型开发和优化，参与需求分析和系统设计，编写高质量的代码。",
    "external_link": "www.baidu.com",
    "status": "active",
    "created_at": "2026-03-14 10:00:00",
    "updated_at": "2026-03-14 14:30:00"
  }
]

const fetchJobs = async () => {
  try {
    const params = {
      pageNum: currentPage.value,
      pageSize: pageSize.value,
      keyword: keyword.value,
      status: status.value
    }

    const response = await getJobPostings(params)
    if (response.data.code === 200) {
      jobList.value = response.data.data.list || response.data.data
      total.value = response.data.data.total || response.data.data.length
    }
  } catch (error) {
    console.error('获取招聘信息失败，使用静态数据:', error)
    jobList.value = mockJobData
    total.value = mockJobData.length
  }
}

const filteredJobs = computed(() => {
  let result = jobList.value

  if (selectedCategories.value.length > 0) {
    result = result.filter(job => selectedCategories.value.includes(job.title))
  }

  return result
})

const clearFilters = () => {
  selectedCategories.value = []
  dateRange.value = []
  keyword.value = ''
  status.value = ''
}

const handleSearch = () => {
  currentPage.value = 1
  fetchJobs()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  fetchJobs()
}

onMounted(() => {
  fetchJobs()
})


</script>

<template>
  <div class="job-container">
    <aside class="filter-sidebar">
      <el-card class="filter-card">
        <el-collapse expand-icon-position="left">
          <el-collapse-item title="搜索条件" name="1">
            <el-input v-model="keyword" placeholder="输入关键词搜索职位" style="margin-bottom: 12px;">
              <template #append>
                <el-button @click="handleSearch">
                  <el-icon>
                    <Search />
                  </el-icon>
                </el-button>
              </template>
            </el-input>
            <el-select v-model="status" placeholder="选择招聘状态" style="width: 100%;" @change="handleSearch">
              <el-option label="全部" value="" />
              <el-option label="招聘中" value="active" />
              <el-option label="已结束" value="inactive" />
            </el-select>
          </el-collapse-item>
          <el-collapse-item title="职位类别" name="2">
            <el-checkbox-group v-model="selectedCategories" size="small">
              <el-checkbox v-for="category in jobCategories" :key="category" :value="category"
                style="display: block; margin: 12px 0;">
                {{ category }}
              </el-checkbox>
            </el-checkbox-group>
          </el-collapse-item>
        </el-collapse>
        <template #header>
          <div class="filter-header">
            <h4 style="margin:0px;">当前筛选</h4>
            <div>
              <el-button @click="clearFilters">
                <el-icon>
                  <Delete />
                </el-icon>
              </el-button>
              <el-button @click="handleSearch">
                <el-icon>
                  <Search />
                </el-icon>
              </el-button>
            </div>
          </div>
          <div class="tag-box">
            <el-tag v-if="keyword" size="small" type="info" effect="plain" class="tag" closable @close="keyword = ''">
              关键词: {{ keyword }}
            </el-tag>
            <el-tag v-if="status" size="small" type="info" effect="plain" class="tag" closable @close="status = ''">
              状态: {{ status === 'active' ? '招聘中' : '已结束' }}
            </el-tag>
            <el-tag v-for="category in selectedCategories" :key="category" size="small" type="info" effect="plain"
              class="tag" closable @close="selectedCategories = selectedCategories.filter(c => c !== category)">
              {{ category }}
            </el-tag>
          </div>
        </template>
      </el-card>
    </aside>
    <main class="job-content">
      <div v-if="jobList.length > 0" class="job-list">
        <jobCard v-for="job in filteredJobs" :key="job.id" :item="job" />
      </div>
      <el-empty v-else :image-size="250" description="暂无符合条件的职位" />
    </main>

    <div class="demo-pb" v-if="total > 0">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" size="large" :background="true"
        layout="prev, pager, next, jumper" :total="total" @current-change="handleCurrentChange" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.job-container {
  display: flex;
  padding: 16px;
  min-height: calc(100vh - 120px);
  background: var(--el-bg-color);
}

.filter-sidebar {
  width: 240px;
  flex-shrink: 0;
  position: sticky;
  top: 16px;
  margin-top: 8px;
  height: fit-content;

  .filter-card {
    border-radius: 8px;
    box-shadow: 0 1px 8px rgba(0, 0, 0, 0.06);
  }
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;

  h4 {
    margin: 0;
    font-size: 16px;
    color: var(--el-text-color-primary);
    font-weight: 600;
  }

  div {
    display: flex;
    gap: 8px;

    .el-button {
      padding: 6px 12px;
    }
  }
}

.tag-box {
  display: flex;
  justify-content: flex-start;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;

  .tag {
    border-radius: 6px;
  }
}

.job-content {
  flex: 1;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);

  .page-title {
    margin: 0;
    font-size: 20px;
    color: #303133;
    font-weight: 600;
  }
}

.job-list {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 5px;
  align-items: start;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.demo-pb {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-around;
  position: fixed;
  bottom: 20px;
  z-index: 1000;
  padding: 10px 0;
}
</style>
