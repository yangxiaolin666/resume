<script setup>
/* eslint-disable prettier/prettier */
import { onMounted, ref, computed, watch } from 'vue'
import { ArrowRightBold, UploadFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { getStudentById, updateResume, completeInfo } from '@/api/student'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { genFileId } from 'element-plus'
const userStore = useUserStore()
const userId = userStore.userId
const router = useRouter()
const all = ref({
  "age": 40,
  "gender": "男",
  "nation": "汉族",
  "major": "计算机科学与技术",
  "degree": "本科生",
  "graduationYear": 2027,
  "internship": 1,
  "languages": [
    {
      "language": "英语",
      "level": "四级"
    }
  ],
  "cities": [
    "大连", "北京"
  ],
  "provinces": [
    "辽宁", "广东"
  ],
  "certificates": [
    {
      "certName": "计算机",
      "level": "二级"
    }
  ],
  "internships": [
    {
      "company": "腾讯",
      "experience": "游戏后端开发"
    },
    {
      "company": "阿里",
      "experience": "游戏后端开发"
    },
  ],
  "specialities": [
    "Java", "C++", "Python"
  ],
  "jobs": [
    "后端开发", "前端开发"
  ]
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
const allSpeciality = ref([])
const languageInput = ref('')
const levelInput = ref('')
const certNameInput = ref('')
const certLevelInput = ref('')
const companyInput = ref('')
const experienceInput = ref('')
const skillNameInput = ref('')
const skillLevelInput = ref('')
const jobInput = ref('')
onMounted(async () => {
  const {
    data: {
      data: { list }
    }
  } = await getStudentById(userId)
  all.value = list[0]
  console.log(all.value);
  allSpeciality.value = all.value.specialities.map(item => item.speciality)

})

watch(allSpeciality, (newV) => {
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
const uploadFile = ref(null)
const handleFileChange = (file) => {
  uploadFile.value = file.raw
  ElMessage({
    message: '简历上传成功',
    type: 'success',
    duration: 1000
  })
}
const handleUpdate = async () => {
  if (uploadFile.value) {
    const res = await updateResume(uploadFile.value)
    console.log(res)
  }
}
const addLanguage = () => {
  if (!languageInput.value.trim() || !levelInput.value.trim()) {
    ElMessage({
      message: '请输入语言和等级',
      type: 'warning',
      duration: 1000
    })
    return
  }
  if (!all.value.languages) {
    all.value.languages = []
  }
  all.value.languages.push({
    language: languageInput.value.trim(),
    level: levelInput.value.trim()
  })
  console.log(all.value.languages);
  languageInput.value = ''
  levelInput.value = ''
  ElMessage({
    message: '语言添加成功',
    type: 'success',
    duration: 1000
  })
}
const removeLanguage = (index) => {
  all.value.languages.splice(index, 1)
}

const addCertificate = () => {
  if (!certNameInput.value.trim() || !certLevelInput.value.trim()) {
    ElMessage({
      message: '请输入证书名称和等级',
      type: 'warning',
      duration: 1000
    })
    return
  }
  if (!all.value.certificates) {
    all.value.certificates = []
  }
  all.value.certificates.push({
    certName: certNameInput.value.trim(),
    level: certLevelInput.value.trim()
  })
  console.log(all.value.certificates);

  certNameInput.value = ''
  certLevelInput.value = ''
  ElMessage({
    message: '证书添加成功',
    type: 'success',
    duration: 1000
  })
}

const removeCertificate = (index) => {
  all.value.certificates.splice(index, 1)
}

const addInternship = () => {
  if (!companyInput.value.trim() || !experienceInput.value.trim()) {
    ElMessage({
      message: '请输入公司名称和实习经历',
      type: 'warning',
      duration: 1000
    })
    return
  }
  if (!all.value.internships) {
    all.value.internships = []
  }
  all.value.internships.push({
    company: companyInput.value.trim(),
    experience: experienceInput.value.trim()
  })
  console.log(all.value.internships);
  companyInput.value = ''
  experienceInput.value = ''
  ElMessage({
    message: '实习经历添加成功',
    type: 'success',
    duration: 1000
  })
}

const removeInternship = (index) => {
  all.value.internships.splice(index, 1)
}

const addSkill = () => {
  if (!skillNameInput.value.trim()) {
    ElMessage({
      message: '请输入技能名称',
      type: 'warning',
      duration: 1000
    })
    return
  }
  if (!all.value.specialities) {
    all.value.specialities = []
  }
  all.value.specialities.push(
    skillNameInput.value.trim(),
  )
  skillNameInput.value = ''
  ElMessage({
    message: '技能添加成功',
    type: 'success',
    duration: 1000
  })
}

const removeSkill = (index) => {
  all.value.specialities.splice(index, 1)
}

const addJob = () => {
  if (!jobInput.value.trim()) {
    ElMessage({
      message: '请输入意向工作',
      type: 'warning',
      duration: 1000
    })
    return
  }
  if (!all.value.jobs) {
    all.value.jobs = []
  }
  all.value.jobs.push({
    post: jobInput.value.trim()
  })
  jobInput.value = ''
  ElMessage({
    message: '意向工作添加成功',
    type: 'success',
    duration: 1000
  })
}

const removeJob = (index) => {
  all.value.jobs.splice(index, 1)
}
const updateChange = async () => {
  if (Array.isArray(all.value.degree)) {
    all.value.degree = all.value.degree[0]
  }
  if (Array.isArray(all.value.major)) {
    all.value.major = all.value.major[0]
  }
  // if(!uploadFile.value){
  //   ElMessage({
  //     message:'请上传简历',
  //     type:'error',
  //     duration:1000
  //   })
  //   return;
  // }
  const dataObj = {
    "age": all.value.age,
    "gender": all.value.gender,
    "nation": all.value.nation,
    "major": all.value.major,
    "degree": all.value.degree,
    "graduationYear": all.value.graduationYear,
    "internship": all.value.internship,
    "languages": all.value.languages,
    "cities": ['北京', '上海'],
    "provinces": ['广东', '河南'],
    "certificates": all.value.certificates,
    "internships": all.value.internships,
    "specialities": all.value.specialities,
    "jobs": all.value.jobs.map(item => item.post)
  }
  console.log(all.value);
  const res = await completeInfo(dataObj)
  await handleUpdate()
  console.log(res.data.data)
  ElMessage({
    message: '修改成功',
    type: 'success',
    duration: 1000
  })
  setTimeout(() => router.push('/selfMessage'), 1100)
}
</script>
<template>
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
          <h4>学号:</h4>
          <el-input v-model="all.stuId" maxlength="12" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入学号" />
        </div>
        <div class="s-line">
          <h4>学位:</h4>
          <el-cascader v-model="all.degree" :options="degreeOptions" style="width: 150px;height: 32px"
            placeholder="请选择学位" popper-style="max-height:80px;" />
        </div>
        <div class="s-line">
          <h4>专业:</h4>
          <el-cascader v-model="all.major"
            :options="all.degree ? (all.degree[0] === '本科生' || all.degree === '本科生' ? SmajorOptions : BmajorOptions) : []"
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
            <div style="display: flex; align-items: center; 
            gap: 5px; margin-bottom: 5px;">
              <el-input v-model="skillNameInput" placeholder="请输入技能名称(例:Java)" size="default" style="width: 305px;" />
              <el-button type="success" size="small" @click="addSkill">+</el-button>
            </div>
            <div class="tag-box">
              <el-tag v-for="(skill, index) in all.specialities" :key="index" size="small" type="primary" effect="plain"
                closable @close="removeSkill(index)" class="tag">
                {{ skill }}
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>意向工作</h4>
            <div style="display: flex; align-items: center; 
            gap: 5px; margin-bottom: 5px;">
              <el-input v-model="jobInput" placeholder="请输入意向工作(例:后端开发)" size="default" style="width: 305px;" />
              <el-button type="success" size="small" @click="addJob">+</el-button>
            </div>
            <div class="tag-box">
              <el-tag v-for="(job, index) in all.jobs" :key="index" size="small" type="danger" effect="plain" closable
                @close="removeJob(index)" class="tag">
                {{ job.post }}
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>语言水平</h4>
            <div style="display: flex; align-items: center; 
            gap: 5px; margin-bottom: 5px;">
              <el-input v-model="languageInput" placeholder="请输入语言(例:英语)" size="default" style="width: 150px;" />
              <el-input v-model="levelInput" placeholder="请输入等级(例:四级)" size="default" style="width: 150px;" />
              <el-button type="success" size="small" @click="addLanguage">+</el-button>
            </div>
            <div class="tag-box">
              <el-tag v-for="(lang, index) in all.languages" :key="index" size="small" type="warning" effect="plain"
                closable @close="removeLanguage(index)" class="tag">
                {{ lang.language }}-{{ lang.level }}
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>证书信息</h4>
            <div style="display: flex; align-items: center; 
            gap: 5px; margin-bottom: 5px;">
              <el-input v-model="certNameInput" placeholder="请输入证书名称(例:计算机)" size="default" style="width: 150px;" />
              <el-input v-model="certLevelInput" placeholder="请输入等级(例:二级)" size="default" style="width: 150px;" />
              <el-button type="success" size="small" @click="addCertificate">+</el-button>
            </div>
            <div class="tag-box">
              <el-tag v-for="(cert, index) in all.certificates" :key="index" size="small" type="success" effect="plain"
                closable @close="removeCertificate(index)" class="tag">
                {{ cert.certName }}-{{ cert.level }}
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>实习经历</h4>
            <div style="display: flex; align-items: center; 
            gap: 5px; margin-bottom: 5px;">
              <el-input v-model="companyInput" placeholder="请输入公司名称(例:阿里)" size="default" style="width: 150px;" />
              <el-input v-model="experienceInput" placeholder="请输入实习经历(例:游戏后端开发)" size="default"
                style="width: 150px;" />
              <el-button type="success" size="small" @click="addInternship">+</el-button>
            </div>
            <div class="tag-box">
              <el-tag v-for="(internship, index) in all.internships" :key="index" size="small" type="danger"
                effect="plain" closable @close="removeInternship(index)" class="tag">
                {{ internship.company }}-{{ internship.experience }}
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <span style="font-size: 15px">
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
      <el-upload ref="upload" class="upload-demo" drag show-file-list :limit="1" :on-change="handleFileChange"
        :auto-upload="false" :on-exceed="handleExceed"
        style="width: 80%;margin-top:100px;margin-left:100px;height: 100%;">
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
    width: 75%;
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