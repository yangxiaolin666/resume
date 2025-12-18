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
    console.log(2);
    Object.keys(filterValues).forEach(key => {
        filterValues[key] = [];
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
    }, 100);
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
            <div class="close-btn clickable un" @click="handleClose"><img class="close-img clickable" src="/close.svg" alt="收起"></div>
        </div>
        <div class="slot-container">
            <slot></slot>
        </div>
        <div class="btn-container">
            <button class="confirm-btn un" @click="handleConfirm">应用筛选</button>
            <button class="reset-btn un" @click="handleReset">重置</button>
        </div>
        </div>
    </div>
    
    <div class="open-btn clickable un"  @click="handleOpen"><img class="open-img un clickable" src="/open.svg" v-if="!showSelector"></div>
    <!-- 占用左侧纵向一部分宽度的空间，避免信息被展开按钮遮挡 -->
    <div class="space" v-if="!showSelector"></div>
</template>

<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->
<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->

<style scoped>

.slot-container {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - var(--height-header) - 52px - 80px);
    padding: 8px 0;
}

.selector-fixed {
    background-color: var(--bg-primary);
    height: calc(100vh - var(--height-header));
    width: 260px;
    position: relative;
    box-shadow: var(--shadow-light);
    z-index: 2;
    padding: 0;
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border-right: 1px solid var(--border-color-light);
    overflow: hidden;
}
.selector-fixed.hidden {
    display: none;
}

.btn-container {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    border-top: 1px solid var(--border-color-light);
    background-color: var(--bg-primary);
    display: flex;
    justify-content: flex-end;
    padding: 16px;
    gap: 12px;
    z-index: 5;
}
.confirm-btn {
    background-color: var(--primary-color);
    min-width: 80px;
    height: 36px;
    border-radius: var(--border-radius-base);
    color: #ffffff;
    font-weight: 500;
    font-size: 14px;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.confirm-btn:hover {
    background-color: var(--primary-hover);
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(24, 144, 255, 0.3);
}

.confirm-btn:active {
    transform: translateY(0);
    box-shadow: var(--shadow-light);
}
.reset-btn {
    background-color: var(--bg-secondary);
    min-width: 60px;
    height: 36px;
    border-radius: var(--border-radius-base);
    color: var(--text-regular);
    font-size: 14px;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.reset-btn:hover {
    background-color: #e8e8e8;
    transform: translateY(-1px);
}

.reset-btn:active {
    transform: translateY(0);
}

.header {
    width: 100%;
    /* border-bottom: 1px solid var(--border-color-light); */
    display: flex;
    align-items: center;
    height: 52px;
    padding: 0 16px;
    box-shadow: var(--shadow-light);
}
.close-btn {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--border-radius-sm);
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.close-btn:hover {
    background-color: var(--bg-secondary);
}

.close-btn:active {
    transform: scale(0.9);
}

.close-img {
    width: 16px;
    height: 16px;
}

.title-container {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    flex: 1;
}

.open-btn {
    position: absolute;
    border-radius: 0 var(--border-radius-lg) var(--border-radius-lg) 0;
    top: 100px;
    height: 40px;
    width: 32px;
    background-color: var(--bg-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--shadow-light);
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    z-index: 1;
}

.open-btn:hover {
    background-color: var(--bg-secondary);
    transform: translateX(2px);
}

.open-btn:active {
    transform: translateX(0) scale(0.95);
}
.open-img {
    width: 18px;
    height: 18px;
}
.space {
    width: 50px;
}


</style>