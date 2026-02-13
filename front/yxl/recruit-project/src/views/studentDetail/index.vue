<script setup>
import { getStudentById } from '@/api/student'
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
const pdfURL=ref()
onMounted(async () => {
  const {
    data: {
      data: { list }
    }
  } = await getStudentById(studentId)
  console.log(list[0])
  student.value = list[0]
  pdfURL.value="http://81.70.199.174:8081"+list[0].resume?.resume
  console.log(pdfURL.value)
  allLanguage.value = list[0].languages.map(item => item.language+item.level)
  allSpeciality.value = list[0].specialities.map(item => item.speciality)
})

</script>
<template>
  <el-menu class="el-menu-demo" mode="horizontal" style="
      display: flex;
      justify-content: space-between;
      height: 80px;
      user-select: none;
    ">
    <div style="display: flex; align-items: center">
      <IconSchool />
      <h1>
        <p>学生招聘平台</p>
      </h1>
    </div>

    <div style="display: flex; align-items: center">
      <el-menu-item @click="router.go(-1)" index="1">
        返回
        <el-icon>
          <ArrowRightBold />
        </el-icon>
      </el-menu-item>
    </div>
  </el-menu>
  <el-splitter v-if="student">
    <el-splitter-panel size="30%" :resizable="false">
      <div class="demo-panel">
        <h1 style="font-size: 60px; margin-bottom: 0px">
          {{ student.name }}
        </h1>
        <h4>{{student.graduationYear-4}}-{{student.graduationYear}}</h4>
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
        <div class="s-line" v-if="student.internships.length>0">
          <div class="skill">
            <h4>实习经历</h4>
            <div class="tag-box">
              <el-tag v-for="item in student.internships" :key="item"
               size="large" type="warning" class="tag"> 
                {{item.company}}-{{ item.experience }} </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line" v-if="allSpeciality.length>0">
          <div class="skill">
            <h4>技能水平</h4>
            <div class="tag-box">
              <el-tag v-for="item in allSpeciality" :key="item"
               size="large" type="primary" class="tag"> {{item}} </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line" v-if="allLanguage.length>0">
          <div class="skill">
            <h4>语言水平</h4>
            <div class="tag-box">
              <el-tag v-for="item in allLanguage" :key="item" 
              size="large" type="info" class="tag"> {{item}} </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line" v-if="student.certificates.length>0">
          <div class="skill">
            <h4>获奖证书</h4>
            <div class="tag-box">
              <el-tag v-for="item in student.certificates"
               :key="item" size="large" type="success" class="tag"> 
                {{item.certName}}-{{item.level}} 
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line" v-if="student.region?.city">
          <div class="skill">
            <h4>意向工作城市</h4>
            <div class="tag-box">
              <el-tag size="large" type="danger" class="tag"> 
                {{student.region.city}}
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line" v-if="student.jobs.length>0">
          <div class="skill">
            <h4>意向工作类别</h4>
            <div class="tag-box">
              <el-tag v-for="item in student.jobs" :key="item" 
              size="large" type="warning" class="tag"> 
                {{item.post}}
              </el-tag>
            </div>
          </div>
        </div>
      </div>
    </el-splitter-panel>
    <el-splitter-panel>
      <div class="pdf-viewer" v-if="student.resume">
        <embed 
          :src="pdfURL"
          type="application/pdf" 
          width="100%" 
          height="1250" />
      </div>
      <div style="padding:100px" v-else>
        <el-empty image-size="250" description="暂无简历信息" />
      </div>
    </el-splitter-panel>
  </el-splitter>
  <div style="padding:100px" v-else>
    <el-empty image-size="250" description="暂无学生信息,请报告给管理员" />
  </div>
</template>

<style scoped lang="scss">
body{
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
