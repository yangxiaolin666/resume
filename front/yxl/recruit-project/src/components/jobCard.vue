<script setup>
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
const router = useRouter()
const userStore = useUserStore()
const { item, onEdit } = defineProps({
  item: {
    type: Object,
    required: true
  },
  onEdit: {
    type: Function,
    default: null
  }
})

const handleDetail = () => {
  let url = item.external_link
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    url = 'https://' + url
  }
  window.open(url, '_blank')
}

const handleEdit = () => {
  if (onEdit) {
    onEdit(item)
  }
}

// 学生可以直接联系发布该岗位的企业
const handleContact = () => {
  router.push({
    path: '/chat',
    query: {
      peerId: item.comUserId,
      peerType: 'company',
      name: item.comName
    }
  })
}
</script>
<template>
  <el-card class="job-card">
    <template #header>
      <div class="card-header">
        <h3>{{ item.title }}</h3>
        <el-tag :type="item.status === 'active' ? 'success' : 'info'">
          {{ item.status === 'active' ? '招聘中' : '已结束' }}
        </el-tag>
      </div>
    </template>
    <div class="company-info">
      <span class="label">公司:</span>
      <span class="value">{{ item.comName }}</span>
    </div>
    <div class="time-info">
      <div class="time-item">
        <span class="label">发布时间:</span>
        <span class="value">{{ item.created_at }}</span>
      </div>
      <div class="time-item">
        <span class="label">更新时间:</span>
        <span class="value">{{ item.updated_at }}</span>
      </div>
    </div>
    <div class="description">
      <span class="label">职位描述:</span>
      <p class="value">{{ item.description }}</p>
    </div>
    <template #footer>
      <div class="card-actions">
        <el-button type="primary" @click="handleDetail">
          查看详情
        </el-button>
        <el-button
          v-if="userStore.userType === 'student' && item.comUserId"
          type="success"
          @click="handleContact"
        >
          联系企业
        </el-button>
        <el-button v-if="typeof onEdit === 'function'" type="warning" @click="handleEdit">
          修改
        </el-button>
      </div>
    </template>
  </el-card>
</template>
<style scoped lang="scss">
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  h3 {
    margin: 0;
    font-size: 18px;
    color: var(--el-text-color-primary);
  }
}

.job-card {
  min-height: 380px;
  max-height: 380px;
  min-width: 280px;
  max-width: 280px;
  margin: 8px;
  border-radius: 12px;
  transition: all 0.3s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    transform: translateY(-2px);
  }
}

.company-info {
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color);

  .label {
    color: var(--el-text-color-secondary);
    margin-right: 8px;
  }

  .value {
    color: var(--el-text-color-primary);
    font-weight: 500;
  }
}

.description {
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color);

  .label {
    display: block;
    color: var(--el-text-color-secondary);
    margin-bottom: 8px;
  }

  .value {
    color: var(--el-text-color-regular);
    font-size: 14px;
    line-height: 1.6;
    margin: 0;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.time-info {
  padding: 12px 0;

  .time-item {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;

    &:last-child {
      margin-bottom: 0;
    }

    .label {
      color: var(--el-text-color-secondary);
      font-size: 13px;
    }

    .value {
      color: var(--el-text-color-regular);
      font-size: 13px;
    }
  }
}

.card-actions {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
