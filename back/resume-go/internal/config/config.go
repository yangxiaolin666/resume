package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config 保存运行时配置，均可通过环境变量覆盖；未设置时使用下面的默认值。
//
// 【学习提示】
// 一个后端服务通常会同时依赖：MySQL（持久化）、Redis（缓存/在线状态）、
// RabbitMQ（异步解耦）、gRPC（内部服务通信）。把它们统一收敛到 Config，
// 好处是：
//  1. 所有可调参数集中在一处，排查问题时不用满项目搜环境变量；
//  2. 单元测试可以构造一份配置，不依赖真实中间件；
//  3. 容器化部署时，只需通过环境变量注入，不用改代码。
type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	UploadDir  string
	JWTSecret  string

	// Redis：缓存热点职位、收藏集合、在线状态、消息消费幂等键
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisEnabled  bool

	// RabbitMQ：聊天消息通知的异步队列
	RabbitMQURL     string
	RabbitMQEnabled bool

	// gRPC：内部服务通信端口（例如通知服务查询发送者昵称）
	GRPCPort     string
	GRPCEnabled  bool
	GRPCInternal string // 本进程作为客户端访问自身 gRPC 服务的地址
}

func Load() Config {
	cfg := Config{
		ServerPort: env("SERVER_PORT", "8081"),
		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "3306"),
		DBUser:     env("DB_USER", "root"),
		DBPassword: env("DB_PASSWORD", "Wangtianye123"),
		DBName:     env("DB_NAME", "resume"),
		UploadDir:  env("UPLOAD_DIR", "uploads/resume"),
		JWTSecret:  env("JWT_SECRET", "resume_secret_key_2024"),

		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       envInt("REDIS_DB", 0),
		RedisEnabled:  envBool("REDIS_ENABLED", true),

		RabbitMQURL:     env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQEnabled: envBool("RABBITMQ_ENABLED", true),

		GRPCPort:    env("GRPC_PORT", "9091"),
		GRPCEnabled: envBool("GRPC_ENABLED", true),
	}
	// 默认让 gRPC 客户端连本机；容器里可以通过 GRPC_INTERNAL 指向独立服务名。
	cfg.GRPCInternal = env("GRPC_INTERNAL", "127.0.0.1:"+cfg.GRPCPort)
	return cfg
}

// DSN 生成 go-sql-driver/mysql 的连接串。
// parseTime=true 让 DATETIME 自动扫描成 time.Time；
// loc=Local 保证时区与服务器一致，避免聊天时间相差 8 小时。
func (c Config) DSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" +
		c.DBName + "?charset=utf8mb4&parseTime=true&loc=Local"
}

// StaticRoot 返回静态资源根目录（即简历目录的上一级）。
func (c Config) StaticRoot() string {
	return filepath.Dir(c.UploadDir)
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// envBool 读取布尔环境变量，支持 1/true/yes/on 等写法。
func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
