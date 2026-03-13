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

// 处理图标路径
const iconUrl = computed(() => {
    if (!props.icon) return '';
    if (props.icon.startsWith('http')) return props.icon;
    if (props.icon.startsWith('/')) {
        const baseUrl = import.meta.env.BASE_URL || '/';
        return baseUrl.replace(/\/$/, '') + props.icon;
    }
    return props.icon;
});

</script>

<template>
    <div class="menu-item" :class="{ active: isActive }">
        <div class="menu-item-icon un">
            <img :src="iconUrl" :alt="text" class="svg-icon" />
        </div>
        <div class="menu-item-text un">
            {{ text }}
        </div>
    </div>
</template>

<style scoped>
.menu-item {
    width: 100px;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    padding: 8px 0;
    margin: 0 8px;
}

.menu-item:hover {
    box-shadow: inset 0 -2px 0 var(--primary-color);
}
.menu-item:hover .menu-item-text {
    color: var(--primary-color);
}

.menu-item-icon {
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 8px;
}

.menu-item-text {
    font-size: 14px;
    color: var(--text-regular);
    transition: color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.svg-icon {
    filter: none;
    /* 移除滤镜过渡以避免颜色异常 */
}
.menu-item:hover .svg-icon {
    filter: invert(15%) sepia(70%) saturate(5000%) hue-rotate(200deg) brightness(100%) contrast(100%);
}

.menu-item.active {
    box-shadow: inset 0 -2px 0 var(--primary-color);
}
.menu-item.active .menu-item-text {
    color: var(--primary-color);
    font-weight: 500;
}
.menu-item.active .svg-icon {
    filter: invert(15%) sepia(70%) saturate(5000%) hue-rotate(200deg) brightness(100%) contrast(100%);
}
</style>