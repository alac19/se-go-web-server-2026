# 第一阶段：编译 Go 代理，编译出可执行文件
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 复制依赖文件并下载依赖（利用 Docker 缓存）
COPY go.mod go.sum ./

# ========== 在下载依赖之前，设置 Go 环境变量 ==========
# 配置使用七牛云提供的代理 goproxy.cn，并配置校验数据库镜像，避免网络问题导致依赖下载失败
RUN go env -w GOPROXY=https://goproxy.cn,direct && \
	go env -w GOSUMDB=sum.golang.google.cn

# 下载依赖项
RUN go mod download
# ========== 环境变量配置结束 ==========

# 复制源代码并编译（关闭 CGO，静态链接，去除调试符号）
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o proxy ./src/main.go

# 第二阶段：生成最小化运行镜像
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求（如需要）
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制编译好的二进制
COPY --from=builder /app/proxy .

# 复制配置文件（用户需挂载或内置默认配置）
# 这里复制默认配置作为示例，实际使用时可挂载外部配置
COPY configs/config.toml ./configs/
# 复制证书（测试用自签名，生产应更换）
COPY cert.pem key.pem ./

# 暴露 HTTP 和 HTTPS 端口
EXPOSE 8080 8443

# 运行代理（默认使用内部配置，也可通过环境变量或挂载覆盖）
CMD ["./proxy"]