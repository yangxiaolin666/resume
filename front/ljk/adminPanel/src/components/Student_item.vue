<script setup>
import { Student } from '../types/student.js'
const props = defineProps({
    studentInfo: {
        type: Object,
        required: true
    },
    isHeader: {
        type: Boolean,
        required: false,
        default: false
    }
})
const student = props.studentInfo;
console.log('Student object in Student_item:', student);

const handleDetailsClick = () => {
    console.log(student);
}


</script>


<template>
<div class="student-item-header" v-if="props.isHeader === true">
    <div class="header-item id"> <!-- 占位，保持与id对应 -->
    </div>
    <div class="header-item stuId">
        学号
    </div>
    <div class="header-item major">
        专业
    </div>
    <div class="header-item name">
        姓名
    </div>
    <div class="header-item gender">
        性别
    </div>
    <div class="header-item andSoOn">
        ...
    </div>
</div>
<div class="student-item-container" v-else>
    <div class="item id">
        {{ student.id }}
    </div>
    <div class="item stuId">
        {{ student.stuId }}
    </div>
    <div class="item major">
        {{ student.major }}
    </div>
    <div class="item name">
        {{ student.name }}
    </div>
    <div class="item gender">
        {{ student.gender }}
    </div>
    <div class="item andSoOn">
        ...
    </div>
    <div class="item details">
        <div class="details-btn clickable" @click="handleDetailsClick">
            详情
        </div>
    </div>
</div>
</template>


<style scoped>
.student-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 15px;
    background-color: var(--bg-secondary);
    border: 1px solid var(--border-color-base);
    border-bottom: none;
    font-weight: 600;
    color: var(--text-primary);
}

.student-item-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 15px;
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color-base);
    border-bottom: none;
    transition: background-color 0.3s ease;
}

.student-item-container:last-child {
    border-bottom: 1px solid var(--border-color-base);
}

.student-item-container:hover {
    background-color: var(--bg-tertiary);
}

.header-item,
.item {
    flex: 1;
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 0 5px;
    font-size: 14px;
}

.header-item.id,
.item.id {
    width: 40px;
    flex: none;
    color: var(--text-secondary);
}

.header-item.details,
.item.details {
    width: 80px;
    flex: none;
}

.details-btn {
    background-color: var(--primary-color);
    color: white;
    padding: 4px 12px;
    border-radius: var(--border-radius-sm);
    font-size: 12px;
    transition: background-color 0.3s ease;
}

.details-btn:hover {
    background-color: var(--primary-hover);
}
</style>

