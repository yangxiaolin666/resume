import { defineStore } from 'pinia'
import { ref } from 'vue'

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
    const students = ref(initialStudents)
    return {
      students
    }
  },
  {
    persist: true
  }
)
