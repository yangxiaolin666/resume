# 核心源码阅读指南

> 配套文档：[README.md](README.md)（全局原理与接口）、[CHAT_GUIDE.md](CHAT_GUIDE.md)（聊天实战教程）。
> 本文只回答一个问题：**哪些源码值得读、按什么顺序读、读到什么程度。**

---

## 0. 先看结论

| 指标 | 数量 |
|---|---|
| Go 文件 | 46 个 |
| 物理行数 | 10,202 行 |
| 去空行/注释后的有效代码 | 7,053 行 |
| 其中 protobuf 生成代码 | 478 行（不用读） |
| 手写有效代码 | 约 6,600 行 |
| **建议精读** | **约 1,200 行** |

不需要通读。先读 7 个必读文件（约 1,241 行，含 api 层），就能理解整个后端的骨架；其余代码在改具体功能时按链路查。

---

## 1. 优先级地图

| 模块 | 文件 | 有效行 | 优先级 | 理由 |
|---|---|---:|---|---|
| 启动 | [main.go](back/resume-go/main.go) | 74 | **必读** | 组件如何组装、依赖强弱、优雅关闭 |
| 配置 | [config.go](back/resume-go/internal/config/config.go) | 86 | **必读** | 所有环境变量与默认值 |
| 鉴权 | [auth.go](back/resume-go/internal/auth/auth.go) | 96 | **必读** | 手写 JWT + bcrypt，短小完整 |
| 路由 | [api.go](back/resume-go/internal/api/api.go) | 141 | **必读** | 路由分组、公开/受保护/管理员边界 |
| 中间件 | [middleware.go](back/resume-go/internal/api/middleware.go) | 87 | **必读** | 验签、无感续期、管理员校验、CORS |
| 聊天 | [hub.go](back/resume-go/internal/chat/hub.go) | 65 | **必读** | 在线连接注册中心 |
| 聊天 | [client.go](back/resume-go/internal/chat/client.go) | 121 | **必读** | 读写泵、心跳、并发安全 |
| 缓存 | [cache.go](back/resume-go/internal/cache/cache.go) | 334 | 选读（高价值） | 降级、防击穿/雪崩、版本号、在线计数 |
| 消息队列 | [mq.go](back/resume-go/internal/mq/mq.go) | 195 | 选读（高价值） | 可靠消费、手动 ACK、断线重连 |
| gRPC | [server.go](back/resume-go/internal/rpc/server.go) + [client.go](back/resume-go/internal/rpc/client.go) | 157 | 选读 | 进程内服务调用与优雅停止 |
| 模型 | [model.go](back/resume-go/internal/model/model.go) | 752 | 当字典查 | DTO、JSON 双命名、参数兼容 |
| 数据层 | `internal/data/*.go` | 1,806 | 改 SQL 时查 | 全部手写 SQL |
| 业务层 | `internal/service/*.go` | 1,738 | 按功能读 | 业务规则、事务、频控 |
| 接口层 | `internal/api/handlers_*.go` | 1,233 | 略读 | 多为参数绑定 + 调 service |
| 生成代码 | `internal/rpc/recruitpb/*` | 396 | **跳过** | protoc 生成 |

---

## 2. 必读模块

### 2.1 `main.go`（74 行）

**一句话**：整个后端的「组装车间」，看它就知道程序启动时发生了什么。

**启动顺序**（对应 `main()` 注释）：

1. `config.Load()` 读取配置；
2. `data.Open(...)` 连接 MySQL —— **强依赖，失败直接退出**；
3. 创建上传目录；
4. `cache.New(...)` 连接 Redis —— **弱依赖，失败自动降级为内存缓存**；
5. `mq.NewPublisher(...)` 连接 RabbitMQ —— 失败降级为跳过；
6. `rpc.NewClient(...)` 连接 gRPC —— 失败降级；
7. `service.New(...)` 创建业务层，同时创建 `chat.Hub`；
8. 注入上下线回调：`OnRegister = svc.HandleUserOnline`、`OnUnregister = svc.HandleUserOffline`；
9. 启动 gRPC 服务端（goroutine）；
10. 启动 RabbitMQ 消费者；
11. 启动 `http.Server`；
12. 监听 `SIGINT/SIGTERM`，给 HTTP 5 秒优雅关闭时间。

