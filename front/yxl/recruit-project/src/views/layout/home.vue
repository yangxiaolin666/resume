<script setup>
import {
  Search,
} from '@element-plus/icons-vue'
import studentCard from '@/components/studentCard.vue'
const students = ref()
import { onMounted, ref } from 'vue'
import { getAllStudent } from '@/api/student'
const currentPage = ref(1)
const pageSize = ref(1000)
const total=ref(10)
onMounted(async ()=>{
  const res=await getAllStudent({
    "pageNum":currentPage.value,
    "pageSize":pageSize.value
  })
  const {data:{data}}=res
  console.log(data)
  total.value=data.size
  students.value=data.list
})
const handleSizeChange = (val) => {
  console.log(`${val} items per page`)
}
const handleCurrentChange =async (val) => {
  console.log(`current page: ${val}`)
  const res=await getAllStudent({
    "pageNum":currentPage.value,
    "pageSize":pageSize.value
  })
  const {data:{data}}=res
  console.log(data)
  students.value=data.list
}
</script>

<template>
  <div class="home-container" style="margin-bottom: 60px;">
    <main class="main-content">
      <aside class="filter-sidebar">
        <el-card class="filter-card">
          <template #header>
            <div class="filter-header">
              <span>筛选条件</span>
              <el-button link>
                重置
              </el-button>
            </div>
          </template>

          <!-- 搜索框 -->
          <div class="filter-section">
            <el-input>
              <template #prefix>
                <el-icon>
                  <Search />
                </el-icon>
              </template>
            </el-input>
          </div>

          <!-- 学位筛选 -->
          <div class="filter-section">
            <h4>学位</h4>
            <el-radio-group size="small">
              <el-radio-button value="">
                全部
              </el-radio-button>
            </el-radio-group>
          </div>
        </el-card>
      </aside>
      <div>
        <el-row>
          <el-col v-for="item in students" 
          :key="item.id" :xs="24" :sm="12" :md="8" :lg="6">
            <studentCard :item="item" />
          </el-col>
        </el-row>
      </div>
    </main>
  </div>
  <div class="demo-pb">
    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      size="large"
      :background="true"
      layout="prev, pager, next, jumper"
      :total="total"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
    />
  </div>
</template>

<style scoped>
.menu-item {
  display: flex;
  align-items: center;
  padding: 3px 0px;
  cursor: pointer;
}

.menu-item:hover {
  color: #409eff;
}

.menu-item span {
  padding-left: 20px;
  font-size: 14px;
}

/* 主要内容样式 */
.main-content {
  display: flex;
  gap: 5px;
  padding: 10px;
}

.filter-sidebar {
  margin-top: 10px;
  width: 280px;
  flex-shrink: 0;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.demo-pb{
  display: flex;
  width:100%;
  align-items: center;
  justify-content: space-around;
  position: fixed;
  bottom:20px;
}
</style>
