import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 学生数据
const initialStudents = [
  {
    id: 1,
    stuId: 20230001,
    name: '张明',
    age: 22,
    gender: '男',
    nation: '汉族',
    major: '计算机科学与技术',
    degree: '本科',
    graduationYear: 2024,
    internship: 2,
    languages: ['英语', '日语'],
    languageLevel: '六级',
    cities: ['北京', '上海'],
    provinces: ['北京市', '上海市'],
    certNames: ['AWS认证', 'PMP认证'],
    certLevel: '高级',
    internshipCompany: '阿里巴巴',
    specialities: ['Java', 'Spring Boot', 'MySQL', 'Redis']
  },
  {
    id: 2,
    stuId: 20230002,
    name: '李华',
    age: 21,
    gender: '女',
    nation: '汉族',
    major: '软件工程',
    degree: '本科',
    graduationYear: 2024,
    internship: 1,
    languages: ['英语'],
    languageLevel: '四级',
    cities: ['深圳'],
    provinces: ['广东省'],
    certNames: ['软件设计师'],
    certLevel: '中级',
    internshipCompany: '腾讯',
    specialities: ['Python', 'Django', 'Vue.js', 'React']
  },
  {
    id: 3,
    stuId: 20230003,
    name: '王芳',
    age: 23,
    gender: '女',
    nation: '回族',
    major: '人工智能',
    degree: '硕士',
    graduationYear: 2025,
    internship: 3,
    languages: ['英语', '德语'],
    languageLevel: '六级',
    cities: ['杭州', '南京'],
    provinces: ['浙江省', '江苏省'],
    certNames: ['TensorFlow认证', '深度学习工程师'],
    certLevel: '高级',
    internshipCompany: '字节跳动',
    specialities: ['Python', '机器学习', '深度学习', 'TensorFlow']
  },
  {
    id: 4,
    stuId: 20230004,
    name: '刘强',
    age: 20,
    gender: '男',
    nation: '汉族',
    major: '网络工程',
    degree: '本科',
    graduationYear: 2024,
    internship: 0,
    languages: ['英语'],
    languageLevel: '四级',
    cities: ['成都'],
    provinces: ['四川省'],
    certNames: ['CCNA', '网络工程师'],
    certLevel: '中级',
    internshipCompany: '',
    specialities: ['网络协议', 'Linux', '网络安全', 'Docker']
  },
  {
    id: 5,
    stuId: 20230005,
    name: '陈静',
    age: 22,
    gender: '女',
    nation: '壮族',
    major: '数据科学',
    degree: '硕士',
    graduationYear: 2025,
    internship: 2,
    languages: ['英语', '法语'],
    languageLevel: '六级',
    cities: ['广州', '武汉'],
    provinces: ['广东省', '湖北省'],
    certNames: ['数据分析师', 'SQL专家'],
    certLevel: '高级',
    internshipCompany: '百度',
    specialities: ['Python', 'SQL', '数据分析', '机器学习']
  }
]

export const useStudentsStore = defineStore(
  'students',
  () => {
    // 状态
    const students = ref(initialStudents)
    const selectedStudent = ref(null)
    const searchQuery = ref('')
    const filterCriteria = ref({
      major: '',
      degree: '',
      internship: null,
      languages: [],
      specialities: []
    })

    // 计算属性
    const allStudents = computed(() => students.value)
    
    const filteredStudents = computed(() => {
      let filtered = students.value
      
      // 姓名模糊搜索
      if (searchQuery.value) {
        filtered = filtered.filter(student => 
          student.name.includes(searchQuery.value)
        )
      }
      
      // 专业模糊搜索
      if (filterCriteria.value.major) {
        filtered = filtered.filter(student => 
          student.major.includes(filterCriteria.value.major)
        )
      }
      
      // 学位筛选
      if (filterCriteria.value.degree) {
        filtered = filtered.filter(student => 
          student.degree === filterCriteria.value.degree
        )
      }
      
      // 实习次数筛选
      if (filterCriteria.value.internship !== null) {
        filtered = filtered.filter(student => 
          student.internship >= filterCriteria.value.internship
        )
      }
      
      // 语言筛选
      if (filterCriteria.value.languages.length > 0) {
        filtered = filtered.filter(student =>
          filterCriteria.value.languages.some(lang => 
            student.languages.includes(lang)
          )
        )
      }
      
      // 技能筛选
      if (filterCriteria.value.specialities.length > 0) {
        filtered = filtered.filter(student =>
          filterCriteria.value.specialities.some(skill => 
            student.specialities.includes(skill)
          )
        )
      }
      
      return filtered
    })
    
    const studentCount = computed(() => students.value.length)
    
    const allMajors = computed(() => {
      return [...new Set(students.value.map(student => student.major))]
    })
    
    const allDegrees = computed(() => {
      return [...new Set(students.value.map(student => student.degree))]
    })
    
    const allLanguages = computed(() => {
      const allLangs = students.value.flatMap(student => student.languages)
      return [...new Set(allLangs)]
    })
    
    const allSkills = computed(() => {
      const allSkills = students.value.flatMap(student => student.specialities)
      return [...new Set(allSkills)]
    })

    // 方法
    function setSelectedStudent(student) {
      selectedStudent.value = student
    }
    
    function clearSelectedStudent() {
      selectedStudent.value = null
    }
    
    function setSearchQuery(query) {
      searchQuery.value = query
    }
    
    function setFilterCriteria(criteria) {
      filterCriteria.value = { ...filterCriteria.value, ...criteria }
    }
    
    function clearFilters() {
      searchQuery.value = ''
      filterCriteria.value = {
        major: '',
        degree: '',
        internship: null,
        languages: [],
        specialities: []
      }
    }
    
    function getStudentById(id) {
      return students.value.find(student => student.id === id)
    }
    
    function addStudent(newStudent) {
      const student = {
        id: Math.max(...students.value.map(s => s.id)) + 1,
        ...newStudent
      }
      students.value.push(student)
      return student
    }
    
    function updateStudent(id, updatedData) {
      const index = students.value.findIndex(student => student.id === id)
      if (index !== -1) {
        students.value[index] = { ...students.value[index], ...updatedData }
        return students.value[index]
      }
      return null
    }
    
    function deleteStudent(id) {
      const index = students.value.findIndex(student => student.id === id)
      if (index !== -1) {
        students.value.splice(index, 1)
        return true
      }
      return false
    }

    return {
      // 状态
      students,
      selectedStudent,
      searchQuery,
      filterCriteria,
      
      // 计算属性
      allStudents,
      filteredStudents,
      studentCount,
      allMajors,
      allDegrees,
      allLanguages,
      allSkills,
      
      // 方法
      setSelectedStudent,
      clearSelectedStudent,
      setSearchQuery,
      setFilterCriteria,
      clearFilters,
      getStudentById,
      addStudent,
      updateStudent,
      deleteStudent
    }
  },
  {
    persist: true
  }
)
