<script setup>
import {reactive, ref } from 'vue';
const props = defineProps({
    show: {
        type: Boolean,
        required: false,
        default: true
    },
    title: {
        type: String,
        required: false,
        default: '类名'
    },
    showTitle: {
        type: Boolean,
        required: false,
        default: false
    },
    modelValue: {
        type: Object,
        default: () => ({})
    },
})

const emit = defineEmits(['confirm', 'reset', 'update:modelValue']);

const filterValues = reactive({ ...props.modelValue });

const handleConfirm = () => {
    console.log(1);
    emit('confirm', {...filterValues});
}
const handleReset = () => {
    Object.keys(filterValues).forEach(key => {
        filterValues[key] = '';
    });
    emit('reset');
    emit('update:modelValue', { ...filterValues });
}

const showSelector = ref(true);
const showText = ref(true);
const handleClose = () => {
    showSelector.value = false;
    setTimeout(() => {
        showText.value = false;
    }, 120);
}
const handleOpen = () => {
    showSelector.value = true;
    setTimeout(() => {
        showText.value = true;
    }, 200);
}

</script>


<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->
<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->


<template>
    <div id="selector" class="selector-fixed" :class="{'no-width': !showSelector, 'no-padding': !showText}">
        <div v-if="showText">
        <div class="header">
            <div class="title-container">{{ title }}</div>
            <div class="close-btn clickable un" @click="handleClose"><img class="close-img" src="/close.svg"></div>
        </div>
        <div class="slot-container">
            <slot></slot>
        </div>
        <div class="btn-container">
            <button class="confirm-btn" @click="handleConfirm">应用筛选</button>
            <button class="reset-btn" @click="handleReset">重置</button>
        </div>
        </div>
    </div>
    
    <div class="open-btn clickable un"><img class="open-img un" src="/open.svg" @click="handleOpen" v-if="!showSelector"></div>
</template>

<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->
<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->

<style scoped>

.slot-container {
    overflow-y: auto;
    /* padding: 2px; */
    display: flex;
    flex-direction: column;
    max-height: 79vh;
}

.selector-fixed {
    background-color: #fff;
    height: calc(100vh - var(--height-header));
    width: 250px;
    position: relative;
    box-shadow: 0 0 1px rgba(0, 0, 0, 0.1);
    z-index: 2;
    padding: 5px;
    transition: width 0.3s ease-in-out;
}
.selector-fixed.hidden {
    display: none;
}

.btn-container {
    position: absolute;
    bottom: 20px;
    border-top: #e0e0e0 solid 2px;
    margin-top: 5px;
    /* padding: 10px; */
    width: 96%;
    background-color: #ffffff;
    display: flex;
    justify-content: center;
    padding: 5px 0;
    gap: 30px;
    z-index: 5;
}
.confirm-btn {
    background-color: #409EFF;
    width: 125px;
    height: 40px;
    border-radius: 4px;
    /* margin-left: 15px; */
    /* margin-right: 10px; */

    color: #ffffff;

}
.confirm-btn:hover {
    background-color: #3b93eb;
}
.reset-btn {
    background-color: #ececec;
    height: 40px;
    width: 60px;
    border-radius: 4px;

    color: #666666;
}
.reset-btn:hover {
    background-color: #e6e6e6;
}

.header {
    width: 100%;
    border-bottom: #c1c1c1 solid 2px;
    display: flex;
    height: 41px;
}
.close-btn {
    width: 30px;
    height: 41px;
    right: 15px;
    position: absolute;
    display: flex;
    align-items: center;
    justify-content: center;
}

.title-container {
    font-size: 20px;
    font-weight: 600;
    margin: 5px 10px;
}

.open-btn {
    position: absolute;
    border-radius: 0 20px 20px 0;
    top: 100px;
    height: 40px;
    width: 45px;
    background-color: #ffffff;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
.open-img {
    width: 26px;
    height: 26px;
}

</style>