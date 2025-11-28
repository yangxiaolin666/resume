<script setup>
import { ref, computed } from 'vue'
const props = defineProps({
    title: { // 筛选类名
        type: String,
        default: ''
    },
    label: {
        type: Array,
        required: true
    },
    options: {
        type: Array,
        required: true,
        default: () => []
    },
    modelValue: {
        type: [Array, String, Number],
        default: () => []
    },
    filterKey: {
        type: String,
        required: true
    },
})

const emit = defineEmits(['update:modelValue', 'change']);

const selectedValue = computed({
    get() {
        return props.modelValue;
    },
    set(value) {
        emit('update:modelValue', value);
        emit('change', {
            key: props.filterKey,
            value: value
        });
    }
});

const handleReset = () => {
    selectedValue.value = [];
};

const handleCheckboxChange = () => {
    emit('change', {
        key: props.filterKey,
        value: selectedValue.value
    });
};
</script>

<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->
<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->

<template>
    <div class="select-item">
        <div class="select-header">
            <div class="section-title">{{ title }}</div>
            <div class="reset-btn" @click="handleReset">清除</div>
        </div>
        <div class="select-body">
            <div class="select-list" v-for="(item, index) in label" :key="index">
                <el-checkbox 
                    v-model="selectedValue" 
                    :label="item.value" 
                    @change="handleCheckboxChange"
                    class="item-text"
                >
                    {{ item.text }}
                </el-checkbox>
            </div>
        </div>
    </div>
</template>

<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->
<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->

<style scoped>
.select-item {
    display: flex;
    flex-direction: column;
    padding: 10px 15px;
    
    border-bottom: #e2e2e2 solid 1px;
}
.select-header {
    display: flex;
    justify-content: space-between;
}
.section-title {
    font-weight: 600;
}
.reset-btn {
    color: #3498db;
    cursor: pointer;
    font-size: 12px;
}
:deep(.el-checkbox__label) {
  font-size: 14px;
}

.select-list:hover {
    background-color: #f0f7ff;
} 

</style>