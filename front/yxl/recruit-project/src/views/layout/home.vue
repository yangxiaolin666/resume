<script>
import IconSchool from '@/components/IconSchool.vue'
export default {
  components: { IconSchool },
  data() {
    return {
      students: [
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
      ],
      // 筛选选项
      educationOptions: ['本科', '研究生'],
      skillOptions: [
        'JavaScript',
        'Vue',
        'React',
        'TypeScript',
        'Node.js',
        'Python',
        'C++',
        'Java',
        'Spring Boot',
        'MySQL',
        'HTML/CSS',
        'UI设计',
        'Photoshop',
        'Figma',
        '机器学习'
      ],
      experienceOptions: ['应届毕业生', '1年经验', '2年经验', '3年以上'],
      languageOptions: [
        '英语四级',
        '日语N1',
        '日语N2',
        '韩语TOPIK I',
        '韩语TOPIK II'
      ],
      certificateOptions: ['软件设计师', '程序员', '网络管理员'],
      internshipOptions: ['0次', '1次', '2次及以上'],
      computerSkillOptions: ['Java'],
      jobRegionOptions: ['辽宁省大连市'],
      jobPositionOptions: ['网络安全维护', '软件开发', '软件技术服务'],
      // 自定义标签相关
      customLanguage: '',
      customCertificate: '',
      customComputerSkill: '',
      customJobRegion: '',
      customJobPosition: '',
      // 选中的筛选条件
      selectedFilters: {
        education: '',
        majors: [],
        skills: [],
        experience: [],
        languages: [],
        certificates: [],
        internships: [],
        computerSkills: [],
        jobRegions: [],
        jobPositions: []
      },
      // 原始专业选项
      originalMajorOptions: [
        '计算机科学与技术',
        '软件工程',
        '网络工程',
        '人工智能'
      ],
      // 研究生专业选项
      postgraduateMajorOptions: ['计算机科学与技术', '人工智能'],
      // 当前查看详情的学生
      currentStudent: null,
      // 动画版本号，用于强制重新渲染动画
      animationVersion: 0,
      // 用于触发动画的key
      animationKey: 0
    }
  },
  computed: {
    // 根据筛选条件过滤学生并添加动画索引
    filteredStudents() {
      return this.students.filter((student) => {
        // 学历筛选
        if (
          this.selectedFilters.education &&
          this.selectedFilters.education !== student.education
        ) {
          return false
        }
        // 专业筛选
        if (
          this.selectedFilters.majors.length > 0 &&
          !this.selectedFilters.majors.includes(student.major)
        ) {
          return false
        }
        // 技能筛选
        if (this.selectedFilters.skills.length > 0) {
          const hasSkill = this.selectedFilters.skills.some((skill) =>
            student.skills.includes(skill)
          )
          if (!hasSkill) {
            return false
          }
        }
        // 工作经验筛选
        if (
          this.selectedFilters.experience.length > 0 &&
          !this.selectedFilters.experience.includes(student.experience)
        ) {
          return false
        }
        // 语言掌握情况筛选
        if (this.selectedFilters.languages.length > 0) {
          const hasLanguage = this.selectedFilters.languages.some((lang) =>
            student.languages.includes(lang)
          )
          if (!hasLanguage) {
            return false
          }
        }
        // 证书筛选
        if (this.selectedFilters.certificates.length > 0) {
          const hasCertificate = this.selectedFilters.certificates.some(
            (cert) => student.certificates.includes(cert)
          )
          if (!hasCertificate) {
            return false
          }
        }
        // 实习次数筛选
        if (this.selectedFilters.internships.length > 0) {
          let matchesInternship = false
          if (
            this.selectedFilters.internships.includes('0次') &&
            student.internships === 0
          ) {
            matchesInternship = true
          }
          if (
            this.selectedFilters.internships.includes('1次') &&
            student.internships === 1
          ) {
            matchesInternship = true
          }
          if (
            this.selectedFilters.internships.includes('2次及以上') &&
            student.internships >= 2
          ) {
            matchesInternship = true
          }
          if (!matchesInternship) {
            return false
          }
        }
        // 计算机专业技能筛选
        if (this.selectedFilters.computerSkills.length > 0) {
          const hasComputerSkill = this.selectedFilters.computerSkills.some(
            (skill) => student.computerSkills.includes(skill)
          )
          if (!hasComputerSkill) {
            return false
          }
        }
        // 就业意向地区筛选
        if (this.selectedFilters.jobRegions.length > 0) {
          const hasJobRegion = this.selectedFilters.jobRegions.some((region) =>
            student.jobRegions.includes(region)
          )
          if (!hasJobRegion) {
            return false
          }
        }
        // 就业岗位筛选
        if (this.selectedFilters.jobPositions.length > 0) {
          const hasJobPosition = this.selectedFilters.jobPositions.some(
            (position) => student.jobPositions.includes(position)
          )
          if (!hasJobPosition) {
            return false
          }
        }
        return true
      })
    },
    // 根据选择的学历动态决定显示的专业选项
    majorOptions() {
      if (this.selectedFilters.education === '研究生') {
        return this.postgraduateMajorOptions
      }
      return this.originalMajorOptions
    },
    // 计算已选择的筛选条件数量
    appliedFilters() {
      let count = 0
      Object.values(this.selectedFilters).forEach((filter) => {
        if (Array.isArray(filter)) {
          count += filter.length
        } else if (filter) {
          count++
        }
      })
      return count
    }
  },
  methods: {
    resetFilters() {
      this.selectedFilters = {
        education: '',
        majors: [],
        skills: [],
        experience: [],
        languages: [],
        certificates: [],
        internships: [],
        computerSkills: [],
        jobRegions: [],
        jobPositions: []
      }
    },
    // 处理学历选择变化，清空已选专业
    handleEducationChange() {
      this.selectedFilters.majors = []
      // 更新动画版本号，强制重新渲染动画
      this.triggerAnimation()
    },
    // 触发动画重新渲染
    triggerAnimation() {
      // 先移除所有行的动画类
      this.animationKey++
    },
    // 重置新学生表单
    resetNewStudentForm() {
      this.newStudent = {
        name: '',
        age: '',
        education: '',
        major: '',
        school: '',
        experience: '',
        skills: [],
        languages: [],
        certificates: [],
        internships: 0,
        computerSkills: [],
        jobRegions: [],
        jobPositions: [],
        bio: '',
        contactInfo: {
          phone: '',
          email: ''
        },
        photo: null
      }
      // 清空自定义标签输入框
      this.customLanguage = ''
      this.customCertificate = ''
      this.customComputerSkill = ''
      this.customJobRegion = ''
      this.customJobPosition = ''
      // 清空文件输入框
      const fileInput = document.getElementById('photo-upload')
      if (fileInput) {
        fileInput.value = ''
      }
    }
  }
}
</script>
<template>
  <div id="app">
    <header class="header">
      <div class="one" style="display: flex; flex-direction: row">
        <IconSchool />
        <h1>
          <p>学生招聘平台</p>
        </h1>
      </div>

      <div class="icons-right" style="
          width: 100px;
          display: flex;
          justify-content: space-between;
          margin-top: 30px;
          margin-right: 80px;
        ">
        <div style="display: flex; flex-direction: column">
          <svg class="icon loginHead" t="1763469640639" viewBox="0 0 1024 1024" version="1.1"
            xmlns="http://www.w3.org/2000/svg" p-id="2286" width="40" height="40"
            @click="$router.push('/updateResume')">
            <path
              d="M618.666667 288a32 32 0 0 1 0 64H320a32 32 0 0 1 0-64h298.666667zM704 480a32 32 0 0 1 0 64H320a32 32 0 0 1 0-64h384zM533.333333 672a32 32 0 0 1 0 64H320a32 32 0 0 1 0-64h213.333333z"
              fill="#ffffff" p-id="2287" />
            <path
              d="M768 85.333333a128 128 0 0 1 128 128v597.333334a128 128 0 0 1-128 128H256a128 128 0 0 1-128-128V213.333333a128 128 0 0 1 128-128h512z m0 64H256a64 64 0 0 0-64 64v597.333334a64 64 0 0 0 64 64h512a64 64 0 0 0 64-64V213.333333a64 64 0 0 0-64-64z"
              fill="#ffffff" p-id="2288" />
          </svg>
          <p>简历</p>
        </div>
        <div style="display: flex; flex-direction: column" @click="$router.push('/login')">
          <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" t="1763469466651"
            viewBox="0 0 1024 1024" version="1.1" p-id="3498" width="40" height="40" class="icon">
            <path
              d="M896 448h-128a32 32 0 0 0 0 64h128a32 32 0 0 0 0-64M896 320h-128a32 32 0 0 0 0 64h128a32 32 0 0 0 0-64M768 256h128a32 32 0 0 0 0-64h-128a32 32 0 0 0 0 64"
              fill="#fff" p-id="3499" />
            <path
              d="M781.76 640c-5.12-2.272-97.216-42.56-198.72-62.752 11.808-9.632 22.144-20.928 31.104-33.248a174.88 174.88 0 0 0 33.792-103.2v-104.768A176.32 176.32 0 0 0 576 194.336 174.368 174.368 0 0 0 472.192 160c-96.928 0-175.776 78.976-175.776 176.032v104.768c0 56.544 26.88 106.816 68.384 139.072-96.48 20.832-181.344 57.824-185.536 59.712C150.016 651.328 128 684.192 128 716v131.712l1.728 5.024c9.504 27.904 35.136 46.624 63.776 46.624h572.992c33.28 0 60.832-24.864 64.96-56.96l0.544-126.4c0-15.232-5.184-30.72-13.664-44-9.088-14.272-21.952-26.016-36.576-32"
              fill="#fff" p-id="3500" />
            /> /> />
          </svg>登录
        </div>
      </div>
    </header>
    <main class="main-container">
      <aside class="filter-panel">
        <h2>筛选条件</h2>

        <div class="filter-section">
          <h3>学历</h3>
          <div class="filter-options">
            <label v-for="edu in educationOptions" :key="edu">
              <input v-model="selectedFilters.education" type="radio" :value="edu" @change="handleEducationChange">
              {{ edu }}
            </label>
          </div>
        </div>

        <div class="filter-section">
          <h3>专业</h3>
          <div class="filter-options">
            <label v-for="major in majorOptions" :key="major">
              <input v-model="selectedFilters.majors" type="checkbox" :value="major">
              {{ major }}
            </label>
          </div>
        </div>

        <div class="filter-section">
          <h3>技能标签</h3>
          <div class="filter-options">
            <label v-for="skill in skillOptions" :key="skill">
              <input v-model="selectedFilters.skills" type="checkbox" :value="skill">
              {{ skill }}
            </label>
          </div>
        </div>

        <div class="filter-section">
          <h3>工作经验</h3>
          <div class="filter-options">
            <label v-for="exp in experienceOptions" :key="exp">
              <input v-model="selectedFilters.experience" type="checkbox" :value="exp">
              {{ exp }}
            </label>
          </div>
        </div>

        <!-- 语言掌握情况 -->
        <div class="filter-section">
          <h3>语言掌握情况</h3>
          <div class="filter-options">
            <label v-for="lang in languageOptions" :key="lang">
              <input v-model="selectedFilters.languages" type="checkbox" :value="lang">
              {{ lang }}
            </label>
          </div>
        </div>

        <!-- 计算机专业技术证书 -->
        <div class="filter-section">
          <h3>计算机专业技术证书</h3>
          <div class="filter-options">
            <label v-for="cert in certificateOptions" :key="cert">
              <input v-model="selectedFilters.certificates" type="checkbox" :value="cert">
              {{ cert }}
            </label>
          </div>
        </div>

        <!-- 专业实习次数 -->
        <div class="filter-section">
          <h3>专业实习次数</h3>
          <div class="filter-options">
            <label v-for="intern in internshipOptions" :key="intern">
              <input v-model="selectedFilters.internships" type="checkbox" :value="intern">
              {{ intern }}
            </label>
          </div>
        </div>

        <!-- 计算机专业掌握情况 -->
        <div class="filter-section">
          <h3>计算机专业掌握情况</h3>
          <div class="filter-options">
            <label v-for="skill in computerSkillOptions" :key="skill">
              <input v-model="selectedFilters.computerSkills" type="checkbox" :value="skill">
              {{ skill }}
            </label>
          </div>
        </div>

        <!-- 就业意向地区 -->
        <div class="filter-section">
          <h3>就业意向地区</h3>
          <div class="filter-options">
            <label v-for="region in jobRegionOptions" :key="region">
              <input v-model="selectedFilters.jobRegions" type="checkbox" :value="region">
              {{ region }}
            </label>
          </div>
        </div>

        <!-- 就业岗位 -->
        <div class="filter-section">
          <h3>就业岗位</h3>
          <div class="filter-options">
            <label v-for="position in jobPositionOptions" :key="position">
              <input v-model="selectedFilters.jobPositions" type="checkbox" :value="position">
              {{ position }}
            </label>
          </div>
        </div>

        <button class="reset-btn" @click="resetFilters">重置筛选</button>
      </aside>

      <div class="student-list">
        <div v-if="filteredStudents.length > 0" class="students-table-container">
          <table class="students-table">
            <thead>
              <tr>
                <th>姓名</th>
                <th>专业</th>
                <th>学历</th>
                <th>工作经验</th>
                <th>技能</th>
                <th>详情</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(student, index) in filteredStudents" :key="index" class="student-box"
                :data-student-id="index">
                <td>{{ student.name }}</td>
                <td>{{ student.major }}</td>
                <td>{{ student.education }}</td>
                <td>{{ student.experience }}</td>
                <td>
                  <div class="skill-tags-small">
                    <span v-for="skill in student.skills" :key="skill" class="tag-small">{{ skill }}</span>
                  </div>
                </td>
                <td>
                  <button class="contact-btn" @click="viewStudentDetail(student)">
                    详情
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-else class="empty-state">
          <p>没有找到符合条件的学生</p>
          <button class="reset-btn" @click="resetFilters">重置筛选条件</button>
        </div>
      </div>
    </main>
  </div>
