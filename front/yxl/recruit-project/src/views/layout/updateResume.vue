<script setup>
/* eslint-disable prettier/prettier */
import { ref, watch } from 'vue'
import { ArrowRightBold } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
const router = useRouter()
const gender = ref('')
const name = ref('')
const age = ref('')
const tel = ref('')
const major = ref('')
const degree=ref('')
watch(degree,(newV,oldV)=>{
  console.log(newV[0]);
  
})
const degreeOptions=[
  {
    value:'本科',
    label:'本科'
  },
  {
    value:'研究生',
    label:'研究生'
  }
]
const SmajorOptions=[
  {
    value:'计算机科学与技术',
    label:'计算机科学与技术'
  },
  {
    value:'软件工程',
    label:'软件工程'
  },
  {
    value:'人工智能',
    label:'人工智能'
  },
  {
    value:'网络工程',
    label:'网络工程'
  }
]
const BmajorOptions=[
  {
    value:'计算机技术',
    label:'计算机技术'
  },
  {
    value:'人工智能',
    label:'人工智能'
  }
]
const stuId=ref('')
const graduationYear=ref('')
const internship=ref('')
const nation=ref('')
const selfIntroduce=ref('')
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
      <el-menu-item @click="router.go(-1)">
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
          <el-input v-model="name" maxlength="15" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 50px" placeholder="请输入姓名" />
        </h1>
        <div class="s-line">
          <h4>性别:</h4>
          <el-input v-model="gender" maxlength="2" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入性别" />
        </div>
        <div class="s-line">
          <h4>年龄:</h4>
          <el-input v-model="age" maxlength="3" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入年龄" />
        </div>
        <div class="s-line">
          <h4>联系方式:</h4>
          <el-input v-model="tel" maxlength="11" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入联系方式" />
        </div>
        <div class="s-line">
          <h4>学位:</h4>
          <el-cascader v-model="degree" :options="degreeOptions"
           style="width: 150px;height: 32px" placeholder="请选择学位"
           popper-style="max-height:80px;" />
        </div>
        <div class="s-line">
          <h4>专业:</h4>
          <el-cascader v-model="major" :options="degree ? (degree[0] === '本科' ? SmajorOptions : BmajorOptions) : []"
           style="width: 150px;height: 32px" placeholder="请选择专业"
           popper-style="max-height:150px;">
          <template #empty>
            <span style="font-size: 15px;margin-bottom: 55px;">请先选择学位</span>
          </template>
          </el-cascader>
        </div>
        <div class="s-line">
          <h4>学号:</h4>
          <el-input v-model="stuId" maxlength="15" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入学号" />
        </div>
        <div class="s-line">
          <h4>毕业年份</h4>
          <el-input v-model="graduationYear" maxlength="5" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入毕业年份" />
        </div>
        <div class="s-line">
          <h4>实习次数:</h4>
          <el-input v-model="internship" maxlength="10" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入实习次数" />
        </div>
        <div class="s-line">
          <h4>民族:</h4>
          <el-input v-model="nation" maxlength="10" show-word-limit word-limit-position="outside"
            style="width: 150px; height: 32px" placeholder="请输入民族" />
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>技能水平</h4>
            <div class="tag-box">
              <el-tag size="large" type="primary" class="tag">
                Java
              </el-tag>
              <el-tag size="large" type="primary" class="tag">
                C++
              </el-tag>
              <el-tag size="large" type="primary" class="tag">
                Python
              </el-tag>
              <el-tag size="large" type="primary" class="tag">
                GoLang
              </el-tag>
              <el-tag size="large" type="primary" class="tag">
                Redis
              </el-tag>
              <el-tag size="large" type="primary" class="tag">
                Java
              </el-tag>
              <el-tag size="large" type="primary" class="tag">
                C++
              </el-tag>
              <el-tag size="large" type="primary" class="tag">
                Python
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>语言水平</h4>
            <div class="tag-box">
              <el-tag size="large" type="warning" class="tag">
                英语
              </el-tag>
              <el-tag size="large" type="warning" class="tag">
                日语
              </el-tag>
              <el-tag size="large" type="warning" class="tag">
                德语
              </el-tag>
              <el-tag size="large" type="warning" class="tag">
                汉语
              </el-tag>
              <el-tag size="large" type="warning" class="tag">
                西班牙语
              </el-tag>
              <el-tag size="large" type="warning" class="tag">
                德语
              </el-tag>
              <el-tag size="large" type="warning" class="tag">
                汉语
              </el-tag>
              <el-tag size="large" type="warning" class="tag">
                西班牙语
              </el-tag>
            </div>
          </div>
        </div>
        <div class="s-line">
          <div class="skill">
            <h4>个人简介</h4>
            <span style="font-size: 15px">
            <el-input
              v-model="selfIntroduce"
              class="ml-4"
              maxlength="100"
              style="width: 320px;"
              placeholder="介绍一下自己吧:"
              show-word-limit
              word-limit-position="outside"
              type="textarea"
              :rows="9"
            />
            </span>
          </div>
        </div>
      </div>
    </el-splitter-panel>
    <el-splitter-panel>
      <div class="demo-panel">
        <img src="@/assents/简历(3)_01.png" alt="">
      </div>
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
</style>
