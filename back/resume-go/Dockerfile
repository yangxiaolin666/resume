# 多阶段构建：第一阶段编译，第二阶段只保留可执行文件，最终镜像更小。
#
# 【学习提示】
#   - CGO_ENABLED=0：生成纯静态二进制，才能在 alpine/scratch 里直接运行；
#   - go mod download 单独一层：只要 go.mod/go.sum 没变，就能复用依赖缓存；
#   - 最终镜像不包含 Go 源码和编译器，体积从 ~1GB 降到 ~20MB。
FROM golang:1.27-alpine AS builder

WORKDIR /build
ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/resume-go .

FROM alpine:3.20

# tzdata：让容器内的 Asia/Shanghai 时区生效
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /out/resume-go /app/resume-go
RUN mkdir -p /app/uploads/resume /app/uploads/chat

# HTTP 端口；gRPC 端口用于内部服务通信
EXPOSE 8081 9091

ENTRYPOINT ["/app/resume-go"]
