# 学生 / 企业聊天系统学习指南

这篇文档按“从数据库到浏览器”的顺序，讲解本次新增的聊天系统。你可以把它当成一次完整的 WebSocket 实战教程：先看整体结构，再按目录顺序读代码。

> 参考项目：
> - `Go-project-1 - 副本`：Hub / Client 连接管理、私聊点对点推送、回调注入
> - `go-gin-chat-master`：Gin + WebSocket 私聊、心跳检测、消息落库
>
> 本项目借鉴了它们的核心思想，但适配了现有的 **net/http + database/sql + JWT + Vue 3** 技术栈，没有引入 gRPC、RabbitMQ、GORM 这些额外组件。

---

## 1. 功能范围

- 学生 ↔ 企业 一对一私聊（管理员不参与）
- 会话列表：最后一条消息、未读数、在线状态
- 历史消息分页（向上加载更早的消息）
- 实时收发（WebSocket）
- 断线自动重连 + 心跳
- 未读消息标记已读
- 学生可以从招聘卡片直接“联系企业”
- 企业可以从学生列表 / 收藏列表直接“联系学生”

---

## 2. 这次补充的技术栈

| 技术 | 作用 | 为什么选它 |
|---|---|---|
| `github.com/gorilla/websocket` | Go 的 WebSocket 实现 | 生态成熟，和 `net/http` 直接集成，参考资料多 |
| WebSocket | 浏览器与后端的长连接 | HTTP 只能“请求-响应”，聊天需要服务端主动推送 |
| MySQL `chat_message` 表 | 消息持久化 | 离线消息、历史记录都依赖数据库；复用项目已有 MySQL |
| JWT（已有） | WebSocket 握手鉴权 | 复用登录 Token，不新增一套账号体系 |
| `sync.RWMutex` + `map` | 在线连接管理 | Go 并发基础：多 goroutine 读写共享 map 必须加锁 |
| `chan []byte` | 每个连接的发送队列 | gorilla/websocket 不允许并发写，所有写集中到一个 goroutine |
| Vue 3 `class` 封装 | 前端 WebSocket 客户端 | 把重连、心跳、事件订阅封装起来，页面只关心业务 |
| Element Plus | 聊天 UI | 项目已在用，减少样式成本 |

### HTTP 与 WebSocket 的分工

本项目采用 **REST + WebSocket 混合**：

- **REST**：登录、会话列表、历史消息、标记已读、断线时的降级发送
- **WebSocket**：新消息实时推送、在线状态、心跳

这样做的原因：

1. 历史消息天然适合分页查询，用 HTTP 更简单；
2. WebSocket 只负责“实时”，职责单一，断线也不影响历史记录；
3. WebSocket 断开时，前端还能用 REST 发消息，不会直接丢消息。

---

## 3. 目录结构

### 后端

```text
back/resume-go/
├── internal/
│   ├── chat/
│   │   ├── hub.go              # 连接注册中心：谁在线、给谁推送
│   │   └── client.go           # 单个连接：读循环、写循环、心跳
│   ├── model/
│   │   └── chat.go             # 聊天相关结构体
│   ├── data/
│   │   └── chat.go             # 聊天 SQL：消息、会话、联系人
│   ├── service/
│   │   └── chat.go             # 聊天业务规则：谁能发给谁、内容校验
│   └── api/
│       └── handlers_chat.go    # REST 接口 + WebSocket 升级
├── schema.sql                  # chat_message 建表语句
└── seed_accounts.sql           # x1 / q1 / g1 演示账号
```

### 前端

```text
front/yxl/recruit-project/src/
├── api/chat.js                 # 聊天 REST 接口
├── utils/chatSocket.js         # WebSocket 客户端封装（重连 + 心跳）
├── views/chat/index.vue        # 聊天页面
├── router/index.js             # /chat 路由
├── views/layout/index.vue      # 顶部“消息”入口
├── components/jobCard.vue      # 学生“联系企业”
├── components/studentCard.vue  # 企业“联系学生”（收藏页）
└── views/layout/home.vue       # 企业学生列表“联系学生”列
```

---

## 4. 数据库设计

表结构在 `schema.sql`：

```sql
CREATE TABLE chat_message (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  sender_id     INT NOT NULL,
  sender_type   VARCHAR(20) NOT NULL,   -- student / company
  receiver_id   INT NOT NULL,
  receiver_type VARCHAR(20) NOT NULL,
  content       TEXT NOT NULL,
  is_read       TINYINT NOT NULL DEFAULT 0,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_chat_pair  (sender_id, sender_type, receiver_id, receiver_type, id),
  INDEX idx_chat_inbox (receiver_id, receiver_type, is_read, id)
);
```

### 为什么用 `user_id + user_type`

