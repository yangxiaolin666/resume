// Package service 是业务层：所有「业务规则」都集中在这里。
//
// 【这一层负责什么】
//   - 校验参数是否合法（必填、长度、取值范围、角色权限）；
//   - 编排多个下层组件，完成一个完整用例（例如「发消息」要校验、入库、推送、发 MQ）；
//   - 把底层错误翻译成统一的业务错误（model.AppError），供 api 层直接返回给前端。
//
// 【与 api/data/cache/mq/rpc 的关系】
//
//	api（Gin Handler）只做 HTTP 参数绑定和响应封装，不写业务判断；
//	service 承接 api，向下调用：
//	  - data：执行 SQL，service 不直接写 SQL；
//	  - cache：Redis 缓存、在线状态、收藏集合、消费幂等键；
//	  - mq：把「通知」这类非核心动作异步化；
//	  - rpc：通过内部 gRPC 获取昵称等跨服务信息。
//	这样分层后，Handler 很薄，替换传输层（HTTP 换 gRPC）时业务代码不用重写。
//
// 【对应简历能力】Go 后端分层架构、业务流程编排、缓存/MQ/gRPC 综合运用。
package service

import (
	"resume-backend/internal/auth"
	"resume-backend/internal/cache"
	"resume-backend/internal/chat"
	"resume-backend/internal/data"
	"resume-backend/internal/mq"
	"resume-backend/internal/rpc"
)

// Service 是业务层的「依赖容器」。
//
// 【分层回顾】
//
//	api（Gin Handler） -> service（业务规则） -> data（SQL）
//	                       |-> cache（Redis）
//	                       |-> mq（RabbitMQ 发布）
//	                       |-> rpc（内部 gRPC 客户端）
//
// 这里全部使用具体类型而不是 interface，是为了让初学者一眼看清依赖关系；
// 真实项目如果要做单元测试，可以再把 Cache/MQ 抽象成 interface 便于 mock。
type Service struct {
	DB        *data.DB      // MySQL 访问入口，负责所有持久化 SQL
	Auth      *auth.Manager // JWT 签发与校验
	Cache     *cache.Store  // Redis 门面；Redis 不可用时内部自动降级为内存 map
	ChatHub   *chat.Hub     // 本进程的 WebSocket 连接表，用于实时推送
	UploadDir string        // 简历文件落盘目录（聊天图片放在同级的 chat 子目录）
	MQ        *mq.Publisher // RabbitMQ 发布端；通知失败只记日志，不影响主流程
	RPC       *rpc.Client   // 内部 gRPC 客户端，用于跨服务取昵称等，可降级
}

// New 由 main 在启动时组装依赖（依赖注入）。
//
// 好处是依赖关系显式可见：看 main 就知道 Service 用到了哪些基础设施；
// 各层之间也不需要用全局单例互相查找，测试时可以传入不同的实现。
func New(db *data.DB, manager *auth.Manager, cacheStore *cache.Store,
	chatHub *chat.Hub, uploadDir string, publisher *mq.Publisher, rpcClient *rpc.Client) *Service {
	return &Service{
		DB:        db,
		Auth:      manager,
		Cache:     cacheStore,
		ChatHub:   chatHub,
		UploadDir: uploadDir,
		MQ:        publisher,
		RPC:       rpcClient,
	}
}
