<script setup>
import { getStudentById } from '@/api/student'
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'

const student = ref('')
const userStore = useUserStore()
const allLanguage = ref([])
const allSpeciality = ref([])
onMounted(async () => {
  const {
    data: {
      data: { list }
    }
  } = await getStudentById(userStore.userId)
  console.log('学生信息:', list[0])
  student.value = list[0]

  // 处理语言数据显示
  if (Array.isArray(list[0].languages)) {
    allLanguage.value = list[0].languages
  } else {
    allLanguage.value = []
  }

  // 处理专业技能显示
  if (Array.isArray(list[0].specialities)) {
    allSpeciality.value = list[0].specialities
  } else {
    allSpeciality.value = []
  }
})
</script>
<template>
  <el-descriptions title="个人信息" direction="vertical" border style="margin-top: 10px;padding:5px">
    <el-descriptions-item :rowspan="3" :width="140" :column="3" label="Avator" align="center">
      <el-image style="width: 150px; height: 150px"
        src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" />
    </el-descriptions-item>
    <el-descriptions-item label="用户名" :width="120">{{ userStore.username }}</el-descriptions-item>
    <el-descriptions-item label="姓名" :width="120">{{ student.name }}</el-descriptions-item>
    <el-descriptions-item label="用户类型" :width="120">{{ userStore.userType }}</el-descriptions-item>
    <el-descriptions-item label="标签" :width="200">
      <el-tag size="small" style="margin-right:5px">{{ student.graduationYear }}届</el-tag>
      <el-tag size="small" style="margin-right:5px">{{ student.degree }}</el-tag>
    </el-descriptions-item>
    <el-descriptions-item label="专业" :width="200">{{ student.major }}</el-descriptions-item>
    <el-descriptions-item label="实习公司" :width="300">
      <div style="display: flex; flex-wrap: wrap; gap: 5px;">
        <el-tag v-for="item in student.internships" :key="item" type="danger" size="small">
          {{ item.company }}-{{ item.experience }}
        </el-tag>
      </div>
    </el-descriptions-item>
    <el-descriptions-item label="实习次数" :width="120">{{ student.internship }}</el-descriptions-item>
    <el-descriptions-item label="民族" :width="120">{{ student.nation }}</el-descriptions-item>
    <el-descriptions-item label="意向工作城市" :width="300">
      <div style="display: flex; flex-wrap: wrap; gap: 5px;">
        <el-tag type="success" size="small">
          {{ student.region }}
        </el-tag>
      </div>
    </el-descriptions-item>
    <el-descriptions-item label="意向工作省份" :width="300">
      <div style="display: flex; flex-wrap: wrap; gap: 5px;">
        <el-tag type="success" size="small">
          {{ student.region }}
        </el-tag>
      </div>
    </el-descriptions-item>
    <el-descriptions-item label="证书名称" :width="300">
      <div style="display: flex; flex-wrap: wrap; gap: 5px;">
        <el-tag type="primary" v-for="item in student.certificates" :key="item" size="small">
          {{ item.certName }}-{{ item.level }}
        </el-tag>
      </div>
    </el-descriptions-item>
    <el-descriptions-item label="技能水平" :width="300">
      <div style="display: flex; flex-wrap: wrap; gap: 5px;">
        <el-tag v-for="item in student.languages" :key="item" type="warning" size="small">
          {{ item.language }}-{{ item.level }}
        </el-tag>
        <el-tag v-for="item in student.specialities" :key="item" size="small">
          {{ item.speciality }}
        </el-tag>
      </div>
    </el-descriptions-item>

  </el-descriptions>
  <el-calendar controller-type="select" />
</template>
<style scoped lang="scss"></style>