`user` 表里学生、企业、管理员都是用户，`user_id` 全局唯一。但聊天时必须知道“这个 ID 是学生还是企业”，否则查询姓名、校验角色都要额外判断。用两个字段一起标识，逻辑最清晰。

### 为什么不建两张表

一张表就能表达所有私聊消息：

```sql
WHERE (sender_id = A AND receiver_id = B)
   OR (sender_id = B AND receiver_id = A)
```

这比 `student_message` / `company_message` 两张表简单得多，也不会出现跨表排序问题。

### 索引怎么想

- `idx_chat_pair`：查询“我和某人的聊天记录”，按双方 ID + 自增 ID 排序；
- `idx_chat_inbox`：查询“发给我的未读消息”，用于统计未读数。

聊天查询最常见的两个方向就是“某个会话的历史”和“我的未读”，索引与之一一对应。

---

## 5. 一条消息的完整旅程

```text
学生浏览器                          Go 后端                            企业浏览器
    |                                 |                                  |
    | 1. 登录拿到 JWT                  |                                  |
    | 2. ws://.../chat/ws?token=xxx   |                                  |
    |-------------------------------->| 3. 解析 token，取出 userId/type   |
    |                                 | 4. Hub.Register(client)          |
    |                                 |                                  |
    | 5. {"type":"message",           |                                  |
    |     "receiverId":12,            |                                  |
    |     "receiverType":"company",   |                                  |
    |     "content":"你好"}           |                                  |
    |-------------------------------->| 6. service 校验角色/内容/接收者   |
    |                                 | 7. INSERT chat_message           |
    |                                 | 8. 查询完整消息（含双方名称）     |
    |                                 | 9. Hub.SendToUser(接收者)         |
    |                                 |----------------------------->    |
    |                                 | 10. Hub.SendToUser(发送者，回显)  |
    |<--------------------------------|                                  |
```

关键点：

- 消息 **先落库，再推送**。这样即使对方离线，消息也不会丢；
- 同一条消息推给双方：接收者实时看到，发送者用于多标签页同步；
- 前端按消息 `id` 去重，避免“REST 响应 + WebSocket 回显”重复上屏。

---

## 6. 后端代码逐层讲解

### 6.1 连接管理中心：`internal/chat/hub.go`

Hub 只做三件事：

1. `Register` / `Unregister`：上线、下线；
2. `SendToUser`：给某个用户的所有连接推送；
3. `IsOnline`：给前端显示在线状态。

内部结构：

```go
clients map[string]map[*Client]struct{}
// key = "student:11" / "company:12"
// value = 这个用户的所有连接（多标签页）
```

学习重点：

- `sync.RWMutex`：读多写少，用读写锁；
- 一个用户允许多个连接，所以值是集合而不是单个连接；
- Hub 不 import service，而是暴露 `OnMessage` 回调，由 API 层注入业务逻辑，避免循环依赖。

### 6.2 单个连接：`internal/chat/client.go`

每个连接两个 goroutine：

- `readPump`：读客户端消息，处理 `ping` / `message`；
- `writePump`：唯一写出口，负责发消息和定时 Ping。

为什么必须只有一个写 goroutine？

> gorilla/websocket 明确要求：**同一个连接不能并发写**，否则会 panic。
> 所以业务层只把数据放进 `send chan []byte`，真正的 `WriteMessage` 全部在 `writePump` 里执行。

心跳有两层：

1. **协议层**：服务端每 54 秒发 Ping 帧，浏览器自动回 Pong，`readPump` 刷新 60 秒读超时；
2. **应用层**：前端每 25 秒发 `{"type":"ping"}`，服务端回 `{"type":"pong"}`，前端据此判断连接是否存活。

慢客户端保护：

```go
select {
case c.send <- data:
default:
    // 队列满就丢弃，避免一个慢连接拖垮整个 Hub
}
```

### 6.3 业务规则：`internal/service/chat.go`

发送前做完整校验：

1. 发送者必须是 `student` / `company`；
2. 接收者必须是相反角色（学生只能找企业，企业只能找学生）；
3. 不能发给自己；
4. 内容去空格后不能为空，最多 800 个字符；
5. 接收者必须真实存在（查 `user` 表），防止伪造 `receiverId`；
6. 入库后推送双方。

这层不关心 SQL 细节，也不关心 WebSocket 怎么发，只写“规则”。

### 6.4 数据访问：`internal/data/chat.go`

四个核心查询：

| 方法 | 作用 | 关键 SQL |
|---|---|---|
| `ListChatMessages` | 某个会话的历史消息 | 双向条件 + `id < beforeId` + 倒序取 N 条再翻转 |
| `ListConversations` | 会话列表 | 用 CASE 把收发统一成 peer，按 peer 分组取 MAX(id) |
| `ListChatPeers` | 可聊天的人 | 学生查 company，企业查 student |
| `MarkChatRead` | 标记已读 | 把“对方发给我的未读”更新为 1 |

