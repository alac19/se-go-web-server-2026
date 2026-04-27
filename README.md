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
├── configs/                          
│   └── config.toml              # 配置文件
├── docs/                        # 文档
├── src/
│   ├── main.go                  # 程序入口
│   ├── config/                  # 配置加载
│   │   ├── config.go
│   │   └── config_test.go
│   ├── health/                  # 健康检查
│   │   ├── health.go
│   │   └── health_test.go
│   ├── lb/                      # 负载均衡
│   │   ├── load_balancer.go
│   │   └── load_balancer_test.go
│   ├── log/                     # 日志系统
│   │   ├── logger.go
│   │   └── logger_test.go
│   └── proxy/                   # 反向代理
│       ├── proxy.go
│       └── proxy_test.go
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── cert.pem
├── key.pem
├── LICENSE
└── README.md
```

---

## 🚀 快速开始

### 前置条件
- Go 1.25 或更高版本（源码运行）
- Docker 和 Docker Compose（容器运行，推荐）

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
编辑 `configs/config.toml`，按需修改监听地址、后端列表、心跳参数等。示例：

```toml
# 监听端口列表（至少两个：HTTP 和 HTTPS）
[[listen]]
listen_addr = "0.0.0.0"
listen_port = 8080

[[listen]]
listen_addr = "0.0.0.0"
listen_port = 8443

# 后端服务器列表
[[forward]]
forward_addr = "backend1"
forward_port = 9091

[[forward]]
forward_addr = "backend2"
forward_port = 9092

[[forward]]
forward_addr = "backend3"
forward_port = 9093

# 心跳检测配置
[heartbeat]
enabled = true
path = "/heartbeat"
interval_seconds = 30
timeout_seconds = 2

# 代理转发超时（秒）
[proxy]
timeout_seconds = 30

# TLS 证书路径
[tls]
cert_file = "cert.pem"
key_file = "key.pem"

# 日志配置
[log]
level = "info"               # debug, info, warn, error
file_path = "log.txt"        # 留空则只输出到控制台
```

### 4. 运行代理

#### 方式一：源码运行
```bash
go run ./src/main.go
```

#### 方式二：编译后运行
```bash
go build -o proxy ./src/main.go
./proxy
```

#### 方式三：Docker 运行（推荐）
```bash
# 构建镜像
docker build -t go-reverse-proxy .

# 启动容器（需提前配置好后端服务）
docker run -d -p 8080:8080 -p 8443:8443 -v $(pwd)/configs:/root/configs -v $(pwd)/cert.pem:/root/cert.pem -v $(pwd)/key.pem:/root/key.pem go-reverse-proxy
```

### 5. 验证代理
```bash
# HTTP 请求
curl http://localhost:8080/test

# HTTPS 请求（自签名证书需跳过验证）
curl -k https://localhost:8443/test
```

连续多次执行，应该能看到响应内容中的后端端口在 9091、9092、9093 之间轮换。

---

## 🐳 使用 Docker Compose 一键启动演示环境

`docker-compose.yml` 已经定义了一个反向代理 + 三个模拟后端（`hashicorp/http-echo`）。无需额外配置后端，可直接体验完整功能：

```bash
docker-compose up -d
```

代理将监听 `8080`（HTTP）和 `8443`（HTTPS），后端模拟服务会返回固定文本。测试命令同上。停止服务：

```bash
docker-compose down
```

---

## 📊 性能测试结果（摘要）

在 **Linux 虚拟机**（Ubuntu，4 核）上使用 `go-wrk` 进行压测，代理日志级别为 `info`，结果如下：

| 并发数 | 总请求数 | QPS  | 平均延迟(ms) | P99 延迟(ms) | 错误率 |
| ------ | -------- | ---- | ------------ | ------------ | ------ |
| 100    | 250000   | 4687 | 21.3         | 44.2         | 0%     |
| 200    | 250000   | 4719 | 42.3         | 88.1         | 0%     |

- 镜像大小约 **16 MB**（满足 ≤30 MB）
- 健康检查动态剔除/恢复后端功能正常
- TLS 1.2/1.3 支持，HTTP/2 协商成功

详细测试数据见 [测试报告](docs/5-测试报告.md)。

---

## ⚙️ 配置说明

| 配置段      | 字段                                                     | 说明                                       |
| ----------- | -------------------------------------------------------- | ------------------------------------------ |
| `listen`    | `listen_addr`, `listen_port`                             | 代理监听的地址和端口（至少两个）           |
| `forward`   | `forward_addr`, `forward_port`                           | 后端服务的地址和端口                       |
| `heartbeat` | `enabled`, `path`, `interval_seconds`, `timeout_seconds` | 健康检查开关、路径、间隔、超时             |
| `proxy`     | `timeout_seconds`                                        | 转发请求的整体超时（秒）                   |
| `tls`       | `cert_file`, `key_file`                                  | HTTPS 证书和私钥文件路径                   |
| `log`       | `level`, `file_path`                                     | 日志级别和输出文件（留空则只输出到控制台） |

> 注意：健康检查要求后端提供 `path` 对应的接口，返回 2xx 状态码即认为健康。

---

## 🧪 测试与验证

- **单元测试**：每个模块均提供 `*_test.go`，执行 `go test ./... -cover` 查看覆盖率。
- **集成测试**：手动测试了 HTTP/HTTPS 转发、负载均衡、健康检查、503/502 错误处理。
- **压力测试**：使用 `ab` 和 `go-wrk` 在不同并发下验证性能，结果如上。

---

## ❓ 常见问题

### Q1: 代理启动后健康检查一直失败？
- 确保后端服务正在运行，并且提供了 `/heartbeat` 接口（返回 200）。
- 检查 `config.toml` 中的 `forward_addr` 是否正确（容器内需用服务名或 IP）。
- 可临时设置 `heartbeat.enabled = false` 跳过健康检查，验证转发功能。

### Q2: 容器内代理无法连接后端？
- 检查容器网络：`docker-compose.yml` 中所有服务是否在同一 `network` 下。
- 后端服务名（如 `backend1`）需在代理的 `config.toml` 中作为 `forward_addr`。
- 确保后端服务监听 `0.0.0.0`（或者容器内可以通信的地址）。

### Q3: 如何更换真正的后端服务？
- 修改 `config.toml` 中的 `forward_addr` 为真实后端的 IP 或域名，并确保后端提供 `/heartbeat` 健康检查接口。
- 如果不需要模拟后端，可以从 `docker-compose.yml` 中删除 `backend1/2/3` 服务。

### Q4: 日志级别如何影响性能？
- `info` 级别会记录每个请求的详细信息（方法、路径、耗时、后端地址），对性能有一定影响但满足需求。
- 生产环境可将日志级别改为 `warn` 或 `error` 以降低日志开销，但会丢失请求日志。

---

## 📄 文档

- [软件需求规格说明书](docs/1-软件需求规格说明书.md)
- [系统设计文档](docs/2-系统设计文档.md)
- [技术规范文档](docs/3-技术规范文档.md)
- [测试计划](docs/4-测试计划.md)
- [测试报告](docs/5-测试报告.md)
- [用户验收测试计划](docs/6-用户验收测试计划.md)
- [用户验收测试报告](docs/7-用户验收测试报告.md)

---

## 🤝 贡献

本项目为个人重构练习，不开放外部贡献。欢迎提 Issue 交流。

---

## 📜 许可证

MIT © [刘灿阳](https://github.com/alac19)

---