**注意两个细节**：

- `chatHub.OnMessage` 不在 `main.go` 注入，而在 WebSocket 握手时由 [handlers_chat.go](back/resume-go/internal/api/handlers_chat.go) 第 215 行附近注入；
- 所有后台 goroutine 共享同一个 `ctx`，`ctx.Done()` 一关，gRPC 与消费者都会退出。

**读完后应该能回答**：

- 为什么 Redis 挂了网站还能启动？
- 如果 MySQL 连不上，程序会怎样？
- 优雅关闭时，正在处理的 HTTP 请求会怎样？

---

### 2.2 `internal/config/config.go`（86 行）

**一句话**：所有可调参数集中在一处，容器部署只靠环境变量。

**结构**：

- `Config` 字段分四组：Server/DB、Redis、RabbitMQ、gRPC；
- `Load()` 用 `env` / `envBool` / `envInt` 读取环境变量并给默认值；
- `DSN()` 生成 MySQL 连接串（`parseTime=true&loc=Local`）；
- `StaticRoot()` 根据上传目录推导静态资源根目录。

**注意**：

- 默认值里包含数据库密码和 `JWT_SECRET`，生产必须用环境变量覆盖；
- `GRPC_INTERNAL` 默认指向本机 `127.0.0.1:<GRPC_PORT>`，容器里可以改成独立服务名。

**读完后应该能回答**：

- 不设置任何环境变量时，服务会连哪个数据库、哪个 Redis？
- 想关闭 Redis 但保留其他功能，应该改哪个变量？

---

### 2.3 `internal/auth/auth.go`（96 行）

**一句话**：不依赖第三方库，手写 HS256 JWT + bcrypt。

**关键内容**：

| 内容 | 说明 |
|---|---|
| `Claims` | `userId` / `username` / `userType` / `tokenType` / `exp` |
| `AccessTokenHours = 2` | Access Token 有效期 2 小时 |
| `RefreshTokenDays = 7` | Refresh Token 有效期 7 天 |
| `RefreshSoonWindow = 30m` | 剩余不足 30 分钟时中间件下发 `New-Token` |
| `Generate` | 拼 `header.payload.signature`，`json.Marshal` + Base64URL |
| `Parse` | 切三段 → 验签 → 解 payload → 校验 `exp` |
| `sign` | HMAC-SHA256，`hmac.Equal` 常量时间比较 |
| `HashPassword` / `CheckPassword` | bcrypt 生成与校验 |
| `TokenFromHeader` | 从 `Authorization: Bearer xxx` 中取 Token |

**为什么值得读**：

- 70 行左右把 JWT 三段结构、为什么先验签再解析、为什么用 `hmac.Equal` 都讲清楚了；
- 这是全项目最独立、最适合作为「读懂一个模块」起点的文件。

**常见坑**：

- `Parse` 只校验签名和过期，**不校验 `tokenType`**；`tokenType` 的业务约束在刷新逻辑里；
- `secret` 硬编码默认值，生产必须换。

**读完后应该能回答**：

- JWT 的三段分别是什么？签名覆盖哪两段？
- 为什么不能先解析 payload 再验签？
- Access Token 过期后，前端凭什么还能免登录？

---

### 2.4 `internal/chat/hub.go`（65 行）

**一句话**：在线连接的注册中心，负责「谁在线、给谁推」。

**数据结构**：

```go
clients map[string]map[*Client]struct{}
// key = "student:11" / "company:12"
// value = 该用户的所有连接（多标签页）
```