会话列表的 SQL 是本模块最值得读的一段：

```sql
SELECT
  CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS peer_id,
  CASE WHEN sender_id = ? THEN receiver_type ELSE sender_type END AS peer_type,
  MAX(id) AS max_id,
  SUM(CASE WHEN receiver_id = ? AND is_read = 0 THEN 1 ELSE 0 END) AS unread
FROM chat_message
WHERE (sender_id = ?) OR (receiver_id = ?)
GROUP BY peer_id, peer_type
```

它一次性算出了：每个联系人的最后一条消息 ID、未读数量。

### 6.5 HTTP 与 WebSocket 入口：`internal/api/handlers_chat.go`

REST：

- `GET /chat/peers`
- `GET /chat/conversations`
- `GET /chat/messages`
- `POST /chat/messages`
- `POST /chat/read`

WebSocket：

- `GET /chat/ws?token=xxx`

为什么 Token 放 URL？

> 浏览器原生 `new WebSocket(url)` 不支持自定义请求头，无法像 axios 那样带 `Authorization`。
> 所以握手时用 query 参数传 JWT，后端在 `Upgrade` 之前完成校验。

安全提示：生产环境应使用 `wss://`，并尽量使用短有效期 Token；URL 可能被日志记录。

鉴权中间件里把 `/chat/ws` 排除：

```go
protected = /com/ /stu/ /admin/ /chat/
excluded  = /login /refresh-token /stu/statistics /chat/ws
```

其他 `/chat/*` 接口仍然走正常的 `Authorization: Bearer`。

---

## 7. WebSocket 消息协议

### 客户端 → 服务端

发送聊天消息：

```json
{
  "type": "message",
  "receiverId": 12,
  "receiverType": "company",
  "content": "你好，我想了解一下这个岗位"
}
```

应用层心跳：

```json
{ "type": "ping" }
```

### 服务端 → 客户端

新消息：

```json
{
  "type": "message",
  "data": {
    "id": 1,
    "senderId": 11,
    "senderType": "student",
    "receiverId": 12,
    "receiverType": "company",
    "content": "你好，我想了解一下这个岗位",
    "isRead": false,
    "createdAt": "2026-10-21 21:43:27",
    "senderName": "x1",
    "receiverName": "q1公司"
  }
}
```

心跳响应：

```json
{ "type": "pong" }
```

连接成功：

```json
{ "type": "connected", "userId": 11, "userType": "student" }
```

错误：

```json
{ "type": "error", "message": "消息内容不能为空" }
```

---

## 8. 前端代码讲解

### 8.1 `src/utils/chatSocket.js`

这是一个 class，职责：

1. 建立连接：`ws://后端/chat/ws?token=...`
2. 自动重连：1s → 2s → 4s → ... 最大 15s
3. 心跳：25 秒 ping 一次，35 秒没收到 pong 就主动断开重连
4. 事件订阅：`onMessage` / `onStatus`

页面只需要：

```js
socket = new ChatSocket({
  token: userStore.token,
  onMessage: handleSocketMessage,
  onStatus: (status) => (connectionStatus.value = status)
})
socket.connect()
```

### 8.2 `src/views/chat/index.vue`

页面状态：

- `conversations`：左侧会话列表
- `current`：当前聊天对象
- `messages`：当前会话消息
- `peers`：发起聊天弹窗里的联系人
- `connectionStatus`：实时连接状态

关键逻辑：

1. **打开页面**：加载会话、联系人、建立 WebSocket；
2. **选择会话**：加载历史消息 → 标记已读 → 清空未读角标；
3. **收到消息**：
   - 判断是否是当前会话；
   - 是则追加并滚动到底部；
   - 不是则刷新会话列表，让未读数 +1；
   - 按 `id` 去重（REST 响应与 WS 回显可能同时到达）；
4. **发送消息**：优先 WebSocket；断线时降级为 REST；
5. **向上加载**：`beforeId = messages[0].id`，插入后保持滚动位置。

### 8.3 路由与入口

- `/chat` 路由注册在 `router/index.js`
- 顶部导航“消息”在 `views/layout/index.vue`
- 学生“联系企业”在 `components/jobCard.vue`
- 企业“联系学生”在 `views/layout/home.vue` 和 `components/studentCard.vue`

入口统一通过 query 传参：

```text
/chat?peerId=12&peerType=company&name=q1公司
```

聊天页启动时读取 query，自动打开对应会话。

---

## 9. 本地运行与双账号测试

一键启动：

```text
双击 d:\resume\start-all.bat
```

启动内容：

