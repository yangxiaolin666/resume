<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../components/Header.vue'
import SelectorFixed from '../components/SelectorFixed.vue'

const menuSelect = ref(1); // 1: 学生信息, 2: 企业信息
const handleMenuSelect = (index) => {
  menuSelect.value = index;
  console.log("改变");
};
const handleLogout = () => {
  console.log("退出登录");
}



</script>

<template>
  <Header
    @menu-change="handleMenuSelect"
    @logout="handleLogout"
  ></Header>
  <div class="main-container">
    
    <div class="content-container">
      <div class="student-container" v-if="menuSelect === 1">
        <div class="selector-container">
          <SelectorFixed title="筛选" :show="false"></SelectorFixed>
        </div>
      </div>
      <div class="" v-if="menuSelect === 2">
        <h1>企业信息</h1>
      </div>
      <div class="" v-if="menuSelect === 3">
        <h1>统计数据</h1>
      </div>
      <div class="" v-if="menuSelect === 4">
        <h1>设置</h1>
      </div>
    </div>
  </div>
  

</template>

<style scoped>



.main-container {
  display: flex;
  height: calc(100vh - 70px);
}
.menu-container {
  height: 100vh;
  width: 180px;
  background-color: #f2f2f2;
  margin: 2px 0px;
  margin-right: 5px;
}
.content-container {
  width: 100vw;
  height: 100%;
  margin: 2px 2px;
  background-color: #f2f2f2;
}


.student-container {
  display: flex;
}


</style>
