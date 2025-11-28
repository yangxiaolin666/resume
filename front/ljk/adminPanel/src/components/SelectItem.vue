<script setup>
import { ref } from 'vue'
const props = defineProps({
    title: { // 筛选类名
        type: String,
        default: ''
    },
    label: {
        type: String,
        required: true
    },
    options: {
        type: Array,
        required: true,
        default: () => []
    },
})
</script>
<template>
    <div class="select-item">
        <div class="select-header">
            <div class="section-title">{{ title }}</div>
            <div class="reset-btn">清除</div>
        </div>
        <div class="select-body">
            
        </div>
    </div>
</template>
<style scoped>
.select-item {
    display: flex;
    flex-direction: column;
}

</style>