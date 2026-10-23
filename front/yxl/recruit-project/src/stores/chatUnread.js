import { ref } from 'vue'

// 全局未读消息总数。
// 布局页顶部“消息”红点负责展示，聊天页在已读后负责清零，
// 两边共用同一个 ref，红点才能立即消失而不是等下一次轮询。
export const unreadTotal = ref(0)

export function setUnreadTotal(value) {
  unreadTotal.value = Math.max(0, Number(value) || 0)
}
