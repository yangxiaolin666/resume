<script setup>
import { ref, reactive } from 'vue'
import { User, Avatar, Location, Clock, Phone, Message, Notebook, Briefcase, Document, Download, Plus, Delete, Medal, Upload, Check } from '@element-plus/icons-vue'
import { ElInput, ElInputNumber, ElUpload, ElMessage, ElTooltip, ElButton, ElDatePicker } from 'element-plus'
import html2pdf from 'html2pdf.js'

const personalInfo = reactive({
  name: '小丽',
  age: '24岁',
  gender: '男',
  location: '北京',
  experience: '2年经验',
  phone: '13888888888',
  email: 'zhangming.dev@qq.com'
})

const skills = reactive({
  language: '英语CET-6，具备良好的英文文档阅读能力，能够阅读英文技术文档和进行基本的技术交流。',
  technical: '熟练掌握HTML5、CSS3、JavaScript、TypeScript等前端核心技术，熟悉Vue.js、React等主流框架，了解Webpack、Vite等构建工具。',
  team: '具备良好的团队协作能力，熟悉敏捷开发流程，能够与UI设计师、后端开发人员高效协作。'
})

const progress = reactive({
  computer: 90,
  english: 85
})

const hobbies = ref('技术博客写作、开源项目贡献、前端技术研究、参加技术分享会、阅读技术书籍、关注前端技术发展趋势')

const jobInfo = reactive({
  position: '前端开发程序员',
  city: '上海',
  salary: '26000/月',
  onboardTime: '一个月内到岗'
})

const educationList = reactive([
  {
    id: 1,
    startDate: '2024-09',
    endDate: '2028-07',
    university: '大连民族大学',
    major: '计算机科学与技术（本科）',
    details: '专业成绩：GPA 3.66/4（专业前5%）\n主修课程：数据结构与算法、计算机网络、操作系统、Web前端开发技术、JavaScript程序设计、Vue.js框架应用、React开发实践、移动端开发技术、软件工程、数据库原理与应用等。'
  }
])

const workList = reactive([
  {
    id: 1,
    startDate: '2026-01',
    endDate: '至今',
    company: '字节跳动科技有限公司',
    position: '前端开发工程师',
    responsibilities: '负责公司核心产品的前端架构设计和开发工作\n使用Vue3+TypeScript技术栈开发高性能Web应用\n参与组件库建设，提升团队开发效率和代码质量\n优化前端性能，提升用户体验和页面加载速度\n与产品、设计、后端团队紧密协作，确保项目顺利交付'
  },
])

const projectList = reactive([
  {
    id: 1,
    startDate: '2025-03',
    endDate: '2025-08',
    projectName: '电商平台前端重构项目',
    role: '前端开发工程师',
    description: '负责电商平台前端架构重构，采用Vue3+TypeScript技术栈\n实现组件化开发，提升代码复用性和可维护性\n优化页面加载性能，减少首屏加载时间30%\n集成微前端架构，支持多团队协作开发\n实现响应式设计，适配PC端和移动端'
  },
])

const awardList = reactive([
  {
    id: 1,
    awardName: '第十五届蓝桥杯国家一等奖'
  },
  {
    id: 2,
    awardName: '2025ICPC亚洲区域赛金奖'
  },
  {
    id: 3,
    awardName: 'GitHub开源项目贡献者，参与多个前端开源项目维护'
  }
])

const avatarUrl = ref('https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png')
const editingField = ref('')
const originalValue = ref('')
const editingEducation = ref(null)
const editingEducationField = ref('')

const startEdit = (field) => {
  editingField.value = field

  // 保存当前字段的原值
  if (field.includes('.')) {
    const [section, subField] = field.split('.')
    if (section === 'personalInfo') {
      originalValue.value = personalInfo[subField]
    } else if (section === 'skills') {
      originalValue.value = skills[subField]
    } else if (section === 'progress') {
      originalValue.value = progress[subField]
    } else if (section === 'jobInfo') {
      originalValue.value = jobInfo[subField]
    }
  } else if (field === 'hobbies') {
    originalValue.value = hobbies.value
  }
}

const handleAvatarChange = (file) => {
  if (file.raw) {
    avatarUrl.value = URL.createObjectURL(file.raw)
  }
}

const beforeAvatarUpload = (file) => {
  const isJPG = file.type === 'image/jpeg'
  const isPNG = file.type === 'image/png'
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isJPG && !isPNG) {
    ElMessage.error('头像图片只能是 JPG 或 PNG 格式!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('头像图片大小不能超过 2MB!')
    return false
  }
  return true
}

const saveEdit = (field, value) => {
  // 如果输入为空，恢复为原值
  if (!value || value.trim() === '') {
    if (field.includes('.')) {
      const [section, subField] = field.split('.')
      if (section === 'personalInfo') {
        personalInfo[subField] = originalValue.value
      } else if (section === 'skills') {
        skills[subField] = originalValue.value
      } else if (section === 'progress') {
        progress[subField] = originalValue.value
      } else if (section === 'jobInfo') {
        jobInfo[subField] = originalValue.value
      }
    } else if (field === 'hobbies') {
      hobbies.value = originalValue.value
    }
    editingField.value = ''
    return
  }

  // 特殊验证：进度值必须是数字且在0-100之间
  if (field.includes('progress.')) {
    const numValue = parseInt(value)
    if (isNaN(numValue) || numValue < 0 || numValue > 100) {
      ElMessage.error('进度值必须是0-100之间的数字')
      return
    }
  }

  if (field.includes('.')) {
    const [section, subField] = field.split('.')
    if (section === 'personalInfo') {
      personalInfo[subField] = value
    } else if (section === 'skills') {
      skills[subField] = value
    } else if (section === 'progress') {
      progress[subField] = parseInt(value) || 0
    } else if (section === 'jobInfo') {
      jobInfo[subField] = value
    }
  } else if (field === 'hobbies') {
    hobbies.value = value
  }

  editingField.value = ''
}