</template>
<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family:
    -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue',
    Arial, sans-serif;
  line-height: 1.6;
  color: #333;
  background-color: #f5f5f5;
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 头部样式 */
.header {
  background: var(--main-blue);
  color: white;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  text-align: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  position: fixed;
  top: 0;
  right: 0;
  left: 0;
  z-index: 1000;
}

@keyframes move {
  0% {
    stroke-dasharray: 0 3000;
  }

  100% {
    stroke-dasharray: 3000 0;
  }
}

@keyframes jump {
  0% {
    transform: translateY(0);
  }

  100% {
    transform: translateY(-5px);
  }
}

.header h1 {
  font-size: 2.5rem;
  padding: 1.7rem 0rem;
  font-weight: 600;
}

.header .icons-right .icon:hover {
  animation: jump 0.2s linear forwards;
}

/* 主容器 */
.main-container {
  flex: 1;
  display: flex;
  gap: 2rem;
  max-width: 1400px;
  margin: 180px auto 2rem;
  padding: 0 1rem;
  width: 100%;
}

/* 筛选面板 */
.filter-panel {
  width: 350px;
  background: white;
  padding: 1.5rem;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
  height: fit-content;
  position: sticky;
  top: 2rem;
  max-height: calc(100vh - 4rem);
  overflow-y: auto;
}

