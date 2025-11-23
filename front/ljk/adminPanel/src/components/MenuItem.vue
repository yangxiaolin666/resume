<script setup>
import { computed } from 'vue';
const props = defineProps({
    text: {
        type: String,
        required: true,
        default: ""
    },
    icon: {
        type: String,
        required: false,
        default: ""
    },
    index: {
        type: Number,
        required: true,
        default: 1
    },
    activeIndex: {
        type: Number,
        required: false,
        default: 0
    }
});
const isActive = computed(() => {
    return props.index === props.activeIndex;
});


</script>

<template>
    <div class="menu-item" :class="{ active: isActive }">
        <div class="menu-item-icon un">
            <img :src="icon" :alt="text" class="svg-icon" />
        </div>
        <div class="menu-item-text un">
            {{ text }}
        </div>
    </div>
</template>

<style scoped>
* {
    font-size: 16px;
}
img {
    width: 100%;
    height: 100%;
}
.menu-item {
    width: 100px;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: box-shadow 0.2s ease;
}

.menu-item:hover {
    box-shadow: inset 0 -2px 0 #007bff;
}
.menu-item:hover .menu-item-text {
    color: #007bff;
}



.menu-item-icon {
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 2px;
}

.menu-item-text {
    transition: color 0.2s ease;
}

.svg-icon {
    filter: none;
    
    transition: filter 0.1s ease;
}
.menu-item:hover .svg-icon {
    filter: invert(15%) sepia(70%) saturate(5000%) hue-rotate(200deg) brightness(100%) contrast(100%);
}

.menu-item.active {
    box-shadow: inset 0 -2px 0 #007bff;
}
.menu-item.active .menu-item-text {
    color: #007bff;
}
.menu-item.active .svg-icon {
    filter: invert(15%) sepia(70%) saturate(5000%) hue-rotate(200deg) brightness(100%) contrast(100%);
}
</style>