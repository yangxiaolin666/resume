<script setup>
import { ref, reactive } from 'vue'
import { User, Avatar, Location, Clock, Phone, Message, Notebook, Briefcase, Medal } from '@element-plus/icons-vue'
import { ElInput, ElInputNumber } from 'element-plus'

const personalInfo = reactive({
  age: '33岁',
  gender: '男',
  location: '上海',
  experience: '6年经验',
  phone: '15888888888',
  email: 'qminjianli@qq.com'
})

const skills = reactive({
  language: '大学英语6级证书，荣获全国大学生英语竞赛一等奖，能够熟练的进行交流、读写。',
  technical: '计算机二级证书，熟练操作Windows平台的各类应用软件，如Word、Excel。',
  team: '具有丰富的团队组建与才经验和项目管理与协调经验。'
})

const progress = reactive({
  computer: 90,
  english: 85
})

const hobbies = ref('')

const editingField = ref('')

const startEdit = (field) => {
  editingField.value = field
}

const saveEdit = (field, value) => {
  if (field.includes('.')) {
    const [section, subField] = field.split('.')
    if (section === 'personalInfo') {
      personalInfo[subField] = value
    } else if (section === 'skills') {
      skills[subField] = value
    } else if (section === 'progress') {
      progress[subField] = parseInt(value) || 0
    }
  } else if (field === 'hobbies') {
    hobbies.value = value
  }

  editingField.value = ''
}
</script>
<template>
  <div class="resume-container">
    <div class="resume-wrapper">
      <el-splitter>
        <el-splitter-panel :resizable="false" size="30%" style="background-color: #254665;">
          <div class="left-panel">
            <div class="avatar-container">
              <img src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" alt="头像" class="avatar">
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
                  <span v-if="editingField !== `progress.${item.field}`" class="progress-value">{{ progress[item.field]
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
              <h1>全民简历</h1>
              <div class="job-info">
                <div class="info-row">
                  <span class="label">求职意向：</span>
                  <span class="value">行政专员</span>
                  <span class="label">意向城市：</span>
                  <span class="value">上海</span>
                </div>
                <div class="info-row">
                  <span class="label">期望薪资：</span>
                  <span class="value">11000/月</span>
                  <span class="label">入职时间：</span>
                  <span class="value">一个月内到岗</span>
                </div>
              </div>
            </div>
            <div class="section">
              <h2><el-icon>
                  <Notebook />
                </el-icon> 教育背景</h2>
              <div class="section-content">
                <div class="timeline-item">
                  <div class="timeline-date">2012-09 ~ 2016-07</div>
                  <div class="timeline-content">
                    <h3>全民简历师范大学</h3>
                    <span class="major">工商管理（本科）</span>
                    <p class="description">专业成绩：GPA 3.66/4（专业前5%）</p>
                    <p class="description">
                      主修课程：基础会计学、货币银行学、统计学、经济学概论、财务会计学、管理学原理、组织行为学、市场营销学、国际贸易理论、国际贸易实务、人力资源开发与管理、财务管理学、企业经营战略概论、质量管理学、西方经济学等等。
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <div class="section">
              <h2><el-icon>
                  <Briefcase />
                </el-icon> 工作经验</h2>
              <div class="section-content">
                <div class="timeline-item">
                  <div class="timeline-date">2018-09 ~ 至今</div>
                  <div class="timeline-content">
                    <h3>全民简历科技有限公司</h3>
                    <span class="position">行政专员</span>
                    <ul class="responsibilities">
                      <li>
                        负责本部门的行政人事管理和日常事务，协助总监做好各部门之间的综合协调，落实公司规章制度，沟通内外联系，保证上情下达和下情上报，负责对会议文件决定的事项进行催办、查办和落实，负责公司组织系统研讨和修订。
                      </li>
                      <li>编制公司人事管理制度，规避各项人事风险。</li>
                    </ul>
                  </div>
                </div>
                <div class="timeline-item">
                  <div class="timeline-date">2016-09 ~ 2018-08</div>
                  <div class="timeline-content">
                    <h3>上海柠岸网络科技有限公司</h3>
                    <span class="position">行政专员</span>
                    <ul class="responsibilities">
                      <li>负责中心简单财务管理，资产管控；</li>
                      <li>负责公司总部的来访客户接待工作，负责引导和介绍公司的分布情况；</li>
                      <li>负责中心的行政事务，人才班车管理、人才建立员工归宿感及前台管理；</li>
                      <li>负责招聘工作，确保人才团队管理和人才储备及培养。</li>
                      <li>督导公司各项行政、人事制度、员工福利、生日以及公司各种宴会活动的执行。</li>
                      <li>负责招聘工作，制定公司的人力资源发展计划，确保人才梯队发展和人才储备及培养。</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
            <div class="section">
              <h2><el-icon>
                  <Medal />
                </el-icon> 荣誉证书</h2>
              <div class="section-content">
                <ul class="certificates">
                  <li>英语四级，听说读写能力良好，能流利的用英语进行日常交流，能快速浏览英文文档和书籍；</li>
                  <li>通过全国计算机二级考试，熟练运用office等常用的办公软件。</li>
                </ul>
              </div>
            </div>
          </div>
        </el-splitter-panel>
      </el-splitter>
    </div>
  </div>
</template>

<style scoped>
.resume-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 20px;
  background-color: #f5f5f5;
}

.resume-wrapper {
  width: 1000px;
  height: 1400px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  overflow: hidden;
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
}

.responsibilities,
.certificates {
  font-size: 14px;
  line-height: 1.5;
  color: #666;
  padding-left: 20px;
}

.responsibilities li,
.certificates li {
  margin-bottom: 8px;
}
</style>