<script setup>
import Option from './Option.vue'
</script>

<template>
    <div class="header-container">
        <div class="img-container"><img src="" alt="大连民族大学校徽"></div>
        <div class="title-container">后台管理系统</div>
        <div class="logout-container">
            <div class="logout-icon">
                <el-tooltip class="item" effect="dark" content="退出登录" placement="bottom">
                    <img src="/logout.png" alt="退出登录">
                </el-tooltip>
            </div>
        </div>
    </div>
</template>

<style scoped>
img {
    width: 100%;
    height: 100%;
}
.header-container {
    position: relative;
    display: flex;
    align-items: center;
    padding: 4px;
    height: 70px;
    background: linear-gradient(to right, var(--color-background) 0% 60%, #2AD1D7);
    /* background-color: var(--color-background); */
    width: 100%;
    margin: 0;
}
.img-container {
    width: 60px;
    height: 60px;
    margin: 0px 20px;
    background-color: #eeeeee;
}
.title-container {
    margin-left: 10px;
    font-size: 24px;
    font-weight: bold;
    color: var(--color-white);
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
    width: 40px;
    height: 40px;
}
</style>