**方法**：

| 方法 | 行为 |
|---|---|
| `Register` | 加入连接；集合空则建 key；异步回调 `OnRegister` |
| `Unregister` | 只删当前连接；集合空才删 key；异步回调 `OnUnregister` |
| `SendToUser` | 序列化一次，`RLock` 遍历该用户所有连接推送 |
| `IsOnline` | 查本进程连接表 |

**为什么值得读**：

- `sync.RWMutex` + `map` 的典型并发写法；
- 通过回调注入业务，避免 `chat → service → chat` 循环依赖；
- 上下线回调是「异步 goroutine」，即使 Redis 慢也不会卡住 WebSocket 读写。

**读完后应该能回答**：

- 为什么 `clients` 的值是集合而不是单个连接？
- 为什么删除连接时要判断集合是否为空？
- Hub 为什么不直接调用 service？

---

### 2.5 `internal/chat/client.go`（121 行）

**一句话**：一个 WebSocket 连接对应两个 goroutine —— 一个只读，一个只写。

**常量**：

| 常量 | 值 | 作用 |
|---|---:|---|
| `writeWait` | 10s | 单次写超时 |
| `pongWait` | 60s | 读超时 |
| `pingPeriod` | 54s | 服务端 Ping 间隔（略小于 pongWait） |
| `maxMessageSize` | 4KB | 单条消息上限 |
| `sendBufferSize` | 128 | 发送队列容量 |

**`readPump`**：

- 设置读上限与读超时；
- Pong 到来时续期；
- 解析 JSON：`ping → pong`、`message → Hub.OnMessage`、其他 → 错误帧；
- 退出时按顺序清理：`Unregister` → `close(send)` → `conn.Close()`（顺序不能反）。

**`writePump`**：

- 唯一写出口，解决 gorilla/websocket **不允许并发写**的问题；
- 两个 `select` 分支：`send` 队列、54 秒 Ping 定时器；
- 队列满时非阻塞丢弃（消息已入库，重拉历史可恢复）。

**读完后应该能回答**：

- 为什么不能多个 goroutine 直接 `WriteMessage`？
- 两层心跳分别是谁发、解决什么问题？
- 客户端很慢时，服务端如何避免被拖垮？

---

## 3. 选读但高价值

### 3.1 `internal/cache/cache.go`（334 行）—— 缓存设计样板

**一句话**：Redis 可用就用 Redis，不可用自动切进程内 map，业务代码无感知。

**按主题读**：

| 主题 | 关键方法 | 设计点 |
|---|---|---|
| 降级 | `New` | Ping 失败打印日志并返回内存 Store，不阻塞启动 |
| 基础读写 | `Get` / `Set` / `Del` | `Set` 的 TTL 自动加最多 10% 随机抖动，防雪崩 |
| 防击穿 | `GetOrLoad[T]` | `singleflight` 让同 key 并发只回源一次 |
| 缓存一致性 | `JobVersion` / `BumpJobVersion` | 版本号拼进 key，写操作自增版本，旧 key 自然过期 |
| 收藏 | `AddFavorite` / `RemoveFavorite` / `BatchIsFavorite` | Redis Set，`SADD/SREM` |
| 在线状态 | `SetOnline` / `SetOffline` / `IsOnline` | `INCR/DECR` 计数 + 5 分钟 TTL，支持多标签页 |
| 幂等 | `MarkOnce` | `SETNX` 占位；Redis 故障返回 true，交给 MySQL 唯一键兜底 |

**重点读 `GetOrLoad`**：泛型 + singleflight + JSON 序列化的组合，是「读缓存 → 回源 → 写缓存」的标准模板。

**读完后应该能回答**：

- 缓存击穿、雪崩、穿透分别对应哪段代码？
- 为什么更新职位不直接删缓存，而是自增版本号？
- Redis 挂了以后，收藏和在线状态还能用吗？

---

