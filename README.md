# **Go 反向代理服务器**

>**High-Concurrency Reverse Proxy in Go (Refactoring Project)**  

[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

从高并发 Rust Web 服务器项目重构而来，使用 Go 标准库实现 HTTP/HTTPS 代理、轮询负载均衡、健康检查、容器化部署。

---

## ✨ 特性

- 🚀 **高并发**：基于 Go goroutine，每个请求独立处理，支持数千并发连接
- 🔁 **反向代理**：支持 HTTP/HTTPS 双协议监听，自动转发到后端服务
- ⚖️ **负载均衡**：轮询（Round‑Robin）算法，原子操作保证线程安全
- 💓 **健康检查**：定期探测后端 `/heartbeat` 接口，自动剔除/恢复故障节点
- 📝 **结构化日志**：使用 `log/slog` 输出 JSON 格式日志，便于接入日志系统
- 🔒 **TLS 支持**：基于 `crypto/tls`，最低 TLS 1.2，安全传输
- ⚙️ **配置驱动**：TOML 格式配置文件，支持监听地址、后端列表、心跳间隔等
- 🐳 **容器化**：提供 Dockerfile 和 docker-compose，一键启动代理 + 模拟后端

---

## 🗂️ 项目结构

```
se-go-web-server-2026/
├── configs/               # 配置文件示例
│   └── config.toml
├── docs/                  # 项目文档
│   ├── 1-软件需求规格说明书.md
│   ├── 2-系统设计文档.md
│   └── 3-技术规范文档.md
├── internal/              # 内部包（待实现）
│   ├── config/
│   ├── lb/
│   ├── health/
│   └── proxy/
├── cmd/                   # 入口（可选）
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

> **注**：代码尚未完全实现，目前已完成文档设计，核心代码将于 2026.04.20 开始编写。

---

## 🚀 快速开始

### 前置条件
- Go 1.25 或更高版本
- （可选）Docker 和 Docker Compose

### 1. 克隆仓库
```bash
git clone https://github.com/alac19/se-go-web-server-2026.git
cd se-go-web-server-2026
```

### 2. 生成测试证书（用于 HTTPS）
```bash
openssl req -x509 -newkey rsa:4096 -nodes -keyout key.pem -out cert.pem -days 365 -subj "/CN=localhost"
```

### 3. 修改配置文件
复制并编辑 `configs/config.toml`，配置监听端口和后端地址：
```toml
[[listen]]
addr = "127.0.0.1"
port = 8080

[[listen]]
addr = "127.0.0.1"
port = 8443

[[backends]]
addr = "127.0.0.1"
port = 9091

[[backends]]
addr = "127.0.0.1"
port = 9092

[heartbeat]
enabled = true
path = "/heartbeat"
interval_seconds = 30
timeout_seconds = 2
```

### 4. 运行代理
```bash
go run main.go -config ./configs/config.toml
```

### 5. 测试
```bash
# HTTP 请求
curl http://localhost:8080/test

# HTTPS 请求（忽略证书验证）
curl -k https://localhost:8443/test
```

---

## 🐳 使用 Docker

### 构建镜像
```bash
docker build -t go-reverse-proxy .
```

### 使用 docker-compose 启动（代理 + 三个模拟后端）
```bash
docker-compose up -d
```

模拟后端会自动响应 `/heartbeat` 和任意路径（返回 "ok"）。

---

## 📊 性能测试（待补充）

使用 `wrk` 或 `hey` 进行压测，结果将记录在此处。

---

## 📄 文档

- [软件需求规格说明书](docs/1-软件需求规格说明书.md)
- [系统设计文档](docs/2-系统设计文档.md)
- [技术规范文档](docs/3-技术规范文档.md)

---

## 🤝 贡献

本项目为个人重构练习，不开放外部贡献。但欢迎提 Issue 交流。

---

## 📜 许可证

MIT © [刘灿阳](https://github.com/alac19)

---