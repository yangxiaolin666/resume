<script setup>
/* eslint-disable prettier/prettier */
import { onMounted, ref, computed, watch } from 'vue'
import { ArrowRightBold, UploadFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { getStudentById, updateResume, completeInfo } from '@/api/student'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { genFileId } from 'element-plus'
const userStore=useUserStore()
const userId=userStore.userId
const router = useRouter()
const all = ref({
  gender: '',
  name: '',
  age: '',
  tel: '130xxxxxxxx',
  degree: '',
  stuId: '',
  graduationYear: '',
  internship: '',
  nation: '',
  major: '',
  selfIntroduce: '',
  //languages: [],
  specialities: [],
})

const degreeOptions = [
  {
    value: '本科生',
    label: '本科生'
  },
  {
    value: '研究生',
    label: '研究生'
  }
]
const SmajorOptions = [
  {
    value: '计算机科学与技术',
    label: '计算机科学与技术'
  },
  {
    value: '软件工程',
    label: '软件工程'
  },
  {
    value: '人工智能',
    label: '人工智能'
  },
  {
    value: '网络工程',
    label: '网络工程'
  }
]
const BmajorOptions = [
  {
    value: '计算机技术',
    label: '计算机技术'
  },
  {
    value: '人工智能',
    label: '人工智能'
  }
]
const allLanguage = ref([])
const allSpeciality = ref([])
onMounted(async () => {
  const {
    data: {
      data: { list }
    }
  } = await getStudentById(userId)
  //console.log(list[0])
  Object.assign(all.value, list[0], {
    tel: '130xxxxxxxx'
  })
  Object.keys(list[0]).forEach(key => {
    all.value[key] = list[0][key]
  })
  all.value.tel = '130xxxxxxxx'
  allLanguage.value = all.value.languages.map(item => item.language+item.level)
  allSpeciality.value = all.value.specialities.map(item => item.speciality)
})

watch(allSpeciality,(newV)=>{
  all.value.specialities = newV
  console.log(newV)
})
const upload = ref()
const handleExceed = (files) => {
  upload.value.clearFiles()
  const file = files[0]
  file.uid = genFileId()
  upload.value.handleStart(file)
}
const uploadFile=ref(null)
const handleFileChange=(file)=>{
  uploadFile.value=file.raw
  ElMessage({
    message: '简历上传成功',
    type: 'success',
    duration: 1000
  })
}
const handleUpdate = async () => {
  if(uploadFile.value){
    const res = await updateResume(uploadFile.value)
    console.log(res)
  }
}
const updateChange = async () => {
  if(Array.isArray(all.value.degree)){
    all.value.degree=all.value.degree[0]
  }
  if(Array.isArray(all.value.major)){
    all.value.major=all.value.major[0]
  }
  // if(!uploadFile.value){
  //   ElMessage({
  //     message:'请上传简历',
  //     type:'error',
  //     duration:1000
  //   })
  //   return;
  // }
  const res = await completeInfo(all.value)
  await handleUpdate()
  console.log(res)
  ElMessage({
    message: '修改成功',
    type: 'success',
    duration: 1000
  })
  setTimeout(() => router.push('/home'), 1100)
}


</script>
<template>
  <el-menu :router="true" class="el-menu-demo" mode="horizontal" style="
      display: flex;
      justify-content: space-between;
      height: 80px;
      user-select: none;
    ">
    <div style="display: flex; align-items: center">
      <IconSchool />
      <h1>
        <p>简历修改</p>
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
  <el-splitter>
    <el-splitter-panel size="30%" :resizable="false">
      <div class="demo-panel">
        <h1 style="font-size: 60px; margin-bottom: 0px">
          {{ all.name }}
        </h1>
        <div class="s-line">
          <h4>性别:</h4>
          <el-input v-model="all.gender" maxlength="2" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入性别" />
        </div>
        <div class="s-line">
          <h4>年龄:</h4>
          <el-input v-model="all.age" maxlength="3" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入年龄" />
        </div>
        <div class="s-line">
          <h4>民族:</h4>
          <el-input v-model="all.nation" maxlength="10" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入民族" />
        </div>
        <div class="s-line">
          <h4>学位:</h4>
          <el-cascader v-model="all.degree" :options="degreeOptions" style="width: 150px;height: 32px"
            placeholder="请选择学位" popper-style="max-height:80px;" />
        </div>
        <div class="s-line">
          <h4>专业:</h4>
          <el-cascader v-model="all.major"
            :options="all.degree ? (all.degree[0] === '本科生'||all.degree==='本科生' ? SmajorOptions : BmajorOptions) : []"
            style="width: 150px;height: 32px" placeholder="请选择专业" popper-style="max-height:150px;">
            <template #empty>
              <span style="font-size: 15px;margin-bottom: 55px;">请先选择学位</span>
            </template>
          </el-cascader>
        </div>
        <div class="s-line">
          <h4>毕业年份</h4>
          <el-input v-model="all.graduationYear" maxlength="5" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入毕业年份" />
        </div>
        <div class="s-line">
          <h4>实习次数:</h4>
          <el-input v-model="all.internship" maxlength="10" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入实习次数" />
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>技能水平</h4>
            <div class="tag-box">
              <el-input-tag v-model="allSpeciality" clearable draggable placeholder="请输入标签,按下Enter确认(例:Java,C++)"
                tag-type="primary" size="large" style="width:320px;" />
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>语言水平</h4>
            <el-input-tag v-model="allLanguage" clearable draggable placeholder="请输入标签,按下Enter确认(例:英语四级,日语N2)"
              tag-type="warning" size="large" style="width:320px;" />
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>个人简介</h4>
            <span style="font-size: 15px">
              <el-input v-model="all.selfIntroduce" class="ml-4" maxlength="100" style="width: 320px;"
                placeholder="介绍一下自己吧:" show-word-limit word-limit-position="outside" type="textarea" :rows="8" />
              <el-button @click="updateChange" type="success" size="large"
                style="width:320px;margin-top: 30px;">提交修改</el-button>
            </span>
          </div>
        </div>
      </div>
    </el-splitter-panel>
    <el-splitter-panel>
      <!-- <div class="demo-panel">
        <img src="@/assents/简历(3)_01.png" alt="">
      </div> -->
      <el-upload ref="upload" class="upload-demo" drag
       show-file-list :limit="1"
      :on-change="handleFileChange" :auto-upload="false"
        :on-exceed="handleExceed" style="width: 80%;margin-top:100px;margin-left:100px;height: 100%;">
        <el-icon class="el-icon--upload" style="font-size: 200px;">
          <upload-filled /></el-icon>
        <div class="el-upload__text" style="height:100px;font-size: 30px;">
          将简历文件拖拽至此上传或 <em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip" style="margin-right: 100px;">
            <p>可提交格式:PDF、DOC、DOCX</p>
            <p>限制提交一个文件,新文件将覆盖旧文件</p>
          </div>
        </template>
      </el-upload>

    </el-splitter-panel>
  </el-splitter>
</template>

<style scoped lang="scss">
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

.right-box {
  display: flex;
  justify-content: space-between;

  .updateBtn {
    display: flex;
    flex-direction: column;

    .el-button {
      width: 200px;
      height: 200px;
      font-size: 100px;
    }

    span {
      font-size: 30px;
    }
  }
}
</style>
