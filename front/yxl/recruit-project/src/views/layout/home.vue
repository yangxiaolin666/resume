<script setup>
  import {
    Search,
  } from '@element-plus/icons-vue'
  import { useStudentsStore } from '@/stores/students'
  import studentCard from '@/components/studentCard.vue'
  const StudentsStore = useStudentsStore()
  const students = StudentsStore.students
</script>

<template>
  <div class="home-container">
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
          :key="item.id" :xs="24" :sm="12" :md="8" :lg="8">
            <studentCard :item="item" />
          </el-col>
        </el-row>
      </div>
    </main>
  </div>
</template>

<style scoped>
.menu-item {
  display: flex;
  align-items: center;
  padding: 8px 0px;
  cursor: pointer;
}

.menu-item:hover {
  color: #409eff;
}

.menu-item span {
  padding-left: 30px;
  font-size: 14px;
}

/* 主要内容样式 */
.main-content {
  display: flex;
  gap: 20px;
  padding: 20px;
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
</style>
