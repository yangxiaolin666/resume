<script setup>
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { onMounted, ref } from 'vue'
import { addLike, deleteLike, queryComId, queryLike } from '@/api/student';
import { useUserStore } from '@/stores/user';
const router = useRouter()
const { item } = defineProps({
  item: {
    type: Object,
    required: true
  }
})
const isLike = ref(false)
const userStore=useUserStore()
const com_id=ref(5)
const stuId=item.id
onMounted(async ()=>{
  const res=await queryComId()
  com_id.value=res.data.data
  IdList.value=[
  {
    "comId": com_id.value,
    "stuId": stuId,
  }
  ]
  const {data:{data:{list}}}=await queryLike({
    "com_id":com_id.value,
    "pageNum":pageNum,
    "pageSize":pageSize
  })
  for(let i=0;i<list.length;i++){
    if(list[i].stuId===stuId){
      isLike.value=true
    }
  }
})
const IdList=ref([
  {
    "comId": com_id.value,
    "stuId": stuId,
  }
])
const pageNum=1
const pageSize=10

const toggleLike = async () => {
  if (isLike.value) {
    const res=await deleteLike(IdList.value)
    ElMessage({
      message:'取消收藏成功',
      type:'success',
      duration:500
    })
    console.log(res)
  } else {
    const res=await addLike(IdList.value)
    ElMessage({
      message:'收藏成功',
      type:'success',
      duration:500
    })
    console.log(res)
  }
  isLike.value = !isLike.value
  if(item.isRefresh){
    setTimeout(()=>{
      router.go(0)
    },600)
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
      <span>学位</span>
      <span>{{ item.degree }}</span>
    </div>
    <div class="text item">
      <span>专业</span>
      <span>{{ item.major }}</span>
    </div>
    
    <div class="skill" >
      <h4>能力标签</h4>
      <div class="tag-box" v-if="item.languages?.length||item.specialities?.length">
        <el-tag v-for="(item, index) in item.specialities" :key="index" size="small" type="primary" class="tag">
          {{ item.speciality }}
        </el-tag>
        <el-tag v-for="(item, index) in item.languages" :key="index" size="small" type="warning" class="tag">
          {{ item.language }}-{{ item.level }}
        </el-tag>
      </div>
      <span v-else>暂无标签</span>
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
        <el-button type="success" v-if="!isLike" @click="toggleLike" >
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
  min-height: 380px;
  max-height: 380px;
  min-width:280px;
  max-width: 280px;
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
  margin: 8px;
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
