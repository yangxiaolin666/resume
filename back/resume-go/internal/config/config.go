package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Config 保存运行时配置，均可通过环境变量覆盖，默认值与 Java 版 application.yml 保持一致。
type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	UploadDir  string
	JWTSecret  string
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
	}
	return cfg
}

// DSN 生成 go-sql-driver/mysql 的连接串。
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
