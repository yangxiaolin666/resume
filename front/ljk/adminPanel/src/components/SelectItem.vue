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
            <div class="reset-btn un" @click="handleReset">清除</div>
        </div>
        <div class="select-body">
            <el-checkbox-group 
        v-model="selectedValue" 
        @change="handleCheckboxChange"
      >
        <div class="select-list" v-for="(item, index) in label" :key="index">
          <el-checkbox 
            :label="item.value" 
            class="item-text"
          >
            {{ item.text }}
          </el-checkbox>
        </div>
      </el-checkbox-group>
        </div>
    </div>
</template>

<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->
<!-- ///////////////////////////////////////////////////////////////////////////////////////////// -->

<style scoped>
.select-item {
    display: flex;
    flex-direction: column;
    padding: 16px;
    border-bottom: 1px solid var(--border-color-light);
}
.select-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
}
.section-title {
    font-weight: 600;
    color: var(--text-primary);
    font-size: 14px;
}
.reset-btn {
    /* color: var(--text-secondary); */
    font-size: 12px;
    padding: 4px 8px;
    border-radius: var(--border-radius-sm);
    transition: all 0.3s ease;
    cursor: pointer;
}

.reset-btn:hover {
    color: var(--text-primary);
    background-color: var(--bg-secondary);
    font-size: 12px;
}
:deep(.el-checkbox__label) {
  font-size: 14px;
  font-weight: 400;
  color: var(--text-regular);
  cursor: pointer;
  transition: color 0.3s ease;
}

.select-list {
    margin-bottom: 8px;
    display: flex;
    align-items: center;
}

.select-list:hover {
    background-color: #f0f7ff;
}

.select-body {
    margin-top: 8px;
}

</style>