- MySQL（3306 或本地测试库 3307）
- Go 后端 `http://localhost:8081`
- 主前端 `http://127.0.0.1:5173`
- 管理端 `http://127.0.0.1:5174/admin`

演示账号：

| 身份 | 账号 | 密码 |
|---|---|---|
| 学生 | x1 | x1 |
| 企业 | q1 | q1 |
| 管理员 | g1 | g1 |

### 如何测试“实时”效果

浏览器的 `localStorage` 是同一个域名共享的，同一个普通窗口里不能同时登录两个账号。推荐：

1. 打开一个**普通窗口**登录 x1；
2. 打开一个**无痕/隐私窗口**登录 q1；
3. 两个窗口都进入“消息”；
4. 任意一方发送消息，另一方应立即出现，不需要刷新。

如果对方不在线，消息仍然会写入数据库；对方下次打开聊天页会从历史记录中看到。

---

## 10. 与两个参考项目的对应关系

| 参考项目 | 它的做法 | 本项目 |
|---|---|---|
| `Go-project-1` | Hub + Client + 房间/私聊；gRPC + RabbitMQ 异步落库 | 保留 Hub/Client/回调注入；直接用 MySQL 同步落库，适合单机学习 |
| `Go-project-1` | GORM | 使用项目已有的 `database/sql`，SQL 更直观 |
| `go-gin-chat` | Gin + 全局 channel + 心跳 | 使用 `net/http` + 每个连接独立读写 goroutine |
| `go-gin-chat` | HTML 模板 + jQuery | Vue 3 + Element Plus，组件化 |
| 两者 | 群聊 / 房间 | 只做学生 ↔ 企业私聊，贴合招聘业务 |

---

## 11. 常见坑与排查

### 1）并发写 WebSocket 导致 panic

错误做法：多个 goroutine 直接 `conn.WriteMessage`。
正确做法：所有写操作进入 `send` channel，由 `writePump` 单线程写。

### 2）Hub map 并发读写

必须用 `sync.RWMutex`；不要在持锁时做耗时操作（例如写数据库）。

### 3）连接关闭顺序

`readPump` 退出时：

1. `Hub.Unregister`
2. `close(send)`
3. `conn.Close()`

如果先 close channel 再 unregister，可能向已关闭的 channel 发送数据。

### 4）慢客户端

`send` 队列满时直接丢弃并打日志，避免整个 Hub 被一个慢连接卡住。

### 5）Token 过期

前端请求返回 401 时会提示；WebSocket 握手失败则会自动重连。生产环境建议在重连前先调用 `/refresh-token`。

### 6）XSS

Vue 的 `{{ content }}` 会自动转义，所以消息按纯文本渲染是安全的。**不要**把消息改成 `v-html`，除非你自己做了内容清洗。

---

## 12. 进阶练习（建议按顺序做）

1. **消息去重**：客户端生成 `clientMsgId`，数据库加唯一索引，防止网络抖动导致重复发送。
2. **已读回执**：对方读取后，通过 WebSocket 推一个 `{type:"read", peerId}`，发送方显示“已读”。
3. **图片/文件消息**：给 `chat_message` 加 `message_type`、`file_url` 字段，前端加图片预览。
4. **顶部未读总数**：在导航“消息”旁显示未读总数，需要新增一个 `GET /chat/unread` 接口。
5. **多实例部署**：单机 Hub 只在当前进程有效。多台后端时需要 Redis Pub/Sub 或消息队列广播，参考 `Go-project-1` 的 MQ 设计。
6. **消息撤回/删除**：加 `recalled_at` 字段，推送 `{type:"recall", id}`。
7. **限流与防刷**：按用户限制每秒消息数，避免恶意刷屏。
8. **Nginx 代理 WebSocket**：生产环境需要在 Nginx 增加：

```nginx
location /chat/ws {
    proxy_pass http://127.0.0.1:8081;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 3600s;
}
```

---

## 13. 推荐阅读顺序

1. `schema.sql` 里的 `chat_message`：先理解数据长什么样；
2. `internal/chat/hub.go`：理解在线连接怎么管理；
3. `internal/chat/client.go`：理解读写 goroutine 和心跳；
4. `internal/service/chat.go`：理解业务规则；
5. `internal/data/chat.go`：理解会话列表和历史消息 SQL；
6. `internal/api/handlers_chat.go`：理解 REST 与 WebSocket 如何接入；
7. `src/utils/chatSocket.js`：理解前端如何封装长连接；
8. `src/views/chat/index.vue`：理解页面状态与消息流；
9. 用两个浏览器窗口实际收发一次。

读完后，建议先做“消息去重”和“已读回执”两个练习，它们最贴近真实聊天系统，也最能巩固 WebSocket + 数据库的配合。
