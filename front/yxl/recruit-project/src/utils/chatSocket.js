import { API_BASE_URL } from '@/utils/request'

// ==================================================================
// 浏览器端 WebSocket 封装
// ==================================================================
// 参考了两个示例项目里的 ws.js / websocket-heartbeat.js，做了简化：
//   1. 自动重连（指数退避：1s -> 2s -> 4s ... 最大 15s）
//   2. 应用层心跳：每 25s 发送 {type:'ping'}，服务端回 {type:'pong'}
//   3. 超过 35s 没收到 pong，认为连接已死，主动关闭并重连
//   4. 对外只暴露 onMessage / onStatus 两个回调
//
// 注意：浏览器原生 WebSocket 不能自定义 Authorization 请求头，
// 所以 token 通过 URL 查询参数传给后端。
// ==================================================================

const HEARTBEAT_INTERVAL = 25000
const HEARTBEAT_TIMEOUT = 35000

function buildWsUrl(token) {
  // API_BASE_URL 可能是绝对地址（http://127.0.0.1:8081），
  // 也可能是开发环境的代理前缀（/api-backend）。
  // 相对地址需要补上当前页面 origin，再拼 WebSocket 路径。
  let base = API_BASE_URL
  if (!/^https?:\/\//.test(base)) {
    base = window.location.origin + (base.startsWith('/') ? base : `/${base}`)
  }
  const wsBase = base.replace(/^http/, 'ws')
  return `${wsBase}/chat/ws?token=${encodeURIComponent(token)}`
}

export class ChatSocket {
  constructor({ token, onMessage, onStatus }) {
    this.token = token
    this.onMessage = onMessage || (() => {})
    this.onStatus = onStatus || (() => {})

    this.ws = null
    this.connected = false
    this.manualClose = false
    this.heartbeatTimer = null
    this.reconnectTimer = null
    this.reconnectDelay = 1000
    this.lastPongAt = 0
  }

  connect() {
    if (!this.token) {
      this.onStatus('closed')
      return
    }
    this.manualClose = false
    this.onStatus('connecting')

    try {
      this.ws = new WebSocket(buildWsUrl(this.token))
    } catch (e) {
      this.onStatus('closed')
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      this.connected = true
      this.lastPongAt = Date.now()
      this.reconnectDelay = 1000
      this.onStatus('open')
      this.startHeartbeat()
    }

    this.ws.onmessage = (event) => {
      let msg
      try {
        msg = JSON.parse(event.data)
      } catch (e) {
        return
      }
      // pong 只用于心跳，不交给业务层
      if (msg.type === 'pong') {
        this.lastPongAt = Date.now()
        return
      }
      this.onMessage(msg)
    }

    this.ws.onclose = () => {
      this.connected = false
      this.stopHeartbeat()
      this.onStatus('closed')
      if (!this.manualClose) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = () => {
      if (this.ws) {
        this.ws.close()
      }
    }
  }

  startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      if (!this.connected) return
      // 长时间没收到 pong，说明连接实际上已经失效
      if (Date.now() - this.lastPongAt > HEARTBEAT_TIMEOUT) {
        this.ws && this.ws.close()
        return
      }
      this.send({ type: 'ping' })
    }, HEARTBEAT_INTERVAL)
  }

  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  scheduleReconnect() {
    if (this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, 15000)
      this.connect()
    }, this.reconnectDelay)
  }

  send(payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(payload))
      return true
    }
    return false
  }

  // 发送聊天消息（msgType: text / image / resume）
  sendMessage(receiverId, receiverType, content, msgType = 'text') {
    return this.send({
      type: 'message',
      receiverId: Number(receiverId),
      receiverType,
      content,
      msgType
    })
  }

  close() {
    this.manualClose = true
    this.stopHeartbeat()
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.connected = false
    this.onStatus('closed')
  }
}
