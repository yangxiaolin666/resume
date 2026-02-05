<script setup>
import { ref, onMounted } from 'vue'

const students = [
  {
    id: 1,
    name: '杨潇林',
    age: 19,
    education: '本科',
    major: '计算机科学与技术',
    school: '大连民族大学',
    experience: '应届毕业生',
    skills: ['JavaScript', 'Vue', 'HTML/CSS'],
    languages: ['英语四级'],
    certificates: ['软件设计师'],
    internships: 0,
    computerSkills: ['Java'],
    jobRegions: ['辽宁省大连市'],
    jobPositions: ['软件开发'],
    bio: '热爱前端开发，有多个项目经验，熟悉现代前端框架。',
    contactInfo: { phone: '13800138001', email: 'yangxl@example.com' },
    photo: ''
  },
  {
    id: 2,
    name: '刘世鹏',
    age: 19,
    education: '研究生',
    major: '软件工程',
    school: '大连民族大学',
    experience: '1年经验',
    skills: ['React', 'TypeScript', 'Node.js'],
    languages: ['英语四级', '日语N2'],
    certificates: ['程序员'],
    internships: 1,
    computerSkills: ['Java'],
    jobRegions: ['辽宁省大连市'],
    jobPositions: ['软件开发', '软件技术服务'],
    bio: '有1年全栈开发经验，擅长React生态系统，对后端也有一定了解。',
    contactInfo: { phone: '13800138002', email: 'liusp@example.com' },
    photo: ''
  },
  {
    id: 3,
    name: '范茂阳',
    age: 19,
    education: '本科',
    major: '软件工程',
    school: '大连民族大学',
    experience: '应届毕业生',
    skills: ['Python', 'C++', '机器学习'],
    languages: ['英语四级'],
    certificates: [],
    internships: 0,
    computerSkills: [],
    jobRegions: [],
    jobPositions: [],
    bio: '对人工智能和机器学习有浓厚兴趣，参与过相关项目开发。',
    contactInfo: { phone: '13800138003', email: 'fanmy@example.com' },
    photo: ''
  },
  {
    id: 4,
    name: '人机',
    age: 19,
    education: '本科',
    major: '人工智能',
    school: '大连民族大学',
    experience: '2年经验',
    skills: ['Java', 'Spring Boot', 'MySQL'],
    languages: ['英语四级', '日语N1'],
    certificates: ['网络管理员'],
    internships: 1,
    computerSkills: ['Java'],
    jobRegions: ['辽宁省大连市'],
    jobPositions: ['软件开发', '网络安全维护'],
    bio: '后端开发工程师，有扎实的Java基础，熟悉主流框架和数据库。',
    contactInfo: { phone: '13800138004', email: 'renji@example.com' },
    photo: ''
  },
  {
    id: 5,
    name: '网络',
    age: 19,
    education: '本科',
    major: '网络工程',
    school: '大连民族大学',
    experience: '应届毕业生',
    skills: ['UI设计', 'Photoshop', 'Figma'],
    languages: ['英语四级', '韩语TOPIK I'],
    certificates: [],
    internships: 0,
    computerSkills: [],
    jobRegions: [],
    jobPositions: ['软件技术服务'],
    bio: 'UI设计师，有良好的美学素养和用户体验意识，熟练使用设计工具。',
    contactInfo: { phone: '13800138005', email: 'wangluo@example.com' },
    photo: ''
  }
]

const showCards = ref(false)

onMounted(() => {
  // 组件挂载后延迟显示动画
  setTimeout(() => {
    showCards.value = true
  }, 100)
})
</script>

<template>
  <div class="waterfall-container">
    <el-row :gutter="20" class="waterfall-row">
      <el-col 
        v-for="student in students" 
        :key="student.id" 
        :xs="24" :sm="12" :md="8" :lg="6" 
        class="waterfall-col"
      >
        <div 
          class="card-wrapper" 
          :class="{ 'show': showCards }"
          :style="{ 'animation-delay': `${student.id * 0.1}s` }"
        >
          <el-card class="student-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <h3>{{ student.name }}</h3>
                <el-tag 
                  :type="student.experience.includes('应届') ? 'info' : 'success'" 
                  size="small"
                >
                  {{ student.experience }}
                </el-tag>
              </div>
            </template>
            
            <div class="student-info">
              <div class="info-item">
                <span class="label">年龄：</span>
                <span>{{ student.age }}</span>
              </div>
              <div class="info-item">
                <span class="label">学历：</span>
                <span>{{ student.education }}</span>
              </div>
              <div class="info-item">
                <span class="label">专业：</span>
                <span>{{ student.major }}</span>
              </div>
              <div class="info-item">
                <span class="label">学校：</span>
                <span>{{ student.school }}</span>
              </div>
              
              <div class="skills-section">
                <h4>技能特长</h4>
                <div class="skills-tags">
                  <el-tag 
                    v-for="skill in student.skills" 
                    :key="skill" 
                    type="primary" 
                    size="small"
                    class="skill-tag"
                  >
                    {{ skill }}
                  </el-tag>
                </div>
              </div>
              
              <div class="languages-section">
                <h4>语言能力</h4>
                <div class="languages-tags">
                  <el-tag 
                    v-for="lang in student.languages" 
                    :key="lang" 
                    type="warning" 
                    size="small"
                  >
                    {{ lang }}
                  </el-tag>
                </div>
              </div>
            </div>
            
            <template #footer>
              <div class="card-footer">
                <el-button type="primary" size="small">查看详情</el-button>
                <el-button type="success" size="small">收藏简历</el-button>
              </div>
            </template>
          </el-card>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.waterfall-container {
  padding: 20px;
  min-height: 100vh;
}

.waterfall-row {
  margin: 0 !important;
}

.waterfall-col {
  margin-bottom: 20px;
}

.card-wrapper {
  opacity: 0;
  transform: translateY(100px);
  transition: all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

.card-wrapper.show {
  opacity: 1;
  transform: translateY(0);
}

.student-card {
  height: 100%;
  border-radius: 12px;
  border: none;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.student-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0;
}

.card-header h3 {
  margin: 0;
  color: #2c3e50;
  font-weight: 600;
}

.student-info {
  padding: 0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 4px 0;
  border-bottom: 1px solid #f0f0f0;
}

.label {
  font-weight: 500;
  color: #666;
}

.skills-section,
.languages-section,
.bio-section,
.contact-section {
  margin-top: 15px;
}

.skills-section h4,
.languages-section h4,
.bio-section h4,
.contact-section h4 {
  margin: 0 0 8px 0;
  color: #34495e;
  font-size: 14px;
  font-weight: 600;
}

.skills-tags,
.languages-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.skill-tag {
  margin: 2px;
}

.bio-text {
  font-size: 13px;
  line-height: 1.5;
  color: #666;
  margin: 0;
}

.contact-info {
  font-size: 13px;
  color: #666;
}

.contact-info div {
  margin-bottom: 4px;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  padding: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .waterfall-container {
    padding: 10px;
  }
  
  .card-wrapper {
    transform: translateY(50px);
  }
}

/* 瀑布流动画延迟 */
.card-wrapper:nth-child(1) { animation-delay: 0.1s; }
.card-wrapper:nth-child(2) { animation-delay: 0.2s; }
.card-wrapper:nth-child(3) { animation-delay: 0.3s; }
.card-wrapper:nth-child(4) { animation-delay: 0.4s; }
.card-wrapper:nth-child(5) { animation-delay: 0.5s; }
</style>
