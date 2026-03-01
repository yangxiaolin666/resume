<script setup>
import {
  Delete,
  Search,
} from '@element-plus/icons-vue'
import studentCard from '@/components/studentCard.vue'
import { onMounted, ref } from 'vue'
import { getAllStudent, getStatics } from '@/api/student'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useFilterStore } from '@/stores/students'
import { storeToRefs } from 'pinia'
const students = ref()
const router = useRouter()
const currentPage = ref(1)
const pageSize = ref(1000)
const total = ref(10)
const allTag = ref()
const filterStore = useFilterStore()
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
const getStudent = async () => {
  const res = await getAllStudent({
    "gender": checkGender.value,
    "degree": checkDegree.value,
    "major": checkMajor.value,
    "specialities": checkSpecial.value,
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
  //console.log(allTag.value)
  allLanguage.value = [...new Set(allTag.value.languages.map(item => item.language + "-" + item.level))]
  allJob.value = [...new Set(allTag.value.jobs.map(item => item.name))]
  allSpecial.value = [...new Set(allTag.value.specialities.map(item => item.name))]
  allCity.value = ['北京', '上海', '广州', '深圳', '武汉', '西安', '南京', '沈阳', '大连', '杭州']
  getStudent()
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
  ElMessage({ message: '搜索成功', type: 'success', duration: 800 })
  getStudent()
}
</script>

<template>
  <div class="home-container" style="margin-bottom: 60px;">
    <main class="main-content">
      <aside class="filter-sidebar">
        <el-card class="filter-card">
          <el-input placeholder="输入标签">
            <template #prefix>
              <el-icon>
                <Search />
              </el-icon>
            </template>
          </el-input>
          <el-collapse expand-icon-position="left">
            <el-collapse-item title="性别" name="-2">
              <el-radio-group v-model="checkGender">
                <el-radio value="男" size="small">男</el-radio>
                <el-radio value="女" size="small">女</el-radio>
              </el-radio-group>
            </el-collapse-item>
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
                <el-checkbox size="small" v-for="item in allLanguage" :label="item" :value="item" />
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
                {{ item }}
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
      <div v-if="students?.length > 0">
        <div style="display: grid; grid-template-columns: repeat(4, 1fr);">
          <div v-for="item in students" :key="item.id">
            <studentCard :item="item" />
          </div>
        </div>
      </div>
      <el-empty style="left:800px;top:170px;position: absolute;" v-else :image-size="250" description="暂无学生信息" />
    </main>
  </div>
  <div class="demo-pb">
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

.menu-item:hover {
  color: #409eff;
}

.menu-item span {
  padding-left: 20px;
  font-size: 14px;
}

.main-content {
  display: flex;
  gap: 0px;
  padding: 5px;
}

.filter-sidebar {
  margin-top: 10px;
  width: 280px;
  flex-shrink: 0;
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
}

.tag-box {
  display: flex;
  justify-content: flex-start;
  gap: 5px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
</style>
