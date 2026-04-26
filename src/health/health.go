package health

import (
	"context" // 用于控制健康检查协程的生命周期
	"fmt"
	"log/slog" // 使用 slog 进行结构化日志记录
	"net/http"
	"net/url"
	"time"
)

// Backend 代表一个后端服务（地址 + HTTP 客户端）
type Backend struct {
	URL    *url.URL
	Client *http.Client
}

// Checker 负责定期检查后端健康状态
type Checker struct {
	backends   []*Backend // 所有配置的后端（固定）
	healthPath string     // 健康检查路径
	interval   time.Duration
	timeout    time.Duration
	onHealthy  func([]*url.URL) // 回调：当健康列表变化时调用（更新负载均衡器）
}

// / 创建健康检查器
// / # 参数
// / - backendURLs: 后端 URL 列表
// / - healthPath: 健康检查路径
// / - intervalSec: 检查间隔（秒）
// / - timeoutSec: 单次请求超时（秒）
// / - onHealthy: 回调函数，参数为当前健康的后端 URL 列表
// /
// / # 返回
// / - *Checker: 创建的 Checker 实例
func NewChecker(
	backendURLs []*url.URL,
	healthPath string,
	intervalSec, timeoutSec int,
	onHealthy func([]*url.URL),
) *Checker {
	// 初始化 Backend 列表
	backends := make([]*Backend, len(backendURLs)) // 使用 make 创建切片

	for i, u := range backendURLs {
		// 每个后端使用独立的 HTTP 客户端（可复用，但为了超时控制，单独创建）
		client := &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		}
		backends[i] = &Backend{URL: u, Client: client}
	}

	return &Checker{
		backends:   backends,
		healthPath: healthPath,
		interval:   time.Duration(intervalSec) * time.Second, // 转换为 time.Duration 类型
		timeout:    time.Duration(timeoutSec) * time.Second,
		onHealthy:  onHealthy,
	}
}

// / 启动后台健康检查循环，结构体方法
// / # 参数
// / - ctx: 上下文，用于控制协程生命周期（取消时停止检查）
func (c *Checker) Start(ctx context.Context) {
	// 立即执行一次初始检测
	c.checkAndUpdate()

	ticker := time.NewTicker(c.interval) // 需要 time.Duration 类型

	go func() {
		for {
			select {
			case <-ctx.Done(): // 主程序退出或需要重新加载配置时调用
				ticker.Stop() // 停止定时器，释放资源
				slog.Info("健康检查协程停止")
				return
			case <-ticker.C:
				c.checkAndUpdate()
			}
		}
	}()
}

// / checkAndUpdate 执行一次所有后端的健康检测，并调用回调更新健康列表
func (c *Checker) checkAndUpdate() {
	var healthyURLs []*url.URL

	for _, backend := range c.backends {
		healthURL := fmt.Sprintf("%s://%s%s", backend.URL.Scheme, backend.URL.Host, c.healthPath) // 构造健康检查 URL
		req, err := http.NewRequest(http.MethodGet, healthURL, nil)                               // 构造 GET 请求

		if err != nil {
			slog.Warn("构造健康检查请求失败", "url", healthURL, "error", err)
			continue
		}

		resp, err := backend.Client.Do(req) // 执行请求，内部会尝试建立 TCP 连接

		if err != nil {
			slog.Warn("健康检查失败", "backend", backend.URL.String(), "error", err)
			continue
		}

		resp.Body.Close() // 及时关闭响应体，避免资源泄漏

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			slog.Info("健康检查成功", "backend", backend.URL.String())
			healthyURLs = append(healthyURLs, backend.URL)
		} else {
			slog.Warn("健康检查返回非成功状态码", "backend", backend.URL.String(), "status", resp.StatusCode)
		}
	}

	// 调用回调更新负载均衡器的后端列表
	if c.onHealthy != nil {
		c.onHealthy(healthyURLs)
	}
}