const startEditEducation = (education, field) => {
  editingEducation.value = education
  editingEducationField.value = field
  originalValue.value = education[field]
}

const saveEditEducation = (education, field, value) => {
  if (!value || value.trim() === '') {
    education[field] = originalValue.value
  } else {
    education[field] = value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const cancelEditEducation = () => {
  if (editingEducation.value && editingEducationField.value) {
    editingEducation.value[editingEducationField.value] = originalValue.value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const addEducation = () => {
  educationList.push({
    id: Date.now(),
    startDate: '',
    endDate: '',
    university: '',
    major: '',
    details: ''
  })
}

const removeEducation = (index) => {
  if (educationList.length > 1) {
    educationList.splice(index, 1)
  } else {
    ElMessage.warning('至少需要保留一条教育背景信息')
  }
}

const startEditWork = (work, field) => {
  editingEducation.value = work
  editingEducationField.value = field
  originalValue.value = work[field]
}

const saveEditWork = (work, field, value) => {
  if (!value || value.trim() === '') {
    work[field] = originalValue.value
  } else {
    work[field] = value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const cancelEditWork = () => {
  if (editingEducation.value && editingEducationField.value) {
    editingEducation.value[editingEducationField.value] = originalValue.value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const addWork = () => {
  workList.push({
    id: Date.now(),
    startDate: '',
    endDate: '',
    company: '',
    position: '',
    responsibilities: ''
  })
}

const removeWork = (index) => {
  if (workList.length > 1) {
    workList.splice(index, 1)
  } else {
    ElMessage.warning('至少需要保留一条工作经历信息')
  }
}

const startEditProject = (project, field) => {
  editingEducation.value = project
  editingEducationField.value = field
  originalValue.value = project[field]
}

const saveEditProject = (project, field, value) => {
  if (!value || value.trim() === '') {
    project[field] = originalValue.value
  } else {
    project[field] = value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const cancelEditProject = () => {
  if (editingEducation.value && editingEducationField.value) {
    editingEducation.value[editingEducationField.value] = originalValue.value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const addProject = () => {
  projectList.push({
    id: Date.now(),
    startDate: '',
    endDate: '',
    projectName: '',
    role: '',
    description: ''
  })
}

const removeProject = (index) => {
  if (projectList.length > 1) {
    projectList.splice(index, 1)
  } else {
    ElMessage.warning('至少需要保留一条项目经历信息')
  }
}

const startEditAward = (award, field) => {
  editingEducation.value = award
  editingEducationField.value = field
  originalValue.value = award[field]
}

const saveEditAward = (award, field, value) => {
  if (!value || value.trim() === '') {
    award[field] = originalValue.value
  } else {
    award[field] = value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const cancelEditAward = () => {
  if (editingEducation.value && editingEducationField.value) {
    editingEducation.value[editingEducationField.value] = originalValue.value
  }
  editingEducation.value = null
  editingEducationField.value = ''
}

const addAward = () => {
  awardList.push({
    id: Date.now(),
    awardName: ''
  })
}

const removeAward = (index) => {
  if (awardList.length > 1) {
    awardList.splice(index, 1)
  } else {
    ElMessage.warning('至少需要保留一条获奖经历信息')
  }
}

const downloadPDF = async () => {
  // 只选择简历内容部分，排除控制面板
  const resumeContent = document.querySelector('.resume-wrapper')
  if (!resumeContent) {
    ElMessage.error('未找到简历内容')
    return
  }

  // 创建PDF生成选项 - 保持原样，不修改任何样式
  const options = {
    margin: [0, 0, 0, 0], // 完全去除边距
    filename: `${personalInfo.name}_简历.pdf`,
    image: {
      type: 'jpeg',
      quality: 0.95
    },
    html2canvas: {
      scale: 1.8, // 提高缩放比例确保清晰度
      useCORS: true,
      logging: false,
      backgroundColor: '#ffffff',
      width: resumeContent.offsetWidth,
      height: resumeContent.scrollHeight,
      windowWidth: resumeContent.offsetWidth,
      scrollX: 0,
      scrollY: 0,
      ignoreElements: (element) => {
        // 忽略控制面板
        return element.classList && element.classList.contains('control-panel')
      }
    },
    jsPDF: {
      unit: 'px',
      format: [resumeContent.offsetWidth, resumeContent.scrollHeight],
      orientation: 'portrait'
    }
  }

  try {
    await html2pdf()
      .set(options)
      .from(resumeContent)
      .save()

    ElMessage.success('简历下载成功')
  } catch (error) {
    console.error('PDF生成失败:', error)
    ElMessage.error('简历下载失败，请重试')
  }
}

</script>
<template>
  <el-splitter>
    <el-splitter-panel size="10%" :resizable="false">
      <div class="control-panel">
        <div class="circle-buttons">
          <el-tooltip content="简历预览" placement="top">
            <el-button class="circle-button" type="info" :icon="Document" circle />
          </el-tooltip>
          <el-tooltip content="简历下载" placement="top">
            <el-button class="circle-button" type="success" :icon="Download" circle @click="downloadPDF" />
          </el-tooltip>
          <el-tooltip content="简历保存" placement="top">
            <el-button class="circle-button" type="warning" :icon="Check" circle />
          </el-tooltip>
          <el-tooltip content="简历上传" placement="top">
            <el-button class="circle-button" type="primary" :icon="Upload" circle />
          </el-tooltip>
        </div>
      </div>
    </el-splitter-panel>
    <el-splitter-panel :min="200">
      <div class="resume-container">
        <div class="resume-wrapper" id="resume-content">
          <el-splitter>
            <el-splitter-panel :resizable="false" size="30%" style="background-color: #254665;">
              <div class="left-panel">
                <div class="avatar-container">
                  <el-tooltip content="点击更换头像" placement="top" effect="dark">
                    <el-upload class="avatar-uploader" :show-file-list="false" :before-upload="beforeAvatarUpload"
                      :on-change="handleAvatarChange" :auto-upload="false">
                      <img v-if="avatarUrl" :src="avatarUrl" alt="头像" class="avatar">
                      <div v-else class="avatar-uploader-icon">
                        <el-icon>
                          <Avatar />
                        </el-icon>
                        <span>上传头像</span>
                      </div>
                    </el-upload>
                  </el-tooltip>
                </div>
                <div class="personal-info">
                  <div v-for="item in [
                    { field: 'age', icon: User, label: '年龄' },
                    { field: 'gender', icon: Avatar, label: '性别' },
                    { field: 'location', icon: Location, label: '地点' },
                    { field: 'experience', icon: Clock, label: '工作经验' },
                    { field: 'phone', icon: Phone, label: '电话' },
                    { field: 'email', icon: Message, label: '邮箱' }
                  ]" :key="item.field" class="info-item" @dblclick="startEdit(`personalInfo.${item.field}`)">
                    <el-icon>
                      <component :is="item.icon" />
                    </el-icon>
                    <span v-if="editingField !== `personalInfo.${item.field}`">{{ personalInfo[item.field] }}</span>
                    <el-input v-else v-model="personalInfo[item.field]"
                      @blur="saveEdit(`personalInfo.${item.field}`, personalInfo[item.field])"
                      @keyup.enter="saveEdit(`personalInfo.${item.field}`, personalInfo[item.field])" size="small"
                      :placeholder="`请输入${item.label}`" autofocus />
                  </div>
                </div>
                <div class="skills-section">
                  <div v-for="item in [
                    { field: 'language', label: '语言能力' },
                    { field: 'technical', label: '技能能力' },
                    { field: 'team', label: '团队能力' }
                  ]" :key="item.field">
                    <h3>{{ item.label }}</h3>
                    <p @dblclick="startEdit(`skills.${item.field}`)">
                      <span v-if="editingField !== `skills.${item.field}`">{{ skills[item.field] }}</span>
                      <el-input v-else v-model="skills[item.field]" type="textarea" :rows="3"
                        @blur="saveEdit(`skills.${item.field}`, skills[item.field])"
                        @keyup.enter="saveEdit(`skills.${item.field}`, skills[item.field])" size="small"
                        :placeholder="`请输入${item.label}`" autofocus />
                    </p>
                  </div>
                </div>
                <div class="skills-progress">
                  <div v-for="item in [
                    { field: 'computer', label: '计算机' },
                    { field: 'english', label: '英语' }
                  ]" :key="item.field">
                    <h3>{{ item.label }}</h3>
                    <div class="progress-container" @dblclick="startEdit(`progress.${item.field}`)">
                      <div class="progress-bar">
                        <div class="progress" :style="{ width: progress[item.field] + '%' }"></div>
                      </div>
                      <span v-if="editingField !== `progress.${item.field}`" class="progress-value">{{
                        progress[item.field]
                      }}%</span>
                      <el-input-number v-else v-model="progress[item.field]"
                        @blur="saveEdit(`progress.${item.field}`, progress[item.field])"
                        @keyup.enter="saveEdit(`progress.${item.field}`, progress[item.field])" :min="0" :max="100"
                        size="small" controls-position="right" autofocus />
                    </div>
                  </div>
                </div>
                <div class="hobbies">
                  <h3>兴趣爱好</h3>
                  <div class="hobbies-content" @dblclick="startEdit('hobbies')">
                    <p v-if="editingField !== 'hobbies'">{{ hobbies || '双击添加兴趣爱好' }}</p>
                    <el-input v-else v-model="hobbies" type="textarea" :rows="3" @blur="saveEdit('hobbies', hobbies)"
                      @keyup.enter="saveEdit('hobbies', hobbies)" size="small" placeholder="请输入兴趣爱好" autofocus />
                  </div>
                </div>
              </div>
            </el-splitter-panel>
            <el-splitter-panel :min="200" style="background-color: white;">
              <div class="right-panel">
                <div class="header">
                  <h1 @dblclick="startEdit('personalInfo.name')" class="name-header">
                    <span v-if="editingField !== 'personalInfo.name'">{{ personalInfo.name }}</span>
                    <el-input v-else v-model="personalInfo.name"
                      @blur="saveEdit('personalInfo.name', personalInfo.name)"
                      @keyup.enter="saveEdit('personalInfo.name', personalInfo.name)" size="small" placeholder="请输入姓名"
                      autofocus />
                  </h1>
                  <div class="job-info">
                    <div class="info-row">
                      <span class="label">求职意向：</span>
                      <span class="value" @dblclick="startEdit('jobInfo.position')">
                        <span v-if="editingField !== 'jobInfo.position'">{{ jobInfo.position }}</span>
                        <el-input v-else v-model="jobInfo.position"
                          @blur="saveEdit('jobInfo.position', jobInfo.position)"
                          @keyup.enter="saveEdit('jobInfo.position', jobInfo.position)" size="small"
                          placeholder="请输入求职意向" autofocus />
                      </span>
                      <span class="label">意向城市：</span>
                      <span class="value" @dblclick="startEdit('jobInfo.city')">
                        <span v-if="editingField !== 'jobInfo.city'">{{ jobInfo.city }}</span>
                        <el-input v-else v-model="jobInfo.city" @blur="saveEdit('jobInfo.city', jobInfo.city)"
                          @keyup.enter="saveEdit('jobInfo.city', jobInfo.city)" size="small" placeholder="请输入意向城市"
                          autofocus />
                      </span>
                    </div>
                    <div class="info-row">
                      <span class="label">期望薪资：</span>
                      <span class="value" @dblclick="startEdit('jobInfo.salary')">
                        <span v-if="editingField !== 'jobInfo.salary'">{{ jobInfo.salary }}</span>
                        <el-input v-else v-model="jobInfo.salary" @blur="saveEdit('jobInfo.salary', jobInfo.salary)"
                          @keyup.enter="saveEdit('jobInfo.salary', jobInfo.salary)" size="small" placeholder="请输入期望薪资"
                          autofocus />
                      </span>
                      <span class="label">入职时间：</span>
                      <span class="value" @dblclick="startEdit('jobInfo.onboardTime')">
                        <span v-if="editingField !== 'jobInfo.onboardTime'">{{ jobInfo.onboardTime }}</span>
                        <el-input v-else v-model="jobInfo.onboardTime"
                          @blur="saveEdit('jobInfo.onboardTime', jobInfo.onboardTime)"
                          @keyup.enter="saveEdit('jobInfo.onboardTime', jobInfo.onboardTime)" size="small"
                          placeholder="请输入入职时间" autofocus />
                      </span>
                    </div>
                  </div>
                </div>
                <div class="section">
                  <div class="section-header">
                    <h2><el-icon>
                        <Notebook />
                      </el-icon> 教育背景</h2>
                    <el-button type="primary" size="small" @click="addEducation" class="add-btn">
                      <el-icon>
                        <Plus />
                      </el-icon>
                    </el-button>
                  </div>
                  <div class="section-content">
                    <div v-for="(education, index) in educationList" :key="education.id"
                      class="timeline-item education-item">
                      <div class="education-actions">
                        <el-button v-if="educationList.length > 1" type="danger" size="small"
                          @click="removeEducation(index)" class="remove-btn">
                          <el-icon>
                            <Delete />
                          </el-icon>
                        </el-button>
                      </div>
                      <div class="timeline-date" @dblclick="startEditEducation(education, 'date')">
                        <span v-if="!(editingEducation === education && editingEducationField === 'date')">
                          {{ education.startDate }} ~ {{ education.endDate }}
                        </span>
                        <div v-else class="date-editor">
                          <el-date-picker v-model="education.startDate" type="month" placeholder="开始时间" format="YYYY-MM"
                            value-format="YYYY-MM"
                            @blur="saveEditEducation(education, 'startDate', education.startDate)"
                            @keyup.enter="saveEditEducation(education, 'startDate', education.startDate)"
                            size="small" />
                          <span class="date-separator">~</span>
                          <el-date-picker v-model="education.endDate" type="month" placeholder="结束时间" format="YYYY-MM"
                            value-format="YYYY-MM" @blur="saveEditEducation(education, 'endDate', education.endDate)"
                            @keyup.enter="saveEditEducation(education, 'endDate', education.endDate)" size="small" />
                          <div class="editor-actions">
                            <el-button type="primary" size="small"
                              @click="saveEditEducation(education, 'startDate', education.startDate)">保存</el-button>
                            <el-button size="small" @click="cancelEditEducation">取消</el-button>
                          </div>
                        </div>
                      </div>
                      <div class="timeline-content">
                        <h3 @dblclick="startEditEducation(education, 'university')">
                          <span v-if="!(editingEducation === education && editingEducationField === 'university')">
                            {{ education.university || '双击输入大学名称' }}
                          </span>
                          <el-input v-else v-model="education.university"
                            @blur="saveEditEducation(education, 'university', education.university)"
                            @keyup.enter="saveEditEducation(education, 'university', education.university)"
                            placeholder="请输入大学名称" size="small" autofocus />
                        </h3>
                        <span class="major" @dblclick="startEditEducation(education, 'major')">
                          <span v-if="!(editingEducation === education && editingEducationField === 'major')">
                            {{ education.major || '双击输入专业' }}
                          </span>
                          <el-input v-else v-model="education.major"
                            @blur="saveEditEducation(education, 'major', education.major)"
                            @keyup.enter="saveEditEducation(education, 'major', education.major)" placeholder="请输入专业"
                            size="small" autofocus />
                        </span>
                        <div class="description" @dblclick="startEditEducation(education, 'details')">
                          <div v-if="!(editingEducation === education && editingEducationField === 'details')"
                            style="white-space: pre-wrap;">{{ education.details || '双击输入详细信息' }}</div>
                          <div v-else class="textarea-editor">
                            <el-input v-model="education.details" type="textarea" :rows="6"
                              placeholder="请输入详细信息（如专业成绩、主修课程等）"
                              @blur="saveEditEducation(education, 'details', education.details)"
                              @keyup.enter="saveEditEducation(education, 'details', education.details)" autofocus />
                            <div class="editor-actions">
                              <el-button type="primary" size="small"
                                @click="saveEditEducation(education, 'details', education.details)">保存</el-button>
                              <el-button size="small" @click="cancelEditEducation">取消</el-button>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="section">
                  <div class="section-header">
                    <h2><el-icon>
                        <Briefcase />
                      </el-icon> 工作经验</h2>
                    <el-button type="primary" size="small" @click="addWork" class="add-btn">
                      <el-icon>
                        <Plus />
                      </el-icon>

                    </el-button>
                  </div>
                  <div class="section-content">
                    <div v-for="(work, index) in workList" :key="work.id" class="timeline-item work-item">
                      <div class="work-actions">
                        <el-button v-if="workList.length > 1" type="danger" size="small" @click="removeWork(index)"
                          class="remove-btn">
                          <el-icon>
                            <Delete />
                          </el-icon>
                        </el-button>
                      </div>
                      <div class="timeline-date" @dblclick="startEditWork(work, 'date')">
                        <span v-if="!(editingEducation === work && editingEducationField === 'date')">
                          {{ work.startDate }} ~ {{ work.endDate }}
                        </span>
                        <div v-else class="date-editor">
                          <el-date-picker v-model="work.startDate" type="month" placeholder="开始时间" format="YYYY-MM"
                            value-format="YYYY-MM" @blur="saveEditWork(work, 'startDate', work.startDate)"
                            @keyup.enter="saveEditWork(work, 'startDate', work.startDate)" size="small" />
                          <span class="date-separator">~</span>
                          <el-input v-model="work.endDate" placeholder="结束时间（如：至今）"
                            @blur="saveEditWork(work, 'endDate', work.endDate)"
                            @keyup.enter="saveEditWork(work, 'endDate', work.endDate)" size="small" />
                          <div class="editor-actions">
                            <el-button type="primary" size="small"
                              @click="saveEditWork(work, 'startDate', work.startDate)">保存</el-button>
                            <el-button size="small" @click="cancelEditWork">取消</el-button>
                          </div>
                        </div>
                      </div>
                      <div class="timeline-content">
                        <h3 @dblclick="startEditWork(work, 'company')">
                          <span v-if="!(editingEducation === work && editingEducationField === 'company')">
                            {{ work.company || '双击输入公司名称' }}
                          </span>
                          <el-input v-else v-model="work.company" @blur="saveEditWork(work, 'company', work.company)"
                            @keyup.enter="saveEditWork(work, 'company', work.company)" placeholder="请输入公司名称"
                            size="small" autofocus />
                        </h3>
                        <span class="position" @dblclick="startEditWork(work, 'position')">
                          <span v-if="!(editingEducation === work && editingEducationField === 'position')">
                            {{ work.position || '双击输入职位' }}
                          </span>
                          <el-input v-else v-model="work.position" @blur="saveEditWork(work, 'position', work.position)"
                            @keyup.enter="saveEditWork(work, 'position', work.position)" placeholder="请输入职位"
                            size="small" autofocus />
                        </span>
                        <div class="responsibilities" @dblclick="startEditWork(work, 'responsibilities')">
                          <div v-if="!(editingEducation === work && editingEducationField === 'responsibilities')"
                            style="white-space: pre-wrap;">{{ work.responsibilities || '双击输入工作职责' }}</div>
                          <div v-else class="textarea-editor">
                            <el-input v-model="work.responsibilities" type="textarea" :rows="6"
                              placeholder="请输入工作职责（每项职责用换行分隔）"
                              @blur="saveEditWork(work, 'responsibilities', work.responsibilities)"
                              @keyup.enter="saveEditWork(work, 'responsibilities', work.responsibilities)" autofocus />
                            <div class="editor-actions">
                              <el-button type="primary" size="small"
                                @click="saveEditWork(work, 'responsibilities', work.responsibilities)">保存</el-button>
                              <el-button size="small" @click="cancelEditWork">取消</el-button>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="section">
                  <div class="section-header">
                    <h2><el-icon>
                        <Briefcase />
                      </el-icon> 项目经历</h2>
                    <el-button type="primary" size="small" @click="addProject" class="add-btn">
                      <el-icon>
                        <Plus />
                      </el-icon>
                    </el-button>
                  </div>
                  <div class="section-content">
                    <div v-for="(project, index) in projectList" :key="project.id" class="timeline-item project-item">
                      <div class="project-actions">
                        <el-button v-if="projectList.length > 1" type="danger" size="small"
                          @click="removeProject(index)" class="remove-btn">
                          <el-icon>
                            <Delete />
                          </el-icon>
                        </el-button>
                      </div>
                      <div class="timeline-date" @dblclick="startEditProject(project, 'date')">
                        <span v-if="!(editingEducation === project && editingEducationField === 'date')">
                          {{ project.startDate }} ~ {{ project.endDate }}
                        </span>
                        <div v-else class="date-editor">
                          <el-date-picker v-model="project.startDate" type="month" placeholder="开始时间" format="YYYY-MM"
                            value-format="YYYY-MM" @blur="saveEditProject(project, 'startDate', project.startDate)"
                            @keyup.enter="saveEditProject(project, 'startDate', project.startDate)" size="small" />
                          <span class="date-separator">~</span>
                          <el-date-picker v-model="project.endDate" type="month" placeholder="结束时间" format="YYYY-MM"
                            value-format="YYYY-MM" @blur="saveEditProject(project, 'endDate', project.endDate)"
                            @keyup.enter="saveEditProject(project, 'endDate', project.endDate)" size="small" />
                          <div class="editor-actions">
                            <el-button type="primary" size="small"
                              @click="saveEditProject(project, 'startDate', project.startDate)">保存</el-button>
                            <el-button size="small" @click="cancelEditProject">取消</el-button>
                          </div>
                        </div>
                      </div>
                      <div class="timeline-content">
                        <h3 @dblclick="startEditProject(project, 'projectName')">
                          <span v-if="!(editingEducation === project && editingEducationField === 'projectName')">
                            {{ project.projectName || '双击输入项目名称' }}
                          </span>
                          <el-input v-else v-model="project.projectName"
                            @blur="saveEditProject(project, 'projectName', project.projectName)"
                            @keyup.enter="saveEditProject(project, 'projectName', project.projectName)"
                            placeholder="请输入项目名称" size="small" autofocus />
                        </h3>
                        <span class="role" @dblclick="startEditProject(project, 'role')">
                          <span v-if="!(editingEducation === project && editingEducationField === 'role')">
                            {{ project.role || '双击输入担任角色' }}
                          </span>
                          <el-input v-else v-model="project.role" @blur="saveEditProject(project, 'role', project.role)"
                            @keyup.enter="saveEditProject(project, 'role', project.role)" placeholder="请输入担任角色"
                            size="small" autofocus />
                        </span>
                        <div class="description" @dblclick="startEditProject(project, 'description')">
                          <div v-if="!(editingEducation === project && editingEducationField === 'description')"
                            style="white-space: pre-wrap;">{{ project.description || '双击输入项目描述' }}</div>
                          <div v-else class="textarea-editor">
                            <el-input v-model="project.description" type="textarea" :rows="6"
                              placeholder="请输入项目描述（每项描述用换行分隔）"
                              @blur="saveEditProject(project, 'description', project.description)"
                              @keyup.enter="saveEditProject(project, 'description', project.description)" autofocus />
                            <div class="editor-actions">
                              <el-button type="primary" size="small"
                                @click="saveEditProject(project, 'description', project.description)">保存</el-button>
                              <el-button size="small" @click="cancelEditProject">取消</el-button>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="section">
                  <div class="section-header">
                    <h2><el-icon>
                        <Medal />
                      </el-icon> 获奖经历</h2>
                    <el-button type="primary" size="small" @click="addAward" class="add-btn">
                      <el-icon>
                        <Plus />
                      </el-icon>
                    </el-button>
                  </div>
                  <div class="section-content">
                    <div v-for="(award, index) in awardList" :key="award.id" class="timeline-item award-item">
                      <div class="award-actions">
                        <el-button v-if="awardList.length > 1" type="danger" size="small" @click="removeAward(index)"
                          class="remove-btn">
                          <el-icon>
                            <Delete />
                          </el-icon>
                        </el-button>
                      </div>
                      <div class="award-content" @dblclick="startEditAward(award, 'awardName')">
                        <div v-if="!(editingEducation === award && editingEducationField === 'awardName')"
                          style="white-space: pre-wrap;">{{ award.awardName || '双击输入获奖经历' }}</div>
                        <div v-else class="textarea-editor">
                          <el-input v-model="award.awardName" type="textarea" :rows="3" placeholder="请输入获奖经历"
                            @blur="saveEditAward(award, 'awardName', award.awardName)"
                            @keyup.enter="saveEditAward(award, 'awardName', award.awardName)" autofocus />
                          <div class="editor-actions">
                            <el-button type="primary" size="small"
                              @click="saveEditAward(award, 'awardName', award.awardName)">保存</el-button>
                            <el-button size="small" @click="cancelEditAward">取消</el-button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </el-splitter-panel>
          </el-splitter>
        </div>
      </div>
    </el-splitter-panel>
  </el-splitter>
</template>

<style scoped>
.resume-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 20px;
  background-color: var(--el-bg-color);
}

.resume-wrapper {
  width: 794px;
  min-height: 1123px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
}


.left-panel {
  padding: 40px 20px;
  color: white;
  height: 100%;
  box-sizing: border-box;
}

.avatar-container {
  display: flex;
  justify-content: center;
  margin-bottom: 30px;
}

.avatar {
  width: 120px;
  height: 160px;
  border-radius: 5px;
  object-fit: cover;
}

.avatar-uploader {
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.avatar-uploader:hover .avatar {
  opacity: 0.7;
}

.avatar-uploader-icon {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 120px;
  height: 160px;
  border: 1px dashed #d9d9d9;
  border-radius: 5px;
  color: #8c939d;
  font-size: 12px;
  background-color: rgba(255, 255, 255, 0.1);
}

.avatar-uploader-icon:hover {
  border-color: #409eff;
}

.avatar-uploader-icon .el-icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.personal-info {
  margin-bottom: 40px;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.3s;
  padding: 5px;
  border-radius: 4px;
}

.info-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.info-item i {
  margin-right: 10px;
  width: 16px;
  text-align: center;
}

/* Element Plus 组件样式优化 */
:deep(.el-input__wrapper) {
  background: transparent !important;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.5) inset !important;
}

:deep(.el-input__inner) {
  color: white !important;
}

:deep(.el-textarea__inner) {
  background: transparent !important;
  color: white !important;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.5) inset !important;
}

:deep(.el-input-number) {
  width: 80px;
}

.skills-section {
  margin-bottom: 30px;
}

.skills-section h3 {
  font-size: 16px;
  margin-bottom: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.3);
  padding-bottom: 5px;
}

.skills-section p {
  font-size: 12px;
  line-height: 1.5;
  margin-bottom: 15px;
  cursor: pointer;
  transition: background-color 0.3s;
  padding: 8px;
  border-radius: 4px;
  position: relative;
  word-wrap: break-word;
  word-break: break-all;
  white-space: pre-wrap;
}

.skills-section p:hover {
  background-color: rgba(255, 255, 255, 0.1);
}



.skills-progress {
  margin-bottom: 30px;
}

.skills-progress h3 {
  font-size: 14px;
  margin-bottom: 5px;
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
  cursor: pointer;
  transition: background-color 0.3s;
  padding: 8px;
  border-radius: 4px;
}

.progress-container:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.progress-bar {
  width: 100%;
  height: 8px;
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  overflow: hidden;
  flex: 1;
}

.progress {
  height: 100%;
  background-color: white;
  border-radius: 4px;
}

.progress-value {
  font-size: 12px;
  min-width: 30px;
}



.hobbies h3 {
  font-size: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.3);
  padding-bottom: 5px;
}

.hobbies-content {
  cursor: pointer;
  transition: background-color 0.3s;
  padding: 8px;
  border-radius: 4px;
  position: relative;
}

.hobbies-content:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.hobbies-content p {
  font-size: 12px;
  line-height: 1.5;
  margin: 0;
  color: rgba(255, 255, 255, 0.8);
  word-wrap: break-word;
  word-break: break-all;
  white-space: pre-wrap;
}



.right-panel {
  padding: 40px;
  height: 100%;
  box-sizing: border-box;
  overflow-y: auto;
}

.header {
  margin-bottom: 40px;
}

.header h1 {
  font-size: 28px;
  margin-bottom: 20px;
  color: #254665;
  cursor: pointer;
  transition: background-color 0.3s;
  padding: 5px 10px;
  border-radius: 4px;
  display: inline-block;
  margin-left: 0;
  padding-left: 0;
}

.header h1:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.header h1 .el-input {
  font-size: 22px;
  font-weight: bold;
  height: 60px;
  width: 200px;
}

.header h1 .el-input .el-input__wrapper {
  height: 100%;
  padding: 10px 15px;
}

.header h1 .el-input .el-input__inner {
  height: 100%;
  line-height: 1.2;
}

.job-info {
  margin-bottom: 20px;
}

.info-row {
  margin-bottom: 10px;
  font-size: 14px;
}

.label {
  font-weight: bold;
  margin-right: 5px;
  color: black;
}

.value {
  margin-right: 30px;
  color: black;
  cursor: pointer;
  padding: 2px 5px;
  border-radius: 3px;
  display: inline-block;
  min-width: 80px;
}

.value:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

/* 求职意向表单输入框样式 */
.job-info :deep(.el-input__wrapper) {
  background: white !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
  width: auto !important;
  min-width: 80px;
}

.job-info :deep(.el-input__inner) {
  color: #606266 !important;
  width: 100% !important;
}

.job-info .value .el-input {
  width: auto !important;
  max-width: 100px;
}


.section {
  margin-bottom: 40px;
}

.section h2 {
  font-size: 18px;
  color: #254665;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 2px solid #e0e0e0;
  display: flex;
  align-items: center;
}

.section h2 i {
  margin-right: 10px;
}

.timeline-item {
  margin-bottom: 30px;
}

.timeline-date {
  font-weight: bold;
  margin-bottom: 10px;
  color: #666;
}

.timeline-content h3 {
  font-size: 16px;
  margin-bottom: 5px;
  color: #333;
}

.major,
.position {
  color: #666;
  font-size: 14px;
  margin-bottom: 10px;
  display: block;
}

.description {
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 10px;
  color: #666;
  word-wrap: break-word;
  word-break: break-all;
  white-space: pre-wrap;
}

.responsibilities,
.certificates {
  font-size: 14px;
  line-height: 1.5;
  color: #666;
  padding-left: 20px;
  word-wrap: break-word;
  word-break: break-all;
  white-space: pre-wrap;
}

.responsibilities li,
.certificates li {
  margin-bottom: 8px;
}

/* 教育背景编辑样式 */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  position: relative;
}

.section-header h2 {
  margin-bottom: 0;
  flex: 1;
  position: relative;
}

.add-btn {
  margin-left: auto;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.section-header:hover .add-btn {
  opacity: 1;
}

.education-item {
  position: relative;
  padding: 7px;
  border: 1px solid transparent;
  border-radius: 8px;
  background-color: transparent;
  transition: all 0.3s ease;
}

.education-item:hover {
  border-color: #e0e0e0;
  background-color: #fafafa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.education-actions {
  position: absolute;
  top: 10px;
  right: 10px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.education-item:hover .education-actions {
  opacity: 1;
}

.remove-btn {
  padding: 4px 8px;
  min-width: auto;
}

.timeline-date {
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
  font-weight: bold;
  color: #666;
  background-color: #fafafa;
  margin-left: 0;
  padding-left: 0;
  width: 90%;
}

.timeline-date:hover {
  background-color: #e8f4fd;
}


.date-editor {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.date-separator {
  font-weight: bold;
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
}

.timeline-content h3 {
  cursor: pointer;
  padding: 0px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
  text-align: left;
  margin-left: 0;
  padding-left: 0;
  width: 100%;
}

.timeline-content h3:hover {
  background-color: #e8f4fd;
}

.major {
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.3s;
  display: block;
  text-align: left;
  margin-left: 0;
  padding-left: 0;
  width: 100%;
}

.major:hover {
  background-color: #e8f4fd;
}

/* 专门针对教育背景的样式 */
.education-item .timeline-content h3 {
  text-align: left !important;
  margin-left: 0 !important;
  padding-left: 0 !important;
  width: 90% !important;
}

.education-item .major {
  text-align: left !important;
  margin-left: 0 !important;
  padding-left: 0 !important;
  width: 90% !important;
}

.description {
  cursor: pointer;
  padding: 5px;
  border-radius: 8px;
  transition: all 0.3s;
  background-color: transparent;
  border: 1px solid transparent;
  min-height: 60px;
}

.description:hover {
  background-color: #f8f9fa;
  border-color: #e0e0e0;
}

.textarea-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: white;
}

/* 日期选择器样式 */
:deep(.el-date-editor) {
  width: 140px;
}

:deep(.el-date-editor .el-input__wrapper) {
  background: white !important;
}

/* 教育背景输入框样式 */
.education-item :deep(.el-input__wrapper) {
  background: white !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

.education-item :deep(.el-input__inner) {
  color: #606266 !important;
}

.education-item :deep(.el-textarea__inner) {
  background: white !important;
  color: #606266 !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

/* 工作经历编辑样式 */
.work-item {
  position: relative;
  padding: 7px;
  border: 1px solid transparent;
  border-radius: 8px;
  background-color: transparent;
  transition: all 0.3s ease;
}

.work-item:hover {
  border-color: #e0e0e0;
  background-color: #fafafa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.work-actions {
  position: absolute;
  top: 10px;
  right: 10px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.work-item:hover .work-actions {
  opacity: 1;
}

.responsibilities {
  cursor: pointer;
  padding: 5px;
  border-radius: 8px;
  transition: all 0.3s;
  background-color: transparent;
  border: 1px solid transparent;
  min-height: 60px;
  font-size: 14px;
  line-height: 1.5;
  color: #666;
}

.responsibilities:hover {
  background-color: #f8f9fa;
  border-color: #e0e0e0;
}

/* 专门针对工作经历的样式 */
.work-item .timeline-content h3 {
  text-align: left !important;
  margin-left: 0 !important;
  padding-left: 0 !important;
  width: 90% !important;
  cursor: pointer;
  padding: 0px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
  margin-bottom: 10px;
}

.work-item .timeline-content h3:hover {
  background-color: #e8f4fd;
}

.work-item .position {
  text-align: left !important;
  margin-left: 0 !important;
  padding-left: 0 !important;
  width: 90% !important;
  cursor: pointer;
  padding: 0px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
  display: block;
  margin-bottom: 15px;
  color: #666;
  font-size: 14px;
}

.work-item .position:hover {
  background-color: #e8f4fd;
}

/* 工作经历输入框样式 */
.work-item :deep(.el-input__wrapper) {
  background: white !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

.work-item :deep(.el-input__inner) {
  color: #606266 !important;
}

.work-item :deep(.el-textarea__inner) {
  background: white !important;
  color: #606266 !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

/* 项目经历编辑样式 */
.project-item {
  position: relative;
  padding: 7px;
  border: 1px solid transparent;
  border-radius: 8px;
  background-color: transparent;
  transition: all 0.3s ease;
}

.project-item:hover {
  border-color: #e0e0e0;
  background-color: #fafafa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.project-actions {
  position: absolute;
  top: 10px;
  right: 10px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.project-item:hover .project-actions {
  opacity: 1;
}

/* 专门针对项目经历的样式 */
.project-item .timeline-content h3 {
  text-align: left !important;
  margin-left: 0 !important;
  padding-left: 0 !important;
  width: 90% !important;
  cursor: pointer;
  padding: 0px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
  margin-bottom: 10px;
}

.project-item .timeline-content h3:hover {
  background-color: #e8f4fd;
}

.project-item .role {
  text-align: left !important;
  margin-left: 0 !important;
  padding-left: 0 !important;
  width: 90% !important;
  cursor: pointer;
  padding: 0px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
  display: block;
  margin-bottom: 15px;
  color: #666;
  font-size: 14px;
}

.project-item .role:hover {
  background-color: #e8f4fd;
}

/* 项目经历输入框样式 */
.project-item :deep(.el-input__wrapper) {
  background: white !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

.project-item :deep(.el-input__inner) {
  color: #606266 !important;
}

.project-item :deep(.el-textarea__inner) {
  background: white !important;
  color: #606266 !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

/* 获奖经历编辑样式 */
.award-item {
  position: relative;
  border: 1px solid transparent;
  border-radius: 8px;
  background-color: transparent;
  transition: all 0.3s ease;
  margin-bottom: 0px;
}

.award-item:hover {
  border-color: #e0e0e0;
  background-color: #fafafa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.award-actions {
  position: absolute;
  top: 2px;
  right: 10px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.award-item:hover .award-actions {
  opacity: 1;
}

.award-content {
  cursor: pointer;
  padding: 5px;
  border-radius: 8px;
  transition: all 0.3s;
  background-color: transparent;
  border: 1px solid transparent;
  font-size: 14px;
  line-height: 1.5;
  color: #666;
}

.award-content:hover {
  background-color: #f8f9fa;
  border-color: #e0e0e0;
}

/* 专门针对获奖经历的样式 */
.award-item .award-content {
  text-align: left !important;
  margin-left: 0 !important;
  padding-left: 0 !important;
  width: 90% !important;
}

/* 获奖经历输入框样式 */
.award-item :deep(.el-textarea__inner) {
  background: white !important;
  color: #606266 !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

/* 圆圈按钮样式 */
.control-panel {
  display: flex;
  justify-content: center;
  height: 100%;
  background-color: var(--el-bg-color);
}

.circle-buttons {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  padding: 20px;
}

.circle-button {
  width: 60px !important;
  height: 60px !important;
  font-size: 24px !important;
  transition: all 0.3s ease;
}

/* 第一个按钮往右偏移 */
.circle-button:first-child {
  margin-left: 10px;
}

.circle-button:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
</style>