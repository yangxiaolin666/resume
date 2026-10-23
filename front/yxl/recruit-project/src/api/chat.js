import request from '@/utils/request'

// 聊天相关的 REST 接口。
// 实时消息走 WebSocket，这里的接口主要负责：
//   1. 拉取会话列表 / 历史消息 / 可聊天的人
//   2. WebSocket 断开时的降级发送
//   3. 标记已读
export const getChatPeers = (keyword = '') =>
  request.get('/chat/peers', { params: { keyword } })

export const getConversations = () => request.get('/chat/conversations')

export const getChatMessages = (params) =>
  request.get('/chat/messages', { params })

// 点击聊天头像查看对方资料（学生=简历，企业=简介 + 在招岗位）
export const getPeerProfile = (peerId, peerType) =>
  request.get('/chat/profile', { params: { peerId, peerType } })

export const sendChatMessage = (data) => request.post('/chat/messages', data)

// 聊天图片上传：字段名必须是 image
export const uploadChatImage = (file) => {
  const formData = new FormData()
  formData.append('image', file)
  return request.post('/chat/uploadImage', formData)
}

export const markChatRead = (data) => request.post('/chat/read', data)