### 3.2 `internal/mq/mq.go`（195 行）—— 消息队列样板

**一句话**：只负责「可靠地把事件送进/送出队列」，幂等逻辑不在这一层。

**关键内容**：

| 部分 | 说明 |
|---|---|
| `Event` | `eventId`、`type`、消息双方、内容、`createdAt` |
| 队列 | `resume.chat.notify`，`durable=true` |
| `Publisher` | 长连接 + channel + `sync.Mutex`（AMQP channel 不能并发写） |
| `Publish` | JSON、持久化消息、`MessageId=eventId` |
| `Consumer.Start` | 外层 `for` + 断线 5 秒重连 |
| `Qos(20)` | 每个消费者最多 20 条未 ACK |
| ACK 策略 | 成功 `Ack`；解析失败/处理失败 `Nack(requeue=false)`，直接丢弃 |

**重要区分**：真正的「三层幂等」在 [service/chat.go](back/resume-go/internal/service/chat.go) 的 `HandleChatNotify`（约 388 行）：

1. MySQL `INSERT IGNORE mq_consumed_event`；
2. Redis `MarkOnce`（24h）；
3. `chat_notification` 唯一键兜底。

且顺序是「先 MySQL 后 Redis」，避免先占 Redis 而库写失败导致永久丢通知。

**读完后应该能回答**：

- 为什么消费失败不重新入队？
- `autoAck=false` 和 `Qos(20)` 分别解决什么问题？
- 同一条消息被投递两次，最终会生成几条通知？

---

### 3.3 `internal/rpc`（157 行，不含 pb）—— 内部服务调用

**一句话**：把「查用户昵称/职位摘要」做成进程内 gRPC 服务，演示服务拆分的最小形态。

**关键内容**：

- `proto/recruit/v1/recruit.proto` 定义 `RecruitInternal`：
  - `GetUserBrief(user_id, user_type)`；
  - `GetJobBrief(job_id)`。
- `server.go`：`Serve(ctx, ":9091", db)`，日志拦截器，`GracefulStop()` + 5 秒兜底 `Stop()`。
- `client.go`：`grpc.NewClient` 懒连接，2 秒超时，失败返回 `ok=false`。
- 唯一实际调用点：离线通知里查发送者昵称，失败降级查 MySQL，再失败用 `"对方"`。

**读完后应该能回答**：

- 为什么 gRPC 挂了不影响主流程？
- 为什么客户端要设置 2 秒超时？
- 生成代码 `recruit.pb.go` / `recruit_grpc.pb.go` 为什么不用读？

---

## 4. 最小阅读路径（约 1,241 行）

按顺序读，读完即可掌握主流程：

| 步骤 | 文件 | 有效行 | 目标 |
|---:|---|---:|---|
| 1 | `main.go` | 74 | 知道启动了哪些组件 |
| 2 | `internal/config/config.go` | 86 | 知道配置从哪来 |
| 3 | `internal/auth/auth.go` | 96 | 理解 JWT 与密码 |
| 4 | `internal/api/api.go` + `middleware.go` | 228 | 理解路由与鉴权边界 |
| 5 | `internal/chat/hub.go` + `client.go` | 186 | 理解 WebSocket 连接管理 |
| 6 | `internal/service/chat.go` | 352 | 理解聊天业务规则 |
| 7 | `internal/data/chat.go` | 219 | 理解会话/历史/未读 SQL |

完成后如果还想深入中间件设计，再补：

| 步骤 | 文件 | 有效行 |
|---:|---|---:|
| 8 | `internal/cache/cache.go` | 334 |
| 9 | `internal/mq/mq.go` | 195 |
| 10 | `internal/rpc/server.go` + `client.go` | 157 |

---

## 5. 其余代码怎么查：垂直切片

不要按目录通读，按「一个功能一条链」查：

