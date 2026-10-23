<template>
  <div class="chat-page">
    <div class="chat-card">
      <!-- 左侧：会话列表 -->
      <aside class="chat-sidebar">
        <div class="sidebar-header">
          <div>
            <h3>消息</h3>
            <p class="sidebar-subtitle">学生与企业实时沟通</p>
          </div>
          <el-button type="primary" size="small" @click="openPeerDialog">
            发起聊天
          </el-button>
        </div>

        <el-scrollbar class="conversation-scroll">
          <div v-if="conversations.length === 0" class="sidebar-empty">
            暂无会话，点击右上角“发起聊天”
          </div>
          <div
            v-for="item in conversations"
            :key="item.peerType + ':' + item.peerId"
            class="conversation-item"
            :class="{ active: isCurrent(item.peerId, item.peerType) }"
            @click="selectPeer(item)"
          >
            <div class="avatar" :style="{ background: avatarColor(item.peerId) }">
              {{ initial(item.peerName) }}
            </div>
            <div class="conversation-main">
              <div class="conversation-top">
                <span class="peer-name">
                  <i class="online-dot" :class="{ online: item.online }"></i>
                  {{ item.peerName }}
                </span>
                <span class="time">{{ shortTime(item.lastTime) }}</span>
              </div>
              <div class="conversation-bottom">
                <span class="last-message">
                  {{ lastMessageText(item) }}
                </span>
                <span v-if="item.unread > 0" class="unread-badge">
                  {{ item.unread > 99 ? '99+' : item.unread }}
                </span>
              </div>
            </div>
          </div>
        </el-scrollbar>
      </aside>

      <!-- 右侧：聊天窗口 -->
      <section class="chat-main">
        <template v-if="current">
          <header class="chat-header">
            <div class="peer-title">
              <div
                class="avatar small clickable-avatar"
                :style="{ background: avatarColor(current.peerId) }"
                title="查看资料"
                @click="openProfile(current.peerId, current.peerType)"
              >
                {{ initial(current.peerName) }}
              </div>
              <div>
                <h3>{{ current.peerName }}</h3>
                <span class="peer-type">
                  {{ current.peerType === 'student' ? '学生' : '企业' }}
                </span>
              </div>
            </div>
            <el-tag size="small" effect="plain" :type="connectionType">
              {{ connectionText }}
            </el-tag>
          </header>

          <div ref="messageBox" class="message-box">
            <div v-if="hasMore" class="load-more">
              <el-button link type="primary" :loading="loadingMore" @click="loadMore">
                加载更早的消息
              </el-button>
            </div>

            <el-empty
              v-if="messages.length === 0"
              description="还没有消息，发送第一条吧"
              :image-size="90"
            />

            <div
              v-for="msg in messages"
              :key="msg.id"
              class="message-row"
              :class="{ mine: isMine(msg) }"
            >
              <div
                class="avatar small"
                :class="{ 'clickable-avatar': !isMine(msg) }"
                :style="{ background: avatarColor(msg.senderId) }"
                :title="isMine(msg) ? '' : '查看资料'"
                @click="openSenderProfile(msg)"
              >
                {{ initial(msg.senderName) }}
              </div>
              <div class="message-body">
                <div class="message-meta">
                  <span>{{ isMine(msg) ? '我' : msg.senderName }}</span>
                  <span>{{ shortTime(msg.createdAt) }}</span>
                </div>
                <div v-if="msg.msgType === 'image'" class="image-bubble">
                  <el-image
                    class="chat-image"
                    :src="imageUrl(msg.content)"
                    :preview-src-list="[imageUrl(msg.content)]"
                    fit="contain"
                    preview-teleported
                  />
                </div>
                <div v-else class="message-bubble">{{ msg.content }}</div>
              </div>
            </div>
          </div>

          <footer class="composer">
            <div v-if="showEmoji" ref="emojiPanelRef" class="emoji-panel">
              <button
                v-for="emoji in emojis"
                :key="emoji"
                type="button"
                class="emoji-item"
                @click="insertEmoji(emoji)"
              >
                {{ emoji }}
              </button>
            </div>
            <el-input
              ref="composerRef"
              v-model="draft"
              type="textarea"
              :rows="3"
              maxlength="800"
              show-word-limit
              resize="none"
              placeholder="输入消息，Enter 发送"
              @keydown="onComposerKeydown"
            />
            <div class="composer-actions">
              <div class="composer-tools">
                <button
                  ref="emojiBtnRef"
                  type="button"
                  class="tool-btn emoji-btn"
                  :class="{ active: showEmoji }"
                  title="表情"
                  @click="toggleEmoji"
                >
                  <el-icon><Sunny /></el-icon>
                </button>
                <button
                  type="button"
                  class="tool-btn"
                  title="发送图片"
                  :disabled="uploading"
                  @click="triggerImage"
                >
                  <el-icon><Picture /></el-icon>
                </button>
                <span class="composer-hint">Enter 发送，Ctrl + Enter 换行</span>
              </div>
              <el-button type="primary" :loading="sending" @click="handleSend">
                发送
              </el-button>
            </div>
            <input
              ref="fileInputRef"
              class="hidden-file"
              type="file"
              accept="image/png,image/jpeg,image/gif,image/webp"
              @change="onImageSelected"
            />
          </footer>
        </template>

        <div v-else class="chat-empty">
          <el-empty description="选择左侧会话，或点击“发起聊天”" />
        </div>
      </section>
    </div>

    <!-- 发起聊天：选择联系人 -->
    <el-dialog v-model="peerDialogVisible" title="发起聊天" width="480px">
      <el-input
        v-model="peerKeyword"
        placeholder="搜索学生姓名 / 企业名称"
        clearable
        @input="onPeerKeywordInput"
      />
      <el-scrollbar height="340px" class="peer-scroll">
        <div
          v-for="peer in peers"
          :key="peer.userType + ':' + peer.userId"
          class="peer-item"
          @click="startChat(peer)"
        >
          <div class="avatar" :style="{ background: avatarColor(peer.userId) }">
            {{ initial(peer.name) }}
          </div>
          <div class="peer-info">
            <div class="peer-name">
              <i class="online-dot" :class="{ online: peer.online }"></i>
              {{ peer.name }}
            </div>
            <div class="peer-desc">
              {{ peer.description || (peer.userType === 'student' ? '学生' : '企业') }}
            </div>
          </div>
          <el-button size="small" type="primary" plain>开始聊天</el-button>
        </div>
        <el-empty v-if="peers.length === 0" description="没有找到可聊天的用户" :image-size="80" />
      </el-scrollbar>
    </el-dialog>

    <!-- 对方资料：学生看简历，企业看简介与在招岗位 -->
    <el-dialog
      v-model="profileVisible"
      :title="profile?.userType === 'company' ? '企业主页' : '个人信息'"
      width="900px"
      top="5vh"
      destroy-on-close
    >
      <div v-loading="profileLoading" class="profile-dialog-body">
        <!-- 学生：布局参照 /studentDetail -->
        <div v-if="profile?.userType === 'student'" class="profile-student">
          <div class="student-side">
            <h2>{{ profile.student.name }}</h2>
            <p class="student-edu">
              {{ profile.student.graduationYear ? profile.student.graduationYear - 4 : '' }}-{{ profile.student.graduationYear }}
            </p>
            <div class="s-line"><span>学号</span><b>{{ profile.student.stuId }}</b></div>
            <div class="s-line"><span>性别</span><b>{{ profile.student.gender || '未填写' }}</b></div>
            <div class="s-line"><span>年龄</span><b>{{ profile.student.age || '未填写' }}</b></div>
            <div class="s-line"><span>民族</span><b>{{ profile.student.nation || '未填写' }}</b></div>
            <div class="s-line"><span>专业</span><b>{{ profile.student.major || '未填写' }}</b></div>
            <div class="s-line"><span>学位</span><b>{{ profile.student.degree || '未填写' }}</b></div>
            <div class="s-line"><span>毕业年份</span><b>{{ profile.student.graduationYear || '未填写' }}</b></div>
            <div class="s-line"><span>实习次数</span><b>{{ profile.student.internship ?? 0 }}</b></div>

            <div v-if="profile.student.internships?.length" class="tag-block">
              <h4>实习经历</h4>
              <div class="tag-box">
                <el-tag v-for="item in profile.student.internships" :key="item.company + item.experience" type="warning">
                  {{ item.company }}-{{ item.experience }}
                </el-tag>
              </div>
            </div>
            <div v-if="profile.student.specialities?.length" class="tag-block">
              <h4>技能水平</h4>
              <div class="tag-box">
                <el-tag v-for="item in profile.student.specialities" :key="item.speciality" type="primary">
                  {{ item.speciality }}
                </el-tag>
              </div>
            </div>
            <div v-if="profile.student.languages?.length" class="tag-block">
              <h4>语言水平</h4>
              <div class="tag-box">
                <el-tag v-for="item in profile.student.languages" :key="item.language" type="info">
                  {{ item.language }}-{{ item.level }}
                </el-tag>
              </div>
            </div>
            <div v-if="profile.student.certificates?.length" class="tag-block">
              <h4>获奖证书</h4>
              <div class="tag-box">
                <el-tag v-for="item in profile.student.certificates" :key="item.certName" type="success">
                  {{ item.certName }}-{{ item.level }}
                </el-tag>
              </div>
            </div>
            <div v-if="profile.student.region?.city || profile.student.region?.province" class="tag-block">
              <h4>意向工作城市</h4>
              <div class="tag-box">
                <el-tag type="danger">
                  {{ profile.student.region?.city || profile.student.region?.province }}
                </el-tag>
              </div>
            </div>
            <div v-if="profile.student.jobs?.length" class="tag-block">
              <h4>意向工作类别</h4>
              <div class="tag-box">
                <el-tag v-for="item in profile.student.jobs" :key="item.post" type="warning">
                  {{ item.post }}
                </el-tag>
              </div>
            </div>
          </div>

          <div class="student-resume">
            <embed
              v-if="profile.student.resume"
              :src="resumeUrl(profile.student.resume.resume)"
              type="application/pdf"
            />
            <el-empty v-else description="暂无简历文件" :image-size="110" />
          </div>
        </div>

        <!-- 企业：简介 + 在招岗位 -->
        <div v-else-if="profile?.userType === 'company'" class="profile-company">
          <div class="company-head">
            <div class="avatar preview" :style="{ background: avatarColor(profile.company.userId) }">
              {{ initial(profile.company.comName) }}
            </div>
            <div>
              <h2>{{ profile.company.comName }}</h2>
              <p>{{ profile.company.comRegion || '地区未填写' }} · 在招 {{ profile.jobs?.length || 0 }} 个岗位</p>
            </div>
          </div>

          <p class="company-intro">
            {{ profile.company.comName }} 位于{{ profile.company.comRegion || '未知地区' }}，
            目前在平台发布了 {{ profile.jobs?.length || 0 }} 个岗位，欢迎在聊天中进一步沟通。
          </p>

          <h3 class="section-title">在招岗位</h3>
          <div v-if="profile.jobs?.length" class="job-list">
            <div v-for="job in profile.jobs" :key="job.id" class="job-item">
              <div class="job-top">
                <strong>{{ job.title }}</strong>
                <span class="job-time">{{ (job.createdAt || '').slice(0, 10) }}</span>
              </div>
              <p class="job-desc">{{ job.description || '暂无岗位描述' }}</p>
              <div class="job-foot">
                <el-tag size="small" type="success">招聘中</el-tag>
                <el-link v-if="job.externalLink" type="primary" :href="job.externalLink" target="_blank">
                  投递链接
                </el-link>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无在招岗位" :image-size="90" />
        </div>

        <el-empty v-else-if="!profileLoading" description="暂无资料" :image-size="100" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Picture, Sunny } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { setUnreadTotal } from '@/stores/chatUnread'
