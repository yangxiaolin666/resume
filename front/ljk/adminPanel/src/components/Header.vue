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
        <div class="img-container un"><img src="/logo.png" alt="大连民族大学校徽" ></div>
        <div class="title-container">后台管理系统</div>

        <div class="menu-container">
            <MenuItem text="学生信息" icon="/student_icon.svg" :index="1" :activeIndex="curIndex" @click="menuItemClick(1)"></MenuItem>
            <MenuItem text="企业信息" icon="/company_icon.png" :index="2" :activeIndex="curIndex" @click="menuItemClick(2)"></MenuItem>
            <MenuItem text="统计数据" icon="/stats_icon.svg" :index="3" :activeIndex="curIndex" @click="menuItemClick(3)"></MenuItem>
            <MenuItem text="设置" icon="/setting_icon.svg" :index="4" :activeIndex="curIndex" @click="menuItemClick(4)"></MenuItem>
        </div>

        <div class="logout-container" @click="openConfirm">
            <div class="logout-icon">
                <el-tooltip class="item" effect="light" content="退出登录" placement="bottom">
                    <img class="un" src="/logout.svg" alt="退出登录">
                </el-tooltip>
            </div>
        </div>
    </div>
</template>

<style scoped>

/* 头部组件样式 */
img {
    width: 100%;
    height: 100%;
}
.header-container {
    position: relative;
    display: flex;
    align-items: center;
    padding: 0 24px;
    height: var(--height-header);
    background-color: var(--bg-primary);
    width: 100%;
    margin: 0;
    box-shadow: var(--shadow-light);
    z-index: 100;
}
.img-container {
    width: 40px;
    height: 40px;
    margin-right: 16px;
    border-radius: var(--border-radius-sm);
    overflow: hidden;
}
.title-container {
    font-size: 20px;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
}
.logout-container {
    position: absolute;
    top: 50%;
    right: 24px;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    transform: translateY(-50%);
    border-radius: var(--border-radius-base);
    transition: all 0.3s ease;
}

.logout-container:hover {
    background-color: var(--bg-secondary);
}
/* .logout-container:hover {

} */
.logout-icon {
    width: 24px;
    height: 24px;
}

.menu-container {
    display: flex;
    gap: 16px;
    align-items: center;
    margin-left: 32px;
    height: 100%;
}
</style>