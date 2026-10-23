<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { changePwd, getMyProfile, renameSelf } from '@/api/login'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const roleText = computed(() => (userStore.userType === 'student' ? '学生' : '企业'))

// ---------- 基本信息：昵称 ----------
const nickname = ref(userStore.displayName || userStore.username)
const savingName = ref(false)

onMounted(async () => {
  try {
    const { data } = await getMyProfile()
    userStore.setDisplayName(data.data.displayName)
    nickname.value = data.data.displayName
  } catch (e) {
    // 获取失败时沿用本地缓存的昵称
  }
})

async function saveNickname() {
  const name = nickname.value.trim()
  if (!name) {
    ElMessage.warning('昵称不能为空')
    return
  }
  if (name === userStore.displayName) {
    ElMessage.info('昵称没有变化')
    return
  }
  savingName.value = true
  try {
    const { data } = await renameSelf(name)
    userStore.setDisplayName(data.data.displayName)
    nickname.value = data.data.displayName
    ElMessage.success('昵称已更新')
  } finally {
    savingName.value = false
  }
}

// ---------- 修改密码 ----------
const pwdFormRef = ref(null)
const changingPwd = ref(false)
const pwdForm = reactive({
  oldPwd: '',
  newPwd: '',
  rePwd: ''
})
const pwdRules = {
  oldPwd: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPwd: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度应为 6~20 位', trigger: 'blur' }
  ],
  rePwd: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== pwdForm.newPwd) {
          callback(new Error('两次填写的新密码不一致'))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

async function submitPassword() {
  if (!pwdFormRef.value) return
  const valid = await pwdFormRef.value.validate().catch(() => false)
  if (!valid) return

  changingPwd.value = true
  try {
    await changePwd(pwdForm.oldPwd, pwdForm.newPwd, pwdForm.rePwd)
    ElMessage.success('密码修改成功，请重新登录')
    userStore.clear()
    setTimeout(() => router.push('/login'), 1000)
  } finally {
    changingPwd.value = false
  }
}
</script>

<template>
  <div class="profile-page">
    <div class="profile-card">
      <header class="profile-header">
        <h2>修改个人信息</h2>
        <p>维护你在平台上的昵称与登录密码</p>
      </header>

      <section class="profile-section">
        <h3>基本信息</h3>
        <el-form label-width="90px" label-position="left">
          <el-form-item label="登录账号">
            <el-input :model-value="userStore.username" disabled />
          </el-form-item>
          <el-form-item label="身份">
            <el-tag effect="plain">{{ roleText }}</el-tag>
          </el-form-item>
          <el-form-item label="昵称">
            <el-input
              v-model="nickname"
              maxlength="20"
              show-word-limit
              placeholder="请输入昵称"
              @keyup.enter="saveNickname"
            />
            <span class="field-tip">昵称会显示在右上角和聊天窗口中，登录账号不会改变</span>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="savingName" @click="saveNickname">
              保存昵称
            </el-button>
          </el-form-item>
        </el-form>
      </section>

      <el-divider />

      <section class="profile-section">
        <h3>修改密码</h3>
        <el-form
          ref="pwdFormRef"
          :model="pwdForm"
          :rules="pwdRules"
          label-width="90px"
          label-position="left"
        >
          <el-form-item label="原密码" prop="oldPwd">
            <el-input v-model="pwdForm.oldPwd" type="password" show-password placeholder="请输入原密码" />
          </el-form-item>
          <el-form-item label="新密码" prop="newPwd">
            <el-input v-model="pwdForm.newPwd" type="password" show-password placeholder="6~20 位新密码" />
          </el-form-item>
          <el-form-item label="确认密码" prop="rePwd">
            <el-input v-model="pwdForm.rePwd" type="password" show-password placeholder="请再次输入新密码" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="changingPwd" @click="submitPassword">
              确认修改
            </el-button>
          </el-form-item>
        </el-form>
      </section>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  display: flex;
  justify-content: center;
  padding: 32px 20px 48px;
}

.profile-card {
  width: 100%;
  max-width: 720px;
  padding: 28px 32px 12px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
  box-sizing: border-box;
}

.profile-header {
  margin-bottom: 24px;
}

.profile-header h2 {
  margin: 0;
  font-size: 22px;
  color: var(--el-text-color-primary);
}

.profile-header p {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.profile-section h3 {
  margin: 0 0 18px;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.profile-section :deep(.el-input) {
  max-width: 360px;
}

.profile-section :deep(.el-form-item) {
  margin-bottom: 18px;
}

.field-tip {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.profile-section :deep(.el-button) {
  width: 120px;
}

@media (max-width: 600px) {
  .profile-card {
    padding: 20px 16px 8px;
  }

  .profile-section :deep(.el-input) {
    max-width: none;
  }
}
</style>
