<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import jobCard from '@/components/jobCard.vue'
import {
  getMyJobPostings,
  createJobPosting,
  updateJobPosting
} from '@/api/recruitment'

const router = useRouter()
const jobList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const addDialogVisible = ref(false)
const editFormRef = ref(null)
const addFormRef = ref(null)
const currentEditJob = ref(null)

const editForm = ref({
  title: '',
  comName: '',
  description: '',
  external_link: '',
  status: 'active'
})

const addForm = ref({
  title: '',
  comName: '',
  description: '',
  external_link: '',
  status: 'active'
})

const formRules = {
  title: [
    { required: true, message: '请输入职位名称', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入职位描述', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ]
}

const mockJobData = [
  {
    "id": 1,
    "com_id": 100,
    "comName": "某某科技有限公司",
    "title": "Java开发工程师",
    "description": "负责后端系统开发，参与需求分析和系统设计，编写高质量的代码。",
    "external_link": "https://example.com/jobs/123",
    "status": "active",
    "created_at": "2026-03-17 10:00:00",
    "updated_at": "2026-03-17 14:30:00"
  },
  {
    "id": 2,
    "com_id": 100,
    "comName": "某某科技有限公司",
    "title": "前端开发工程师",
    "description": "负责前端页面开发，使用Vue框架进行开发。",
    "external_link": "https://example.com/jobs/789",
    "status": "inactive",
    "created_at": "2026-03-15 09:00:00",
    "updated_at": "2026-03-15 09:00:00"
  },
  {
    "id": 3,
    "com_id": 100,
    "comName": "某某科技有限公司",
    "title": "Python开发工程师",
    "description": "负责Python后端开发，参与需求分析和系统设计，编写高质量的代码。",
    "external_link": "https://example.com/jobs/456",
    "status": "active",
    "created_at": "2026-03-16 11:00:00",
    "updated_at": "2026-03-16 15:30:00"
  }
]

const fetchJobs = async () => {
  try {
    const response = await getMyJobPostings()
    if (response.data.code === 200) {
      jobList.value = response.data.data
      total.value = response.data.data.length
    }
  } catch (error) {
    console.error('获取招聘信息失败，使用静态数据:', error)
    ElMessage.warning('网络连接失败，使用演示数据')
    jobList.value = mockJobData
    total.value = mockJobData.length
  }
}

const handleCurrentChange = (val) => {
  console.log(`current page: ${val}`)
  currentPage.value = val
}

const openEditDialog = (job) => {
  currentEditJob.value = job
  editForm.value = {
    title: job.title,
    description: job.description,
    external_link: job.external_link ? job.external_link.replace('https://', '').replace('http://', '') : '',
    status: job.status
  }
  dialogVisible.value = true
}

const handleDeleteJob = async (job) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除职位"${job.title}"吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    try {
      const response = await deleteJobPosting(job.id)
      if (response.data.code === 200) {
        ElMessage({
          message: '职位删除成功',
          type: 'success',
          duration: 2000
        })
        await fetchJobs()
      }
    } catch (apiError) {
      console.error('删除职位失败，使用本地操作:', apiError)
      ElMessage.warning('网络连接失败，使用本地删除')

      // 本地删除数据
      jobList.value = jobList.value.filter(item => item.id !== job.id)
      total.value = jobList.value.length

      ElMessage({
        message: '职位删除成功',
        type: 'success',
        duration: 2000
      })
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除职位失败:', error)
      ElMessage.error('职位删除失败')
    }
  }
}

