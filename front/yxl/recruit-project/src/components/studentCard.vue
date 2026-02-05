<script setup>
const { item } = defineProps({
  item: {
    type: Object,
    required: true
  }
})
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
        <el-tag size="small" type="primary" 
        class="tag" v-for="(specialitie,index) in item.specialities" 
        :key="index">{{ specialitie }}</el-tag>    
             
      </div>
    </div>
    <div class="skill">
      <h4>语言水平</h4>
      <div class="tag-box">
        <el-tag size="small" type="warning" 
        class="tag" v-for="(language,index) in item.languages" 
        :key="index">{{ language }}</el-tag>    
      </div>
    </div>
    <template #footer>
      <div style="display: flex;justify-content: space-between;">
        <el-button type="primary">查看简历</el-button>
        <el-button type="success">收藏简历</el-button>
      </div>
    </template>
  </el-card>
</template>
<style scoped lang="scss">
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.box-card{
  min-height: 400px;
}
.text {
  font-size: 14px;
}

.item {
  display: flex;
  justify-content: space-between;
  padding: 10px;
  border-bottom: 1px solid #e4e7ed;
}

.box-card {
  margin: 10px;
  border-radius: 20px;
}

.skill {
  padding: 5px;
  font-size: 15px;
  h4{
    margin: 10px 0;
  }
  .tag-box {
    display: flex;
    justify-content: flex-start;
    gap:10px;
    flex-wrap: wrap;
  }
}
</style>
