<script setup></script>
<template>
  <el-card class="box-card">
    <template #header>
      <div class="card-header">
        <span>卡片名称</span>
        <el-button class="button" type="text">
          操作
        </el-button>
      </div>
    </template>
    <div v-for="o in 4" :key="o" class="text item">
      {{ 'List item ' + o }}
    </div>
  </el-card>
</template>
<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.text {
  font-size: 14px;
}

.item {
  margin-bottom: 18px;
}

.box-card {
  width: 300px;
}</style>