| 要改什么 | 调用链 | 大约行数 |
|---|---|---:|
| 登录 / 改密 / 昵称 | `handlers_login.go` → `service/user.go` → `data/user.go` → `auth` | 300–450 |
| 学生画像 / 简历 | `handlers_student.go` → `service/student.go` → `data/student.go` | 800–1,000 |
| 职位发布 / 列表 | `handlers_job.go` → `service/job.go` → `data/job.go` | 600–800 |
| 收藏 | `handlers_like.go` → `service/like.go` → `data/like.go` | 300–400 |
| 学生检索 | `handlers_search.go` → `service/search.go` → `data/search.go` | 400–500 |
| 简历投递 | `handlers_application.go` → `service/application.go` → `data/application.go` | 300–400 |
| 后台管理 / Excel | `handlers_admin.go` → `service/admin.go` → `data/admin.go` | 900–1,100 |
| 鉴权 / 续期 | `middleware.go` → `auth.go` | 180 |

**查法**：先看 handler 的请求/响应，再看 service 的业务校验，最后看 data 的 SQL。三层职责非常固定，读一条就懂一类。

---

## 6. 可以放心跳过的部分

| 内容 | 原因 |
|---|---|
| `internal/rpc/recruitpb/*.pb.go` | protoc 生成，读 `.proto` 即可 |
| `handlers_*.go` 中的 `ShouldBindJSON`、参数默认值 | 模式重复 |
| `model.go` 全部 752 行 | 当字典查字段即可 |
| `seed_accounts.sql`、`seed_statistics.sql` | 演示数据 |
| `schema.sql` 的索引/建表细节 | 改表时再看 |
| `uploads/` 下的示例文件 | 运行产物 |

---

## 7. 建议动手做的 3 个实验

1. **关 Redis 启动**
   设置 `REDIS_ENABLED=false`，观察日志降级；再测试职位列表、收藏、在线状态是否仍可用。
   → 验证「弱依赖降级」不是纸面设计。

2. **改心跳参数**
   把 `client.go` 的 `pingPeriod` 改小，同时把前端 `chatSocket.js` 的 25s 调小，观察重连与 `pong` 行为。
   → 理解两层心跳的配合。

3. **人为制造重复事件**
   在 `HandleChatNotify` 里临时重复调用一次（或手动向队列发同一 `eventId`），确认只生成一条通知。
   → 理解 MySQL → Redis → 唯一键的三层幂等。

---

## 8. 阅读检查清单

读完必读部分后，自查是否能解释：

- [ ] 为什么 MySQL 是强依赖，Redis / RabbitMQ / gRPC 是弱依赖？
- [ ] JWT 的三段结构、验签顺序、`hmac.Equal` 的意义
- [ ] Access Token 与 Refresh Token 的有效期与使用场景
- [ ] 一个用户多标签页时，Hub 与在线计数如何工作
- [ ] 为什么 WebSocket 只能有一个写 goroutine
- [ ] 会话列表如何用一条 SQL 算出「最后一条消息 + 未读数」
- [ ] 历史消息为什么用 `id < beforeId` 而不是 OFFSET
- [ ] 缓存击穿 / 雪崩 / 一致性分别怎么解决
- [ ] MQ 消费失败为什么丢弃而不是重试
- [ ] 离线通知的三层幂等分别在哪
- [ ] gRPC 调用失败时如何降级
- [ ] 服务退出时如何优雅关闭

---

## 9. 与 README 章节对照

| 想了解 | 看 README |
|---|---|
| 技术栈 / 架构 | 第 3、4 章 |
| 全部接口 | 第 11 章 |
| 聊天原理 | 第 12 章 |
| 缓存 / MQ / gRPC | 第 18、19、20 章 |
| 部署 | 第 22 章 |
| 已知问题 | 第 23 章 |

---

**一句话总结**：先读 1,241 行的主链路，再按功能查垂直切片；`cache`、`mq`、`rpc` 是加分项，`recruitpb` 和绝大多数 handler 样板可以直接跳过。
