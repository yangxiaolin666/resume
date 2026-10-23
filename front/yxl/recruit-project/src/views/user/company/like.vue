<script setup>
import studentCard from '@/components/studentCard.vue'
import { ref, onMounted } from 'vue';
import { queryComId, queryLike } from '@/api/student';
import { useUserStore } from '@/stores/user';
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(10)
const handleCurrentChange = async (val) => {
  console.log(`current page: ${val}`)
  const { data: { data } } = await queryLike({
    "com_id": com_id.value,
    "pageNum": currentPage.value,
    "pageSize": pageSize.value
  })
  //console.log(data)
  total.value = data.total
  students.value = data.list
}
const com_id = ref(0)
const students = ref()
onMounted(async () => {
  const res = await queryComId()
  com_id.value = res.data.data
  console.log(com_id.value)
  const { data: { data } } = await queryLike({
    "com_id": com_id.value,
    "pageNum": currentPage.value,
    "pageSize": pageSize.value
  })

  //console.log(data)
  total.value = data.total
  students.value = data.list
})

</script>
<template>
  <div class="like-container">
    <main class="like-content">
      <div v-if="total > 0" class="student-list">
        <div v-for="item in students" :key="item.id">
          <studentCard :item="{
            isRefresh: true,
            id: item.stuId,
            userId: item.userId,
            name: item.studentName,
            graduationYear: item.graduationYear,
            gender: item.gender,
            major: item.major,
            degree: item.degree,
            specialities: item.specialities?.split('、').map(item => ({ speciality: item })),
            languages: item.languages?.split('、').map((lang, index) => ({
              language: lang,
              level: item.level?.split('、')[index] || ''
            })),
            jobs: item.post?.split('、').map(item => ({ post: item }))
          }" />
        </div>
      </div>
      <el-empty v-else :image-size="250" description="暂无收藏简历" />
    </main>
  </div>
  <div class="demo-pb" v-if="total > 0">
    <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" size="large" :background="true"
      layout="prev, pager, next, jumper" :total="total" @current-change="handleCurrentChange" />
  </div>
</template>
<style scoped lang="scss">
.like-container {
  min-height: calc(100vh - 80px);
  background: var(--el-bg-color);
}

.like-content {
  flex: 1;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  background: var(--el-bg-color);
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);

  .page-title {
    margin: 0;
    font-size: 20px;
    color: #303133;
    font-weight: 600;
  }
}

.student-list {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 0px;
  align-items: start;
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
</style>
