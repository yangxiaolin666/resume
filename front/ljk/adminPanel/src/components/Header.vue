<script setup>

import { ElMessageBox, ElMessage } from 'element-plus';
import MenuItem from './MenuItem.vue';
import { computed, ref } from 'vue';
const openConfirm = () => {
  ElMessageBox.confirm(
    '即将退出登录, 是否继续?',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    // 退出登录逻辑
    emit('logout');
  }).catch(() => {
    // 正常来说，这里不会有任何操作
  });
};

const props = defineProps({
    defaultActive: {
        type: Number,
        required: false,
        default: 1
    }
})

const curIndex = ref(props.defaultActive);

const emit = defineEmits(['menu-change', 'logout']);

const menuItemClick = (index) => {
  curIndex.value = index;
  emit('menu-change', index);
};

</script>

<template>
    <div class="comfirm-container">

    </div>
    <div class="header-container">
        <div class="img-container un hide"><img src="/logo.png" alt="大连民族大学校徽" ></div>
        <div class="title-container">后台管理系统</div>

        <div class="menu-container">
            <MenuItem text="学生信息" icon="/student_icon.svg" :index="1" :activeIndex="curIndex" @click="menuItemClick(1)"></MenuItem>
            <MenuItem text="企业信息" icon="/company_icon.png" :index="2" :activeIndex="curIndex" @click="menuItemClick(2)"></MenuItem>
            <MenuItem text="统计数据" icon="/stats_icon.svg" :index="3" :activeIndex="curIndex" @click="menuItemClick(3)"></MenuItem>
            <MenuItem text="设置" icon="/setting_icon.svg" :index="4" :activeIndex="curIndex" @click="menuItemClick(4)"></MenuItem>
        </div>

        <div class="logout-container" @click="openConfirm">
            <div class="logout-icon">
                <el-tooltip class="item" effect="dark" content="退出登录" placement="bottom">
                    <img class="un" src="/logout.svg" alt="退出登录">
                </el-tooltip>
            </div>
        </div>
    </div>
</template>

<style scoped>

* {
    color: var(--font-color);
}
img {
    width: 100%;
    height: 100%;
}
.header-container {
    position: relative;
    display: flex;
    align-items: center;
    /* padding: 4px; */
    height: 70px;
    background-color: #ffffff;
    width: calc(100%);
    margin: 0;
    box-shadow: 0 2px 4px rgba(0, 0, 0, .12), 0 0 6px rgba(0, 0, 0, .04)
}
.img-container {
    width: 60px;
    height: 60px;
    margin: 0px 20px;
    /* background-color: #eeeeee; */
}
.title-container {
    margin-left: 6px;
    font-size: 24px;
    font-weight: bold;
}
.logout-container {
    position: absolute;
    top: auto;
    right: 25px;
    width: 50px;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    cursor: pointer;
}
/* .logout-container:hover {

} */
.logout-icon {
    width: 36px;
    height: 36px;
}

.menu-container {
    width: auto;
    height: 100%;

    display: flex;
    gap: 10px;

    align-items: center;
    justify-content: center;

    margin-left: 45px;
}
</style>