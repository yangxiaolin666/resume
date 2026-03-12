<script setup>
import {
  Delete,
  Search,
} from '@element-plus/icons-vue'
import studentCard from '@/components/studentCard.vue'
import { onMounted, ref, watch } from 'vue'
import { getAllStudent, getStatics, addLike, deleteLike, queryComId, queryLike } from '@/api/student'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useFilterStore } from '@/stores/students'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
const students = ref()
const router = useRouter()
const currentPage = ref(1)
const pageSize = ref(1000)
const total = ref(10)
const allTag = ref()
const filterStore = useFilterStore()
const userStore = useUserStore()
const { clearAllFilters } = filterStore
const {
  checkDegree,
  checkMajor,
  checkGender,
  checkLanguage,
  checkJob,
  checkSpecial,
  checkCity
} = storeToRefs(filterStore)
const allLanguage = ref()
const allJob = ref()
const allSpecial = ref()
const allCity = ref()
const com_id = ref(5)
const likedStudents = ref(new Set())
const getStudent = async () => {
  const res = await getAllStudent({
    "gender": checkGender.value,
    "degree": checkDegree.value,
    "major": checkMajor.value,
    "languages": checkLanguage.value,
    "specialities": checkSpecial.value,
    "jobs": checkJob.value,
    "cities": checkCity.value,
    "pageNum": currentPage.value,
    "pageSize": pageSize.value
  })
  const { data: { data } } = res
  console.log(data)
  total.value = data.size
  students.value = data.list
}
onMounted(async () => {
  const { data: { data: data1 } } = await getStatics()
  allTag.value = data1
  console.log(allTag.value)
  allLanguage.value = [...new Set(allTag.value.languages)]
  allJob.value = [...new Set(allTag.value.jobs.map(item => item.name))]
  allSpecial.value = [...new Set(allTag.value.specialities.map(item => item.name))]
  allCity.value = ['北京', '上海', '广州', '深圳', '武汉', '西安', '南京', '沈阳', '大连', '杭州']

  const res = await queryComId()
  com_id.value = res.data.data

  await getStudent()
  await initLikeStatus()
  console.log(checkSpecial.value);

})
const handleSizeChange = (val) => {
  console.log(`${val} items per page`)
}
const handleCurrentChange = async (val) => {
  console.log(`current page: ${val}`)
  const res = await getAllStudent({
    "pageNum": currentPage.value,
    "pageSize": pageSize.value
  })
  const { data: { data } } = res
  console.log(data)
  students.value = data.list
}
const clearTag = () => {
  clearAllFilters()
  ElMessage({
    message: '清空标签成功',
    type: 'success',
    duration: 800
  })
  setTimeout(() => {
    getStudent()
  }, 100)
}
const search = async () => {
  getStudent()
  ElMessage({ message: '搜索成功', type: 'success', duration: 800 })
}
watch(checkLanguage, (newV) => {
  console.log(newV);
})

const initLikeStatus = async () => {
  try {
    const { data: { data: { list } } } = await queryLike({
      "com_id": com_id.value,
      "pageNum": 1,
      "pageSize": 1000
    })

    likedStudents.value = new Set(list.map(item => item.stuId))
  } catch (error) {
    console.error('初始化收藏状态失败:', error)
  }
}

const toggleLike = async (student) => {
  try {
    const IdList = [{
      "comId": com_id.value,
      "stuId": student.id,
    }]

    if (likedStudents.value.has(student.id)) {
      await deleteLike(IdList)
      likedStudents.value.delete(student.id)
      ElMessage({
        message: '取消收藏成功',
        type: 'success',
        duration: 500
      })
    } else {
      await addLike(IdList)
      likedStudents.value.add(student.id)
      ElMessage({
        message: '收藏成功',
        type: 'success',
        duration: 500
      })
    }
  } catch (error) {
    console.error('收藏操作失败:', error)
    ElMessage({
      message: '操作失败，请重试',
      type: 'error',
      duration: 1000
    })
  }
}
</script>

