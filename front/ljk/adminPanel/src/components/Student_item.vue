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
    <div class="item id"> <!-- 占位，保持与id对应 -->>
        &nbsp;&nbsp;
    </div>
    <div class="item stuId">
        学号
    </div>
    <div class="item major">
        专业
    </div>
    <div class="item name">
        姓名
    </div>
    <div class="item gender">
        性别
    </div>
    <div class="item andSoOn">
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
    <div class="details">
        <div class="details-btn clickable" @click="handleDetailsClick">
            详情
        </div>
    </div>
</div>
</template>


<style scoped>
.student-item-container {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    height: 50px;
    border-bottom: 1px solid #e5e5e5;
    box-sizing: border-box;
    padding: 0 10px;
}

.student-item-container .item {
    flex: 1;
    padding: 0 10px;
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.student-item-container .details {
    margin-left: auto;
    padding-left: 20px;
}

.student-item-container .details-btn {
    padding: 6px 12px;
    background-color: #409eff;
    color: white;
    border-radius: 4px;
    cursor: pointer;
}

.student-item-container .details-btn:hover {
    background-color: #66b1ff;
}
</style>