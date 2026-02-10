<script setup>
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { ref } from 'vue'

const router = useRouter()
const { item } = defineProps({
  item: {
    type: Object,
    required: true
  }
})
const isLike = ref(!item.isLike || true)
const allLanguage = ref(item.languages.map(item => item.language+item.level))
const allSpeciality = ref(item.specialities.map(item => item.speciality))
const toggleLike = () => {
  isLike.value = !isLike.value
  if (isLike.value) {
    ElMessage({
      message: '已取消收藏',
      type: 'info'
    })
  } else {
    ElMessage({
      message: '收藏成功',
      type: 'success'
    })
  }
}
</script>
<template>
  <el-card class="box-card">
    <template #header>
      <div class="card-header">
        <h3>{{ item.name }}</h3>
        <el-tag type="success"> {{ item.graduationYear }}届 </el-tag>
      </div>
    </template>
    <div class="text item">
      <span>性别</span>
      <span>{{ item.gender }}</span>
    </div>
    <div class="text item">
      <span>专业</span>
      <span>{{ item.major }}</span>
    </div>
    <div class="text item">
      <span>学位</span>
      <span>{{ item.degree }}</span>
    </div>
    <div class="skill">
      <h4>技能水平</h4>
      <div class="tag-box">
        <el-tag v-for="(speciality, index) in allSpeciality" :key="index" size="small" type="primary" class="tag">
          {{ speciality }}
        </el-tag>
        <el-tag v-for="(language, index) in allLanguage" :key="index" size="small" type="warning" class="tag">
          {{ language }}
        </el-tag>
      </div>
    </div>
    <template #footer>
      <div style="display: flex; justify-content: space-between">
        <el-button type="primary" @click="
          router.push({
            path: '/studentDetail',
            query: {
              id: item.id
            }
          })
          ">
          查看简历
        </el-button>
        <el-button type="success" v-if="isLike" @click="toggleLike" >
          收藏简历
        </el-button>
        <el-button type="info" v-else @click="toggleLike">
          已收藏
        </el-button>
      </div>
    </template>
  </el-card>
</template>
<style scoped lang="scss">
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  h3{
    margin: 3px;
  }
}

.box-card {
  min-height: 420px;
  min-width:300px;
}

.text {
  font-size: 18px;
}

.item {
  display: flex;
  justify-content: space-between;
  padding: 10px;
  border-bottom: 1px solid var(--el-border-color);
}

.box-card {
  margin: 10px;
  border-radius: 20px;
}

.skill {
  padding: 5px;
  font-size: 15px;

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
