<script setup>
import { addLike, deleteLike, getStudentById, queryComId, queryLike } from '@/api/student'
import { fileUrl } from '@/utils/request'
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowRightBold } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
const router = useRouter()
const route = useRoute()
const studentId = route.query.id
const student = ref('')
const allLanguage = ref([])
const allSpeciality = ref([])
const pdfURL = ref()
const pageNum = 1
const pageSize = 10
const comId = 5
const stuId = studentId
const IdList = [
  {
    "comId": comId,
    "stuId": stuId,
  }
]
const isLike = ref(false)
const toggleLike = async () => {
  if (isLike.value) {
    const deleteData = [{
      comId: com_id.value,
      stuId: stuId
    }]
    const res = await deleteLike(deleteData)
    ElMessage({
      message: '取消收藏成功',
      type: 'success',
      duration: 500
    })
    console.log(res)
  } else {
    const addData = [{
      comId: com_id.value,
      stuId: stuId
    }]
    const res = await addLike(addData)
    ElMessage({
      message: '收藏成功',
      type: 'success',
      duration: 500
    })
  }
  isLike.value = !isLike.value
}
const com_id = ref()
onMounted(async () => {
  const res = await queryComId()
  com_id.value = res.data.data
  const { data: { data: { list: list1 } } } = await queryLike({
    "com_id": com_id.value,
    "pageNum": pageNum,
    "pageSize": pageSize
  })
  console.log(list1);
  console.log(stuId);

  for (let i = 0; i < list1.length; i++) {
    if (Number(list1[i].stuId) === Number(stuId)) {
      isLike.value = true
    }
  }
  const {
    data: {
      data: { list }
    }
  } = await getStudentById(studentId)
  //console.log(list[0])
  student.value = list[0]
  pdfURL.value = fileUrl(list[0].resume?.resume)
  //console.log(pdfURL.value)
  allLanguage.value = list[0].languages.map(item => item.language + item.level)
  allSpeciality.value = list[0].specialities.map(item => item.speciality)

})

</script>
<template>
  <div style="width:100%; min-width:max-content;">
    <el-menu class="fixed-menu" mode="horizontal">
      <div class="menu-left">
        <IconSchool />
        <h1>
          <p>学生招聘平台</p>
        </h1>
      </div>

      <div class="menu-right">
        <el-menu-item @click="router.go(-1)" index="1">
          返回
          <el-icon>
            <ArrowRightBold />
          </el-icon>
        </el-menu-item>
      </div>
    </el-menu>
    <div class="content-area">
      <el-splitter v-if="student">
        <el-splitter-panel size="30%" :resizable="false">
          <div class="demo-panel">
            <h1 style="font-size: 60px; margin-bottom: 0px">
              {{ student.name }}
            </h1>
            <h4>{{ student.graduationYear - 4 }}-{{ student.graduationYear }}</h4>
            <div class="s-line">
              <h4>学号:</h4>
              <span>{{ student.stuId }}</span>
            </div>
            <div class="s-line">
              <h4>性别:</h4>
              <span>{{ student.gender }}</span>
            </div>
            <div class="s-line">
              <h4>年龄:</h4>
              <span>{{ student.age }}</span>
            </div>
            <div class="s-line">
              <h4>民族:</h4>
              <span>{{ student.nation }}</span>
            </div>
            <div class="s-line">
              <h4>专业:</h4>
              <span>{{ student.major }}</span>
            </div>
            <div class="s-line">
              <h4>学位:</h4>
              <span>{{ student.degree }}</span>
            </div>
            <div class="s-line">
              <h4>毕业年份</h4>
              <span>{{ student.graduationYear }}</span>
            </div>
            <div class="s-line">
              <h4>实习次数:</h4>
              <span>{{ student.internship }}</span>
            </div>
            <div class="s-line" v-if="student.internships.length > 0">
              <div class="skill">
                <h4>实习经历</h4>
                <div class="tag-box">
                  <el-tag v-for="item in student.internships" :key="item" size="large" type="warning" class="tag">
                    {{ item.company }}-{{ item.experience }} </el-tag>
                </div>
              </div>
            </div>
            <div class="s-line" v-if="allSpeciality.length > 0">
              <div class="skill">
                <h4>技能水平</h4>
                <div class="tag-box">
                  <el-tag v-for="item in allSpeciality" :key="item" size="large" type="primary" class="tag"> {{ item }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="s-line" v-if="allLanguage.length > 0">
              <div class="skill">
                <h4>语言水平</h4>
                <div class="tag-box">
                  <el-tag v-for="item in allLanguage" :key="item" size="large" type="info" class="tag"> {{ item }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="s-line" v-if="student.certificates.length > 0">
              <div class="skill">
                <h4>获奖证书</h4>
                <div class="tag-box">
                  <el-tag v-for="item in student.certificates" :key="item" size="large" type="success" class="tag">
                    {{ item.certName }}-{{ item.level }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="s-line" v-if="student.region?.city">
              <div class="skill">
                <h4>意向工作城市</h4>
                <div class="tag-box">
                  <el-tag size="large" type="danger" class="tag">
                    {{ student.region.city }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="s-line" v-if="student.jobs.length > 0">
              <div class="skill">
                <h4>意向工作类别</h4>
                <div class="tag-box">
                  <el-tag v-for="item in student.jobs" :key="item" size="large" type="warning" class="tag">
                    {{ item.post }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="s-line">
              <el-button style="width:340px" type="success" v-if="!isLike" @click="toggleLike">
                收藏简历
              </el-button>
              <el-button style="width:340px" type="info" v-else @click="toggleLike">
                已收藏
              </el-button>
            </div>
          </div>
        </el-splitter-panel>
        <el-splitter-panel>
          <div class="pdf-viewer" v-if="student.resume">
            <embed :src="pdfURL" type="application/pdf" width="100%" height="1250" />
          </div>
          <div style="padding:100px" v-else>
            <el-empty :image-size="250" description="暂无简历信息" />
          </div>
        </el-splitter-panel>
      </el-splitter>
      <div style="padding:100px" v-else>
        <el-empty :image-size="250" description="暂无学生信息,请报告给管理员" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.fixed-menu {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 80px;
  background: var(--el-bg-color);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  user-select: none;
}

.menu-left {
  display: flex;
  align-items: center;
}

.menu-right {
  display: flex;
  align-items: center;
}

.content-area {
  margin-top: 80px;
  min-height: calc(100vh - 80px);
}

body {
  background-color: var(--el-bg-color-page);
}

.demo-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  height: 100%;

  .s-line {
    font-size: 20px;
    display: flex;
    justify-content: space-between;
    width: 65%;
    padding: 15px;
    border-bottom: 2px solid var(--el-border-color);
  }
}

h4 {
  margin: 0;
}

.skill {
  font-size: 20px;

  h4 {
    margin: 5px 0;
  }

  .tag-box {
    display: flex;
    justify-content: flex-start;
    gap: 5px;
    flex-wrap: wrap;
  }
}

.pdf-viewer {
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}
</style>