import { ChatSocket } from '@/utils/chatSocket'
import { API_BASE_URL } from '@/utils/request'
import {
  getChatPeers,
  getChatMessages,
  getConversations,
  getPeerProfile,
  markChatRead,
  sendChatMessage,
  uploadChatImage
} from '@/api/chat'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const myId = computed(() => Number(userStore.userId))
const myType = computed(() => userStore.userType)

const conversations = ref([])
const peers = ref([])
const current = ref(null)
const messages = ref([])
const draft = ref('')
const hasMore = ref(false)
const loadingMore = ref(false)
const sending = ref(false)
const messageBox = ref(null)
const composerRef = ref(null)

const peerDialogVisible = ref(false)
const peerKeyword = ref('')
const connectionStatus = ref('closed')

// ---------- 表情与图片 ----------
const showEmoji = ref(false)
const uploading = ref(false)
const fileInputRef = ref(null)
const emojiPanelRef = ref(null)
const emojiBtnRef = ref(null)

// ---------- 对方资料弹窗 ----------
const profileVisible = ref(false)
const profileLoading = ref(false)
const profile = ref(null)
const emojis = [
  '😀', '😄', '😁', '😆', '😅', '😂', '🤣', '😊',
  '😍', '😘', '😎', '🤔', '😐', '😴', '😭', '😡',
  '👍', '👎', '👌', '🙏', '👏', '💪', '🤝', '✌️',
  '❤️', '💔', '🔥', '✨', '🎉', '🎁', '☕', '🌹',
  '😉', '😋', '🤗', '🤩', '😇', '🙃', '😜', '🤫',
  '💯', '✅', '❌', '⏰', '📌', '📎', '💡', '🚀'
]

