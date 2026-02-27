<script setup>
import { useStudentsStore } from '@/stores/students'
import studentCard from '@/components/studentCard.vue'
import { ref,onMounted } from 'vue';
import { queryLike } from '@/api/student';
const currentPage = ref(1)
const pageSize = ref(10)
const total=ref(10)
const handleCurrentChange = async (val) => {
  console.log(`current page: ${val}`)
  const {data:{data}}=await queryLike({
    "com_id":com_id,
    "pageNum":currentPage.value,
    "pageSize":pageSize.value
  })
  console.log(data)
  total.value=data.total
  students.value=data.list
}
const StudentsStore = useStudentsStore()
const com_id=5
const students=ref()
onMounted(async ()=>{
  const {data:{data}}=await queryLike({
    "com_id":com_id,
    "pageNum":currentPage.value,
    "pageSize":pageSize.value
  })
  console.log(data)
  total.value=data.total
  students.value=data.list
})

</script>
<template>
  <div style="margin:10px 0px;padding-bottom: 60px;" v-if="total>0">
    <div style="display: grid; grid-template-columns: repeat(5, 1fr);">
      <div v-for="item in students" :key="item.id">
        <studentCard :item="{
          isRefresh:true,
          id:item.stuId,
          name:item.studentName,
          graduationYear:item.graduationYear,
          gender:item.gender,
          major:item.major,
          degree:item.degree,
          specialities:item.specialities?.split('、').map(item=>({speciality:item})),
          languages:item.languages?.split('、').map(item=>({language:item,level:''})),
        }" />
      </div>
    </div>
  </div>
  <div style="padding:100px" v-else>
    <el-empty :image-size="250" description="暂无收藏简历" />
  </div>
  <div class="demo-pb" v-if="total>0">
    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      size="large"
      :background="true"
      layout="prev, pager, next, jumper"
      :total="total"
      @current-change="handleCurrentChange"
    />
  </div>
</template>
<style scoped>
.demo-pb{
  display: flex;
  width:100%;
  align-items: center;
  justify-content: space-around;
  position: fixed;
  bottom:20px;
}
</style>