.filter-panel h2 {
  font-size: 1.3rem;
  margin-bottom: 1.5rem;
  color: #333;
  border-bottom: 2px solid var(--main-blue);;
  padding-bottom: 0.5rem;
}

.filter-section {
  margin-bottom: 1.5rem;
}

.filter-section h3 {
  font-size: 1rem;
  color: #666;
  font-weight: 500;
  margin: 0;
}

.filter-options {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.filter-options label {
  display: flex;
  align-items: center;
  cursor: pointer;
  font-size: 0.9rem;
  transition: color 0.2s;
}

.filter-options label:hover {
  color:var(--main-blue);
}

.filter-options input[type='checkbox'] {
  margin-right: 0.5rem;
  cursor: pointer;
}

.reset-btn {
  width: 100%;
  padding: 0.8rem;
  background-color: #f0f0f0;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
  margin-top: 1rem;
}

.reset-btn:hover {
  background-color: #e0e0e0;
  transform: translateY(-1px);
}

/* 学生列表 */
.student-list {
  flex: 1;
  background: white;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

@keyframes Up {
  0% {
    transform: translateY(1000px);
  }

  100% {
    transform: translateY(0px);
    opacity: 1;
  }
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 4rem 1rem;
  color: #999;
}

.empty-state p {
  font-size: 1.1rem;
  margin-bottom: 1.5rem;
}

/* 学生表格样式 */
.students-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
  /* 略微减小字体大小 */
  background-color: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  table-layout: fixed;
}

/* 表格头部样式 */
.students-table thead {
  background: var(--main-blue);
  color: white;
}

.students-table th {
  padding: 0.75rem 1rem;
  /* 减小内边距 */
  text-align: center;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  font-size: 0.8rem;
  /* 略微减小字体大小 */
  /* 为每一列设置固定宽度 */
}

/* 表格主体样式 */
.students-table tbody tr[data-student-id] {
  /* 使用attr()函数获取data-student-id属性的值 */
  opacity: 0;
  animation: Up 0.5s calc(attr(data-student-id number) * 0.1s) forwards;
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.08);
  height: auto;
}