let socket = null
let peerSearchTimer = null
let refreshTimer = null
let presenceTimer = null

// 右上角标签表示“对方”的在线状态，与会话列表里的绿点保持一致
const connectionText = computed(() => (current.value?.online ? '在线' : '离线'))

const connectionType = computed(() => (current.value?.online ? 'success' : 'info'))

function isMine(message) {
  return Number(message.senderId) === myId.value && message.senderType === myType.value
}

function isCurrent(peerId, peerType) {
  return !!current.value &&
    Number(current.value.peerId) === Number(peerId) &&
    current.value.peerType === peerType
}

function findConversation(peerId, peerType) {
  return conversations.value.find(
    (c) => Number(c.peerId) === Number(peerId) && c.peerType === peerType
  )
}

// ---------- 展示辅助 ----------
const avatarColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#9b59b6', '#1abc9c']

function avatarColor(id) {
  return avatarColors[Number(id) % avatarColors.length]
}

function initial(name) {
  const text = String(name || '?').trim()
  return text ? text.charAt(0).toUpperCase() : '?'
}

function shortTime(value) {
  if (!value) return ''
  const date = new Date(String(value).replace(' ', 'T'))
  if (Number.isNaN(date.getTime())) return String(value)
  const now = new Date()
  const sameDay = date.toDateString() === now.toDateString()
  const pad = (n) => String(n).padStart(2, '0')
  if (sameDay) return `${pad(date.getHours())}:${pad(date.getMinutes())}`
  return `${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function imageUrl(path) {
  if (!path) return ''
  return /^https?:\/\//.test(path) ? path : API_BASE_URL + path
}

// 简历文件在数据库里可能是 uploads\resume\x.pdf 这种反斜杠路径，统一转成 URL
function resumeUrl(path) {
  if (!path) return ''
  if (/^https?:\/\//.test(path)) return path
  return API_BASE_URL + '/' + String(path).replace(/\\/g, '/').replace(/^\/+/, '')
}

async function openProfile(peerId, peerType) {
  if (!peerId || !peerType) return
  profileVisible.value = true
  profileLoading.value = true
  profile.value = null
  try {
    const { data } = await getPeerProfile(peerId, peerType)
    profile.value = data.data
  } finally {
    profileLoading.value = false
  }
}

// 只允许点对方的头像查看资料，自己的头像不弹窗
function openSenderProfile(msg) {
  if (isMine(msg)) return
  openProfile(msg.senderId, msg.senderType)
}

function lastMessageText(item) {
  const body = item.lastMsgType === 'image' ? '[图片]' : item.lastMessage
  return `${item.lastFromMe ? '我：' : ''}${body}`
}

// ---------- 表情与图片发送 ----------
function toggleEmoji() {
  showEmoji.value = !showEmoji.value
}

// 点击表情面板以外的任何位置，自动收起面板
function onDocumentMouseDown(event) {
  if (!showEmoji.value) return
  if (emojiPanelRef.value?.contains(event.target) || emojiBtnRef.value?.contains(event.target)) {
    return
  }
  showEmoji.value = false
}

function insertEmoji(emoji) {
  const el = composerRef.value?.textarea
  if (!el) {
    draft.value += emoji
    return
  }
  const start = el.selectionStart ?? draft.value.length
  const end = el.selectionEnd ?? start
  draft.value = `${draft.value.slice(0, start)}${emoji}${draft.value.slice(end)}`
  nextTick(() => {
    el.focus()
    const pos = start + emoji.length
    el.selectionStart = pos
    el.selectionEnd = pos
  })
}

function triggerImage() {
  fileInputRef.value?.click()
}

async function onImageSelected(event) {
  const file = event.target.files?.[0]
  // 允许下次选择同一张图片
  event.target.value = ''
  if (!file || !current.value) return

  if (!/^image\/(png|jpeg|gif|webp)$/.test(file.type)) {
    ElMessage.error('仅支持 JPG、PNG、GIF、WEBP 图片')
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片不能超过 5MB')
    return
  }

  uploading.value = true
  try {
    const { data } = await uploadChatImage(file)
    await sendPayload(data.data.url, 'image')
  } finally {
    uploading.value = false
  }
}

// ---------- 数据加载 ----------
async function loadConversations() {
  const res = await getConversations()
  conversations.value = res.data.data.list || []
  syncUnreadTotal()
  if (current.value) {
    const item = conversations.value.find(
      (c) => Number(c.peerId) === Number(current.value.peerId) && c.peerType === current.value.peerType
    )
    if (item) current.value.online = item.online
  }
}

async function loadPeers() {
  const res = await getChatPeers(peerKeyword.value)
  peers.value = res.data.data.list || []
}

async function loadMessages(reset) {
  if (!current.value) return
  const params = {
    peerId: current.value.peerId,
    peerType: current.value.peerType,
    limit: 30
  }
  if (!reset && messages.value.length > 0) {
    params.beforeId = messages.value[0].id
  }
  const res = await getChatMessages(params)
  const data = res.data.data
  const list = data.messages || []
  messages.value = reset ? list : [...list, ...messages.value]
  hasMore.value = !!data.hasMore

  // 从入口按钮进来时可能只带了 peerId，没有名字，这里用消息里的名字补全
  if (current.value.peerName === '聊天' && messages.value.length > 0) {
    const sample = messages.value[0]
    const fromPeer = Number(sample.senderId) === Number(current.value.peerId) &&
      sample.senderType === current.value.peerType
    current.value.peerName = fromPeer ? sample.senderName : sample.receiverName
  }
}

async function selectPeer(peer) {
  const peerId = Number(peer.peerId ?? peer.userId)
  const peerType = peer.peerType ?? peer.userType
  const existing = conversations.value.find(
    (c) => Number(c.peerId) === peerId && c.peerType === peerType
  )
  current.value = {
    peerId,
    peerType,
    peerName: existing?.peerName || peer.peerName || peer.name || '聊天',
    online: existing?.online ?? !!peer.online
  }
  messages.value = []
  hasMore.value = false

  await loadMessages(true)
  await markCurrentRead()
  clearUnread(peerId, peerType)
  scrollToBottom()

  router.replace({
    path: '/chat',
    query: {
      peerId,
      peerType,
      name: current.value.peerName && current.value.peerName !== '聊天'
        ? current.value.peerName
        : undefined
    }
  })
}

async function markCurrentRead() {
  if (!current.value) return
  try {
    await markChatRead({
      peerId: current.value.peerId,
      peerType: current.value.peerType
    })
  } catch (e) {
    // 标记已读失败不影响聊天
  }
}

function clearUnread(peerId, peerType) {
  const item = conversations.value.find(
    (c) => Number(c.peerId) === Number(peerId) && c.peerType === peerType
  )
  if (item) item.unread = 0
  syncUnreadTotal()
}

// 把会话列表里的未读数汇总给顶部“消息”红点
function syncUnreadTotal() {
  setUnreadTotal(
    conversations.value.reduce((sum, item) => sum + (Number(item.unread) || 0), 0)
  )
}

async function loadMore() {
  if (!current.value || loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  const box = messageBox.value
  const oldHeight = box ? box.scrollHeight : 0
  const oldTop = box ? box.scrollTop : 0
  try {
    await loadMessages(false)
    await nextTick()
    if (box) {
      // 插入旧消息后保持当前阅读位置
      box.scrollTop = box.scrollHeight - oldHeight + oldTop
    }
  } finally {
    loadingMore.value = false
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (messageBox.value) {
      messageBox.value.scrollTop = messageBox.value.scrollHeight
    }
  })
}

// ---------- WebSocket ----------
function connectSocket() {
  socket = new ChatSocket({
    token: userStore.token,
    onMessage: handleSocketMessage,
    onStatus: (status) => {
      connectionStatus.value = status
      // 重连成功后立即刷新一次在线状态
      if (status === 'open') refreshPresence()
    }
  })
  socket.connect()
}

function handleSocketMessage(msg) {
  if (msg.type === 'error') {
    ElMessage.error(msg.message || '消息发送失败')
    return
  }
  if (msg.type !== 'message' || !msg.data) return

  const message = msg.data
  const peerId = isMine(message) ? message.receiverId : message.senderId
  const peerType = isMine(message) ? message.receiverType : message.senderType

  if (isCurrent(peerId, peerType)) {
    // 同一个消息可能既来自 WebSocket 推送，又来自 REST 降级响应，按 id 去重
    if (!messages.value.some((item) => item.id === message.id)) {
      messages.value.push(message)
      scrollToBottom()
    }
    if (!isMine(message)) {
      markCurrentRead()
    }
  }
  scheduleRefreshConversations()
}

function scheduleRefreshConversations() {
  if (refreshTimer) return
  refreshTimer = setTimeout(async () => {
    refreshTimer = null
    try {
      await loadConversations()
    } catch (e) {
      // 刷新失败不打扰用户
    }
  }, 300)
}

// 对端上下线时后端不会主动推送，所以定时拉一次会话列表来更新在线状态
async function refreshPresence() {
  if (document.hidden) return
  try {
    await loadConversations()
    // 刚发起、还没有历史消息的会话不在会话列表里，单独查一次联系人状态
    if (current.value && !findConversation(current.value.peerId, current.value.peerType)) {
      const res = await getChatPeers('')
      const peer = (res.data.data.list || []).find(
        (p) => Number(p.userId) === Number(current.value.peerId) && p.userType === current.value.peerType
      )
      if (peer) current.value.online = !!peer.online
    }
  } catch (e) {
    // 轮询失败静默忽略，等下一轮
  }
}

// ---------- 发送 ----------
function onComposerKeydown(event) {
  if (event.key !== 'Enter') return
  // 中文输入法正在选词时按 Enter 只是确认候选词，不发送
  if (event.isComposing || event.keyCode === 229) return
  // Shift + Enter 交给浏览器原生换行
  if (event.shiftKey) return

  event.preventDefault()
  // Ctrl / Cmd + Enter 在光标处插入换行
  if (event.ctrlKey || event.metaKey) {
    insertLineBreak()
    return
  }
  handleSend()
}

function insertLineBreak() {
  const el = composerRef.value?.textarea
  if (!el) {
    draft.value += '\n'
    return
  }
  const start = el.selectionStart ?? draft.value.length
  const end = el.selectionEnd ?? start
  draft.value = `${draft.value.slice(0, start)}\n${draft.value.slice(end)}`
  // 等 v-model 更新后再把光标移到换行之后
  nextTick(() => {
    el.selectionStart = start + 1
    el.selectionEnd = start + 1
  })
}

// 统一发送入口：优先走 WebSocket，断线时自动降级为 REST 发送
async function sendPayload(content, msgType = 'text') {
  if (!current.value) return false

  if (socket && socket.connected &&
    socket.sendMessage(current.value.peerId, current.value.peerType, content, msgType)) {
    return true
  }

  sending.value = true
  try {
    const res = await sendChatMessage({
      receiverId: current.value.peerId,
      receiverType: current.value.peerType,
      content,
      msgType
    })
    handleSocketMessage({ type: 'message', data: res.data.data })
    return true
  } catch (e) {
    return false
  } finally {
    sending.value = false
  }
}

async function handleSend() {
  const content = draft.value.trim()
  if (!content || !current.value) return

  const ok = await sendPayload(content, 'text')
  if (ok) {
    draft.value = ''
    showEmoji.value = false
  }
}

// ---------- 联系人弹窗 ----------
function openPeerDialog() {
  peerDialogVisible.value = true
  loadPeers()
}

function onPeerKeywordInput() {
  if (peerSearchTimer) clearTimeout(peerSearchTimer)
  peerSearchTimer = setTimeout(loadPeers, 300)
}

async function startChat(peer) {
  peerDialogVisible.value = false
  await selectPeer(peer)
}

// ---------- 生命周期 ----------
onMounted(async () => {
  if (myType.value !== 'student' && myType.value !== 'company') {
    ElMessage.warning('只有学生和企业可以使用聊天')
    return
  }

  await loadConversations()
  await loadPeers()
  connectSocket()
  // 每 8 秒刷新一次对方的在线/离线状态
  presenceTimer = setInterval(refreshPresence, 8000)
  // 从其他标签页切回来时立即刷新
  window.addEventListener('focus', refreshPresence)
  // 点击页面其他位置时收起表情面板
  document.addEventListener('mousedown', onDocumentMouseDown)

  const queryPeerId = Number(route.query.peerId || 0)
  const queryPeerType = route.query.peerType
  if (queryPeerId > 0 && queryPeerType) {
    await selectPeer({
      peerId: queryPeerId,
      peerType: queryPeerType,
      peerName: route.query.name
    })
  }
  // 不自动选中第一个会话，保留未读角标，交给用户主动点击
})

onBeforeUnmount(() => {
  if (socket) socket.close()
  if (peerSearchTimer) clearTimeout(peerSearchTimer)
  if (refreshTimer) clearTimeout(refreshTimer)
  if (presenceTimer) clearInterval(presenceTimer)
  window.removeEventListener('focus', refreshPresence)
  document.removeEventListener('mousedown', onDocumentMouseDown)
})
</script>

<style scoped>
.chat-page {
  /* 顶部导航固定 80px，剩下的区域整块铺满 */
  height: calc(100vh - 80px);
  padding: 0;
  box-sizing: border-box;
}

.chat-card {
  height: 100%;
  min-width: 0;
  display: flex;
  overflow: hidden;
  background: var(--el-bg-color);
  border: none;
  border-radius: 0;
  box-shadow: none;
}

/* 左侧会话列表 */
.chat-sidebar {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color);
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.sidebar-header h3 {
  margin: 0;
  font-size: 17px;
  color: var(--el-text-color-primary);
}

.sidebar-subtitle {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.conversation-scroll {
  flex: 1;
}

.sidebar-empty {
  padding: 40px 20px;
  text-align: center;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.conversation-item {
  display: flex;
  gap: 12px;
  padding: 14px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--el-border-color-lighter);
  transition: background 0.2s;
}

.conversation-item:hover {
  background: var(--el-fill-color-light);
}

.conversation-item.active {
  background: var(--el-color-primary-light-9);
}

.conversation-main {
  flex: 1;
  min-width: 0;
}

.conversation-top,
.conversation-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.peer-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

.conversation-bottom {
  margin-top: 6px;
}

.last-message {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.unread-badge {
  flex-shrink: 0;
  box-sizing: border-box;
  width: 18px;
  height: 18px;
  padding: 0;
  border-radius: 50%;
  background: var(--el-color-danger);
  color: #fff;
  font-size: 10px;
  line-height: 18px;
  text-align: center;
}

/* 右侧聊天区 */
.chat-main {
  flex: 1;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color-page);
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.peer-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.peer-title h3 {
  margin: 0;
  font-size: 16px;
  color: var(--el-text-color-primary);
}

.peer-type {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.message-box {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  /* 始终给滚动条预留位置，避免消息变多时滚动条出现/消失导致内容左右抖动 */
  scrollbar-gutter: stable;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.load-more {
  text-align: center;
}

.message-row {
  display: flex;
  gap: 10px;
  max-width: 72%;
  min-width: 0;
}

.message-row.mine {
  margin-left: auto;
  flex-direction: row-reverse;
}

.message-body {
  min-width: 0;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.message-row.mine .message-meta {
  flex-direction: row-reverse;
}

.message-bubble {
  padding: 10px 14px;
  border-radius: 8px;
  border-top-left-radius: 2px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-light);
  color: var(--el-text-color-primary);
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.message-row.mine .message-bubble {
  border-radius: 8px;
  border-top-right-radius: 2px;
  background: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: #fff;
}

.composer {
  position: relative;
  padding: 14px 20px;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color-lighter);
}

.composer-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
}

.composer-tools {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--el-text-color-regular);
  font-size: 16px;
  cursor: pointer;
  transition: background 0.2s;
}

.tool-btn:hover:not(:disabled) {
  background: var(--el-fill-color-light);
}

.tool-btn.active {
  background: var(--el-color-primary-light-9);
}

.tool-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.emoji-btn {
  font-size: 18px;
  line-height: 1;
}

.emoji-panel {
  position: absolute;
  left: 20px;
  bottom: calc(100% + 6px);
  z-index: 20;
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 2px;
  width: 340px;
  max-height: 220px;
  padding: 8px;
  overflow-y: auto;
  background: var(--el-bg-color-overlay, var(--el-bg-color));
  border: 1px solid var(--el-border-color-light);
  border-radius: 10px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
}

.emoji-item {
  padding: 4px;
  border: none;
  border-radius: 6px;
  background: transparent;
  font-size: 20px;
  line-height: 1.3;
  cursor: pointer;
}

.emoji-item:hover {
  background: var(--el-fill-color-light);
}

.hidden-file {
  display: none;
}

.image-bubble .chat-image {
  display: block;
  max-width: 220px;
  max-height: 240px;
  border-radius: 8px;
  cursor: zoom-in;
}

.image-bubble .chat-image :deep(img) {
  max-width: 100%;
  height: auto;
}

.composer-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.chat-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 头像 */
.avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  font-size: 16px;
}

.avatar.small {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  font-size: 13px;
}

.online-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--el-border-color-darker);
  display: inline-block;
}

.online-dot.online {
  background: var(--el-color-success);
}

/* 联系人弹窗 */
.peer-scroll {
  margin-top: 12px;
}

.peer-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 6px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
}

.peer-item:hover {
  background: var(--el-fill-color-light);
}

.peer-info {
  flex: 1;
  min-width: 0;
}

.peer-desc {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 可点击头像 */
.clickable-avatar {
  cursor: pointer;
  transition: opacity 0.2s;
}

.clickable-avatar:hover {
  opacity: 0.82;
}

/* 对方资料弹窗 */
.profile-dialog-body {
  min-height: 200px;
}

.profile-student {
  display: flex;
  gap: 20px;
}

.student-side {
  width: 320px;
  flex-shrink: 0;
  max-height: 66vh;
  padding-right: 8px;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.student-side h2 {
  margin: 0;
  font-size: 26px;
  color: var(--el-text-color-primary);
}

.student-edu {
  margin: 4px 0 14px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.s-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.s-line b {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.tag-block {
  margin-top: 14px;
}

.tag-block h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.tag-box {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.student-resume {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
}

.student-resume embed {
  width: 100%;
  height: 62vh;
  border: none;
}

.profile-company {
  max-height: 66vh;
  padding-right: 6px;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.company-head {
  display: flex;
  align-items: center;
  gap: 14px;
}

.company-head h2 {
  margin: 0;
  font-size: 22px;
  color: var(--el-text-color-primary);
}

.company-head p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.avatar.preview {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  font-size: 22px;
}

.company-intro {
  margin: 16px 0 6px;
  padding: 12px 14px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--el-text-color-regular);
}

.section-title {
  margin: 16px 0 10px;
  font-size: 15px;
  color: var(--el-text-color-primary);
}

.job-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.job-item {
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
}

.job-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.job-top strong {
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.job-time {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.job-desc {
  margin: 8px 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  word-break: break-word;
}

.job-foot {
  display: flex;
  align-items: center;
  gap: 10px;
}

@media (max-width: 720px) {
  .profile-student {
    flex-direction: column;
  }

  .student-side {
    width: 100%;
    max-height: 40vh;
  }

  .student-resume embed {
    height: 40vh;
  }
}
</style>
