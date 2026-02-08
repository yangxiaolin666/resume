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
onMounted(async () => {
  const {
    data: {
      data: { list }
    }
  } = await getStudentById(studentId)
  console.log(list[0])
  student.value = list[0]
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
      <el-menu-item @click="router.go(-1)">
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
        <h4>2024-2028</h4>
        <div class="s-line">
          <h4>性别:</h4>
          <span>{{ student.gender }}</span>
        </div>
        <div class="s-line">
          <h4>年龄:</h4>
          <span>{{ student.age }}</span>
        </div>
        <div class="s-line">
          <h4>联系方式:</h4>
          <span>130xxxxxxxxx</span>
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
          <h4>学号:</h4>
          <span>{{ student.stuId }}</span>
        </div>
        <div class="s-line">
          <h4>毕业年份</h4>
          <span>{{ student.graduationYear }}</span>
        </div>
        <div class="s-line">
          <h4>实习次数:</h4>
          <span>{{ student.internship }}</span>
        </div>
        <div class="s-line">
          <h4>民族:</h4>
          <span>{{ student.nation }}</span>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>技能水平</h4>
            <div class="tag-box">
              <el-tag size="large" type="primary" class="tag"> Java </el-tag>
              <el-tag size="large" type="primary" class="tag"> C++ </el-tag>
              <el-tag size="large" type="primary" class="tag"> Python </el-tag>
              <el-tag size="large" type="primary" class="tag"> GoLang </el-tag>
              <el-tag size="large" type="primary" class="tag"> Redis </el-tag>
              <el-tag size="large" type="primary" class="tag"> Java </el-tag>
              <el-tag size="large" type="primary" class="tag"> C++ </el-tag>
              <el-tag size="large" type="primary" class="tag"> Python </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>语言水平</h4>
            <div class="tag-box">
              <el-tag size="large" type="warning" class="tag"> 英语 </el-tag>
              <el-tag size="large" type="warning" class="tag"> 日语 </el-tag>
              <el-tag size="large" type="warning" class="tag"> 德语 </el-tag>
              <el-tag size="large" type="warning" class="tag"> 汉语 </el-tag>
              <el-tag size="large" type="warning" class="tag"> 西班牙语 </el-tag>
              <el-tag size="large" type="warning" class="tag"> 德语 </el-tag>
              <el-tag size="large" type="warning" class="tag"> 汉语 </el-tag>
              <el-tag size="large" type="warning" class="tag"> 西班牙语 </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>个人简介</h4>
            <span style="font-size: 15px">
              性格开朗、执行力强，热衷技术钻研，具备扎实的后端基础与项目实战
              经验。在校担任工作室负责人，负责管理工作室团队的算法培训以及算
              法招新，具备良好的沟通与团队协作能力，能够快速适应团队与业务，
              独立承担模块设计与开发任务，有良好的合作能力以及公司开发任务的
              上手迅速
            </span>
          </div>
        </div>
      </div>
    </el-splitter-panel>
    <el-splitter-panel>
      <div class="demo-panel">
        <img src="@/assents/简历(3)_01.png" alt="" />
      </div>
    </el-splitter-panel>
  </el-splitter>
  <div style="font-size: 50px;text-align:center;padding:200px" v-else>
    暂无学生信息
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
    padding: 20px;
    border-bottom: 2px solid var(--el-border-color);
  }
}

h4 {
  margin: 0;
}

.skill {
  font-size: 20px;

  h4 {
    margin: 10px 0;
  }

  .tag-box {
    display: flex;
    justify-content: flex-start;
    gap: 10px;
    flex-wrap: wrap;
  }
}
</style>