<template>
  <div class="home-container" style="margin-bottom: 60px;">
    <main class="main-content">
      <aside class="filter-sidebar">
        <el-card class="filter-card">
          <el-collapse expand-icon-position="left">
            <el-collapse-item title="学位" name="-1">
              <el-radio-group v-model="checkDegree">
                <el-radio value="本科生" size="small">本科生</el-radio>
                <el-radio value="研究生" size="small">研究生</el-radio>
              </el-radio-group>
            </el-collapse-item>
            <el-collapse-item title="专业" name="0">
              <el-radio-group v-model="checkMajor">
                <el-radio value="计算机科学与技术" size="small">计算机科学与技术</el-radio>
                <el-radio value="软件工程" size="small">软件工程</el-radio>
                <el-radio value="人工智能" size="small">人工智能</el-radio>
                <el-radio value="网络工程" size="small">网络工程</el-radio>
                <el-radio value="计算机技术" size="small">计算机技术</el-radio>
              </el-radio-group>
            </el-collapse-item>
            <el-collapse-item title="语言水平" name="1">
              <el-checkbox-group v-model="checkLanguage">
                <el-checkbox size="small" v-for="item in allLanguage" :label="item.language + '-' + item.level"
                  :value="item" />
              </el-checkbox-group>
            </el-collapse-item>
            <el-collapse-item title="技能水平" name="2">
              <el-checkbox-group v-model="checkSpecial">
                <el-checkbox size="small" v-for="item in allSpecial" :label="item" :value="item" />
              </el-checkbox-group>
            </el-collapse-item>
            <el-collapse-item title="意向工作" name="3">
              <el-checkbox-group v-model="checkJob">
                <el-checkbox size="small" v-for="item in allJob" :label="item" :value="item" />
              </el-checkbox-group>
            </el-collapse-item>
            <el-collapse-item title="热门城市" name="4">
              <el-checkbox-group v-model="checkCity">
                <el-checkbox size="small" v-for="item in allCity" :label="item" :value="item" />
              </el-checkbox-group>
            </el-collapse-item>
          </el-collapse>
          <template #header>
            <div class="filter-header">
              <h4 style="margin:0px;">当前已选标签</h4>
              <div>
                <el-button @click="clearTag">
                  <el-icon>
                    <Delete />
                  </el-icon>
                </el-button>
                <el-button @click="search">
                  <el-icon>
                    <Search />
                  </el-icon>
                </el-button>
              </div>
            </div>
            <div class="tag-box">
              <el-tag size="small" type="info" effect="plain" class="tag" v-if="checkGender.length > 0">
                {{ checkGender }}
              </el-tag>
              <el-tag size="small" type="info" effect="plain" class="tag" v-if="checkDegree.length > 0">
                {{ checkDegree }}
              </el-tag>
              <el-tag size="small" type="info" effect="plain" class="tag" v-if="checkMajor.length > 0">
                {{ checkMajor }}
              </el-tag>
              <el-tag v-for="(item, index) in checkLanguage" :key="index" size="small" type="info" effect="plain"
                class="tag">
                {{ item.language + "-" + item.level }}
              </el-tag>
              <el-tag v-for="(item, index) in checkSpecial" :key="index" size="small" type="info" effect="plain"
                class="tag">
                {{ item }}
              </el-tag>
              <el-tag v-for="(item, index) in checkJob" :key="index" size="small" type="info" effect="plain"
                class="tag">
                {{ item }}
              </el-tag>
              <el-tag v-for="(item, index) in checkCity" :key="index" size="small" type="info" effect="plain"
                class="tag">
                {{ item }}
              </el-tag>
            </div>
          </template>
        </el-card>
      </aside>
      <div v-if="students?.length > 0" style="min-height: 400px; flex: 1;margin-top:8px;border-radius: 10px;">
        <el-table :data="students" style="width: 100%" :border="true" stripe>
          <el-table-column prop="name" label="姓名" width="100" />
          <el-table-column prop="gender" label="性别" width="80" />
          <el-table-column prop="degree" label="学位" width="100" />
          <el-table-column prop="major" label="专业" width="150" />
          <el-table-column prop="graduationYear" label="毕业年份" width="100" />
          <el-table-column label="语言能力" width="200">
            <template #default="scope">
              <div
                style="height: 48px; overflow: hidden; display: flex; flex-direction: column; justify-content: center;">
                <template v-if="scope.row.languages?.length > 0">
                  <div v-if="scope.row.languages.length <= 3" style="display: flex; justify-content: center; gap: 4px;">
                    <el-tag v-for="lang in scope.row.languages" :key="lang.language" size="small" type="warning">
                      {{ lang.language }}-{{ lang.level }}
                    </el-tag>
                  </div>
                  <template v-else>
                    <div style="display: flex; justify-content: center; gap: 4px; margin-bottom: 4px;">
                      <el-tag v-for="lang in scope.row.languages.slice(0, 3)" :key="lang.language" size="small"
                        type="warning">
                        {{ lang.language }}-{{ lang.level }}
                      </el-tag>
                    </div>
                    <div style="display: flex; justify-content: center; gap: 4px;">
                      <el-tag v-for="lang in scope.row.languages.slice(3, 6)" :key="lang.language" size="small"
                        type="warning">
                        {{ lang.language }}-{{ lang.level }}
                      </el-tag>
                    </div>
                  </template>
                </template>
                <span v-else
                  style="display: flex; justify-content: center; align-items: center; height: 48px; color: #999;">
                  暂无标签
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="专业技能" width="200">
            <template #default="scope">
              <div
                style="height: 48px; overflow: hidden; display: flex; flex-direction: column; justify-content: center;">
                <template v-if="scope.row.specialities?.length > 0">
                  <div v-if="scope.row.specialities.length <= 3"
                    style="display: flex; justify-content: center; gap: 4px;">
                    <el-tag v-for="skill in scope.row.specialities" :key="skill.speciality" size="small" type="primary">
                      {{ skill.speciality }}
                    </el-tag>
                  </div>
                  <template v-else>
                    <div style="display: flex; justify-content: center; gap: 4px; margin-bottom: 4px;">
                      <el-tag v-for="skill in scope.row.specialities.slice(0, 3)" :key="skill.speciality" size="small"
                        type="primary">
                        {{ skill.speciality }}
                      </el-tag>
                    </div>
                    <div style="display: flex; justify-content: center; gap: 4px;">
                      <el-tag v-for="skill in scope.row.specialities.slice(3, 6)" :key="skill.speciality" size="small"
                        type="primary">
                        {{ skill.speciality }}
                      </el-tag>
                    </div>
                  </template>
                </template>
                <span v-else
                  style="display: flex; justify-content: center; align-items: center; height: 48px; color: #999;">
                  暂无标签
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="查看学生">
            <template #default="scope">
              <el-button type="primary" size="small"
                @click="$router.push({ path: '/studentDetail', query: { id: scope.row.id } })">
                查看简历
              </el-button>
            </template>
          </el-table-column>
          <el-table-column label="收藏学生">
            <template #default="scope">
              <el-button :type="likedStudents.has(scope.row.id) ? 'info' : 'success'" size="small"
                @click="toggleLike(scope.row)">
                {{ likedStudents.has(scope.row.id) ? '已收藏' : '收藏' }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <el-empty style="margin: 100px auto;" v-else :image-size="250" description="暂无学生信息" />
    </main>
  </div>
  <div class="demo-pb" v-if="students?.length > 0">
    <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" size="large" :background="true"
      layout="prev, pager, next, jumper" :total="total" @size-change="handleSizeChange"
      @current-change="handleCurrentChange" />
  </div>
</template>

<style scoped>
.menu-item {
  display: flex;
  align-items: center;
  padding: 3px 0px;
  cursor: pointer;
}

::-webkit-scrollbar {
  width: 0px;
  height: 1px;
}

::-webkit-scrollbar-thumb {
  border-radius: 5px;
  -webkit-box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.2);
  background: rgba(0, 0, 0, 0.2);
}

body {
  overflow-y: scroll;
  overflow-x: scroll;
}

.menu-item:hover {
  color: #409eff;
}

.menu-item span {
  padding-left: 20px;
  font-size: 14px;
}

.main-content {
  display: flex;
  gap: 6px;
  padding: 12px;
  min-height: calc(100vh - 120px);
  padding-bottom: 80px;
}

.filter-sidebar {
  margin-top: 8px;
  width: 260px;
  flex-shrink: 0;
  position: sticky;
  top: 16px;
  height: fit-content;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
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

.tag-box {
  display: flex;
  justify-content: flex-start;
  gap: 4px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.home-container {
  min-height: 100vh;
  background-color: #f5f5f5;
}
</style>