.students-table tbody tr:hover {
  background-color: #f8f9fa;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.students-table tbody tr:last-child {
  border-bottom: none;
}

.students-table td {
  padding: 0.75rem;
  vertical-align: middle;
  /* 保持垂直居中 */
  text-align: center;
  word-break: break-word;
}

/* 小技能标签容器 */
.skill-tags-small {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  /* 减小标签间距 */
  justify-content: center;
  align-items: center;
  min-height: 28px;
  /* 略微减小最小高度 */
}

/* 小技能标签样式 */
.tag-small {
  background-color: var(--main-blue);
  color: white;
  padding: 0.2rem 0.6rem;
  /* 减小标签内边距 */
  border-radius: 20px;
  font-size: 0.75rem;
  /* 略微减小字体大小 */
  font-weight: 500;
  transition: all 0.2s ease;
}

.tag-small:hover {
  background-color: #764ba2;
  transform: scale(1.05);
}

/* 操作按钮样式 */
.students-table .contact-btn {
  padding: 0.35rem 0.75rem;
  /* 减小按钮内边距 */
  background-color: var(--main-blue);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.8rem;
  /* 减小按钮字体大小 */
  font-weight: 500;
  transition: all 0.3s ease;
  text-decoration: none;
  display: inline-block;
  line-height: 1.3;
}

