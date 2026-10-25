# 校园招聘与简历投递平台（Resume Recruit Platform）

一个面向「学生 — 企业 — 管理员」三端的校园招聘系统，包含学生画像与简历、企业职位发布、学生检索与收藏、简历投递、实时聊天、数据看板、后台批量管理等功能。

仓库内同时存在 **两套可互相替代的后端实现**（Java / Go）与 **两套前端应用**（学生+企业主站 / 管理后台）。两套后端共用同一套数据库、同一套 JWT 约定和同一套 HTTP 路径，前端可以无缝切换。

> 本文档以**代码与 SQL 文件为唯一事实来源**，覆盖：技术栈、架构、全部接口、鉴权原理、聊天原理、数据库设计、缓存/消息队列/gRPC、部署方式、已知问题。读完后应能独立启动、调试和二次开发本项目。

---

## 目录

1. [项目概览](#1-项目概览)
2. [功能矩阵](#2-功能矩阵)
3. [技术栈](#3-技术栈)
4. [系统架构](#4-系统架构)
5. [目录结构](#5-目录结构)
6. [快速开始](#6-快速开始)
7. [配置说明](#7-配置说明)
8. [认证与鉴权原理](#8-认证与鉴权原理)
9. [统一响应与错误码](#9-统一响应与错误码)
10. [数据库设计](#10-数据库设计)
11. [接口文档（全量）](#11-接口文档全量)
12. [聊天系统原理（重点）](#12-聊天系统原理重点)
13. [简历与文件存储](#13-简历与文件存储)
14. [搜索、分页与查询原理](#14-搜索分页与查询原理)
15. [收藏模块原理](#15-收藏模块原理)
16. [简历投递模块](#16-简历投递模块)
17. [Excel 批量导入原理](#17-excel-批量导入原理)
18. [Redis 缓存设计](#18-redis-缓存设计)
19. [gRPC 内部服务](#19-grpc-内部服务)
20. [RabbitMQ 离线通知](#20-rabbitmq-离线通知)
21. [前端实现原理](#21-前端实现原理)
22. [部署](#22-部署)
23. [已知问题与注意事项](#23-已知问题与注意事项)
24. [二次开发指南](#24-二次开发指南)
25. [附录：接口速查表](#25-附录接口速查表)

---

## 1. 项目概览

| 维度 | 说明 |
|---|---|
| 业务定位 | 校园招聘平台：学生维护画像/简历并投递，企业检索学生、发布职位、收藏与沟通，管理员维护基础数据 |
| 角色 | `student`（学生）、`company`（企业）、`admin`（管理员），对应 `user.user_type` |
| 学生身份标识 | 登录名 = 学号（`user.username`），学生档案主键 = `student.id`，学号列 = `student.stu_id` |
| 企业身份标识 | 登录名 = 企业名（`user.username`），企业档案主键 = `company.id` |
| 聊天标识 | 一律使用 `(user_id, user_type)` 二元组标识一个人 |
| 两套后端 | `back/resume`（Java 17 + Spring Boot 3）、`back/resume-go`（Go + Gin），默认都监听 `8081`，**二选一运行** |
| 两套前端 | `front/yxl/recruit-project`（学生+企业，Vite 5173）、`front/ljk/adminPanel`（管理端，Vite 5174/`/admin/`） |
| 数据库 | MySQL 8（`resume` 库），两套后端共用 |

### 1.1 两套后端的关系

- **Java 版**（`back/resume`）：项目早期实现，覆盖登录、学生画像、简历上传、职位、收藏、检索、后台管理；**没有聊天与投递**。
- **Go 版**（`back/resume-go`）：功能超集重写，覆盖 Java 版全部接口，并新增**实时聊天（WebSocket）**、**简历投递**、**用户中心**、**Redis 缓存**、**RabbitMQ 离线通知**、**gRPC 内部服务**、**Docker/K8s 部署**。
- 两者**刻意保持兼容**：
  - JWT claims 字段完全一致（`userId` / `username` / `userType` / `tokenType` / `exp`），同一密钥签发，Token 可互认；
  - HTTP 路径一致（`/login`、`/stu/**`、`/com/**`、`/admin/**`）；
  - 统一响应体一致（`{code, message, data}`）；
  - Go 版额外做了**驼峰/下划线双写**，以同时兼容 Java 版前端与自身 DTO。
- 一键启动脚本 [start-all.bat](file:///d:/resume/start-all.bat) 默认启动的是 **Go 版**；`deploy.ps1` 部署的是 **Java 版**（打包 jar 后 systemd 重启）。

```
                  ┌──────────────────────────────┐
                  │  前端 A：学生/企业主站        │
                  │  Vue3 + Element Plus + Pinia  │
                  └──────────────┬───────────────┘
                                 │ HTTP(REST) + WebSocket
                  ┌──────────────▼───────────────┐
                  │  前端 B：管理后台 /admin/      │
                  │  Vue3 + Element Plus          │
                  └──────────────┬───────────────┘
                                 │
        ┌────────────────────────▼────────────────────────┐
        │  API 层（二选一，均 :8081）                       │
        │  A. Spring Boot 3.1.3 / Java 17                  │
        │  B. Go 1.27 / Gin（含 /chat/ws、gRPC :9091）      │
        └───┬──────────────┬──────────────┬───────────────┘
            │              │              │
       ┌────▼────┐   ┌─────▼─────┐  ┌─────▼──────┐
       │ MySQL 8 │   │  Redis    │  │ RabbitMQ   │
       │ 唯一真源 │   │ 缓存/在线 │  │ 离线通知   │
       └─────────┘   └───────────┘  └────────────┘
```

---

## 2. 功能矩阵

| 模块 | 学生 | 企业 | 管理员 | 实现后端 |
|---|---|---|---|---|
| 登录 / 刷新 Token / 校验 Token | ✅ | ✅ | ✅ | Java + Go |
| 查看/修改自己的账号资料、改昵称 | ✅ | ✅ | — | Go |
| 学生完整画像（9 类信息 + 6 张子表） | ✅ | 只读检索 | 只读检索 | Java + Go |
| 简历上传 / 最新简历 | ✅ | — | — | Java + Go |
| 简历模板制作并导出 PDF | ✅（纯前端） | — | — | 前端 |
| 职位发布 / 修改 / 上下架 / 删除 | — | ✅ | — | Java + Go |
| 职位列表 / 详情 | ✅ | ✅ | — | Java + Go |
| 学生多条件检索（含完整画像） | — | ✅ | ✅ | Java + Go |
| 收藏学生 / 取消收藏 / 判断收藏 | — | ✅ | — | Java + Go |
| 简历投递 / 撤回 / 状态流转 | ✅ | ✅ | — | Go |
| 实时聊天（文本/图片/简历） | ✅ | ✅ | —（禁止） | Go |
| 离线消息通知 / 未读数 | ✅ | ✅ | — | Go |
| 统计数据看板 | 公开 | 公开 | — | Java + Go |
| Excel 批量导入学生/企业 | — | — | ✅ | Java + Go |
| 批量删除/更新学生/企业 | — | — | ✅ | Java + Go |
| 语言/证书/技能字典维护 | — | — | ✅ | Java + Go |
| 管理员重置学生密码 | — | — | ✅ | Java + Go |

---

## 3. 技术栈

### 3.1 前端 A：学生 + 企业主站（`front/yxl/recruit-project`）

| 类别 | 技术 | 版本 | 用途 |
|---|---|---|---|
| 框架 | Vue | ^3.5.26 | Composition API + `<script setup>` |
| 构建 | Vite | ^7.3.1 | 开发服务器、HMR、代理 |
| 路由 | Vue Router | ^4.6.4 | `createWebHistory`，全懒加载 |
| 状态 | Pinia | ^3.0.4 | 用户/筛选/简历状态 |
| 持久化 | pinia-plugin-persistedstate | ^4.7.1 | Token 落到 `localStorage` |
| UI | Element Plus | ^2.13.2 | 组件库 + 暗黑主题 |
| 图表 | ECharts / vue-echarts | ^6.0.0 / ^8.0.1 | 统计看板 5 张图 |
| 请求 | axios | ^1.13.4 | 统一实例 + 拦截器 |
| PDF 导出 | html2pdf.js | ^0.14.0 | 简历模板导出 |
| 样式 | Sass | ^1.97.3 | 组件样式 |
| 质量 | ESLint + Oxlint + Prettier | — | `npm run lint` |
| 自动导入 | unplugin-auto-import / unplugin-vue-components | — | Element Plus 按需解析 |

### 3.2 前端 B：管理后台（`front/ljk/adminPanel`）

| 类别 | 技术 | 版本 | 用途 |
|---|---|---|---|
| 框架 | Vue | ^3.5.22 | `<script setup>` |
| 构建 | Vite | ^7.1.11 | `base: '/admin/'` |
| 路由 | Vue Router | ^4.6.3 | `createWebHistory('/admin/')` |
| 状态 | Pinia | ^3.0.3 | 仅示例 store（未使用） |
| UI | Element Plus | ^2.11.8 | 表格、分页、上传、抽屉 |
| 请求 | axios | ^1.8.4 | **页面内直连，无封装** |
| 插件 | @vitejs/plugin-vue-jsx | ^5.1.1 | JSX 支持 |

### 3.3 后端 A：Java（`back/resume`）

| 类别 | 技术 | 版本 | 用途 |
|---|---|---|---|
| 运行时 | Java | 17 | — |
| 框架 | Spring Boot | 3.1.3 | Web MVC、内嵌 Tomcat |
| ORM | MyBatis（XML + 注解） | starter 3.0.0 | 手写 SQL |
| 分页 | PageHelper | 1.4.6 | `startPage` + `PageInfo` |
| 数据库 | MySQL Connector/J | 父管理 | MySQL 8 |
| JWT | com.auth0:java-jwt | 4.4.0 | HS256 签发/校验 |
| 密码 | spring-security-crypto | 父管理 | 仅用 `BCryptPasswordEncoder` |
| 缓存 | Spring Data Redis (Lettuce) | 父管理 | 登录计数、收藏缓存 |
| Excel | Apache POI (`poi-ooxml`) | 5.2.3 | `.xlsx` 导入 |
| 文档 | Knife4j OpenAPI3 | 4.3.0 | `/doc.html` |
| 校验 | spring-boot-starter-validation | 父管理 | `@NotBlank` 等 |
| 其他 | fastjson / MinIO / Java-WebSocket / OkHttp | — | **已声明但代码未使用** |

### 3.4 后端 B：Go（`back/resume-go`）

| 类别 | 技术 | 版本 | 用途 |
|---|---|---|---|
| 运行时 | Go | 1.27 | — |
| HTTP | Gin | v1.12.0 | 路由分组、中间件、静态资源 |
| 数据库 | database/sql + go-sql-driver/mysql | v1.10.1 | 手写 SQL，无 ORM |
| WebSocket | gorilla/websocket | v1.5.3 | 聊天长连接 |
| JWT | 标准库 `crypto/hmac` + `sha256` | — | **手写 HS256，无第三方库** |
| 密码 | golang.org/x/crypto/bcrypt | v0.58.0 | bcrypt |
| 缓存 | redis/go-redis/v9 | v9.23.0 | 缓存/收藏/在线/幂等 |
| 并发 | golang.org/x/sync/singleflight | v0.24.0 | 缓存回源防击穿 |
| MQ | rabbitmq/amqp091-go | v1.16.0 | 离线通知 |
| gRPC | google.golang.org/grpc | v1.84.0 | 内部服务 |
| Excel | xuri/excelize/v2 | v2.11.0 | `.xlsx` 导入 |
| JSON | bytedance/sonic（Gin 间接依赖） | — | 序列化加速 |

### 3.5 基础设施

| 组件 | 版本 | 用途 |
|---|---|---|
| MySQL | 8.0 | 唯一数据源；用到窗口函数 `ROW_NUMBER()`（要求 8.0+） |
| Redis | 7 | 缓存、收藏集合、在线状态、MQ 幂等 |
| RabbitMQ | 3.13（management） | 聊天离线通知 |
| Nginx Ingress | — | K8s 入口，`proxy-body-size: 12m` |

---

## 4. 系统架构

### 4.1 分层

**Go 版**（分层最清晰）：

```
api（HTTP 绑定 / 统一响应 / 中间件）
  └── service（业务规则、事务编排、聊天频控）
        ├── data（SQL、事务）
        ├── cache（Redis / 内存降级）
        ├── mq（RabbitMQ / 空实现降级）
        ├── rpc（gRPC 客户端）
        └── chat.Hub（进程内连接管理）
```

- `chat` 包通过回调（`Hub.OnMessage` / `OnRegister` / `OnUnregister`）反向依赖注入业务，避免 `chat → service → chat` 循环引用。
- 依赖强弱：MySQL **强依赖**（连不上直接退出）；Redis / RabbitMQ / gRPC **弱依赖**（失败自动降级）。

**Java 版**：按业务域分包（`user` / `student` / `company` / `admin`），每域内 controller → service → mapper，公共能力放 `common`。

### 4.2 关键数据流

| 场景 | 流转 |
|---|---|
| 登录 | 前端 → `/login` → 查 `user` → bcrypt 校验 → 签发 access(2h)+refresh(7d) → 前端存 Pinia/localStorage |
| 学生完善资料 | 前端 → `/stu/completeInfo` → 主表局部更新 + 6 张子表「先删后插」→ 重新聚合返回 |
| 企业检索学生 | 前端 → `/com/detailSearch` → 主表动态条件 + PageHelper/COUNT → 批量 IN 查子表 → 聚合 DTO |
| 聊天 | WebSocket → Hub → service 校验 → INSERT `chat_message` → 回查 → Hub 双端推送 → MQ 离线通知 |
| 投递 | `/stu/applications` → 校验职位/简历 → INSERT（唯一键防重）→ 企业端状态流转 |
| 统计看板 | `/stu/statistics`（公开）→ 5 段 `GROUP BY` 聚合 |

---

## 5. 目录结构

```text
d:\resume
├─ back
│  ├─ resume                     # Java 后端（Spring Boot 3）
│  │  ├─ src/main/java/com/resume
│  │  │  ├─ user        # 登录 / JWT
│  │  │  ├─ student     # 学生画像 / 简历
│  │  │  ├─ company     # 职位 / 收藏 / 检索
│  │  │  ├─ admin       # 后台管理 / 字典
│  │  │  ├─ common      # 拦截器 / 全局异常 / CORS
│  │  │  ├─ model       # entity / dto / result
│  │  │  └─ utils       # JwtUtil / PasswordUtil
│  │  ├─ src/main/resources
│  │  │  ├─ application.yml
│  │  │  └─ mapper/*.xml
│  │  ├─ resume.sql              # 基础库（user/student/company/...）
│  │  ├─ templates/*.xlsx        # Excel 导入模板
│  │  └─ pom.xml
│  └─ resume-go                  # Go 后端（Gin + WebSocket + gRPC + MQ）
│     ├─ main.go
│     ├─ internal
│     │  ├─ api                  # 路由 + handlers + 中间件
│     │  ├─ auth                 # 手写 JWT + bcrypt
│     │  ├─ chat                 # Hub / Client（读写泵、心跳）
│     │  ├─ service              # 业务规则
│     │  ├─ data                 # SQL / 事务
│     │  ├─ cache                # Redis + 内存降级
│     │  ├─ mq                   # RabbitMQ 生产/消费
│     │  ├─ rpc                  # gRPC 服务端/客户端 + pb
│     │  ├─ model                # DTO / Entity
│     │  └─ config               # 环境变量
│     ├─ proto/recruit/v1/recruit.proto
│     ├─ schema.sql              # 扩展表（chat_message 等）
│     ├─ seed_accounts.sql       # x1/q1/g1 演示账号
│     ├─ seed_statistics.sql     # 看板演示数据
│     ├─ Dockerfile / docker-compose.yml
│     ├─ deploy/k8s/*.yaml
│     └─ uploads/{resume,chat}
├─ front
│  ├─ yxl/recruit-project        # 学生 + 企业主站
│  │  └─ src/{api,components,router,stores,utils,views}
│  └─ ljk/adminPanel             # 管理后台
│     └─ src/{components,router,views}
├─ 接口文档/                      # 历史接口文档（与代码可能不一致）
├─ start-all.bat                 # Windows 一键启动（MySQL + Go + 两前端）
├─ deploy.ps1                    # Java 版打包 + SFTP + systemd 重启
└─ CHAT_GUIDE.md                 # 聊天系统学习指南
```

---

## 6. 快速开始

### 6.1 一键启动（Windows，推荐）

双击 [start-all.bat](file:///d:/resume/start-all.bat)。脚本按顺序执行：

1. **数据库**：探测 3306，未监听则尝试 3307；若都没有且存在本地测试数据目录，则用 `mysqld --datadir=... --port=3307` 拉起临时实例；随后幂等执行 `schema.sql`、`seed_accounts.sql`、`seed_statistics.sql`。
2. **Go 后端**：`go run .`（`DB_HOST=127.0.0.1`、`DB_PORT=<探测值>`），等待 8081 就绪（最长 90s）。
3. **主前端**：首次自动 `npm install`，然后 `npm run dev -- --host 127.0.0.1 --port 5173 --strictPort`，注入 `VITE_API_BASE_URL=http://127.0.0.1:8081`、`VITE_ADMIN_URL=http://127.0.0.1:5174`。
4. **管理端**：同上，端口 5174，注入 `VITE_LOGIN_URL=http://127.0.0.1:5173`。

启动完成后访问：

| 入口 | 地址 |
|---|---|
| 学生/企业主站 | http://127.0.0.1:5173/ |
| 管理后台 | http://127.0.0.1:5174/admin/ |
| 后端 | http://localhost:8081/ |
| RabbitMQ 控制台（如启用） | http://localhost:15672/ |

演示账号（`seed_accounts.sql`）：

| 角色 | 账号 | 密码 |
|---|---|---|
| 学生 | `x1` ~ `x10` | 与账号相同 |
| 企业 | `q1` ~ `q10` | 与账号相同 |
| 管理员 | `g1` | `g1` |

> 测试实时聊天：普通窗口登录 `x1`，无痕窗口登录 `q1`，双方进入「消息」即可互发。

### 6.2 手动启动 Go 后端

```bash
cd d:\resume\back\resume-go
# 环境变量（PowerShell 示例）
$env:DB_HOST="127.0.0.1"; $env:DB_PORT="3306"; $env:DB_NAME="resume"
$env:DB_USER="root"; $env:DB_PASSWORD="<你的密码>"
$env:REDIS_ENABLED="false"; $env:RABBITMQ_ENABLED="false"; $env:GRPC_ENABLED="false"
go run .
```

关闭 Redis / RabbitMQ / gRPC 后可单机跑通全部核心功能（聊天仅本机实时推送，离线通知不生成）。

### 6.3 手动启动 Java 后端

```bash
cd d:\resume\back\resume
# 先改 src/main/resources/application.yml 的数据源与 Redis 配置
mvn spring-boot:run
# 或
mvn clean package -DskipTests && java -jar target/resume-0.0.1-SNAPSHOT.jar
```

接口文档：http://localhost:8081/doc.html

### 6.4 启动前端

```bash
# 主站
cd d:\resume\front\yxl\recruit-project
npm install
npm run dev            # http://localhost:5173，/api-backend → 127.0.0.1:8081

# 管理端
cd d:\resume\front\ljk\adminPanel
npm install
npm run dev -- --port 5174
# 访问 http://localhost:5174/admin/
```

前端环境变量（无 `.env` 文件时使用默认值）：

| 变量 | 作用 | 默认值 |
|---|---|---|
| `VITE_API_BASE_URL`（主站） | 生产环境后端地址 | `http://182.92.210.198:8081` |
| `VITE_ADMIN_URL`（主站） | 管理员登录后跳转地址 | `http://47.93.133.36` |
| `VITE_API_BASE_URL`（管理端） | 后端地址 | `http://81.70.199.174:8081` |
| `VITE_LOGIN_URL`（管理端） | 无 token 时跳回的登录页 | `http://47.93.133.36` |

开发模式下主站走 Vite 代理，不依赖 `VITE_API_BASE_URL`。

### 6.5 Docker Compose（Go 版）

```bash
cd d:\resume\back\resume-go
docker compose up -d --build
```

起的服务：`mysql:8.0`（3306）、`redis:7-alpine`（6379）、`rabbitmq:3.13-management`（5672/15672）、`app`（8081）。
MySQL 初始化按序挂载 `../resume.sql` → `schema.sql` → `seed_accounts.sql` → `seed_statistics.sql`；`uploads` 目录以卷挂载到 `/app/uploads`。

### 6.6 数据库初始化顺序

```text
resume.sql           # 基础表 + 历史样例数据（user/student/company/stu_*）
schema.sql           # 补充表（job_postings/com_like_stu/chat_message/job_application/...）+ 种子字典
seed_accounts.sql    # x1~x10 / q1~q10 / g1
seed_statistics.sql  # 看板演示数据
```

> `resume.sql` 未包含 `job_postings`、`com_like_stu`、`admin_all_skills`、`language`、`certificate` 的建表语句，由 `resume-go/schema.sql` 补齐（文件头注释已说明）。

---

## 7. 配置说明

### 7.1 Java：`src/main/resources/application.yml`

| 配置 | 值 | 说明 |
|---|---|---|
| `server.port` | `8081` | HTTP 端口 |
| `spring.mvc.static-path-pattern` | `/resume/**` | 静态资源前缀 |
| `spring.web.resources.static-locations` | `classpath:/static/` … `file:/resume/uploads/` | 静态目录 |
| `spring.servlet.multipart.max-file-size` | `10MB` | 单文件上限 |
| `spring.servlet.multipart.max-request-size` | `10MB` | 单请求上限 |
| `spring.datasource.url` | `jdbc:mysql://localhost:3306/resume` | 数据源 |
| `spring.datasource.username/password` | 明文（见文件） | **建议改为环境变量** |
| `spring.data.redis.host/port` | `localhost` / **`6380`** | 注意不是默认 6379 |
| `spring.data.redis.lettuce.pool.max-active` | `8` | 连接池 |
| `spring.jackson.property-naming-strategy` | **`SNAKE_CASE`** | 全局 JSON 下划线命名（关键） |
| `knife4j.enable` | `true` | 接口文档开关 |
| `mybatis.mapper-locations` | `classpath:mapper/*.xml` | XML 位置 |
| `file.upload.resume-path` | `/resume/uploads/resume/` | 简历落盘目录 |

`mybatis-config.xml` 仅开启 `mapUnderscoreToCamelCase=true`。

### 7.2 Go：环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `SERVER_PORT` | `8081` | HTTP 端口 |
| `DB_HOST` / `DB_PORT` | `localhost` / `3306` | MySQL |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `root` / 见代码 / `resume` | MySQL |
| `UPLOAD_DIR` | `uploads/resume` | 简历目录；聊天图片自动放同级 `chat/` |
| `JWT_SECRET` | 固定字符串（见 config.go） | HS256 密钥，需与 Java 一致才能互认 Token |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | `localhost:6379` / 空 / `0` | Redis |
| `REDIS_ENABLED` | `true` | 关掉则纯内存模式 |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | AMQP |
| `RABBITMQ_ENABLED` | `true` | 关掉则跳过通知链路 |
| `GRPC_PORT` | `9091` | gRPC 端口 |
| `GRPC_ENABLED` | `true` | 是否启用 gRPC |
| `GRPC_INTERNAL` | `127.0.0.1:<GRPC_PORT>` | 客户端访问自身 gRPC 的地址 |

DSN 固定为：`user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=true&loc=Local`（`loc=Local` 用于避免时间差 8 小时）。

连接池：`MaxOpenConns=20`、`MaxIdleConns=10`、`ConnMaxLifetime=1h`。

### 7.3 前端 Vite

主站 `vite.config.js`：

```js
base: './',
alias: { '@': './src' },
server: {
  proxy: {
    '/api-backend': {
      target: 'http://127.0.0.1:8081',
      changeOrigin: true,
      ws: true,                       // WebSocket 也走代理
      rewrite: p => p.replace(/^\/api-backend/, '')
    }
  }
}
```

管理端 `vite.config.js`：`base` 生产为 `/admin/`，**无代理**，跨域直连后端（后端已开启 CORS）。

---
## 8. 认证与鉴权原理

### 8.1 双 Token 模型

| Token | 有效期 | 存放 | 用途 |
|---|---|---|---|
| Access Token | 2 小时 | 请求头 `Authorization: Bearer <token>` | 访问业务接口 |
| Refresh Token | 7 天 | 请求头 `Refresh-Token: <token>` | 换取新的 Access Token |

Claims（两套后端一致）：

```json
{
  "userId": 11,
  "username": "x1",
  "userType": "student",
  "tokenType": "access",   // access | refresh
  "exp": 1760000000
}
```

- 算法：`HS256`，密钥为一个固定字符串（Java 在 `JwtUtil`、Go 在 `config.go`，**生产必须改为环境变量**）。
- Java 用 `com.auth0:java-jwt`；Go 用标准库 `crypto/hmac` + `sha256` + `base64.RawURLEncoding` 手写三段式 JWT。
- Go 的校验顺序：切三段 → 验签（`hmac.Equal` 常量时间比较）→ 解 payload → 校验 `exp`。
- 刷新逻辑：校验签名与过期 → 要求 `tokenType == "refresh"` → 用原 claims 重新签发 Access Token。

### 8.2 请求鉴权流程

**Java 版**（三个拦截器按顺序注册）：

| 顺序 | 拦截器 | 作用 | 拦截路径 | 白名单 |
|---|---|---|---|---|
| 1 | `TokenRefreshInterceptor` | Token 剩余 < 30 分钟时自动续签，写入响应头 `New-Token` | `/com/**`、`/stu/**`、`/admin/**` | `/login`、`/refresh-token`、`/stu/statistics`、Swagger 路径 |
| 2 | `AuthInterceptor` | 解析 `Authorization: Bearer`，校验签名与过期，写入 request 属性 | 同上 | 同上 |
| 3 | `AdminInterceptor` | 校验 `userType == "admin"` | `/admin/**` | 无 |

失败响应（HTTP 状态与 body 同时置为 401/403）：

```json
{"code":401,"message":"未提供Token","data":null}
{"code":401,"message":"Token无效或已过期","data":null}
{"code":403,"message":"权限不足，需要管理员权限","data":null}
```

`OPTIONS` 预检请求一律放行；`/validateToken` 不在受保护路径内。

**Go 版**（Gin 中间件）：

- 公开路由：`POST /login`、`POST /refresh-token`、`POST /validateToken`、`GET /stu/statistics`、`GET /chat/ws`。
- 受保护分组 `Group("/", authMiddleware)`：其余 `/user/**`、`/stu/**`、`/com/**`、`/chat/**`。
- 管理分组 `Group("/admin", authMiddleware, adminMiddleware)`。
- `authMiddleware` 在 Token 快过期时同样回写 `New-Token`（前端无需处理，属于“无感续期”）。
- 失败：`401 {"code":401,"message":"未提供Token"}` / `401 "Token无效或已过期"`；非管理员访问后台 `403 "权限不足，需要管理员权限"`。

> **注意**：两套后端的业务中间件都**只验签与过期，不校验 `tokenType`**，因此 Refresh Token 也能直接访问业务接口（只有 `/refresh-token` 强制要求 refresh 类型）。

### 8.3 密码

- 统一使用 **BCrypt**（默认强度 10），数据库中存 `$2a$10$...` 哈希。
- 学生初始密码 = 学号；企业初始密码 = 企业名；管理员初始密码见 `init_admin.sql` / `seed_accounts.sql`。
- 改密接口 `PATCH /stu/updatePwd` 需要 `old_pwd`、`new_pwd`、`re_pwd` 三个字段（下划线命名）。
- 登录失败计数（Java）：Redis `login_attempt:{username}`，30 分钟 TTL，本意是 5 次锁定，**当前版本锁定判断被注释，实际不生效**。

### 8.4 CORS

- Java：`allowedOriginPatterns("*")`、`allowCredentials(true)`、方法 `GET/POST/PUT/DELETE/OPTIONS/PATCH`、`allowedHeaders("*")`、`maxAge 3600`。
- Go：回显 `Origin`（无 Origin 时 `*`），暴露 `New-Token`，允许头包含 `Authorization, Content-Type, Refresh-Token, X-Requested-With, platform`。
- `OPTIONS` 直接返回 204。

### 8.5 前端鉴权

| 项目 | Token 存储 | 守卫逻辑 |
|---|---|---|
| 主站 | Pinia `rec-user`（persist → `localStorage`） | 路由守卫本地解析 JWT 的 `exp`，过期或解析失败则 `clear()` 并跳 `/login` |
| 管理端 | `localStorage['token']` | 只判断 token 是否存在，不校验有效性 |

管理员登录链路：主站 `/login`（`userType=admin`）→ 拿到 token → 跳转 `${VITE_ADMIN_URL}/admin/?token=xxx` → 管理端守卫把 token 写入 `localStorage` 并 `history.replaceState` 去掉 URL 参数。

---

## 9. 统一响应与错误码

### 9.1 响应结构

```json
{ "code": 200, "message": "操作成功", "data": { } }
```

- Java：`Result<T>`，`success()` 的 message 为 `"操作成功"`；HTTP 状态通常也是 200，只有鉴权失败会同时置 401/403。
- Go：`model.Result`；`respond()` **恒用 HTTP 200**，只有中间件与 WebSocket 握手会返回非 200。
- 分页统一使用 PageHelper 兼容结构：

```json
{
  "total": 128,
  "list": [],
  "pageNum": 1,
  "pageSize": 10,
  "pages": 13,
  "size": 10,
  "firstPage": 1,
  "lastPage": 13
}
```

### 9.2 错误码

| code | 含义 | 典型场景 |
|---|---|---|
| 200 | 成功 | — |
| 201 | 创建成功 | 简历上传（Java） |
| 400 | 参数错误 / 业务校验失败 | 缺字段、原密码错误、状态非法 |
| 401 | 未认证 | 无 Token、Token 无效/过期、身份不匹配 |
| 403 | 无权限 | 非管理员访问后台、越权操作他人职位、管理员聊天 |
| 404 | 资源不存在 | 职位/学生/企业/用户不存在 |
| 409 | 冲突 | 重复投递、对方回复前重复发简历 |
| 40005 | 用户名或密码错误 | 登录 |
| 40009 | Excel 内容为空 | 导入 |
| 40101 | 权限不足（冗余校验） | 管理员改密 |
| 40402 / 40403 | 部分学生 / 部分企业不存在 | 批量删除 |
| 41301 | 文件超过大小限制 | 简历 > 10MB；图片 > 5MB |
| 41501 | 文件格式不支持 | 非 pdf/doc/docx；非图片；非 xlsx |
| 42201 | 年龄越界（16–60） | 完善资料 |
| 42202 | 毕业年份不合法 | 完善资料 |
| 42205 / 42206 | 学生 / 企业数据验证失败 | 批量更新 |
| 500 | 服务器内部错误 | 未捕获异常、SQL 异常 |
| 50001 | 数据库写入失败 / 密码错误 | 导入、改密 |

Java 的 `GlobalExceptionHandler` 支持 `"code:message"` 形式的运行时异常（如 `throw new RuntimeException("40005:用户名或密码错误")`），自动拆分为对应业务码。

---

## 10. 数据库设计

### 10.1 表清单

| 表 | 含义 | 使用方 |
|---|---|---|
| `user` | 统一账号（学生/企业/管理员） | 全部 |
| `student` | 学生基本信息 | 全部 |
| `company` | 企业信息 | 全部 |
| `stu_language` | 语言能力 | 画像 |
| `stu_certificate` | 证书 | 画像 |
| `stu_internship` | 实习经历 | 画像 |
| `stu_job` | 求职意向岗位 | 画像 |
| `stu_region` | 意向省/市 | 画像 |
| `stu_speciality` | 专业技能标签 | 画像 |
| `stu_resume` | 简历文件记录 | 画像/聊天/投递 |
| `job_postings` | 招聘职位 | 职位模块 |
| `job_application` | 简历投递 | 投递模块（Go） |
| `com_like_stu` | 企业收藏学生 | 收藏模块 |
| `admin_all_skills` | 技能字典 | 管理端 |
| `language` | 语言字典 | 管理端 |
| `certificate` | 证书字典 | 管理端 |
| `chat_message` | 聊天消息 | 聊天（Go） |
| `chat_notification` | 离线通知 | 聊天（Go） |
| `mq_consumed_event` | MQ 消费幂等记录 | 聊天（Go） |
| `com_collect` | 旧版收藏表 | **未使用（遗留）** |
| `role` | 角色表 | **未使用（遗留）** |

### 10.2 核心表字段

**user**

| 字段 | 类型 | 说明 |
|---|---|---|
| `user_id` | int AI PK | 主键 |
| `username` | varchar(100) | 学生=学号，企业=企业名，管理员=`admin` |
| `password` | varchar(100) | BCrypt 哈希 |
| `user_type` | varchar(25) | `student` / `company` / `admin`（鉴权唯一依据） |
| `create_time` / `update_time` | datetime | 时间戳 |

**student**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int AI PK | **所有 `stu_*` 子表的 `stu_id` 实际指向它** |
| `stu_id` | int | 学号（对应 `user.username`） |
| `name` / `age` / `gender` / `nation` | — | 基本画像 |
| `major` / `degree` / `graduation_year` | — | 专业/学位/毕业年份 |
| `internship` | int | 实习次数 |
| `user_id` | int | 关联 `user.user_id`（普通索引 `user_stu`，无外键） |

**company**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int unsigned ZEROFILL AI PK | 企业 ID |
| `com_name` | varchar(255) | 企业名 |
| `com_region` | varchar(255) | 所在地区 |
| `user_id` | int | 关联账号（索引 `user_com`） |

**画像子表**（字段全部为 `stu_id` + 业务列）：

| 表 | 关键列 | 说明 |
|---|---|---|
| `stu_language` | `language`、`level`、`cert_no`、`issue_date` | 语言能力 |
| `stu_certificate` | `cert_name`、`level`、`cert_no`、`issue_date` | 证书 |
| `stu_internship` | `company`、`experience` | 实习公司 + 经历描述 |
| `stu_job` | `post` | 意向岗位 |
| `stu_region` | `province`、`city` | 一行一个省或市（另一侧为 NULL） |
| `stu_speciality` | `speciality` | 技能标签 |
| `stu_resume` | `resume` | 磁盘路径（追加版本，多条） |

**job_postings**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int AI PK | — |
| `com_id` | int | **唯一物理外键** → `company(id)`，`ON DELETE CASCADE` |
| `title` | varchar(255) NOT NULL | 职位标题 |
| `description` | text | 描述 |
| `external_link` | varchar(500) | 外部链接 |
| `status` | enum/varchar，默认 `active` | `active` / `inactive` |
| `created_at` / `updated_at` | datetime | 自动维护 |
| 索引 | `idx_com_id`、`idx_status_created` | 列表/详情 |

**job_application**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | bigint AI PK | — |
| `job_id` | bigint | 职位 |
| `com_id` | bigint | 企业（冗余，便于企业端查询与鉴权） |
| `student_id` | bigint | 学生 `student.id` |
| `resume_url` | varchar | 投递时的简历快照 |
| `status` | varchar，默认 `pending` | `pending/viewed/interview/offer/rejected` |
| `created_at` / `updated_at` | datetime | — |
| 约束 | `UNIQUE uk_job_student(job_id, student_id)` | 防重复投递 |
| 索引 | `idx_student(student_id,id)`、`idx_company_status(com_id,status,id)` | 列表查询 |

**com_like_stu**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int AI PK | — |
| `stu_id` / `com_id` | int | 学生 / 企业 |
| `create_time` | datetime | 收藏时间 |
| 约束 | `UNIQUE uk_stu_com(stu_id, com_id)` | 防重复收藏 |
| 索引 | `idx_com_time(com_id, create_time)` | 收藏列表排序 |

**chat_message**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | bigint AI PK | 自增，兼作消息游标 |
| `sender_id` / `sender_type` | int / varchar | 发送者 `user_id` + 角色 |
| `receiver_id` / `receiver_type` | int / varchar | 接收者 |
| `content` | text | 文本或图片/简历地址 |
| `msg_type` | varchar，默认 `text` | `text` / `image` / `resume` |
| `is_read` | tinyint | 0/1 |
| `created_at` | datetime | 服务端时间 |
| 索引 | `idx_chat_pair(sender_id,sender_type,receiver_id,receiver_type,id)`、`idx_chat_inbox(receiver_id,receiver_type,is_read,id)` | 历史消息 / 未读数 |

**chat_notification**

| 字段 | 说明 |
|---|---|
| `id` | 主键 |
| `message_id` | 关联消息 |
| `receiver_id` / `receiver_type` | 接收者 |
| `sender_id` / `sender_type` | 发送者 |
| `title` / `content` | 通知标题与摘要（图片 `[图片]`、简历 `[简历]`，文本截断 60 字） |
| `is_read` / `created_at` | 状态与时间 |
| 约束 | `UNIQUE uk_notification_message(message_id, receiver_id, receiver_type)` 幂等兜底 |

**mq_consumed_event**：`event_id`（PK）、`consumed_at`，用于 MQ 消费去重。

**字典表**：`admin_all_skills(skillName, create_time)`、`language(language_name PK)`、`certificate(certificate_name PK)`。

### 10.3 关系图

```text
user 1──1 student 1──N stu_language
                   ├──N stu_certificate
                   ├──N stu_internship
                   ├──N stu_job
                   ├──N stu_region
                   ├──N stu_speciality
                   ├──N stu_resume
                   └──N job_application

user 1──1 company 1──N job_postings 1──N job_application
                   └──N com_like_stu N──1 student

chat_message / chat_notification：以 (user_id, user_type) 逻辑关联 user，无外键
```

> 除 `job_postings.com_id` 外均为逻辑关联（只有索引、无外键），删除学生/企业时需要业务层自行清理子表。

---

## 11. 接口文档（全量）

### 11.1 公共约定

| 项 | 说明 |
|---|---|
| Base URL | `http://<host>:8081` |
| 请求体 | `application/json; charset=utf-8`（上传接口除外） |
| 鉴权 | `Authorization: Bearer <accessToken>` |
| 刷新 | `Refresh-Token: <refreshToken>`（仅 `/refresh-token`） |
| 响应 | `{code, message, data}` |
| 分页参数 | 列表类多用 `pageNum`/`pageSize`；后台用 `page`/`pageSize` |
| 文件字段 | 简历 `resumeFile`；聊天图片 `image`；学生 Excel `stuFile`；企业 Excel `comFile` |
| JSON 命名 | **Java 后端全局 SNAKE_CASE**（`user_id`、`page_num`、`graduation_year`…），唯一例外是登录请求体的 `userType`；Go 后端驼峰/下划线双兼容、双输出 |

**Java 版 JSON 命名示例**：

```json
// POST /login 响应（Java）
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "user_id": 11,
    "username": "x1",
    "user_type": "student",
    "token": "...",
    "refresh_token": "..."
  }
}
```

**Go 版登录响应**：

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "userId": 11,
    "username": "x1",
    "userType": "student",
    "displayName": "王一",
    "token": "...",
    "refreshToken": "..."
  }
}
```

### 11.2 认证与用户

| 方法 | 路径 | 说明 | 参数 | 返回 | 实现 |
|---|---|---|---|---|---|
| POST | `/login` | 登录 | JSON：`username`、`password`、`userType`（`student/company/admin`） | access + refresh | Java + Go |
| POST | `/refresh-token` | 刷新 access | Header `Refresh-Token` | `{token}` | Java + Go |
| POST | `/validateToken` | 校验 Token | Body `{token}`（Go 也接受裸文本） | `{valid,message,userId,userType,needRefresh}` | Java + Go |
| GET | `/user/me` | 当前用户资料 | 无 | `{userId,username,userType,displayName}` | Go |
| PUT | `/user/name` | 改昵称/企业名 | `{name}`（≤20 字） | `UserProfile` | Go |

登录失败返回：`code=40005, message="用户名或密码错误"`；身份不匹配：`code=401, message="用户身份不匹配"`。

### 11.3 学生端

| 方法 | 路径 | 说明 | 参数 | 返回 | 实现 |
|---|---|---|---|---|---|
| GET | `/stu/getAllSkill` | 技能字典 | 无 | 逗号分隔字符串 | Java + Go |
| PUT | `/stu/completeInfo` | 保存完整画像 | `StudentCompleteInfoDTO`（见下） | `StudentDetailDTO` | Java + Go |
| POST | `/stu/uploadResume` | 上传简历 | `multipart`，字段 `resumeFile` | `{resumeId,resumeName,resumeUrl}`（code=201） | Java + Go |
| GET | `/stu/statistics` | 画像统计（免登录） | 无 | 5 组聚合数据 | Java + Go |
| PATCH | `/stu/updatePwd` | 修改密码 | `{old_pwd,new_pwd,re_pwd}` | null | Java + Go |
| GET | `/stu/resume/latest` | 最新简历 | 无 | `{url,name}`；无则 404 | Go |
| POST | `/stu/applications` | 投递职位 | `{jobId,resumeUrl?}` | `JobApplication` | Go |
| GET | `/stu/applications` | 我的投递 | 无 | `{list:[...]}` | Go |
| DELETE | `/stu/applications/:id` | 撤回投递 | 路径 id | null | Go |

`/stu/completeInfo` 请求体：

```json
{
  "age": 21,
  "gender": "女",
  "nation": "汉族",
  "major": "软件工程",
  "degree": "本科生",
  "graduationYear": 2028,
  "internship": 1,
  "languages": [{"language": "英语", "level": "CET-6"}],
  "cities": ["大连"],
  "provinces": ["辽宁"],
  "certificates": [{"certName": "计算机二级", "level": "良好"}],
  "internships": [{"company": "某科技", "experience": "后端开发实习"}],
  "specialities": ["Java", "MySQL"],
  "jobs": ["后端开发"]
}
```

（Java 版字段为下划线：`graduation_year`、`cert_names`、`internship`…；`certificates[].cert_name`。Go 版兼容驼峰与下划线。）

业务校验：年龄 16–60（`42201`）；毕业年份 ∈ `[今年-50, 今年+10]`（`42202`）；关联表“先删后插”，学号/姓名不可改。

`/stu/statistics` 返回：

```json
{
  "languages": [{"value": 12, "language": "英语", "level": "CET-6"}],
  "cities": [{"value": 8, "name": "大连"}],
  "internships": [{"value": 5, "company": "某科技", "experience": "后端开发"}],
  "specialities": [{"value": 20, "name": "Java"}],
  "jobs": [{"value": 15, "name": "后端开发"}]
}
```

### 11.4 企业端：职位

| 方法 | 路径 | 说明 | 参数 | 返回 | 实现 |
|---|---|---|---|---|---|
| POST | `/com/jobPosting` | 发布职位 | `{title,description,external_link,status}`；`com_id` 由 Token 反查覆盖 | `JobPosting` | Java + Go |
| PUT | `/com/jobPosting/{id}` | 修改职位 | 同上（局部更新） | `JobPosting` | Java + Go |
| PATCH | `/com/jobPosting/{id}/status` | 上下架 | `{status:"active|inactive"}` | `JobPosting` | Java + Go |
| DELETE | `/com/jobPosting/{id}` | 删除职位 | 路径 id | `{success,message}` | Java + Go |
| GET | `/com/jobPostings` | 职位分页 | `pageNum`、`pageSize`、`status`、`keyword`、`comId` | `PageInfo<JobPosting>` | Java + Go |
| GET | `/com/jobPosting/{id}` | 职位详情 | 路径 id | `JobPosting` | Java + Go |
| GET | `/com/myJobPostings` | 本企业职位（不分页） | `status?` | `JobPosting[]` | Java + Go |
| GET | `/com/applications` | 收到的投递 | `jobId?`、`status?` | `{list}` | Go |
| PATCH | `/com/applications/:id/status` | 投递状态流转 | `{status}` | `JobApplication` | Go |

越权控制：写操作先由 userId 查 `company.id`，再比对目标职位的 `com_id`，不一致返回 403。

`JobPosting` 主要字段：`id`、`comId`、`comName`、`title`、`description`、`externalLink`、`status`、`createdAt`、`updatedAt`（Go 同时输出 `com_id`/`external_link`/`created_at`）。

### 11.5 企业端：收藏

| 方法 | 路径 | 说明 | 参数 | 返回 | 实现 |
|---|---|---|---|---|---|
| GET | `/com/QueryComIdByUserId` | 当前用户的企业 ID | 无 | int | Java + Go |
| POST | `/com/LikeAdd` | 批量收藏 | `[{stuId,comId}]` | null；有重复则 500 + 提示 | Java + Go |
| POST | `/com/LikeQuery` | 收藏分页 | `{com_id,comName?,pageNum,pageSize}` | `PageInfo<StuFavorite>` | Java + Go |
| DELETE | `/com/LikeDelete` | 批量取消 | `[{stuId,comId}]` | null | Java + Go |
| GET | `/com/LikeIf` | 是否已收藏（查库） | `com_id`、`stu_id` | bool | Java + Go |
| GET | `/com/LikeIfByRedis` | 批量判断（Redis） | `stuIdList=1,2,3` | `{stuId:true}` | Java + Go |

`StuFavorite` 字段：`stuId`、`studentName`、`gender`、`major`、`degree`、`graduationYear`、`languages`、`level`、`specialities`、`post`（语言/技能/岗位由 `GROUP_CONCAT` 拼成字符串）。

### 11.6 企业端 / 管理端：学生检索

| 方法 | 路径 | 说明 | 参数 | 返回 | 实现 |
|---|---|---|---|---|---|
| POST | `/com/search` | 学生列表（精简） | `StudentQueryDTO` | `PageInfo<Student>` | Java + Go |
| POST | `/com/detailSearch` | 学生详情（聚合） | 同上 | `PageInfo<StudentDetailDTO>` | Java + Go |
| POST | `/admin/searchStuDetail` | 后台学生详情检索 | 同上 | 同上 | Java + Go |

`StudentQueryDTO` 支持字段：

| 类型 | 字段 |
|---|---|
| 精确 | `id`、`stu_id`、`age`、`gender`、`degree`、`graduation_year` |
| 模糊 | `name`、`nation`、`major` |
| 范围 | `internship`（>=） |
| 列表 | `cities[]`、`provinces[]`、`languages[{language,level}]`、`cert_names[]`、`specialities[]`、`jobs[]` |
| 其他 | `cert_level`（模糊）、`internship_company`（模糊） |
| 分页 | `page_num`、`page_size`（默认 1 / 10） |

### 11.7 管理端

| 方法 | 路径 | 说明 | 参数 | 返回 |
|---|---|---|---|---|
| PUT | `/admin/updateStuPwd/{stu_id}` | 重置学生密码 | Query `pwd` | 成功提示 |
| POST | `/admin/addStu` | Excel 导入学生 | `stuFile`（.xlsx） | `{successCount,failureCount,errors[]}` |
| POST | `/admin/addCom` | Excel 导入企业 | `comFile`（.xlsx） | 同上 |
| DELETE | `/admin/delStu` | 批量删学生 | `{stuIds:[...]}` | 结果；部分不存在 40402 |
| DELETE | `/admin/delCom` | 批量删企业 | `{comIds:[...]}` | 结果；部分不存在 40403 |
| PUT | `/admin/updateStu` | 批量改学生 | `{students:[{id,...}]}` | 部分失败 42205 |
| PUT | `/admin/updateCom` | 批量改企业 | `{companies:[{id,...}]}` | 部分失败 42206 |
| GET | `/admin/searchStu` | 学生分页 | `username`（按姓名模糊）、`page`、`pageSize` | `{list,pagination}` |
| GET | `/admin/searchCom` | 企业分页 | `companyName`、`page`、`pageSize` | `{list,pagination}` |
| POST | `/admin/searchStuDetail` | 学生详情检索 | `StudentQueryDTO` | `PageInfo` |
| GET | `/admin/getAllLanguage` | 语言字典 | 无 | `[]string` |
| POST | `/admin/addLanguage` | 新增语言 | form `language` | true |
| POST | `/admin/delLanguage` | 删除语言 | form `language` | true |
| GET | `/admin/getAllCertificate` | 证书字典 | 无 | `[]string` |
| POST | `/admin/addCertificate` | 新增证书 | form `certificate` | true |
| POST | `/admin/delCertificate` | 删除证书 | form `certificate` | true |
| POST | `/admin/addSkillAll` | 批量加技能 | 原始文本 `Java,Go` | 提示 |
| GET | `/admin/getAllSkill` | 技能字典 | 无 | 逗号字符串 |
| DELETE | `/admin/deleteSkill` | 删技能 | Query `skills=Java,Go` | 提示 |

批量结果结构：

```json
{
  "successCount": 9,
  "failureCount": 1,
  "errors": [{"identifier": "20240001", "reason": "学号已存在"}]
}
```

### 11.8 聊天（Go 专属）

| 方法 | 路径 | 说明 | 参数 | 返回 |
|---|---|---|---|---|
| GET | `/chat/ws` | WebSocket 握手 | Query `token` | 101 升级 |
| GET | `/chat/peers` | 可聊天对象 | `keyword?` | `{list:[ChatPeer]}` |
| GET | `/chat/conversations` | 会话列表 | 无 | `{list:[Conversation]}` |
| GET | `/chat/messages` | 历史消息 | `peerId`、`peerType`、`beforeId?`、`limit?`（默认 30，上限 100） | `{messages,hasMore}` |
| POST | `/chat/messages` | REST 发消息（降级） | `{receiverId,receiverType,content,msgType}` | `ChatMessage` |
| POST | `/chat/uploadImage` | 聊天图片 | `image`（≤5MB） | `{url:"/uploads/chat/xxx.png"}` |
| POST | `/chat/read` | 会话已读 | `{peerId,peerType}` | null |
| GET | `/chat/resume-permission` | 能否发简历 | `peerId`、`peerType` | `{canSend,waitingReply}` |
| GET | `/chat/profile` | 对方资料 | `peerId`、`peerType` | 学生画像 / 企业信息+在招职位 |
| GET | `/chat/notifications` | 离线通知 | `limit?`（默认 20，上限 100） | `{list}` |
| POST | `/chat/notifications/read` | 全部已读 | 无 | null |
| GET | `/chat/notifications/unread` | 未读通知数 | 无 | `{count}` |

### 11.9 静态资源与兜底

| 路径 | 说明 |
|---|---|
| `GET /uploads/resume/<file>` | 简历文件（Go；Java 映射到 `file:/resume/uploads/`） |
| `GET /uploads/chat/<file>` | 聊天图片（Go） |
| `GET /resume/uploads/<file>` | 旧前缀兼容（两套后端） |
| 其他 | Go `NoRoute` → `{code:404,message:"接口不存在"}`（HTTP 200）；Java 走 Spring 默认 404 |

### 11.10 接口调用示例

```bash
# 登录
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"username":"x1","password":"x1","userType":"student"}'

# 带 Token 查询职位
curl "http://localhost:8081/com/jobPostings?pageNum=1&pageSize=10&status=active" \
  -H "Authorization: Bearer <accessToken>"

# 上传简历
curl -X POST http://localhost:8081/stu/uploadResume \
  -H "Authorization: Bearer <accessToken>" \
  -F "resumeFile=@C:/tmp/resume.pdf"

# 管理端批量删除
curl -X DELETE http://localhost:8081/admin/delStu \
  -H "Authorization: Bearer <adminToken>" \
  -H "Content-Type: application/json" \
  -d '{"stu_ids":[1,2]}'
```

---
## 12. 聊天系统原理（重点）

聊天仅在 **Go 后端** 中实现，前端为主站 `/chat` 页面。

### 12.1 总体设计：REST + WebSocket 混合

| 通道 | 负责 |
|---|---|
| HTTP REST | 登录、联系人、会话列表、历史消息、标记已读、断线时降级发送、图片上传、通知 |
| WebSocket | 新消息实时推送、连接建立通知、心跳 |

这样设计的原因：

1. 历史消息天然适合分页查询，HTTP 更简单；
2. WebSocket 只负责“实时”，职责单一，断线不影响历史数据；
3. 断线时仍可用 REST 发消息，不丢消息。

### 12.2 连接建立与鉴权

```text
ws://<host>:8081/chat/ws?token=<accessToken>
```

浏览器原生 `new WebSocket(url)` **不支持自定义请求头**，因此 Token 只能放在 query 参数里。服务端流程：

1. 读取 `token`，为空 → HTTP 401 `缺少 token`；
2. `auth.Manager.Parse(token)` 验签 + 校验 `exp`，失败 → 401 `token 无效或已过期`；
3. 角色必须是 `student` 或 `company`，`admin` → 403 `只有学生和企业可以使用聊天`；
4. `websocket.Upgrader.Upgrade(...)`（`CheckOrigin` 恒为 true，生产应收紧）；
5. `chat.NewClient(hub, conn, userId, userType, username)` → `client.Start()`；
6. 立即下发 `{"type":"connected","userId":11,"userType":"student"}`。

`/chat/ws` 不经过 `authMiddleware`，鉴权在 handler 内完成；其他 `/chat/*` 接口仍走 `Authorization: Bearer`。

### 12.3 Hub：连接注册中心

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]map[*Client]struct{} // "student:11" -> 该用户的所有连接
    OnMessage    func(*Client, IncomingMessage)
    OnRegister   func(*Client)
    OnUnregister func(*Client)
}
```

| 方法 | 行为 |
|---|---|
| `Register` | 建集合 → 加入连接 → 日志“用户上线 当前连接数 N” → `go OnRegister(c)` 异步 |
| `Unregister` | 只删当前连接；集合空才删外层 key（避免把其他标签页一起下线）→ `go OnUnregister(c)` 异步 |
| `SendToUser(id,type,payload)` | `json.Marshal` 一次，`RLock` 遍历该用户所有连接 `push` |
| `IsOnline(id,type)` | 只查本进程连接表 |

设计要点：

- **一个用户允许多连接**（多标签页/多设备），所以值是集合；
- 所有 map 读写必须加锁（Go map 并发读写会 panic）；
- Hub 不反向 import service，通过注入 `OnMessage` 回调打破循环依赖。

### 12.4 Client：两个 goroutine

常量：

```go
writeWait      = 10 * time.Second
pongWait       = 60 * time.Second
pingPeriod     = 54 * time.Second   // pongWait * 9 / 10
maxMessageSize = 4 * 1024
sendBufferSize = 128
```

**readPump（唯一读入口）**

1. `SetReadLimit(4KB)`、`SetReadDeadline(now+60s)`；
2. `SetPongHandler`：收到 Pong → 续期 60s；
3. 循环 `ReadMessage()`：
   - 出错 → 退出；
   - 收到数据 → 续期读超时；
   - JSON 解析失败 → 推 `{"type":"error","message":"消息格式错误"}`，继续；
   - `type=="ping"` → 推 `{"type":"pong"}`；
   - `type=="message"` → 调 `Hub.OnMessage` 执行业务；
   - 其他 → 推 `{"type":"error","message":"未知的消息类型: X"}`。
4. `defer` 清理顺序：`Hub.Unregister` → `close(send)` → `conn.Close()`（顺序不能反）。

**writePump（唯一写出口）**

```text
业务 goroutine --push--> send chan (128) --> writePump --> conn.WriteMessage
```

- gorilla/websocket **不允许并发写**，所有写必须串行到这里；
- `select` 两个分支：`data, ok := <-send`（`ok=false` 时发 Close 帧退出）、`<-ticker.C`（每 54s 发协议层 Ping）；
- 任意写失败直接 return，`conn.Close()` 会让阻塞的 `ReadMessage` 报错，readPump 随之清理。

**非阻塞入队（保护 Hub）**

```go
select {
case c.send <- data:
default:
    // 队列满：丢弃 + 打日志（消息已入库，用户重拉历史可补齐）
}
```

### 12.5 两层心跳

| 层 | 报文 | 发起方 | 作用 |
|---|---|---|---|
| 协议层 | WebSocket Ping/Pong 控制帧 | 服务端每 54s Ping，浏览器自动 Pong | 维持 TCP 与中间设备连接，续期 60s 读超时 |
| 应用层 | `{"type":"ping"}` / `{"type":"pong"}` | 前端每 25s 主动发 | 前端判断业务通道是否健康，35s 无 pong 则主动断开重连 |

### 12.6 消息协议

客户端 → 服务端：

```json
{"type":"message","receiverId":12,"receiverType":"company","content":"你好","msgType":"text"}
{"type":"ping"}
```

服务端 → 客户端：

| type | 结构 | 时机 |
|---|---|---|
| `connected` | `{"type":"connected","userId":11,"userType":"student"}` | 连接建立 |
| `pong` | `{"type":"pong"}` | 收到应用层 ping |
| `message` | `{"type":"message","data":ChatMessage}` | 业务消息（双端各推一份） |
| `error` | `{"type":"error","message":"..."}` | 格式错误/未知类型/业务校验失败 |

`ChatMessage`：

```json
{
  "id": 101,
  "senderId": 11,
  "senderType": "student",
  "senderName": "王一",
  "receiverId": 12,
  "receiverType": "company",
  "receiverName": "企业一",
  "content": "你好",
  "msgType": "text",
  "isRead": false,
  "createdAt": "2026-10-23 21:43:27"
}
```

### 12.7 一条消息的完整时序

```text
发送方浏览器                    Go 后端                             接收方浏览器
   │ 1. ws?token=xxx              │                                      │
   │─────────────────────────────►│ 2. 验签、Upgrade、Hub.Register       │
   │                              │                                      │
   │ 3. {"type":"message",...}    │                                      │
   │─────────────────────────────►│ 4. service.SendChatMessage           │
   │                              │    - 角色互补校验                     │
   │                              │    - msgType/内容长度/图片/简历频控   │
   │                              │    - 接收者存在性（查 user 表）       │
   │                              │ 5. INSERT chat_message               │
   │                              │ 6. SELECT 回查（拿 id/时间/姓名）     │
   │                              │ 7. Hub.SendToUser(receiver) ────────►│ 实时上屏
   │◄─────────────────────────────│ 8. Hub.SendToUser(sender)            │ 多端同步
   │                              │ 9. MQ.Publish(chat.message.created)  │
   │                              ▼                                      │
   │                     RabbitMQ → HandleChatNotify → 离线通知           │
```

**先落库、再推送**：即使接收者离线，消息也不会丢；下次拉历史即可补齐。

### 12.8 发送前的业务校验（`service.SendChatMessage`）

| 校验 | 失败码 |
|---|---|
| 发送者必须 `student/company` | 403 |
| 接收者必须是互补角色 | 400 |
| 不能发给自己 / `receiverId<=0` | 400 |
| `msgType` ∈ `text/image/resume`（空 → `text`） | 400 |
| 文本去空格后非空、按 **rune** 计数 ≤800 | 400 |
| `image` 的 content 必须以 `/uploads/chat/` 开头（防外链伪造） | 400 |
| `resume` 仅学生可发 | 403 |
| `resume` 必须本站地址（`/uploads/resume/` 或旧 `/api/resume/download/`）、≤500 字符 | 400 |
| 简历频控：对方最后一次回复之后已发过简历 | 409 |
| 接收者必须存在于 `user` 表 | 404 |

简历频控 SQL：

```sql
SELECT COUNT(*) FROM chat_message
WHERE sender_id=? AND sender_type=? AND receiver_id=? AND receiver_type=?
  AND msg_type='resume'
  AND id > COALESCE((
      SELECT MAX(id) FROM chat_message
      WHERE sender_id=? AND sender_type=? AND receiver_id=? AND receiver_type=?
  ), 0)
```

### 12.9 会话模型与列表 SQL

**没有 `conversation` 表**，会话是按对方聚合出来的视图：

```sql
SELECT t.peer_id, t.peer_type, t.unread,
       COALESCE(ss.name, cc.com_name, '') AS peer_name,
       m.content, m.msg_type, m.created_at, m.sender_id, m.sender_type
FROM (
  SELECT
    CASE WHEN sender_id=? AND sender_type=? THEN receiver_id ELSE sender_id END AS peer_id,
    CASE WHEN sender_id=? AND sender_type=? THEN receiver_type ELSE sender_type END AS peer_type,
    MAX(id) AS max_id,
    SUM(CASE WHEN receiver_id=? AND receiver_type=? AND is_read=0 THEN 1 ELSE 0 END) AS unread
  FROM chat_message
  WHERE (sender_id=? AND sender_type=?) OR (receiver_id=? AND receiver_type=?)
  GROUP BY peer_id, peer_type
) t
JOIN chat_message m ON m.id = t.max_id
LEFT JOIN student ss ON t.peer_type='student' AND ss.user_id=t.peer_id
LEFT JOIN company cc ON t.peer_type='company' AND cc.user_id=t.peer_id
ORDER BY t.max_id DESC
```

要点：

- `CASE` 把双向消息归一成“对方”；
- `MAX(id)` 取最后一条（比 `MAX(created_at)` 可靠，秒级时间会撞）；
- `SUM(is_read=0)` 是该会话全部历史未读数；
- `LastFromMe` 必须同时比较 `sender_id` 与 `sender_type`（学生和企业 ID 可能撞号）。

### 12.10 历史消息：游标分页

```sql
SELECT m.*, 发送者姓名, 接收者姓名
FROM chat_message m
LEFT JOIN student ss ON m.sender_type='student'  AND ss.user_id=m.sender_id
LEFT JOIN company cc ON m.sender_type='company'  AND cc.user_id=m.sender_id
LEFT JOIN student rs ON m.receiver_type='student' AND rs.user_id=m.receiver_id
LEFT JOIN company rc ON m.receiver_type='company' AND rc.user_id=m.receiver_id
WHERE ((我发他) OR (他发我))
  AND m.id < ?          -- beforeId 游标
ORDER BY m.id DESC
LIMIT ?
```

- 用 `id < beforeId` 而非 OFFSET，避免深分页与新消息插入导致的重复/漏读；
- 查完在 Go 中翻转为“旧 → 新”；
- `hasMore = len(messages) == limit`；
- 前端向上滚动时取 `messages[0].id` 作为 `beforeId`，插入后按 `scrollHeight - oldHeight + oldTop` 保持阅读位置。

### 12.11 已读 / 未读

```sql
UPDATE chat_message SET is_read = 1
WHERE sender_id=? AND sender_type=? AND receiver_id=? AND receiver_type=? AND is_read=0
```

只把“对方发给我的消息”标已读。前端在选中会话时调用 `/chat/read`，并本地把该会话 `unread` 清零；顶部导航每 15s 调 `/chat/conversations` 汇总未读总数（>99 显示 `99+`）。

### 12.12 离线通知：MQ + 三层幂等

发送成功后（尽力而为）发布事件：

```json
{
  "eventId": "chat.message.created:101",
  "type": "chat.message.created",
  "messageId": 101,
  "senderId": 11,
  "senderType": "student",
  "receiverId": 12,
  "receiverType": "company",
  "content": "你好",
  "msgType": "text",
  "createdAt": "2026-10-23T21:43:27+08:00"
}
```

消费流程 `HandleChatNotify`：

1. **MySQL 幂等**：`INSERT IGNORE INTO mq_consumed_event`，`RowsAffected==1` 才继续；
2. **Redis 幂等**：`SETNX resume:mq:consumed:{eventId}`（24h），Redis 故障时返回 true 交由 MySQL 兜底；
3. **在线则跳过**：接收者在线说明 WS 已实时推送，不再写离线通知；
4. **取发送者昵称**：gRPC `GetUserBrief` → 本地库 → 固定 `"对方"`（三级降级）；
5. **生成摘要**：图片 `[图片]`、简历 `[简历]`、文本按 rune 截断 60 字；
6. **写通知**：`INSERT IGNORE INTO chat_notification`（唯一键第三层兜底）。

> “先库后 Redis”的顺序是刻意的：若先占 Redis 而库写失败，第二次消费会被误判重复而永久丢通知。

### 12.13 在线状态

| 时机 | 动作 |
|---|---|
| 连接注册 | `go OnRegister` → `Cache.SetOnline`：`INCR resume:online:{type}:{id}` + `EXPIRE 300s` |
| 连接断开 | `go OnUnregister` → `Cache.SetOffline`：`DECR`，≤0 时 `DEL` |
| 查询 | Redis `EXISTS` → 回退本地 `Hub.IsOnline` |

- 用计数器而不是布尔值，支持多标签页；
- TTL 兜底进程被 kill 的残留；
- 前端每 8s 刷一次 `/chat/conversations` / `/chat/peers` 获取 `online`，**没有主动 presence 推送**；
- 已知缺陷：Redis 在线 key 的 TTL 只在上线时设置一次，长连接超过 5 分钟会过期（单实例会回退本地 Hub，多实例可能误判）。

### 12.14 图片与简历消息

**图片**：`POST /chat/uploadImage`（字段 `image`，≤5MB，`jpg/jpeg/png/gif/webp`），落盘 `uploads/chat/`，命名 `{userId}_{yyyyMMddHHmmss}_{8位hex}{ext}`，返回 `/uploads/chat/xxx`，随后以 `msgType=image` 发送该 URL。

**简历**：学生先 `GET /stu/resume/latest` 取最新简历 URL，再以 `msgType=resume` 发送；企业端点击可预览 PDF。频控保证“对方回复前只能发一次”。

### 12.15 前端 WebSocket 封装（`src/utils/chatSocket.js`）

```js
const socket = new ChatSocket({
  token: userStore.token,
  onMessage: handleSocketMessage,
  onStatus: s => (connectionStatus.value = s)
})
socket.connect()
```

实现细节：

- URL：取 `API_BASE_URL` → 补 `window.location.origin` → `replace(/^http/, 'ws')` → 拼 `/chat/ws?token=...`；
  开发环境实际为 `ws://localhost:5173/api-backend/chat/ws?token=...`（Vite `ws:true` 代理到 8081）。
- 心跳：25s 发 `{"type":"ping"}`；35s 未收到 pong 主动 `close()` 触发重连。
- 重连：**指数退避 1s → 2s → 4s → 8s → 15s（封顶）**，连上后重置为 1s；`manualClose` 时不重连。
- 消息：`type==='pong'` 只更新 `lastPongAt`，不交给业务；其他交给 `onMessage`。
- 发送：`sendMessage(receiverId, receiverType, content, msgType)`，仅在 `readyState===OPEN` 时返回 true。

### 12.16 聊天页（`views/chat/index.vue`）关键逻辑

1. 进入页面：加载会话 → 加载联系人 → 学生再加载自己的简历 → 连接 WebSocket；
2. 每 8s 刷新在线状态；窗口 focus 立即刷新；
3. 选择会话：清空消息 → 加载 30 条历史 → 刷新简历权限 → 标记已读 → 清空本地未读 → `router.replace('/chat?peerId&peerType&name')`；
4. 收到消息：用消息 `id` 去重；当前会话追加并滚动到底；非当前会话 300ms 防抖刷新会话列表；
5. 发送：优先 WS；WS 未连接时降级 `POST /chat/messages`，再把响应交给同一处理函数去重上屏；
6. 图片：前端校验 MIME 与 ≤5MB → 上传 → 发 `image` 消息；
7. 简历：`getMyResume({silent:true})` → 无简历提示 → 发送 `resume` 消息 → 乐观置 `canSendResume=false`；
8. 置顶浮窗：学生显示自己的最新简历，企业显示当前会话学生的简历（来自 `/chat/profile`）；
9. 输入：`Enter` 发送、`Shift+Enter` 换行、`Ctrl/Cmd+Enter` 插入换行、输入法组合态不发送；
10. 销毁：关闭 socket、清理所有定时器与事件监听。

### 12.17 聊天的边界与扩展

| 限制 | 说明 / 扩展方向 |
|---|---|
| 单实例 Hub | 跨实例实时推送不支持；多实例需 Redis Pub/Sub 或 MQ 广播 |
| MQ 失败即丢弃 | `Nack(requeue=false)`，不重试、无死信队列（通知允许丢失） |
| 无消息撤回/已读回执推送 | 可增加 `recalled_at`、`{type:"read"}` |
| 无客户端消息 ID | 网络抖动可能重复发送，可加 `clientMsgId` + 唯一索引 |
| Token 在 URL | 可能进日志，生产应使用 `wss://` + 短有效期 Token |
| 在线状态不广播 | 只能轮询获取 |

---

## 13. 简历与文件存储

### 13.1 简历上传链路

| 环节 | Java | Go |
|---|---|---|
| 字段 | `resumeFile` | `resumeFile` |
| 大小 | ≤10MB（`41301`） | ≤10MB（`41301`） |
| 类型 | `.pdf/.doc/.docx`（`41501`） | 同左 |
| 命名 | `{学号}_{yyyyMMddHHmmss}_{UUID前8位}{ext}` | `{学号}_{yyyyMMddHHmmss}_{8位hex}{ext}` |
| 落盘 | `file.upload.resume-path` | `UPLOAD_DIR` |
| 入库 | `stu_resume(stu_id=student.id, resume=路径)` | 同左（追加版本） |
| 返回 | `{resumeId,resumeName,resumeUrl}`，code=201 | 同左，`resumeUrl=/api/resume/download/{name}` |
| 最新简历 | `ORDER BY id DESC LIMIT 1` | `GET /stu/resume/latest` |

安全点：**不使用用户原始文件名**，扩展名统一转小写后白名单校验，防止重名覆盖与 `../` 路径穿越。

### 13.2 访问地址

| 后端 | 可用地址 |
|---|---|
| Go | `/uploads/resume/<file>`、`/uploads/chat/<file>`、旧前缀 `/resume/uploads/<file>` |
| Java | `/resume/uploads/**`、`/uploads/**`（映射到 `file:/resume/uploads/`） |

**已知不一致**：

- Java 返回的 `resume_url` 是 `/api/resume/download/<file>`，但代码中**没有该下载接口**；
- Go 的上传接口也返回 `/api/resume/download/...`，同样未注册路由；不过“最新简历”“投递记录”“聊天发简历”走 `publicResumeURL()`，会转换成 `/uploads/resume/<base>`，是可访问的；
- 数据库中历史数据可能是 `uploads\resume\xxx.pdf`（反斜杠），前端 `fileUrl()` 会统一替换为 `/` 并补全前缀。

---

## 14. 搜索、分页与查询原理

### 14.1 动态条件 SQL

以 Go `data/studentWhere(q)` 为例（Java 的 `SearchMapper.xml` 等价）：

| 条件 | 实现 |
|---|---|
| 等值 | `s.id=?`、`s.stu_id=?`、`s.age=?`、`s.gender=?`、`s.degree=?`、`s.graduation_year=?` |
| 模糊 | `s.name LIKE CONCAT('%',?,'%')`（民族、专业同理） |
| 范围 | `s.internship >= ?` |
| 地区 | `sr.city IN (...)`、`sr.province IN (...)` |
| 语言 | 每个语言一个 `EXISTS (SELECT 1 FROM stu_language WHERE stu_id=s.id AND language=? [AND level=?])`，多语言之间 `OR` |
| 证书 | `s.id IN (SELECT stu_id FROM stu_certificate WHERE cert_name IN (...))` |
| 证书等级 | `level LIKE` |
| 实习公司 | `EXISTS (... company LIKE ...)` |
| 技能 | `EXISTS (... speciality IN (...))` |
| 岗位 | `EXISTS (... post IN (...))` |

所有值均走 `?` 绑定，杜绝 SQL 注入。

### 14.2 分页

- **Java**：PageHelper `startPage(pageNum,pageSize)` 后紧接查询，自动改写为 `COUNT` + `LIMIT`，返回 `PageInfo`。
- **Go**：手写 `SELECT COUNT(*) FROM (主查询) t` + `LIMIT ? OFFSET ?`，并复刻 PageHelper 的 `PageInfo` 字段。
- **聊天历史**：不用 OFFSET，用 `id < beforeId` 的游标分页。

### 14.3 避免 N+1

`detailSearch` 两步：

1. 先分页查主表（含地区聚合、最新简历窗口函数）；
2. 收集本页所有 `student.id`，对 5 张子表各执行一次 `WHERE stu_id IN (...)`，在内存用 `map[stuId]` 回填。

```sql
-- 最新简历：MySQL 8 窗口函数
SELECT * FROM (
  SELECT r.*, ROW_NUMBER() OVER (PARTITION BY stu_id ORDER BY id DESC) rn
  FROM stu_resume r
) t WHERE rn = 1
```

### 14.4 统计聚合

`GET /stu/statistics` 由 5 段 `GROUP BY` 组成：

```sql
SELECT COUNT(*) AS value, language, level FROM stu_language GROUP BY language, level ORDER BY value DESC;
SELECT COUNT(*) AS value, city AS name FROM stu_region WHERE city IS NOT NULL GROUP BY city ORDER BY value DESC;
SELECT COUNT(*) AS value, company, MAX(experience) AS experience FROM stu_internship
  WHERE company IS NOT NULL GROUP BY company ORDER BY value DESC;
SELECT COUNT(*) AS value, speciality AS name FROM stu_speciality GROUP BY speciality ORDER BY value DESC;
SELECT COUNT(*) AS value, post AS name FROM stu_job GROUP BY post ORDER BY value DESC;
```

Go 版每段查询后显式 `rows.Close()`，避免同一函数多次查询占满连接池。

---

## 15. 收藏模块原理

### 15.1 数据模型

- 主存储：`com_like_stu(stu_id, com_id, create_time)`，唯一键 `(stu_id, com_id)` 防重复。
- 缓存：Redis Set `resume:favorite:{userId}`（Go）/ `user:favorite:{userId}`（Java），member 为学生 ID 字符串。

### 15.2 收藏流程

1. 校验 `student.id` 与 `company.id` 均存在；
2. 用行构造器批量查已存在组合：`WHERE (stu_id, com_id) IN ((?,?),(?,?))`；
3. 过滤重复后批量插入；
4. 插入行数不等于待插入数则抛异常回滚；
5. 若存在已收藏，返回“添加 N 个学生失败，原因：学生已存在”。

### 15.3 取消收藏

`DELETE FROM com_like_stu WHERE (stu_id, com_id) IN ((?,?),...)`，返回未删除条数；全部不存在时提示“学生不存在”。

> Java 版额外用 `RedisTemplate` 删缓存，但与写入使用的 `StringRedisTemplate` 序列化器不同、且把整个 List 当成一个 member，**取消收藏时 Redis 缓存实际不会被清除**（Go 版使用 Set 的 `SREM` 正常）。

### 15.4 收藏列表查询

```sql
SELECT f.stu_id, s.name, s.gender, s.major, s.degree, s.graduation_year,
       (SELECT GROUP_CONCAT(language SEPARATOR '、') FROM stu_language WHERE stu_id=s.id) AS languages,
       (SELECT GROUP_CONCAT(speciality SEPARATOR '、') FROM stu_speciality WHERE stu_id=s.id) AS specialities,
       (SELECT GROUP_CONCAT(post SEPARATOR '、') FROM stu_job WHERE stu_id=s.id) AS post,
       f.create_time
FROM com_like_stu f
JOIN student s ON s.id = f.stu_id
WHERE f.com_id = ?
ORDER BY f.create_time DESC
LIMIT ? OFFSET ?
```

用**相关子查询 + GROUP_CONCAT**而不是多表 JOIN，避免一对多表连接产生笛卡尔积导致分页数量失真。

### 15.5 两种判断方式

| 接口 | 数据源 | 场景 |
|---|---|---|
| `/com/LikeIf` | MySQL | 强一致，单条判断 |
| `/com/LikeIfByRedis` | Redis Set | 批量、低延迟，允许短暂不一致 |

---

## 16. 简历投递模块

仅 Go 版实现。

### 16.1 状态机

```text
pending（待处理） → viewed（已查看） → interview（面试中） → offer（录用）
                                        └→ rejected（不合适）
```

状态常量：`pending`、`viewed`、`interview`、`offer`、`rejected`。

### 16.2 投递校验（学生端）

1. 职位必须存在（404）；
2. 职位状态必须为 `active`（400）；
3. `resumeUrl` 为空时自动取最新简历；没有简历 → 400；
4. `resumeUrl` 必须是本站地址（400）；
5. `UNIQUE uk_job_student(job_id, student_id)` 防重复；MySQL 错误码 **1062** 被捕获为哨兵错误 `ErrDuplicateApplication` → 返回 **409**；
6. `resume_url` 是投递时的**快照**，学生后续换简历不影响已投记录。

### 16.3 企业端

- 列表：`GET /com/applications?jobId=&status=`，SQL 联查职位标题、企业名、学生姓名/专业/学位；
- 状态推进：`UPDATE ... WHERE id=? AND com_id=?`，`com_id` 条件用于防水平越权；
- 学生撤回：`DELETE ... WHERE id=? AND student_id=?`。

---

## 17. Excel 批量导入原理

### 17.1 模板列顺序

**学生**（第 1 行为表头，从第 2 行开始，0–8 列）：

| 列 | 字段 |
|---|---|
| 0 | 学号 `stu_id` |
| 1 | 姓名 `name` |
| 2 | 年龄 `age` |
| 3 | 性别 `gender` |
| 4 | 民族 `nation` |
| 5 | 专业 `major` |
| 6 | 学位 `degree` |
| 7 | 毕业年份 `graduation_year` |
| 8 | 实习次数 `internship` |

**企业**（0–1 列）：企业名称 `com_name`、所在地区 `com_region`。

模板文件：`back/resume/templates/学生数据导入模板.xlsx`、`企业数据导入模板.xlsx`。

### 17.2 实现差异

| 项 | Java | Go |
|---|---|---|
| 库 | Apache POI `XSSFWorkbook` | excelize v2 |
| 只读 | 第 1 个 sheet，跳过第 1 行 | 同左 |
| 单元格 | 兼容 NUMERIC/STRING/BOOLEAN，数字转 int | 同左 |
| 查重 | 学生按 `stu_id`，企业按 `com_name` | 同左 |
| 事务 | 整体 `@Transactional`，行级错误 `continue` | 事务内插入，错误逐行收集 |
| 默认密码 | 学号 / 企业名（BCrypt） | 同左 |
| 结果 | `{successCount,failureCount,errors[{identifier,reason}]}` | 同左（驼峰） |
| 部分失败 | `code=200, message="部分数据导入失败"` | 同左 |

导入顺序：先插 `user`（`useGeneratedKeys` 回填 `user_id`），再写 `student` / `company` 的 `user_id`。

> 仅实现导入，**没有导出**。

---

## 18. Redis 缓存设计

### 18.1 Key 一览（Go）

| Key | 类型 | TTL | 用途 |
|---|---|---|---|
| `resume:job:version` | string(int) | 无 | 职位缓存代际版本号 |
| `resume:job:v{n}:list:{status}\|{keyword}\|{comId}\|{pageNum}\|{pageSize}` | string(JSON) | 30s + ≤10% 抖动 | 职位列表 |
| `resume:job:v{n}:detail:{id}` | string(JSON) | 5min + 抖动 | 职位详情 |
| `resume:favorite:{userId}` | Set | 无 | 企业收藏学生集合 |
| `resume:online:{type}:{id}` | string(计数) | 5min | 在线状态（多端计数） |
| `resume:mq:consumed:{eventId}` | string | 24h | MQ 快速幂等 |

Java 版：`login_attempt:{username}`（30min）、`user:favorite:{userId}`（Set）。

### 18.2 三大缓存问题

| 问题 | 方案 |
|---|---|
| 缓存击穿 | `singleflight.Group`，同 key 并发只回源一次 |
| 缓存雪崩 | TTL 增加 `rand(ttl/10)` 抖动，打散过期时间 |
| 缓存一致性 | 职位增删改 → `INCR resume:job:version`，读请求拼新版本号；旧 key 等 TTL 自然淘汰，不做 SCAN 删除 |
| 缓存穿透 | 详情不存在直接返回错误，**不缓存空值** |
| 降级 | Redis 不可用时自动切内存 map；`MarkOnce` 返回 true，由 MySQL 唯一键兜底 |

---

## 19. gRPC 内部服务

### 19.1 接口定义（`proto/recruit/v1/recruit.proto`）

```proto
service RecruitInternal {
  rpc GetUserBrief(GetUserBriefRequest) returns (UserBrief);
  rpc GetJobBrief(GetJobBriefRequest) returns (JobBrief);
}
```

| 方法 | 入参 | 出参 |
|---|---|---|
| `GetUserBrief` | `user_id`、`user_type` | `user_id`、`user_type`、`display_name`、`exists` |
| `GetJobBrief` | `job_id` | `job_id`、`title`、`company_name`、`status`、`exists` |

### 19.2 实现与调用

- 服务端 `rpc.Serve(ctx, ":9091", db)`，带日志拦截器（方法名 + 耗时 + err），`ctx.Done()` 时 `GracefulStop()`，5s 未完成则 `Stop()`。
- 客户端 `grpc.NewClient` 懒连接 + `insecure` 凭证，2s 超时。
- **唯一实际调用点**：`HandleChatNotify` 取发送者昵称（失败降级查库）。
- `JobBrief` 目前无业务调用，属预留能力。
- **与 Java 后端没有 gRPC 互通**，只在本进程内自我调用（`127.0.0.1:9091`）。

---

## 20. RabbitMQ 离线通知

| 项 | 值 |
|---|---|
| 队列 | `resume.chat.notify`（durable） |
| 交换机 | 默认 `""`，routing key 即队列名 |
| 事件类型 | `chat.message.created` |
| 消息属性 | JSON、`Persistent`、`MessageId=eventId` |
| 生产 | 仅 `SendChatMessage`，失败只打日志 |
| 消费 | 仅 `HandleChatNotify` |
| 预取 | `Qos(20)` |
| ACK | 成功 `Ack`；JSON 解析失败或 Handler 报错 → `Nack(requeue=false)` **丢弃**（防毒消息死循环，无死信队列） |
| 重连 | 消费者循环，出错 5s 后重连；发布时检测连接关闭并自动重连 |
| 降级 | `RABBITMQ_ENABLED=false` 时 `Publish` 直接返回 nil |

---

## 21. 前端实现原理

### 21.1 主站请求层（`src/utils/request.js`）

| 项 | 实现 |
|---|---|
| baseURL | 开发 `/api-backend`；生产 `VITE_API_BASE_URL`，缺省 `http://182.92.210.198:8081` |
| 超时 | 5000ms |
| 固定头 | `platform: H5` |
| Token | `useUserStore().token` → `Authorization: Bearer ...` |
| Loading | 全局 `ElLoading` 全屏，多个并发请求共享同一实例 |
| 成功判定 | `code === 200 || 201`，否则 `ElMessage.error` + `reject` |
| 错误脱敏 | 正则过滤 `json/sql/mysql/dial tcp/panic/...` 等技术词，统一提示“操作失败，请稍后重试” |
| `silent` | `config.silent === true` 时不弹提示（用于“暂无简历”等预期空结果） |
| 401 | **无特殊处理**，不做跳转/刷新（Token 过期由路由守卫本地解析处理） |
| `fileUrl(path)` | 反斜杠转 `/`，相对路径补 `API_BASE_URL` |

### 21.2 状态管理

| Store | id | 内容 | 持久化 |
|---|---|---|---|
| `stores/user.js` | `rec-user` | `userId/username/displayName/userType/token` | ✅ localStorage |
| `stores/students.js` | `filter` | 学位/专业/性别/语言/岗位/技能/城市筛选 | ✅ |
| `stores/resume.js` | `resume` | 简历草稿 | ✅（实际未使用） |
| `stores/chatUnread.js` | 模块级 `ref` | 全局未读总数 | ❌ |

### 21.3 路由与权限

- `createWebHistory`，全部页面懒加载。
- 白名单 `/login`；其余要求 `token` 且本地 JWT `exp` 未过期。
- **没有角色级路由权限**，角色差异体现在菜单与页面内条件渲染：
  - 企业：应届生列表 `/home`、招聘信息 `/job`；
  - 学生/企业：消息 `/chat`；
  - 所有人：数据看板 `/showData`。

### 21.4 ECharts 看板

- 按需注册 `CanvasRenderer`、`PieChart/BarChart/LineChart/RadarChart`、`Title/Tooltip/Legend/Grid`；
- 数据来自 `/stu/statistics`，取前 10 条；
- 使用 `shallowRef` + `watchEffect` 减少大对象响应式开销；
- 通过 `MutationObserver` 监听 `<html class="dark">` 切换图表配色。

### 21.5 简历模板制作

`makeResume.vue` 为**纯前端**页面：可编辑字段、增删教育/工作/项目/奖项、上传头像（≤2MB 本地预览），最后用 `html2pdf.js` 导出 `${name}_简历.pdf`（`scale 1.8`，排除控制面板）。

### 21.6 管理端

- 无 `src/api`、无 axios 封装，所有请求直接写在 `views/Home.vue`；
- `backendUrl = VITE_API_BASE_URL || 'http://81.70.199.174:8081'`；
- Token 从 `localStorage['token']` 读取，手动拼 `Authorization`；
- 功能：学生/企业列表、分页、搜索、批量删除、Excel 导入；导出、统计、设置、编辑抽屉均为占位。

---

## 22. 部署

### 22.1 Windows 本地

见 [6.1 一键启动](#61-一键启动windows推荐)。

### 22.2 Linux + systemd（Java 版）

仓库内 `.ssh/resume.service` 与 `.ssh/prod_application.yml` 是生产配置样例；`deploy.ps1` 的流程：

1. `mvn clean package -DskipTests` 生成 `target/resume-0.0.1-SNAPSHOT.jar`；
2. Posh-SSH 通过 SFTP 上传到 `/opt/resume/`；
3. SSH 执行 `systemctl restart resume.service`，等待 25s 检查状态与 8081 端口；
4. 调 `POST /login` 做健康验证。

> 脚本中硬编码了服务器地址与明文密码，且已提交到仓库，**建议改为环境变量或密钥管理**。

### 22.3 Docker Compose（Go 版）

`docker-compose.yml` 包含 mysql / redis / rabbitmq / app 四个服务，`app` 依赖三者 `service_healthy`，`uploads` 目录以卷持久化，全部配置由环境变量注入。

### 22.4 Kubernetes（Go 版）

| 资源 | 说明 |
|---|---|
| `Namespace recruit` | 命名空间 |
| `ConfigMap resume-go-config` | 端口、DB/Redis/MQ 地址、`UPLOAD_DIR`、`TZ` |
| `Secret resume-go-secret` | `DB_USER`、`DB_PASSWORD`、`JWT_SECRET` |
| `Deployment resume-go` | 1 副本，RollingUpdate（`maxSurge 1` / `maxUnavailable 0`），端口 8081/9091，挂载 PVC 到 `/app/uploads`，`readiness/liveness` 均探活 `GET /stu/statistics`，资源 `100m/128Mi` ~ `500m/512Mi` |
| `Service resume-go` | ClusterIP，8081 http / 9091 grpc |
| `PVC resume-uploads-pvc` | `ReadWriteOnce`，5Gi（扩容需对象存储/NAS） |
| `Ingress` | `ingressClassName: nginx`，host `recruit.example.com`，`proxy-body-size: 12m` |

> `readinessProbe` 使用公开接口 `/stu/statistics`，因此该接口必须保持免鉴权可用。

### 22.5 Nginx 代理 WebSocket

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

## 23. 已知问题与注意事项

> 以下均为阅读代码与 SQL 得到的事实，不是猜测；使用或二次开发前请务必了解。

### 23.1 安全类

1. **密钥/密码硬编码**：JWT 密钥、数据库密码、部署脚本密码均写在仓库中，生产必须改为环境变量/密钥管理。
2. **Refresh Token 可当 Access Token 用**：业务中间件不校验 `tokenType`。
3. **管理员 Token 经 URL query 传递**（`/admin/?token=...`），可能进入浏览器历史与日志；管理端也不校验 Token 有效性。
4. **CORS 全开 + `allowCredentials(true)`**：生产应收紧 Origin。
5. **WebSocket `CheckOrigin` 恒 true**：生产应校验来源。
6. Java 登录失败锁定逻辑被注释，**暴力破解无限制**。

### 23.2 功能/一致性类

1. **`/api/resume/download/{name}` 不存在**：两套后端的上传接口都返回该地址，直接访问会 404。
2. **没有简历内容解析**：仅做类型/大小校验与落盘，不提取 PDF 文本。
3. **`/stu/getAllSkill` 返回字符串而非数组**（Java/Go 一致），前端需自行 `split(',')`。
4. **Java 全局 SNAKE_CASE 与 `LoginDTO.userType` 的 `@JsonProperty` 例外并存**，前后端字段命名易错。
5. **Java 的收藏缓存删除失效**：`RedisTemplate` 与 `StringRedisTemplate` 序列化器不一致，且 `remove` 传了 List 对象。
6. **Java `SkillsServiceImpl.getAllSkill` 可能抛异常**：`Set.of(split(","))` 遇重复元素抛 `IllegalArgumentException`。
7. **Java 字典写操作用 `@Select` 执行 INSERT/DELETE**，可能导致 MyBatis 走 `executeQuery` 报错（Go 版正常）。
8. **删除学生/企业不清理子表**：会残留 `stu_*`、`stu_resume`、`com_like_stu` 等孤儿数据。
9. **`StudentDetailMapper` 未去重**：`LEFT JOIN stu_region` 无 `DISTINCT`，多地区学生会重复行、`total` 偏大。
10. **样例数据不一致**：`resume.sql` 中 `user_id=2` 的密码是明文，无法通过 BCrypt 校验；`student.id=5` 与部分子表 `stu_id` 不匹配。

### 23.3 前端类

1. 菜单 index 为 `/Job`，路由注册为 `/job`，**大小写不一致导致“招聘信息”点不开**。
2. `myJobs.vue` 的删除/上下架方法**未 import 也未绑定**，属死代码。
3. `makeResume.vue` 的预览/保存/上传按钮无点击处理，仅下载可用。
4. 管理端“导出/统计/设置/查看编辑”均为占位。
5. 两个 Vite 项目默认端口都是 5173，同时启动会冲突（`start-all.bat` 已分别指定 5173/5174）。
6. 主站 axios 超时 5s，对上传偏短；且无 401 统一处理。

### 23.4 部署/运维类

1. 聊天在线状态 Redis TTL 只在上线时设置，长连接 >5 分钟会过期。
2. 跨实例实时推送不支持。
3. MQ 消费失败直接丢弃，无死信队列。
4. `mq_consumed_event` 只增不删，需要定期归档。
5. K8s 使用 `ReadWriteOnce` PVC 且副本数 1，扩容需改对象存储。

---

## 24. 二次开发指南

### 24.1 新增一个后端接口（Go）

1. `internal/model` 定义请求/响应结构（需要兼容前端时实现 `UnmarshalJSON`）；
2. `internal/data` 写 SQL（全部 `?` 绑定，事务用 `BeginTx`）；
3. `internal/service` 写业务规则，错误用 `model.AppError{Code, Msg}`；
4. `internal/api` 写 handler 并注册路由（注意放在 `authorized` 还是 `admin` 分组）；
5. 前端 `src/api/*.js` 增加函数，页面调用。

### 24.2 新增一个前端页面

1. `views/` 新建 `.vue`；
2. `router/index.js` 增加懒加载路由；
3. 需要登录的页面无需额外配置（全局守卫已覆盖）；
4. 需要菜单入口时修改 `views/layout/index.vue`，**注意 `index` 与路由 path 大小写必须一致**。

### 24.3 新增数据库字段

1. `resume.sql` / `schema.sql` 增加列；
2. 实体 + Mapper 更新；
3. 若是画像子表，记得同时更新 `completeInfo` 的“先删后插”逻辑与详情聚合。

### 24.4 切换后端

- 前端不用改代码：两套后端路径一致；只需保证 `VITE_API_BASE_URL` / 代理指向运行中的那一套。
- 若使用 Java 版，聊天与投递相关页面会 404（Java 未实现）。

### 24.5 调试建议

| 问题 | 排查点 |
|---|---|
| 401 | `Authorization` 是否带 `Bearer ` 前缀；Token 是否过期（JWT `exp`） |
| 403 | 是否 `admin` 接口；是否操作了他人职位 |
| 上传失败 | 字段名是否为 `resumeFile` / `image` / `stuFile` / `comFile`；大小与扩展名 |
| 聊天收不到 | WS 是否 101；Token 是否有效；对方是否在线；`/chat/messages` 是否能拉到历史 |
| 聊天重复上屏 | 是否按消息 `id` 去重 |
| 分页总数不对 | 是否用了 `size` 而不是 `total` |
| 文件 404 | 是否访问 `/uploads/...` 而不是 `/api/resume/download/...` |

---

## 25. 附录：接口速查表

### 25.1 按模块统计

| 模块 | 接口数 | 路径前缀 |
|---|---|---|
| 认证/用户 | 5 | `/login`、`/refresh-token`、`/validateToken`、`/user/*` |
| 学生 | 9 | `/stu/*` |
| 职位 | 9 | `/com/jobPosting*`、`/com/applications*` |
| 收藏 | 6 | `/com/Like*`、`/com/QueryComIdByUserId` |
| 检索 | 2（+1 后台复用） | `/com/search`、`/com/detailSearch` |
| 聊天 | 12 | `/chat/*` |
| 管理 | 19 | `/admin/*` |
| 静态 | 3 | `/uploads/*`、`/resume/uploads/*` |

### 25.2 免登录接口

```text
POST /login
POST /refresh-token
POST /validateToken
GET  /stu/statistics
GET  /chat/ws?token=...
```

### 25.3 仅管理员

```text
/admin/**
```

### 25.4 仅 Go 后端

```text
/user/me, PUT /user/name
GET  /stu/resume/latest
/stu/applications*
/com/applications*
/chat/**
```

### 25.5 关键源码索引

| 能力 | 文件 |
|---|---|
| Go 路由总表 | `back/resume-go/internal/api/api.go` |
| Go 鉴权 | `internal/auth/auth.go`、`internal/api/middleware.go` |
| Go 聊天 Hub | `internal/chat/hub.go` |
| Go 聊天连接 | `internal/chat/client.go` |
| Go 聊天业务 | `internal/service/chat.go` |
| Go 聊天 SQL | `internal/data/chat.go` |
| Go 聊天入口 | `internal/api/handlers_chat.go` |
| Go 缓存 | `internal/cache/cache.go` |
| Go MQ | `internal/mq/mq.go` |
| Go gRPC | `internal/rpc/server.go`、`client.go`、`proto/recruit/v1/recruit.proto` |
| Java 鉴权 | `common/interceptor/*.java`、`utils/JwtUtil.java` |
| Java 异常 | `common/exception/GlobalExceptionHandler.java` |
| Java 检索 SQL | `resources/mapper/SearchMapper.xml`、`StudentDetailMapper.xml` |
| Java 收藏 SQL | `resources/mapper/ComLikeStu.xml` |
| 前端请求封装 | `front/yxl/recruit-project/src/utils/request.js` |
| 前端 WS 封装 | `front/yxl/recruit-project/src/utils/chatSocket.js` |
| 前端聊天页 | `front/yxl/recruit-project/src/views/chat/index.vue` |
| 前端路由守卫 | `front/yxl/recruit-project/src/router/index.js` |
| 管理端页面 | `front/ljk/adminPanel/src/views/Home.vue` |
| 聊天教程 | `CHAT_GUIDE.md` |

---

**至此，本文档覆盖了项目的全部技术点、接口、聊天原理与部署方式。** 若只关心聊天，建议直接从第 12 章开始，并配合 `CHAT_GUIDE.md` 阅读代码。