const handleToggleJobStatus = async (job) => {
  try {
    const newStatus = job.status === 'active' ? 'inactive' : 'active'
    const actionText = newStatus === 'active' ? '启用' : '下架'

    await ElMessageBox.confirm(
      `确定要${actionText}职位"${job.title}"吗？`,
      `确认${actionText}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    try {
      const response = await updateJobPostingStatus(job.id, newStatus)
      if (response.data.code === 200) {
        ElMessage({
          message: `职位已${actionText}`,
          type: 'success',
          duration: 2000
        })
        await fetchJobs()
      }
    } catch (apiError) {
      console.error('修改职位状态失败，使用本地操作:', apiError)
      ElMessage.warning('网络连接失败，使用本地修改')

      // 本地修改状态
      const now = new Date()
      const updated_at = now.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
      }).replace(/\//g, '-')

      const jobIndex = jobList.value.findIndex(item => item.id === job.id)
      if (jobIndex !== -1) {
        jobList.value[jobIndex] = {
          ...jobList.value[jobIndex],
          status: newStatus,
          updated_at
        }
      }

      ElMessage({
        message: `职位已${actionText}`,
        type: 'success',
        duration: 2000
      })
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('修改职位状态失败:', error)
      ElMessage.error('修改职位状态失败')
    }
  }
}

const closeDialog = () => {
  dialogVisible.value = false
  resetForm()
}

const resetForm = () => {
  if (editFormRef.value) {
    editFormRef.value.resetFields()
  }
  editForm.value = {
    title: '',
    comName: '',
    description: '',
    external_link: '',
    status: 'active'
  }
  currentEditJob.value = null
}

const submitEdit = async () => {
  if (!editFormRef.value) return

  await editFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const requestData = {
          title: editForm.value.title,
          description: editForm.value.description,
          external_link: editForm.value.external_link ? 'https://' + editForm.value.external_link : '',
          status: editForm.value.status
        }

        const response = await updateJobPosting(currentEditJob.value.id, requestData)

        if (response.data.code === 200) {
          ElMessage({
            message: '职位修改成功',
            type: 'success',
            duration: 2000
          })

          // 重新获取数据以更新界面
          await fetchJobs()
          closeDialog()
        }
      } catch (error) {
        console.error('修改职位失败，使用本地操作:', error)
        ElMessage.warning('网络连接失败，使用本地修改')

        // 本地修改数据
        const now = new Date()
        const updated_at = now.toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          hour12: false
        }).replace(/\//g, '-')

        const jobIndex = jobList.value.findIndex(job => job.id === currentEditJob.value.id)
        if (jobIndex !== -1) {
          jobList.value[jobIndex] = {
            ...jobList.value[jobIndex],
            title: editForm.value.title,
            description: editForm.value.description,
            external_link: editForm.value.external_link ? 'https://' + editForm.value.external_link : '',
            status: editForm.value.status,
            updated_at
          }
        }

        ElMessage({
          message: '职位修改成功',
          type: 'success',
          duration: 2000
        })

        closeDialog()
      }
    }
  })
}

const openAddDialog = () => {
  addDialogVisible.value = true
}

const closeAddDialog = () => {
  addDialogVisible.value = false
  resetAddForm()
}

const resetAddForm = () => {
  if (addFormRef.value) {
    addFormRef.value.resetFields()
  }
  addForm.value = {
    title: '',
    comName: '',
    description: '',
    external_link: '',
    status: 'active'
  }
}

const submitAdd = async () => {
  if (!addFormRef.value) return

  await addFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const requestData = {
          title: addForm.value.title,
          description: addForm.value.description,
          external_link: addForm.value.external_link ? 'https://' + addForm.value.external_link : '',
          status: addForm.value.status
        }

        const response = await createJobPosting(requestData)

        if (response.data.code === 201) {
          ElMessage({
            message: '职位添加成功',
            type: 'success',
            duration: 2000
          })

          // 重新获取数据以更新界面
          await fetchJobs()
          closeAddDialog()
        }
      } catch (error) {
        console.error('添加职位失败，使用本地操作:', error)
        ElMessage.warning('网络连接失败，使用本地添加')

        // 本地添加数据
        const now = new Date()
        const created_at = now.toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          hour12: false
        }).replace(/\//g, '-')

        const newJob = {
          id: Math.max(...jobList.value.map(job => job.id), 0) + 1,
          com_id: 100,
          comName: "某某科技有限公司",
          title: addForm.value.title,
          description: addForm.value.description,
          external_link: addForm.value.external_link ? 'https://' + addForm.value.external_link : '',
          status: addForm.value.status,
          created_at,
          updated_at: created_at
        }

        jobList.value.unshift(newJob)
        total.value = jobList.value.length

        ElMessage({
          message: '职位添加成功',
          type: 'success',
          duration: 2000
        })

        closeAddDialog()
      }
    }
  })
}

onMounted(() => {
  fetchJobs()
})
</script>

<template>
  <div class="my-jobs-container">
    <main class="jobs-content">
      <div class="content-header">
        <h2 class="page-title">已发布职位</h2>
        <el-button type="primary" @click="openAddDialog">
          <el-icon>
            <Plus />
          </el-icon>
          添加职位
        </el-button>
      </div>
      <div v-if="jobList.length > 0" class="job-list">
        <jobCard v-for="job in jobList" :key="job.id" :item="job" :on-edit="openEditDialog" />
      </div>
      <el-empty v-else :image-size="250" description="暂无已发布职位" />
    </main>

    <div class="demo-pb" v-if="total > 0">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" size="large" :background="true"
        layout="prev, pager, next, jumper" :total="total" @current-change="handleCurrentChange" />
    </div>

    <el-dialog v-model="dialogVisible" title="修改职位" width="500px" @close="closeDialog">
      <el-form :model="editForm" :rules="formRules" ref="editFormRef" label-width="100px">
        <el-form-item label="职位名称" prop="title">
          <el-input v-model="editForm.title" placeholder="请输入职位名称" />
        </el-form-item>
        <el-form-item label="公司名称" prop="comName">
          <el-input v-model="editForm.comName" placeholder="请输入公司名称" />
        </el-form-item>
        <el-form-item label="职位描述" prop="description">
          <el-input v-model="editForm.description" type="textarea" :rows="4" placeholder="请输入职位描述" />
        </el-form-item>
        <el-form-item label="外部链接" prop="external_link">
          <el-input v-model="editForm.external_link" placeholder="请输入外部链接">
            <template #prepend>https://</template>
          </el-input>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="editForm.status" placeholder="请选择状态">
            <el-option label="招聘中" value="active" />
            <el-option label="已结束" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeDialog">取消</el-button>
          <el-button type="primary" @click="submitEdit">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="addDialogVisible" title="添加职位" width="500px" @close="closeAddDialog">
      <el-form :model="addForm" :rules="formRules" ref="addFormRef" label-width="100px">
        <el-form-item label="职位名称" prop="title">
          <el-input v-model="addForm.title" placeholder="请输入职位名称" />
        </el-form-item>
        <el-form-item label="公司名称" prop="comName">
          <el-input v-model="addForm.comName" placeholder="请输入公司名称" />
        </el-form-item>
        <el-form-item label="职位描述" prop="description">
          <el-input v-model="addForm.description" type="textarea" :rows="4" placeholder="请输入职位描述" />
        </el-form-item>
        <el-form-item label="外部链接" prop="external_link">
          <el-input v-model="addForm.external_link" placeholder="请输入外部链接">
            <template #prepend>https://</template>
          </el-input>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="addForm.status" placeholder="请选择状态">
            <el-option label="招聘中" value="active" />
            <el-option label="已结束" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeAddDialog">取消</el-button>
          <el-button type="primary" @click="submitAdd">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.my-jobs-container {
  min-height: calc(100vh - 80px);
}

.jobs-content {
  flex: 1;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 20px;

  .page-title {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color);
  }
}

.job-list {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 0px;
  align-items: start;
}

.job-card {
  .el-card {
    min-height: 380px;
    max-height: 380px;
    border-radius: 12px;
    transition: all 0.3s;

    &:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      transform: translateY(-2px);
    }
  }
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