.students-table .contact-btn:hover {
  background-color: var(--main-blue);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

/* 响应式布局 */
@media (max-width: 1024px) {
  .main-container {
    flex-direction: column;
  }

  .filter-panel {
    width: 100%;
    position: static;
  }
}

@media (max-width: 768px) {
  .header h1 {
    font-size: 2rem;
  }

  .students-table {
    font-size: 0.9rem;
    table-layout: fixed;
  }

  .students-table th,
  .students-table td {
    padding: 0.75rem 0.5rem;
  }

  .students-table th {
    font-size: 0.75rem;
  }

  /* 响应式列宽 */
  .students-table th:nth-child(1),
  .students-table td:nth-child(1) {
    width: 15%;
  }

  /* 姓名 */
  .students-table th:nth-child(2),
  .students-table td:nth-child(2) {
    width: 18%;
  }

  /* 专业 */
  .students-table th:nth-child(3),
  .students-table td:nth-child(3) {
    width: 15%;
  }

  /* 学历 */
  .students-table th:nth-child(4),
  .students-table td:nth-child(4) {
    width: 15%;
  }

  /* 工作经验 */
  .students-table th:nth-child(5),
  .students-table td:nth-child(5) {
    width: 22%;
  }

  /* 技能 */
  .students-table th:nth-child(6),
  .students-table td:nth-child(6) {
    width: 15%;
  }

  /* 操作 */

  .tag-small {
    padding: 0.2rem 0.5rem;
    font-size: 0.7rem;
  }

  .students-table .contact-btn {
    padding: 0.4rem 0.75rem;
    font-size: 0.8rem;
  }
}
</style>
